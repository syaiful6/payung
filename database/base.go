package database

import (
	"fmt"
	"path"

	"github.com/spf13/viper"
	"github.com/syaiful6/payung/config"
	"github.com/syaiful6/payung/helper"
	"github.com/syaiful6/payung/logger"
	"github.com/syaiful6/payung/packager"
)

// Base database
type Base struct {
	model    config.ModelConfig
	dbConfig config.SubConfig
	viper    *viper.Viper
	name     string
	dumpPath string
}

// Context database interface
type Context interface {
	perform(*packager.Package) error
}

func newBase(model config.ModelConfig, dbConfig config.SubConfig) (base Base) {
	base = Base{
		model:    model,
		dbConfig: dbConfig,
		viper:    dbConfig.Viper,
		name:     dbConfig.Name,
	}
	base.dumpPath = path.Join(model.DumpPath, dbConfig.Type, base.name)
	helper.MkdirP(base.dumpPath)
	return
}

// New - initialize Database
func runModel(model config.ModelConfig, dbConfig config.SubConfig, backupPackage *packager.Package) (err error) {
	base := newBase(model, dbConfig)
	var ctx Context
	switch dbConfig.Type {
	case "mysql":
		ctx = &MySQL{Base: base}
	case "mariabackup":
		ctx = &MariaBackup{Base: base}
	case "redis":
		ctx = &Redis{Base: base}
	case "postgresql":
		ctx = &PostgreSQL{Base: base}
	case "mongodb":
		ctx = &MongoDB{Base: base}
	default:
		logger.Warn(fmt.Errorf("model: %s databases.%s config `type: %s`, but is not implement", model.Name, dbConfig.Name, dbConfig.Type))
		return
	}

	logger.Info("=> database |", dbConfig.Type, ":", base.name)

	// perform
	err = ctx.perform(backupPackage)
	if err != nil {
		return err
	}
	logger.Info("")

	return
}

// Run databases
func Run(model config.ModelConfig, backupPackage *packager.Package) error {
	if len(model.Databases) == 0 {
		return nil
	}

	logger.Info("------------- Databases -------------")
	for _, dbCfg := range model.Databases {
		err := runModel(model, dbCfg, backupPackage)
		if err != nil {
			return err
		}
	}
	logger.Info("------------- Databases -------------\n")

	return nil
}
