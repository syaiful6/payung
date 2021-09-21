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

// MySQL database
//
// type: mysql
// host: 127.0.0.1
// port: 3306
// database:
// username: root
// password:
// additional_options:
type MySQL struct {
	Base
	host              string
	port              string
	database          string
	username          string
	password          string
	additionalOptions []string
}

func (ctx *MySQL) perform() (err error) {
	viper := ctx.viper
	viper.SetDefault("host", "127.0.0.1")
	viper.SetDefault("username", "root")
	viper.SetDefault("port", 3306)

	ctx.host = viper.GetString("host")
	ctx.port = viper.GetString("port")
	ctx.database = viper.GetString("database")
	ctx.username = viper.GetString("username")
	ctx.password = viper.GetString("password")
	addOpts := viper.GetString("additional_options")
	if len(addOpts) > 0 {
		ctx.additionalOptions = strings.Split(addOpts, " ")
	}

	// mysqldump command
	if len(ctx.database) == 0 {
		return fmt.Errorf("mysql database config is required")
	}

	err = ctx.dump()
	return
}

func (ctx *MySQL) dumpArgs() []string {
	dumpArgs := []string{}
	if len(ctx.host) > 0 {
		dumpArgs = append(dumpArgs, "--host", ctx.host)
	}
	if len(ctx.port) > 0 {
		dumpArgs = append(dumpArgs, "--port", ctx.port)
	}
	if len(ctx.username) > 0 {
		dumpArgs = append(dumpArgs, "-u", ctx.username)
	}
	if len(ctx.password) > 0 {
		dumpArgs = append(dumpArgs, `-p`+ctx.password)
	}
	if len(ctx.additionalOptions) > 0 {
		dumpArgs = append(dumpArgs, ctx.additionalOptions...)
	}

	dumpArgs = append(dumpArgs, ctx.database)
	return dumpArgs
}

func (ctx *MySQL) dump() error {
	logger.Info("-> Dumping MySQL...")
	mysqldump, err := helper.CreateCmd("mysqldump", ctx.dumpArgs()...)
	if err != nil {
		return fmt.Errorf("-> Create dump command line error: %s", err)
	}
	stdoutPipe, err := mysqldump.StdoutPipe()
	if err != nil {
		return fmt.Errorf("-> Can't pipe stdout error: %s", err)
	}

	err = mysqldump.Start()
	if err != nil {
		return fmt.Errorf("-> can't start mysqlump error: %s", err)
	}

	dumpFilePath := path.Join(ctx.dumpPath, ctx.database+".sql")
	err, ext, r := compressor.CompressTo(ctx.model, bufio.NewReader(stdoutPipe))
	if err != nil {
		return fmt.Errorf("-> can't compress mysqldump output: %s", err)
	}
	dumpFilePath = dumpFilePath + ext
	f, err := os.Create(dumpFilePath + ".br")
	defer f.Close()
	_, err = io.Copy(f, r)

	if err = mysqldump.Wait(); err != nil {
		return fmt.Errorf("-> Dump error: %s", err)
	}

	logger.Info("dump path:", ctx.dumpPath)
	return nil
}
