package storage

import (
	"path"

	"github.com/syaiful6/payung/helper"
	"github.com/syaiful6/payung/logger"
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

func (ctx *Local) upload(fileKey string) (err error) {
	_, err = helper.Exec("cp", ctx.archivePath, ctx.destPath)
	if err != nil {
		return err
	}
	logger.Info("Store successed", ctx.destPath)
	return nil
}

func (ctx *Local) delete(fileKey string) (err error) {
	_, err = helper.Exec("rm", path.Join(ctx.destPath, fileKey))
	return
}
