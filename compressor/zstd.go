package compressor

import (
	"io"

	"github.com/klauspost/compress/zstd"
)

type ZStandard struct {
	Base
}

func (ctx *ZStandard) compressTo(r io.Reader) (error, string, io.Reader) {
	ctx.viper.SetDefault("level", 3)
	level := ctx.viper.GetInt("level")
	pr, pw := io.Pipe()

	go func() {
		var err error
		w, err := zstd.NewWriter(pw, zstd.WithEncoderLevel(zstd.EncoderLevelFromZstd(level)))
		if err != nil {
			panic(err)
		}
		if _, err = io.Copy(w, r); err != nil {
			panic(err)
		}
		if err = w.Flush(); err != nil {
			panic(err)
		}
		w.Close()
		pw.Close()
	}()

	return nil, ".zstd", pr
}
