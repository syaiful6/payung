package database

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/syaiful6/payung/compressor"
	"github.com/syaiful6/payung/helper"
	"github.com/syaiful6/payung/logger"
)

type MariaBackup struct {
	Base
	username          string
	password          string
	galeraInfo        bool
	additionalOptions []string
}

func (ctx *MariaBackup) perform() (err error) {
	viper := ctx.viper
	viper.SetDefault("username", "root")
	viper.SetDefault("galera_info", false)

	ctx.username = viper.GetString("username")
	ctx.password = viper.GetString("password")
	ctx.galeraInfo = viper.GetBool("galera_info")
	addOpts := viper.GetString("additional_options")
	if len(addOpts) > 0 {
		ctx.additionalOptions = strings.Split(addOpts, " ")
	}

	return ctx.dump()
}

func (ctx *MariaBackup) dumpArgs() []string {
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

	return dumpArgs
}

func (ctx *MariaBackup) dump() (err error) {
	logger.Info("-> Backup using mariabackup...")
	mariabackup, err := helper.CreateCmd("mariabackup", ctx.dumpArgs()...)
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
	err, ext, r := compressor.CompressTo(ctx.model, bufio.NewReader(stdoutPipe))
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
