package schema

import (
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/cmd/user"
	"log"
)

func NewGenerateImportCommand() *cobra.Command {
	var (
		output string
	)

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Generate JSON schema for the user import/export",
		Run: func(cmd *cobra.Command, args []string) {
			err := generateSchema(generateSchemaParams{
				structure:       &user.ImportOrExportList{},
				rootSchemaTitle: "User import",
				output:          output,
				extractComments: true,
				doNotReference:  false,
				commentPaths:    []string{"cmd/user"},
			})
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	cmd.Flags().StringVarP(
		&output, "output",
		"o",
		"json_schema/hanko.user_import.json",
		"Output file")

	return cmd
}
