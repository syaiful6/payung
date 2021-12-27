package compressor

import (
	"github.com/golang/snappy"
	"io"
)

type Snappy struct {
	Base
}

func (ctx *Snappy) compressTo(r io.Reader) (error, string, io.Reader) {
	pr, pw := io.Pipe()

	go func() {
		var err error
		w := snappy.NewBufferedWriter(pw)
		if _, err = io.Copy(w, r); err != nil {
			panic(err)
		}
		if err = w.Flush(); err != nil {
			panic(err)
		}
		w.Close()
		pw.Close()
	}()

	return nil, ".snappy", pr
}
