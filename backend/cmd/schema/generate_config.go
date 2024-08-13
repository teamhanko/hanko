package schema

import (
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/config"
	"log"
)

func NewGenerateConfigCommand() *cobra.Command {
	var (
		output string
	)

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Generate JSON schema for the backend config",
		Run: func(cmd *cobra.Command, args []string) {
			err := generateSchema(generateSchemaParams{
				structure:       &config.Config{},
				rootSchemaTitle: "Config",
				output:          output,
				extractComments: true,
				doNotReference:  false,
				commentPaths:    []string{"config", "ee"},
			})
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	cmd.Flags().StringVarP(
		&output, "output",
		"o",
		"json_schema/hanko.config.json",
		"Output file")

	return cmd
}
