package database

import (
	"fmt"
	"path"
	"strings"

	"github.com/syaiful6/payung/helper"
	"github.com/syaiful6/payung/logger"
)

type MariaBackup struct {
	Base
	username          string
	password          string
	additionalOptions []string
}

func (ctx *MariaBackup) perform() (err error) {
	viper := ctx.viper
	viper.SetDefault("username", "root")

	ctx.username = viper.GetString("username")
	ctx.password = viper.GetString("password")
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
	if ctx.username != "" {
		dumpArgs = append(dumpArgs, "--user", ctx.username)
	}

	if ctx.password != "" {
		dumpArgs = append(dumpArgs, "--password", ctx.password)
	}

	path := path.Join(ctx.dumpPath, "mariabackup")
	dumpArgs = append(dumpArgs, "--target-dir", path)

	return dumpArgs
}

func (ctx *MariaBackup) prePareBackup() []string {
	path := path.Join(ctx.dumpPath, "mariabackup")
	return []string{
		"--prepare",
		"--target-dir",
		path,
	}
}

func (ctx *MariaBackup) dump() (err error) {
	logger.Info("-> Backup using mariabackup...")
	_, err = helper.Exec("mariabackup", ctx.dumpArgs()...)
	if err != nil {
		return fmt.Errorf("-> Dump error: %s", err)
	}
	_, err = helper.Exec("mariabackup", ctx.prePareBackup()...)
	if err != nil {
		return fmt.Errorf("-> prepare failed: %s", err)
	}

	_, err = helper.Exec("tar", "--remove-files -cf -", "-C", ctx.dumpPath, "mariabackup")
	if err != nil {
		return fmt.Errorf("-> tar failed: %s", err)
	}

	return
}
