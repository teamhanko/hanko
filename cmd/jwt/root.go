package jwt

import (
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/config"
	"github.com/teamhanko/hanko/persistence"
)

func NewJwtCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "jwt",
		Short: "Tools for handling JSON Web Tokens",
		Long:  ``,
	}
}

func RegisterCommands(parent *cobra.Command, cfg *config.Config, persister persistence.Persister) {
	cmd := NewJwtCmd()
	parent.AddCommand(cmd)
	cmd.AddCommand(NewCreateCommand(cfg, persister))
}
