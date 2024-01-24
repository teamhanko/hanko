/*
Copyright © 2022 Hanko GmbH <developers@hanko.io>
*/
package serve

import (
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/mapper"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/server"
	"log"
	"sync"
)

func NewServePublicCommand() *cobra.Command {
	var (
		configFile                string
		authenticatorMetadataFile string
	)

	cmd := &cobra.Command{
		Use:   "public",
		Short: "Start the public portion of the hanko server",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.Load(&configFile)
			if err != nil {
				log.Fatal(err)
			}

			authenticatorMetadata := mapper.LoadAuthenticatorMetadata(&authenticatorMetadataFile)

			persister, err := persistence.New(cfg.Database)
			if err != nil {
				log.Fatal(err)
			}
			var wg sync.WaitGroup
			wg.Add(1)

			go server.StartPublic(cfg, &wg, persister, nil, authenticatorMetadata)

			wg.Wait()
		},
	}

	cmd.Flags().StringVar(&configFile, "config", config.DefaultConfigFilePath, "config file")
	cmd.Flags().StringVar(&authenticatorMetadataFile, "auth-meta", "", "authenticator metadata file")

	return cmd
}
