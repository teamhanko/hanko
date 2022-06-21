package jwt

import (
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/config"
)

func NewJwtCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "jwt",
		Short: "Tools for handling JSON Web Tokens",
		Long:  ``,
	}
}

func RegisterCommands(parent *cobra.Command, cfg *config.Config) {
	cmd := NewJwtCmd()
	parent.AddCommand(cmd)
	cmd.AddCommand(NewCreateCommand(cfg))
}
