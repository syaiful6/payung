package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"

	"github.com/andybalholm/brotli"
	"github.com/golang/snappy"
	"github.com/klauspost/compress/zstd"
	"github.com/spf13/cobra"
)

func uncompressFormat(format string, printToStdout bool, outputFile string, inputFile string) error {
	var (
		input  io.Reader
		output io.Writer
		err    error
	)
	input = os.Stdin
	output = os.Stdout

	if inputFile != "" {
		fr, err := os.Open(inputFile)
		if err != nil {
			return err
		}
		defer fr.Close()
		input = fr
	}

	var sr io.Reader
	switch format {
	case "snappy":
		sr = snappy.NewReader(input)
	case "zstd":
		sr, err = zstd.NewReader(input)
		if err != nil {
			return err
		}
	case "brotli":
		sr = brotli.NewReader(input)
	case "gzip":
		sr, err = gzip.NewReader(input)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("%s invalid format", format)
	}

	if printToStdout {
		output = os.Stdout
	} else if outputFile != "" {
		output, err := os.Open(outputFile)
		if err != nil {
			return err
		}
		defer output.Close()
	}

	_, err = io.Copy(output, sr)
	return err
}

func decompressCommand() *cobra.Command {
	var (
		printToStdout bool
		outputFile    string
		format        string
	)
	var uncompressCmd = &cobra.Command{
		Use:   "decompress",
		Short: `decompress`,
		Long:  `decompress backup file`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return uncompressFormat(format, printToStdout, outputFile, args[0])
			}
			return uncompressFormat(format, printToStdout, outputFile, "")
		},
	}

	uncompressCmd.Flags().StringVarP(&format, "format", "f", "snappy", "The compress format, eg snappy, brotli, zstd or gzip")
	uncompressCmd.Flags().BoolVarP(&printToStdout, "stdout", "c", false, "write on standard output")
	uncompressCmd.Flags().StringVarP(&outputFile, "output", "o", "", "output file; valid only if there is a single input entry")

	return uncompressCmd
}
