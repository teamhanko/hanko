package schema

import "github.com/spf13/cobra"

func NewMarkdownCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "markdown",
		Short: "Generate markdown from JSON Schema",
		Long:  ``,
	}

	cmd.AddCommand(NewMarkdownConfigCommand())
	cmd.AddCommand(NewMarkdownImportCommand())

	return cmd
}
