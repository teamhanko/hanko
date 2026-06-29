package config

import (
	"github.com/spf13/cobra"
)

func NewConfigCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Configuration related commands",
		Long:  ``,
	}
}

func RegisterCommands(parent *cobra.Command) {
	cmd := NewConfigCommand()
	cmd.AddCommand(NewShowCommand())
	parent.AddCommand(cmd)
}
