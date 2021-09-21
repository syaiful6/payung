package model

import (
	"os"
	"time"

	"github.com/syaiful6/payung/archive"
	"github.com/syaiful6/payung/config"
	"github.com/syaiful6/payung/database"
	"github.com/syaiful6/payung/logger"
	"github.com/syaiful6/payung/packager"
	"github.com/syaiful6/payung/storage"
)

// Model class
type Model struct {
	Config config.ModelConfig
}

func (ctx Model) Perform() {
	quit := hookSignals()
	serveErr := make(chan error)

	go func() {
		serveErr <- ctx.run()
		logger.Info("backup terminated")
	}()

	select {
	case err := <-serveErr:
		if err != nil {
			logger.Error(err)
		}
		ctx.cleanup()
	case <-quit:
		logger.Info("Backup interupted")
		ctx.cleanup()
	}
}

// Perform model
func (ctx Model) run() (err error) {
	logger.Info("======== " + ctx.Config.Name + " ========")
	logger.Info("WorkDir:", ctx.Config.DumpPath+"\n")

	backupPackage := packager.NewPackage(ctx.Config.Name, time.Now())

	err = database.Run(ctx.Config)
	if err != nil {
		logger.Error(err)
		return
	}

	if ctx.Config.Archive != nil {
		err = archive.Run(ctx.Config)
		if err != nil {
			logger.Error(err)
			return
		}
	}

	packager := &packager.Packager{Config: ctx.Config}

	if err = packager.Run(backupPackage); err != nil {
		logger.Error(err)
		return
	}

	err = storage.Run(ctx.Config, backupPackage)
	if err != nil {
		logger.Error(err)
		return
	}

	return
}

// Cleanup model temp files
func (ctx Model) cleanup() {
	logger.Info("Cleanup temp: " + ctx.Config.TempPath + "/\n")
	err := os.RemoveAll(ctx.Config.TempPath)
	if err != nil {
		logger.Error("Cleanup temp dir "+ctx.Config.TempPath+" error:", err)
	}
	logger.Info("======= End " + ctx.Config.Name + " =======\n\n")
}
