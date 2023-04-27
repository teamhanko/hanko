/*
Copyright © 2022 Hanko GmbH <developers@hanko.io>

*/
package serve

import (
	"github.com/spf13/cobra"
)

func NewServeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the hanko server",
		Long:  ``,
	}
}

func RegisterCommands(parent *cobra.Command) {
	cmd := NewServeCommand()
	parent.AddCommand(cmd)
	cmd.AddCommand(NewServePublicCommand())
	cmd.AddCommand(NewServeAdminCommand())
	cmd.AddCommand(NewServeAllCommand())
}
