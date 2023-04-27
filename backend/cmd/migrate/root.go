/*
Copyright © 2022 Hanko GmbH <developers@hanko.io>

*/

package migrate

import (
	"github.com/spf13/cobra"
)

//var persister *persistence.Persister

func NewMigrateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Database migration helpers",
		Long:  ``,
	}
}

func RegisterCommands(parent *cobra.Command) {
	cmd := NewMigrateCmd()
	parent.AddCommand(cmd)
	cmd.AddCommand(NewMigrateUpCommand())
	cmd.AddCommand(NewMigrateDownCommand())
}
