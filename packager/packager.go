package packager

import (
	"path"
	"time"

	"github.com/syaiful6/payung/config"
	"github.com/syaiful6/payung/helper"
	"github.com/syaiful6/payung/logger"
)

func Run(model config.ModelConfig) (archivePath string, err error) {
	logger.Info("------------ Packaging the backup files -------------")

	filePath := archiveFilePath(model, ".tar")
	opts := packagerOptions(model, filePath)

	_, err = helper.Exec("tar", opts...)
	if err == nil {
		archivePath = filePath
		return
	}
	logger.Info("------------ Packaging Complete! -------------")
	return
}

func archiveFilePath(model config.ModelConfig, ext string) string {
	return path.Join(model.TempPath, time.Now().Format("2006.01.02.15.04.05")+ext)
}

func packagerOptions(model config.ModelConfig, filePath string) (opts []string) {
	if helper.IsGnuTar {
		opts = append(opts, "--ignore-failed-read")
	}

	opts = append(opts, "-cf", filePath)
	opts = append(opts, "-C", model.TempPath, model.Name)

	return
}
