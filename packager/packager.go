package packager

import (
	"path/filepath"

	"github.com/syaiful6/payung/config"
	"github.com/syaiful6/payung/encryptor"
	"github.com/syaiful6/payung/helper"
	"github.com/syaiful6/payung/logger"
)

type Packager struct {
	Config config.ModelConfig
}

func (p *Packager) Run(backupPackage *Package) (err error) {
	logger.Info("------------ Packaging the backup files -------------")

	filePath := filepath.Join(p.Config.TempPath, backupPackage.BaseName())
	opts := packagerOptions(p.Config, filePath)

	_, err = helper.Exec("tar", opts...)
	if err != nil {
		return
	}

	filePath, err = encryptor.Run(filePath, p.Config)
	if err != nil {
		logger.Error(err)
		return
	}
	splitter := Splitter{
		Config:       p.Config,
		ChunkSize:    p.Config.SplitIntoChunksOf,
		SuffixLength: 3,
		Package:      backupPackage,
	}

	if err = splitter.Split(); err != nil {
		return
	}
	logger.Info("------------ Packaging Complete! -------------")
	return

}

func packagerOptions(model config.ModelConfig, filePath string) (opts []string) {
	if helper.IsGnuTar {
		opts = append(opts, "--ignore-failed-read")
	}

	opts = append(opts, "-cf", filePath)
	opts = append(opts, "-C", model.TempPath, model.Name)

	return
}
