package config

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/v3/config"
)

func NewShowCommand() *cobra.Command {
	var (
		configFile string
	)

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Parse config file and output as compacted JSON",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.Load(&configFile)
			if err != nil {
				log.Fatal(err)
			}

			err = cfg.PostProcess()
			if err != nil {
				log.Fatal(fmt.Errorf("failed to post process config: %w", err))
			}

			if err = cfg.Validate(); err != nil {
				log.Fatalf("failed to validate config: %s", err)
			}

			jsonBytes, err := json.Marshal(cfg)
			if err != nil {
				log.Fatal(fmt.Errorf("failed to marshal config to JSON: %w", err))
			}

			fmt.Println(string(jsonBytes))
		},
	}

	cmd.Flags().StringVar(&configFile, "config", config.DefaultConfigFilePath, "config file")

	return cmd
}
