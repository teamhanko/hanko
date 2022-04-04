/*
Copyright © 2022 Hanko GmbH <developers@hanko.io>

*/
package serve

import (
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/config"
	"github.com/teamhanko/hanko/persistence"
)

func NewServeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the hanko server",
		Long:  ``,
	}
}

func RegisterCommands(parent *cobra.Command, config *config.Config, persister *persistence.Persister) {
	cmd := NewServeCommand()
	parent.AddCommand(cmd)
	cmd.AddCommand(NewServePublicCommand(config, persister))
	cmd.AddCommand(NewServePrivateCommand(config, persister))
	cmd.AddCommand(NewServeAllCommand(config, persister))
}
