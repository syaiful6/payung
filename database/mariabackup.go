package database

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/syaiful6/payung/compressor"
	"github.com/syaiful6/payung/config"
	"github.com/syaiful6/payung/helper"
	"github.com/syaiful6/payung/logger"
	"github.com/syaiful6/payung/packager"
)

var (
	mariabackupMetadataPath = path.Join(config.HomeDir, ".payung/mariabackup-metadata")
)

type MariaBackup struct {
	Base
	username          string
	password          string
	galeraInfo        bool
	enableIncremental bool
	additionalOptions []string
	metadata          *MariaBackupFull
}

type MariaBackupIncremental struct {
	Name    string    `json:"name"`
	Time    time.Time `json:"time"`
	LSNPath string    `json:"lsn_path"`
}

type MariaBackupFull struct {
	Name         string                   `json:"name"`
	Time         time.Time                `json:"time"`
	LSNPath      string                   `json:"lsn_path"`
	Incrementals []MariaBackupIncremental `json:"incremental"`
}

func (ctx *MariaBackup) perform(backupPackage *packager.Package) (err error) {
	viper := ctx.viper
	viper.SetDefault("username", "root")
	viper.SetDefault("galera_info", false)
	viper.SetDefault("enable_incremental", false)

	ctx.username = viper.GetString("username")
	ctx.password = viper.GetString("password")
	ctx.galeraInfo = viper.GetBool("galera_info")
	addOpts := viper.GetString("additional_options")
	if len(addOpts) > 0 {
		ctx.additionalOptions = strings.Split(addOpts, " ")
	}
	ctx.enableIncremental = viper.GetBool("enable_incremental")
	if ctx.enableIncremental {
		return ctx.incrementalBackup(backupPackage)
	}
	return ctx.fullBackup("")
}

func (ctx *MariaBackup) incrementalBackup(backupPackage *packager.Package) error {
	metadataFileName := path.Join(mariabackupMetadataPath, ctx.name+".json")
	ctx.metadata = ctx.loadMetadata(metadataFileName)
	if ctx.metadata != nil && ctx.metadata.Time.Add(time.Hour*24*5).After(time.Now()) {
		// first
		lsnPath := path.Join(ctx.metadata.LSNPath, backupPackage.Time.Format("2006.01.02.15.04.05"))
		baseLsn := ctx.metadata.LSNPath
		if len(ctx.metadata.Incrementals) > 0 {
			baseLsn = ctx.metadata.Incrementals[len(ctx.metadata.Incrementals)-1].LSNPath
		}
		ctx.metadata.Incrementals = append(ctx.metadata.Incrementals, MariaBackupIncremental{
			Name:    ctx.name,
			Time:    backupPackage.Time,
			LSNPath: lsnPath,
		})
		if err := ctx.takeIncrementalBackup(lsnPath, baseLsn); err != nil {
			return err
		}
	} else {
		if ctx.metadata != nil && ctx.metadata.LSNPath != "" {
			if err := os.RemoveAll(ctx.metadata.LSNPath); err != nil {
				logger.Warn("can't delete %s path, operation return error: %v", ctx.metadata.LSNPath, err)
			}
		}

		lsnPath := path.Join(mariabackupMetadataPath, backupPackage.Time.Format("2006.01.02.15.04.05"))
		ctx.metadata = &MariaBackupFull{
			Name:    ctx.name,
			Time:    backupPackage.Time,
			LSNPath: lsnPath,
		}
		if err := ctx.fullBackup(lsnPath); err != nil {
			return err
		}
	}

	if err := ctx.saveMetadata(metadataFileName); err != nil {
		return err
	}
	// also save in dump path
	if err := ctx.saveMetadata(path.Join(ctx.dumpPath, "metadata.json")); err != nil {
		return err
	}

	return nil
}

func (ctx *MariaBackup) baseMariaBackupOptions(extraLsnDir string) []string {
	dumpArgs := []string{
		"--backup",
	}
	if ctx.galeraInfo {
		dumpArgs = append(dumpArgs, "--galera-info")
	}
	dumpArgs = append(dumpArgs, "--stream=xbstream")

	if ctx.username != "" {
		dumpArgs = append(dumpArgs, "--user", ctx.username)
	}

	if ctx.password != "" {
		dumpArgs = append(dumpArgs, "--password", ctx.password)
	}

	if len(ctx.additionalOptions) > 0 {
		dumpArgs = append(dumpArgs, ctx.additionalOptions...)
	}
	if extraLsnDir != "" {
		dumpArgs = append(dumpArgs, "--extra-lsndir", extraLsnDir)
	}

	return dumpArgs
}

func (ctx *MariaBackup) fullBackup(extraLsnDir string) (err error) {
	logger.Info("-> Backup using mariabackup (full backup)...")
	return ctx.takeBackup(ctx.baseMariaBackupOptions(extraLsnDir))
}

func (ctx *MariaBackup) takeIncrementalBackup(lsnPath, baseLsnPath string) error {
	logger.Info("-> Backup using mariabackup (incremental backup)...")
	options := ctx.baseMariaBackupOptions(lsnPath)
	options = append(options, "--incremental-basedir", baseLsnPath)

	return ctx.takeBackup(options)
}

func (ctx *MariaBackup) takeBackup(options []string) error {
	mariabackup, err := helper.CreateCmd("mariabackup", options...)
	if err != nil {
		return fmt.Errorf("-> Create dump command line error: %s", err)
	}
	stdoutPipe, err := mariabackup.StdoutPipe()
	if err != nil {
		return fmt.Errorf("-> Can't pipe stdout error: %s", err)
	}

	err = mariabackup.Start()
	if err != nil {
		return fmt.Errorf("-> can't start mariabackup error: %s", err)
	}
	dumpFilePath := path.Join(ctx.dumpPath, "mariabackup.xb")
	ext, r, err := compressor.CompressTo(ctx.model, bufio.NewReader(stdoutPipe))
	if err != nil {
		return fmt.Errorf("-> can't compress mariabackup.xb output: %s", err)
	}
	dumpFilePath = dumpFilePath + ext
	f, err := os.Create(dumpFilePath)
	if err != nil {
		return fmt.Errorf("-> error: can't create file for database dump: %s", err)
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	if err != nil {
		return fmt.Errorf("-> error: can't copy dump output to file: %s", err)
	}

	if err = mariabackup.Wait(); err != nil {
		return fmt.Errorf("-> Dump error: %s", err)
	}

	logger.Info("dump path:", ctx.dumpPath)
	return nil
}

func (ctx *MariaBackup) loadMetadata(metadataFileName string) *MariaBackupFull {
	helper.MkdirP(mariabackupMetadataPath)
	if !helper.IsExistsPath(metadataFileName) {
		return nil
	}

	f, err := ioutil.ReadFile(metadataFileName)
	if err != nil {
		return nil
	}
	var metadata MariaBackupFull
	err = json.Unmarshal(f, &metadata)
	if err != nil {
		return nil
	}

	return &metadata
}

func (c *MariaBackup) saveMetadata(metadataFileName string) error {
	data, err := json.Marshal(&c.metadata)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(metadataFileName, data, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}
