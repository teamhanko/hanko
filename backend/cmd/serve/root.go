/*
Copyright © 2022 Hanko GmbH <developers@hanko.io>

*/
package serve

import (
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/config"
)

func NewServeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the hanko server",
		Long:  ``,
	}
}

func RegisterCommands(parent *cobra.Command, config *config.Config) {
	cmd := NewServeCommand()
	parent.AddCommand(cmd)
	cmd.AddCommand(NewServePublicCommand(config))
	cmd.AddCommand(NewServePrivateCommand(config))
	cmd.AddCommand(NewServeAllCommand(config))
}
