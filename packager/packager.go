package packager

import (
	"io"
	"os"
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
	var (
		output io.Reader
		encExt string
	)
	logger.Info("------------ Packaging the backup files -------------")

	opts := packagerOptions(p.Config)
	tarCmd, err := helper.CreateCmd("tar", opts...)
	if err != nil {
		return
	}
	output, err = tarCmd.StdoutPipe()
	if err != nil {
		return
	}
	err = tarCmd.Start()
	if err != nil {
		return err
	}

	output, encExt, err = encryptor.Run(output, p.Config)
	if err != nil {
		logger.Error(err)
		return
	}
	backupPackage.Extension += encExt

	if p.Config.SplitIntoChunksOf <= 0 {
		packageFile := filepath.Join(p.Config.TempPath, backupPackage.BaseName())
		f, err := os.Create(packageFile)
		if err != nil {
			return err
		}
		if _, err = io.Copy(f, output); err != nil {
			return err
		}
	} else {
		splitter := Splitter{
			Config:       p.Config,
			ChunkSize:    p.Config.SplitIntoChunksOf,
			SuffixLength: 3,
			Package:      backupPackage,
		}

		if err = splitter.Split(output); err != nil {
			return
		}
	}

	logger.Info("------------ Packaging Complete! -------------")
	return

}

func packagerOptions(model config.ModelConfig) (opts []string) {
	if helper.IsGnuTar {
		opts = append(opts, "--ignore-failed-read")
	}

	opts = append(opts, "-cf", "-")
	opts = append(opts, "-C", model.TempPath, model.Name)

	return
}
