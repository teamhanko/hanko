/*
Copyright © 2022 Hanko GmbH <developers@hanko.io>
*/
package serve

import (
	"log"
	"sync"

	"github.com/gobuffalo/pop/v6"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/crypto/jwk/local_db"
	"github.com/teamhanko/hanko/backend/v3/mapper"
	"github.com/teamhanko/hanko/backend/v3/persistence"
	"github.com/teamhanko/hanko/backend/v3/saml"
	"github.com/teamhanko/hanko/backend/v3/server"
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
			pop.Debug = true
			if err != nil {
				log.Fatal(err)
			}
			persister := persistence.New(dbConnection)

			// Sync SAML providers from config to database for backward compatibility
			if cfg.Saml.Enabled {
				log.Println("Syncing SAML providers from config to database...")
				err := saml.SyncProviderConfigToDatabase(cfg, persister)
				if err != nil {
					log.Printf("SAML config sync failed: %v", err)
					// Don't fail startup - just log the error
				}
			}

			err = local_db.SyncSecretKeys(cfg, persister)
			if err != nil {
				log.Fatalf("Failed to sync secret keys: %v", err)
			}

			var wg sync.WaitGroup
			wg.Add(3)

			prometheus := echoprometheus.NewMiddleware("hanko")

			go server.StartPublic(cfg, &wg, persister, prometheus, authenticatorMetadata)
			go server.StartAdmin(cfg, &wg, persister, prometheus)
			go server.StartManagement(cfg, &wg, persister)

			wg.Wait()
		},
	}

	cmd.Flags().StringVar(&configFile, "config", config.DefaultConfigFilePath, "config file")
	cmd.Flags().StringVar(&authenticatorMetadataFile, "auth-meta", "", "authenticator metadata file")

	return cmd
}
