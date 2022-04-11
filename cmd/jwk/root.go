package jwk

import (
	"github.com/spf13/cobra"
)


func NewMigrateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "jwk",
		Short: "Tools for handling JSON Web Keys",
		Long:  ``,
	}
}

func RegisterCommands(parent *cobra.Command) {
	cmd := NewMigrateCmd()
	parent.AddCommand(cmd)
}
