package packager

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/syaiful6/payung/config"
	"github.com/syaiful6/payung/helper"
	"github.com/syaiful6/payung/logger"
)

type Splitter struct {
	Config       config.ModelConfig
	Package      *Package
	ChunkSize    int
	SuffixLength int
}

func (s *Splitter) Split(r io.Reader) (err error) {
	if s.ChunkSize <= 0 {
		return nil
	}
	logger.Info("------------- Splitter -------------")
	splitCmd, err := helper.CreateCmd("split", s.options()...)
	if err != nil {
		return err
	}
	splitCmd.Stdin = r

	if err = splitCmd.Start(); err != nil {
		return err
	}

	if err = splitCmd.Wait(); err != nil {
		return err
	}

	if err = s.AfterPackaging(); err != nil {
		return err
	}

	logger.Info("------------- Splitter completed! -------------\n")

	return
}

func (s *Splitter) options() (opts []string) {
	opts = append(opts, "-b", fmt.Sprintf("%dm", s.ChunkSize), "-")
	filename := filepath.Join(s.Config.TempPath, s.Package.BaseName())
	opts = append(opts, filename+"-")

	return opts
}

func (s *Splitter) AfterPackaging() error {
	suffixes, err := s.chunkSuffixes()
	if err != nil {
		return err
	}
	firstSuffix := strings.Repeat("a", s.SuffixLength)
	if len(suffixes) == 1 && suffixes[0] == firstSuffix {
		return os.Rename(
			filepath.Join(s.Config.TempPath, fmt.Sprintf("%s-%s", s.Package.BaseName(), firstSuffix)),
			filepath.Join(s.Config.TempPath, s.Package.BaseName()),
		)
	}
	s.Package.ChunkSuffixes = suffixes
	return nil
}

func (s *Splitter) chunkSuffixes() ([]string, error) {
	globPattern := filepath.Join(s.Config.TempPath, s.Package.BaseName()) + "-*"
	chunks, err := filepath.Glob(globPattern)
	if err != nil {
		return []string{}, err
	}
	sort.Strings(chunks)

	suffixes := []string{}
	for i := range chunks {
		chunk := chunks[i]
		ext := filepath.Ext(chunk)
		patterns := strings.Split(ext, "-")
		suffixes = append(suffixes, patterns[len(patterns)-1])
	}
	return suffixes, nil
}
