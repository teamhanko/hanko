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

func NewServePrivateCommand(config *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "private",
		Short: "Start the private portion of the hanko server",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			persister, err := persistence.New(config.Database)
			if err != nil {
				log.Fatal(err)
			}
			var wg sync.WaitGroup
			wg.Add(1)

			go server.StartPrivate(config, &wg, persister)

			wg.Wait()
		},
	}
}
