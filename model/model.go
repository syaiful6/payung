package model

import (
	"os"

	"github.com/syaiful6/payung/archive"
	"github.com/syaiful6/payung/config"
	"github.com/syaiful6/payung/database"
	"github.com/syaiful6/payung/encryptor"
	"github.com/syaiful6/payung/logger"
	"github.com/syaiful6/payung/packager"
	"github.com/syaiful6/payung/splitter"
	"github.com/syaiful6/payung/storage"
)

// Model class
type Model struct {
	Config config.ModelConfig
}

// Perform model
func (ctx Model) Perform() {
	var err error
	logger.Info("======== " + ctx.Config.Name + " ========")
	logger.Info("WorkDir:", ctx.Config.DumpPath+"\n")

	defer func() {
		if r := recover(); r != nil {
			ctx.cleanup()
		}

		ctx.cleanup()
	}()

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

	archivePath, err := packager.Run(ctx.Config)
	if err != nil {
		logger.Error(err)
		return
	}

	archivePath, err = encryptor.Run(archivePath, ctx.Config)
	if err != nil {
		logger.Error(err)
		return
	}

	archivePaths, err := splitter.Run(archivePath, ctx.Config)
	if err != nil {
		logger.Error(err)
		return
	}

	err = storage.Run(ctx.Config, archivePaths)
	if err != nil {
		logger.Error(err)
		return
	}
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
