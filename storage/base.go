package storage

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/syaiful6/payung/config"
	"github.com/syaiful6/payung/logger"
)

// Base storage
type Base struct {
	model       config.ModelConfig
	archivePath string
	viper       *viper.Viper
	keep        int
}

// Context storage interface
type Context interface {
	open() error
	close()
	upload(fileKey string) error
	delete(fileKey string) error
}

func newBase(model config.ModelConfig, archivePath string) (base Base) {
	base = Base{
		model:       model,
		archivePath: archivePath,
		viper:       model.StoreWith.Viper,
	}

	if base.viper != nil {
		base.keep = base.viper.GetInt("keep")
	}

	return
}

// Run storage
func Run(model config.ModelConfig, archivePaths []string) (err error) {
	logger.Info("------------- Storage --------------")
	logger.Info("=> Storage | " + model.StoreWith.Type)

	for i := range archivePaths {
		archivePath := archivePaths[i]
		if err = upload(model, archivePath); err != nil {
			return
		}
	}

	logger.Info("------------- Storage --------------\n")
	return nil
}

func upload(model config.ModelConfig, archivePath string) (err error) {
	newFileKey := filepath.Base(archivePath)
	base := newBase(model, archivePath)
	var ctx Context
	switch model.StoreWith.Type {
	case "local":
		ctx = &Local{Base: base}
	case "ftp":
		ctx = &FTP{Base: base}
	case "scp":
		ctx = &SCP{Base: base}
	case "s3":
		ctx = &S3{Base: base}
	case "oss":
		ctx = &OSS{Base: base}
	default:
		return fmt.Errorf("[%s] storage type has not implement", model.StoreWith.Type)
	}

	err = ctx.open()
	if err != nil {
		return err
	}
	defer ctx.close()

	err = ctx.upload(newFileKey)
	if err != nil {
		return err
	}

	cycler := Cycler{}
	cycler.run(model.Name, newFileKey, base.keep, ctx.delete)
	return
}
