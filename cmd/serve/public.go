/*
Copyright © 2022 Hanko GmbH <developers@hanko.io>

*/
package serve

import (
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/config"
	"github.com/teamhanko/hanko/persistence"
	"github.com/teamhanko/hanko/server"
	"log"
	"sync"
)

func NewServePublicCommand(config *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "public",
		Short: "Start the public portion of the hanko server",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			persister, err := persistence.New(config.Database)
			if err != nil {
				log.Fatal(err)
			}
			var wg sync.WaitGroup
			wg.Add(1)

			go server.StartPublic(config, &wg, persister)

			wg.Wait()
		},
	}
}
