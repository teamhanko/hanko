/*
Copyright © 2022 Hanko GmbH <developers@hanko.io>
*/
package serve

import (
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/server"
	"log"
	"sync"
)

func NewServeAllCommand() *cobra.Command {
	var (
		configFile string
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

			persister, err := persistence.New(cfg.Database)
			if err != nil {
				log.Fatal(err)
			}
			var wg sync.WaitGroup
			wg.Add(2)

			prometheus := echoprometheus.NewMiddleware("hanko")

			go server.StartPublic(cfg, &wg, persister, prometheus)
			go server.StartAdmin(cfg, &wg, persister, prometheus)

			wg.Wait()
		},
	}

	cmd.Flags().StringVar(&configFile, "config", config.DefaultConfigFilePath, "config file")

	return cmd
}
