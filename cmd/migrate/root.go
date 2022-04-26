/*
Copyright © 2022 Hanko GmbH <developers@hanko.io>

*/

package migrate

import (
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/persistence"
)

//var persister *persistence.Persister

func NewMigrateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Database migration helpers",
		Long:  ``,
	}
}

func RegisterCommands(parent *cobra.Command, persister persistence.Migrator) {
	cmd := NewMigrateCmd()
	parent.AddCommand(cmd)
	cmd.AddCommand(NewMigrateUpCommand(persister))
	cmd.AddCommand(NewMigrateDownCommand(persister))
}
