package saml

import (
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/v3/cmd/saml/certificate"
)

func NewSamlCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "saml",
		Short: "Tools for managing SAML configuration",
		Long:  ``,
	}
}

func RegisterCommands(parent *cobra.Command) {
	cmd := NewSamlCmd()
	parent.AddCommand(cmd)
	certificate.RegisterCommands(cmd)
}
