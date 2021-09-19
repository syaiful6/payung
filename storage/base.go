package storage

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
	"github.com/syaiful6/payung/config"
	"github.com/syaiful6/payung/logger"
	"github.com/syaiful6/payung/packager"
)

// Base storage
type Base struct {
	model         config.ModelConfig
	time          time.Time
	backupPackage *packager.Package
	viper         *viper.Viper
	keep          int
}

func (b *Base) RemotePath(path string, backupPackage *packager.Package) string {
	timestr := backupPackage.Time.Format("2006.01.02.15.04.05")
	if path != "" {
		return filepath.Join(path, backupPackage.Name, timestr)
	}
	return filepath.Join(backupPackage.Name, timestr)
}

// Context storage interface
type Context interface {
	open() error
	close()
	upload(backupPackage *packager.Package) error
	delete(backupPackage *packager.Package) error
}

func newBase(model config.ModelConfig, backupPackage *packager.Package) (base Base) {
	base = Base{
		model:         model,
		time:          time.Now(),
		backupPackage: backupPackage,
		viper:         model.StoreWith.Viper,
	}

	if base.viper != nil {
		base.keep = base.viper.GetInt("keep")
	}

	return
}

// Run storage
func Run(model config.ModelConfig, backupPackage *packager.Package) (err error) {
	logger.Info("------------- Storage --------------")
	logger.Info("=> Storage | " + model.StoreWith.Type)

	if err = upload(model, backupPackage); err != nil {
		return
	}

	logger.Info("------------- Storage --------------\n")
	return nil
}

func upload(model config.ModelConfig, backupPackage *packager.Package) (err error) {
	base := newBase(model, backupPackage)
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

	err = ctx.upload(backupPackage)
	if err != nil {
		return err
	}

	cycler := Cycler{}
	cycler.run(model.Name, *backupPackage, base.keep, ctx.delete)

	return
}
