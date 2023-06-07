package jwt

import (
	"github.com/spf13/cobra"
)

func NewJwtCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "jwt",
		Short: "Tools for handling JSON Web Tokens",
		Long:  ``,
	}
}

func RegisterCommands(parent *cobra.Command) {
	cmd := NewJwtCmd()
	parent.AddCommand(cmd)
	cmd.AddCommand(NewCreateCommand())
}
