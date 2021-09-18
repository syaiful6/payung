package splitter

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/syaiful6/payung/config"
	"github.com/syaiful6/payung/helper"
	"github.com/syaiful6/payung/logger"
)

// Run split
func Run(archivePath string, model config.ModelConfig) (archivePaths []string, err error) {
	// Return archivePath by default
	archivePaths = append(archivePaths, archivePath)

	if model.SplitIntoChunksOf <= 0 {
		return
	}

	opts := options(archivePath, model.SplitIntoChunksOf)

	logger.Info("------------- Splitter -------------")
	helper.Exec("split", opts...)

	logger.Info("=> split", opts)

	// Return chunk path after success
	archivePaths = chunks(archivePath)

	logger.Info("=> into", len(archivePaths), "chunks")

	logger.Info("------------- Splitter -------------\n")

	return
}

func options(archivePath string, chunkSize int) (opts []string) {
	opts = append(opts, "-b", fmt.Sprintf("%dm", chunkSize))
	opts = append(opts, archivePath, archivePath+"-")

	return opts
}

func chunks(archivePath string) (chunkPaths []string) {
	chunkPaths, err := filepath.Glob(archivePath + "-*")
	if err != nil {
		logger.Error("Get split chunks files error:", err)
	}

	sort.Strings(chunkPaths)

	return
}
