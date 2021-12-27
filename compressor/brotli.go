package compressor

import (
	"io"

	"github.com/andybalholm/brotli"
)

type Brotli struct {
	Base
}

func (ctx *Brotli) compressTo(r io.Reader) (error, string, io.Reader) {
	ctx.viper.SetDefault("level", brotli.DefaultCompression)
	level := ctx.viper.GetInt("level")

	pr, pw := io.Pipe()

	go func() {
		w := brotli.NewWriterLevel(pw, level)
		if _, err := io.Copy(w, r); err != nil {
			panic(err)
		}
		w.Close()
		pw.Close()
	}()

	return nil, ".br", pr
}
