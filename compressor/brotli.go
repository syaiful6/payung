package compressor

import (
	"io"
	"os"

	"github.com/andybalholm/brotli"
)

type Brotli struct {
	Base
}

func (ctx *Brotli) compressTo(r io.Reader, target string) error {
	f, err := os.Create(target + ".br")
	ctx.viper.SetDefault("level", brotli.DefaultCompression)
	level := ctx.viper.GetInt("level")
	if err != nil {
		return err
	}
	w := brotli.NewWriterLevel(f, level)
	_, err = io.Copy(w, r)
	w.Close()
	return err
}
