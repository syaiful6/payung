package encryptor

import (
	"fmt"
	"io"

	"github.com/syaiful6/payung/helper"
)

// OpenSSL encryptor for use openssl aes-256-cbc
//
// - base64: false
// - salt: true
// - password:
type OpenSSL struct {
	Base
	salt     bool
	base64   bool
	password string
}

func (ctx *OpenSSL) perform() (io.Reader, string, error) {
	sslViper := ctx.viper
	sslViper.SetDefault("salt", true)
	sslViper.SetDefault("base64", false)

	ctx.salt = sslViper.GetBool("salt")
	ctx.base64 = sslViper.GetBool("base64")
	ctx.password = sslViper.GetString("password")

	if len(ctx.password) == 0 {
		err := fmt.Errorf("password option is required")
		return nil, "", err
	}

	pr, pw := io.Pipe()

	go func() {
		opts := ctx.options()
		cmd, err := helper.CreateCmd("openssl", opts...)
		// set the stdin
		cmd.Stdin = ctx.r

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			panic(err)
		}
		if err = cmd.Start(); err != nil {
			panic(err)
		}

		if _, err := io.Copy(pw, stdout); err != nil {
			panic(err)
		}
		if err := cmd.Wait(); err != nil {
			panic(err)
		}
		pw.Close()
	}()

	return pr, ".enc", nil
}

func (ctx *OpenSSL) options() (opts []string) {
	opts = append(opts, "aes-256-cbc")
	if ctx.base64 {
		opts = append(opts, "-base64")
	}
	if ctx.salt {
		opts = append(opts, "-salt")
	}
	opts = append(opts, `-k`, ctx.password)
	return opts
}
