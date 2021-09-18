package compressor

import (
	"compress/gzip"
	"io"
	"os"
)

type Gzip struct {
	Base
}

func (ctx *Gzip) compressTo(r io.Reader, target string) error {
	f, err := os.Create(target + ".gz")

	if err != nil {
		return err
	}
	w := gzip.NewWriter(f)
	_, err = io.Copy(w, r)
	w.Close()
	return err
}
