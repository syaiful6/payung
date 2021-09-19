package storage

import (
	"os"
	"path"

	"github.com/syaiful6/payung/helper"
	"github.com/thatique/awan/verr"

	// "crypto/tls"
	"time"

	"github.com/secsy/goftp"
	"github.com/syaiful6/payung/logger"
	"github.com/syaiful6/payung/packager"
)

// FTP storage
//
// type: ftp
// path: /backups
// host: ftp.your-host.com
// port: 21
// timeout: 30
// username:
// password:
type FTP struct {
	Base
	path     string
	host     string
	port     string
	username string
	password string

	client *goftp.Client
}

func (ctx *FTP) open() (err error) {
	ctx.viper.SetDefault("port", "21")
	ctx.viper.SetDefault("timeout", 300)

	ctx.host = helper.CleanHost(ctx.viper.GetString("host"))
	ctx.port = ctx.viper.GetString("port")
	ctx.path = ctx.viper.GetString("path")
	ctx.username = ctx.viper.GetString("username")
	ctx.password = ctx.viper.GetString("password")

	ftpConfig := goftp.Config{
		User:     ctx.viper.GetString("username"),
		Password: ctx.viper.GetString("password"),
		Timeout:  ctx.viper.GetDuration("timeout") * time.Second,
	}
	ctx.client, err = goftp.DialConfig(ftpConfig, ctx.host+":"+ctx.port)
	if err != nil {
		return err
	}
	return
}

func (ctx *FTP) close() {
	ctx.client.Close()
}

func (ctx *FTP) upload(backupPackage *packager.Package) (err error) {
	logger.Info("-> Uploading...")
	_, err = ctx.client.Stat(ctx.path)
	if os.IsNotExist(err) {
		if _, err := ctx.client.Mkdir(ctx.path); err != nil {
			return err
		}
	}

	remotePath := ctx.RemotePath(ctx.path, backupPackage)
	_, err = ctx.client.Stat(remotePath)
	if os.IsNotExist(err) {
		if _, err := ctx.client.Mkdir(remotePath); err != nil {
			return err
		}
	}

	fileNames := backupPackage.FileNames()
	// close files
	var files []*os.File
	defer func() {
		for i := range files {
			files[i].Close()
		}
	}()

	for i := range fileNames {
		file, err := os.Open(path.Join(ctx.model.TempPath, fileNames[1]))
		if err != nil {
			return err
		}
		files = append(files, file)
		fileName := path.Join(remotePath, fileNames[i])
		err = ctx.client.Store(fileName, file)
		if err != nil {
			return err
		}
	}

	logger.Info("Store successed")
	return nil
}

func (ctx *FTP) delete(backupPackage *packager.Package) error {
	remotePath := ctx.RemotePath(ctx.path, backupPackage)
	fileNames := backupPackage.FileNames()

	var errlist []error

	for i := range fileNames {
		err := ctx.client.Delete(path.Join(remotePath, fileNames[i]))
		if err != nil {
			errlist = append(errlist, err)
		}
	}

	if len(errlist) == 0 {
		err := ctx.client.Rmdir(remotePath)
		if err != nil {
			errlist = append(errlist, err)
		}
	}

	if len(errlist) > 0 {
		return verr.NewAggregate(errlist)
	}
	return nil
}
