package compressor

import (
	"io"

	"github.com/spf13/viper"
	"github.com/syaiful6/payung/config"
)

// Base compressor
type Base struct {
	model config.ModelConfig
	viper *viper.Viper
}

type Compressor interface {
	compressTo(r io.Reader) (string, io.Reader, error)
}

func newBase(model config.ModelConfig) (base Base) {
	base = Base{
		model: model,
		viper: model.CompressWith.Viper,
	}
	return
}

func CompressTo(model config.ModelConfig, r io.Reader) (string, io.Reader, error) {
	base := newBase(model)
	var ctx Compressor
	switch model.CompressWith.Type {
	case "gzip":
		ctx = &Gzip{Base: base}
	case "brotli":
		ctx = &Brotli{Base: base}
	case "snappy":
		ctx = &Snappy{Base: base}
	case "zstd":
		ctx = &ZStandard{Base: base}
	default:
		return "", r, nil
	}

	return ctx.compressTo(r)
}
