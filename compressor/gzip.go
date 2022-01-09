package compressor

import (
	"compress/gzip"
	"io"
)

type Gzip struct {
	Base
}

func (ctx *Gzip) compressTo(r io.Reader) (string, io.Reader, error) {
	ctx.viper.SetDefault("level", gzip.DefaultCompression)
	level := ctx.viper.GetInt("level")

	pr, pw := io.Pipe()

	go func() {
		w, err := gzip.NewWriterLevel(pw, level)
		if err != nil {
			panic(err)
		}
		if _, err := io.Copy(w, r); err != nil {
			panic(err)
		}
		w.Close()
		pw.Close()
	}()

	return ".gz", pr, nil
}
