package archive

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/syaiful6/payung/compressor"
	"github.com/syaiful6/payung/config"
	"github.com/syaiful6/payung/helper"
	"github.com/syaiful6/payung/logger"
)

// Run archive
func Run(model config.ModelConfig) (err error) {
	if model.Archive == nil {
		return nil
	}

	logger.Info("------------- Archives -------------")

	helper.MkdirP(model.DumpPath)

	includes := model.Archive.GetStringSlice("includes")
	includes = cleanPaths(includes)

	excludes := model.Archive.GetStringSlice("excludes")
	excludes = cleanPaths(excludes)

	if len(includes) == 0 {
		return fmt.Errorf("archive.includes have no config")
	}
	logger.Info("=> includes", len(includes), "rules")

	opts := options(excludes, includes)

	tarCmd, err := helper.CreateCmd("tar", opts...)
	if err != nil {
		return err
	}
	stdoutPipe, err := tarCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("-> Can't pipe stdout error: %s", err)
	}
	err = tarCmd.Start()
	if err != nil {
		return fmt.Errorf("-> can't archive files: %s", err)
	}
	archiveFilePath := path.Join(model.DumpPath, "archive.tar")
	err, ext, r := compressor.CompressTo(model, bufio.NewReader(stdoutPipe))
	if err != nil {
		return fmt.Errorf("-> can't compress tar output: %s", err)
	}
	archiveFilePath = archiveFilePath + ext
	f, err := os.Create(archiveFilePath)
	if err != nil {
		return err
	}

	defer f.Close()
	_, err = io.Copy(f, r)
	if err != nil {
		return err
	}

	if err = tarCmd.Wait(); err != nil {
		return fmt.Errorf("-> archive error: %s", err)
	}

	logger.Info("------------- Archives -------------\n")

	return nil
}

func options(excludes, includes []string) (opts []string) {
	if helper.IsGnuTar {
		opts = append(opts, "--ignore-failed-read")
	}
	opts = append(opts, "-cPf", "-")

	for _, exclude := range excludes {
		opts = append(opts, "--exclude="+filepath.Clean(exclude))
	}

	opts = append(opts, includes...)

	return opts
}

func cleanPaths(paths []string) (results []string) {
	for _, p := range paths {
		results = append(results, filepath.Clean(p))
	}
	return
}
