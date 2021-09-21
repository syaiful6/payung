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
	compressTo(r io.Reader) (error, string, io.Reader)
}

func newBase(model config.ModelConfig) (base Base) {
	base = Base{
		model: model,
		viper: model.CompressWith.Viper,
	}
	return
}

func CompressTo(model config.ModelConfig, r io.Reader) (error, string, io.Reader) {
	base := newBase(model)
	var ctx Compressor
	switch model.CompressWith.Type {
	case "gzip":
		ctx = &Gzip{Base: base}
	case "brotli":
		ctx = &Brotli{Base: base}
	default:
		return nil, "", r
	}

	return ctx.compressTo(r)
}
