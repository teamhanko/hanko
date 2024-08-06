package schema

import "github.com/spf13/cobra"

func NewSchemaCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "schema",
		Short: "JSONSchema related commands",
		Long:  ``,
	}
}

func RegisterCommands(parent *cobra.Command) {
	cmd := NewSchemaCommand()
	parent.AddCommand(cmd)
	cmd.AddCommand(NewJson2MdCommand())
}
