/*
Copyright © 2022 Hanko GmbH <developers@hanko.io>
*/
package serve

import (
	"log"
	"sync"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/mapper"
	"github.com/teamhanko/hanko/backend/v2/persistence"
	"github.com/teamhanko/hanko/backend/v2/server"
)

func NewServeAllCommand() *cobra.Command {
	var (
		configFile                string
		authenticatorMetadataFile string
	)

	cmd := &cobra.Command{
		Use:   "all",
		Short: "Start the public and admin portion of the hanko server",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.Load(&configFile)
			if err != nil {
				log.Fatal(err)
			}

			authenticatorMetadata := mapper.LoadAuthenticatorMetadata(&authenticatorMetadataFile)

			dbConnection, err := persistence.NewConnection(cfg.Database)
			if err != nil {
				log.Fatal(err)
			}
			persister := persistence.New(dbConnection, nil)
			var wg sync.WaitGroup
			wg.Add(2)

			prometheus := echoprometheus.NewMiddleware("hanko")

			go server.StartPublic(cfg, &wg, persister, prometheus, authenticatorMetadata)
			go server.StartAdmin(cfg, &wg, persister, prometheus)

			wg.Wait()
		},
	}

	cmd.Flags().StringVar(&configFile, "config", config.DefaultConfigFilePath, "config file")
	cmd.Flags().StringVar(&authenticatorMetadataFile, "auth-meta", "", "authenticator metadata file")

	return cmd
}
