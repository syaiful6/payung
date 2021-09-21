package encryptor

import (
	"io"

	"github.com/spf13/viper"
	"github.com/syaiful6/payung/config"
	"github.com/syaiful6/payung/logger"
)

// Base encryptor
type Base struct {
	model config.ModelConfig
	viper *viper.Viper
	r     io.Reader
}

// Context encryptor interface
type Context interface {
	perform() (r io.Reader, ext string, err error)
}

func newBase(r io.Reader, model config.ModelConfig) (base Base) {
	base = Base{
		r:     r,
		model: model,
		viper: model.EncryptWith.Viper,
	}
	return
}

// Run compressor
func Run(r io.Reader, model config.ModelConfig) (io.Reader, string, error) {
	base := newBase(r, model)
	var ctx Context
	switch model.EncryptWith.Type {
	case "openssl":
		ctx = &OpenSSL{Base: base}
	default:
		return r, "", nil
	}

	logger.Info("------------ Encryptor -------------")

	logger.Info("=> Encrypt | " + model.EncryptWith.Type)
	r, ext, err := ctx.perform()
	if err != nil {
		return nil, "", err
	}

	logger.Info("------------ Encryptor -------------\n")

	return r, ext, nil
}
