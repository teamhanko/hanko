package certificate

import (
	"github.com/spf13/cobra"
)

func NewCertificateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "certificate",
		Short: "Tools for managing SAML certificates",
		Long:  ``,
	}
}

func RegisterCommands(parent *cobra.Command) {
	cmd := NewCertificateCmd()
	parent.AddCommand(cmd)
	cmd.AddCommand(NewCreateCommand())
	cmd.AddCommand(NewCreateAllCommand())
}
