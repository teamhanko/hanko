package schema

import (
	"github.com/spf13/cobra"
)

func NewSchemaCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "schema",
		Short: "JSONSchema related commands",
		Long:  ``,
	}
}

func RegisterCommands(parent *cobra.Command) {
	cmd := NewSchemaCommand()
	cmd.AddCommand(NewGenerateCommand())
	cmd.AddCommand(NewMarkdownCommand())
	parent.AddCommand(cmd)
}
