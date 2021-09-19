package storage

import (
	"path"

	"github.com/syaiful6/payung/helper"
	"github.com/syaiful6/payung/logger"
	"github.com/syaiful6/payung/packager"
)

// Local storage
//
// type: local
// path: /data/backups
type Local struct {
	Base
	destPath string
}

func (ctx *Local) open() (err error) {
	ctx.destPath = ctx.model.StoreWith.Viper.GetString("path")
	helper.MkdirP(ctx.destPath)
	return
}

func (ctx *Local) close() {}

func (ctx *Local) upload(backupPackage *packager.Package) (err error) {
	remotePath := ctx.RemotePath(ctx.destPath, backupPackage)
	helper.MkdirP(remotePath)

	fileNames := backupPackage.FileNames()
	for i := range fileNames {
		src := path.Join(ctx.model.TempPath, fileNames[i])
		dest := path.Join(remotePath, fileNames[i])
		if _, err = helper.Exec("cp", src, dest); err != nil {
			return err
		}
	}

	logger.Info("Store successed", ctx.destPath)
	return nil
}

func (ctx *Local) delete(backupPackage *packager.Package) (err error) {
	remotePath := ctx.RemotePath(ctx.destPath, backupPackage)

	_, err = helper.Exec("rm", "-rf", remotePath)
	return
}
