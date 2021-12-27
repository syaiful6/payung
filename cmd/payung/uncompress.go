package main

import (
	"compress/gzip"
	"fmt"
	"github.com/andybalholm/brotli"
	"github.com/golang/snappy"
	"github.com/klauspost/compress/zstd"
	"github.com/spf13/cobra"
	"io"
	"os"
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
		outWrite, err := os.Open(outputFile)
		if err != nil {
			return err
		}
		defer outWrite.Close()

		outWrite = outWrite
	}

	_, err = io.Copy(output, sr)
	return err
}

func uncompressCommand() *cobra.Command {
	var (
		printToStdout bool
		outputFile    string
	)
	var uncompressCmd = &cobra.Command{
		Use:   "uncompress",
		Short: `Uncompress`,
		Long:  `Uncompress backup file`,
		Run: func(cmd *cobra.Command, arg []string) {
			cmd.Usage()
		},
	}
	var snappyCmd = &cobra.Command{
		Use:   "snappy",
		Short: `uncompress snappy`,
		Long:  `uncompress snappy`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return uncompressFormat("snappy", printToStdout, outputFile, args[0])
			}
			return uncompressFormat("snappy", printToStdout, outputFile, "")
		},
	}
	var zstdCmd = &cobra.Command{
		Use:   "zstd",
		Short: `uncompress zstd`,
		Long:  `uncompress zstd`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return uncompressFormat("snappy", printToStdout, outputFile, args[0])
			}
			return uncompressFormat("snappy", printToStdout, outputFile, "")
		},
	}
	var brotliCmd = &cobra.Command{
		Use:   "brotli",
		Short: `uncompress brotli`,
		Long:  `uncompress brotli`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return uncompressFormat("snappy", printToStdout, outputFile, args[0])
			}
			return uncompressFormat("snappy", printToStdout, outputFile, "")
		},
	}
	var gzipCmd = &cobra.Command{
		Use:   "gzip",
		Short: `uncompress gzip`,
		Long:  `uncompress gzip`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return uncompressFormat("snappy", printToStdout, outputFile, args[0])
			}
			return uncompressFormat("snappy", printToStdout, outputFile, "")
		},
	}

	uncompressCmd.PersistentFlags().BoolVarP(&printToStdout, "stdout", "c", false, "write on standard output")
	uncompressCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "", "output file; valid only if there is a single input entry")

	uncompressCmd.AddCommand(snappyCmd)
	uncompressCmd.AddCommand(zstdCmd)
	uncompressCmd.AddCommand(brotliCmd)
	uncompressCmd.AddCommand(gzipCmd)

	return uncompressCmd
}
