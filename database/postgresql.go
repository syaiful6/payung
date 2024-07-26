package database

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/syaiful6/payung/compressor"
	"github.com/syaiful6/payung/helper"
	"github.com/syaiful6/payung/logger"
	"github.com/syaiful6/payung/packager"
)

// PostgreSQL database
//
// type: postgresql
// host: localhost
// port: 5432
// database: test
// username:
// password:
// additional_options:
type PostgreSQL struct {
	Base
	host              string
	port              string
	database          string
	username          string
	password          string
	additionalOptions []string
}

func (ctx PostgreSQL) perform(backupPackage *packager.Package) (err error) {
	viper := ctx.viper
	viper.SetDefault("host", "localhost")
	viper.SetDefault("port", 5432)

	ctx.host = viper.GetString("host")
	ctx.port = viper.GetString("port")
	ctx.database = viper.GetString("database")
	ctx.username = viper.GetString("username")
	ctx.password = viper.GetString("password")

	addOpts := viper.GetString("additional_options")
	if len(addOpts) > 0 {
		ctx.additionalOptions = strings.Split(addOpts, " ")
	}

	err = ctx.dump()
	return
}

func (ctx *PostgreSQL) dumpArgs() []string {
	// mysqldump command
	dumpArgs := []string{}

	if len(ctx.host) > 0 {
		dumpArgs = append(dumpArgs, "--host="+ctx.host)
	}
	if len(ctx.port) > 0 {
		dumpArgs = append(dumpArgs, "--port="+ctx.port)
	}
	if len(ctx.username) > 0 {
		dumpArgs = append(dumpArgs, "--username="+ctx.username)
	}
	if len(ctx.additionalOptions) > 0 {
		dumpArgs = append(dumpArgs, ctx.additionalOptions...)
	}

	dumpArgs = append(dumpArgs, "-d", ctx.database)

	return dumpArgs
}

func (ctx *PostgreSQL) dump() error {
	logger.Info("-> Dumping PostgreSQL...")
	if len(ctx.password) > 0 {
		os.Setenv("PGPASSWORD", ctx.password)
	}

	pgDump, err := helper.CreateCmd("pg_dump", ctx.dumpArgs()...)
	if err != nil {
		return err
	}
	var errOut bytes.Buffer
	pgDump.Stderr = &errOut
	stdoutPipe, err := pgDump.StdoutPipe()
	if err != nil {
		return fmt.Errorf("-> Can't pipe stdout error: %s", err)
	}

	err = pgDump.Start()
	if err != nil {
		return fmt.Errorf("-> can't start pg_dump error: %s", err)
	}

	dumpFilePath := path.Join(ctx.dumpPath, ctx.database+".sql")

	ext, r, err := compressor.CompressTo(ctx.model, bufio.NewReader(stdoutPipe))
	if err != nil {
		return fmt.Errorf("-> can't compress pg_dump output: %s", err)
	}
	dumpFilePath = dumpFilePath + ext
	f, err := os.Create(dumpFilePath)
	if err != nil {
		return fmt.Errorf("-> can't dump to file output: %s", err)
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	if err != nil {
		return fmt.Errorf("-> error: can't copy dump output to file: %s", err)
	}

	if err = pgDump.Wait(); err != nil {
		return fmt.Errorf("-> Dump error: %s, stderr: %s", err, errOut.String())
	}

	logger.Info("dump path:", ctx.dumpPath)
	return nil
}
