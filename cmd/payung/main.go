package main

import (
	"github.com/spf13/cobra"
	"github.com/syaiful6/payung/version"
)

// show application version number
var showVersion bool

func main() {
	execute()
}

func execute() error {
	// Rootcmd is the rootcommand for
	var rootCmd = &cobra.Command{
		Use:   "Payung",
		Short: "Easy full stack backup operations on UNIX-like systems.",
		Long:  `Easy full stack backup operations on UNIX-like systems.`,
		Run: func(cmd *cobra.Command, args []string) {
			if showVersion {
				version.PrintVersion()
				return
			}
			cmd.Usage()
		},
	}
	rootCmd.AddCommand(performCmd())
	rootCmd.AddCommand(uncompressCommand())

	return rootCmd.Execute()
}
