/*
Copyright © 2022 Hanko GmbH <developers@hanko.io>

*/

package migrate

import (
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/config"
)

//var persister *persistence.Persister

func NewMigrateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Database migration helpers",
		Long:  ``,
	}
}

func RegisterCommands(parent *cobra.Command, config *config.Config) {
	cmd := NewMigrateCmd()
	parent.AddCommand(cmd)
	cmd.AddCommand(NewMigrateUpCommand(config))
	cmd.AddCommand(NewMigrateDownCommand(config))
}
