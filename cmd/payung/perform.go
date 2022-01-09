package main

import (
	"github.com/spf13/cobra"
	"github.com/syaiful6/payung/config"
	"github.com/syaiful6/payung/model"
)

func performCmd() *cobra.Command {
	var (
		modelName  string
		configFile string
		dumpPath   string
	)
	var performCommand = &cobra.Command{
		Use:   "perform",
		Short: `Perform backup`,
		Long:  `Perform backup`,
		RunE: func(cmd *cobra.Command, args []string) error {
			config.Init(configFile, dumpPath)
			if len(modelName) == 0 {
				performAll()
			} else {
				performOne(modelName)
			}

			return nil
		},
	}
	performCommand.Flags().StringVarP(&modelName, "model", "m", "", "Model name that you want execute")
	performCommand.Flags().StringVarP(&configFile, "config", "c", "", "Special a config file")
	performCommand.Flags().StringVarP(&dumpPath, "dumpfile", "d", "", "special Dump path folder")
	return performCommand
}

func performAll() {
	for _, modelConfig := range config.Models {
		m := model.Model{
			Config: modelConfig,
		}
		m.Perform()
	}
}

func performOne(modelName string) {
	for _, modelConfig := range config.Models {
		if modelConfig.Name == modelName {
			m := model.Model{
				Config: modelConfig,
			}
			m.Perform()
			return
		}
	}
}
