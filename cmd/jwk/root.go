package jwk

import (
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/config"
	"github.com/teamhanko/hanko/persistence"
)


func NewMigrateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "jwk",
		Short: "Tools for handling JSON Web Keys",
		Long:  ``,
	}
}

func RegisterCommands(parent *cobra.Command, cfg *config.Config, persister *persistence.Persister) {
	cmd := NewMigrateCmd()
	parent.AddCommand(cmd)
	cmd.AddCommand(NewCreateCommand(cfg, persister))
}
