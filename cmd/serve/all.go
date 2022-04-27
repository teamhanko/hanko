/*
Copyright © 2022 Hanko GmbH <developers@hanko.io>

*/
package serve

import (
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/config"
	"github.com/teamhanko/hanko/persistence"
	"github.com/teamhanko/hanko/server"
	"sync"
)

func NewServeAllCommand(config *config.Config, persister persistence.Persister) *cobra.Command {
	return &cobra.Command{
		Use:   "all",
		Short: "Start the public and private portion of the hanko server",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			var wg sync.WaitGroup
			wg.Add(2)

			go server.StartPublic(config, &wg, persister)
			go server.StartPrivate(config, &wg, persister)

			wg.Wait()
		},
	}
}
