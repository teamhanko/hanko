package schema

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/cmd/user"
	"log"
	"os/exec"
	"path/filepath"
)

func NewMarkdownImportCommand() *cobra.Command {
	var (
		outPath string
	)

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Generate markdown for the user import schema",
		Run: func(cmd *cobra.Command, args []string) {
			err := generateSchema(generateSchemaParams{
				structure:       &user.ImportOrExportList{},
				rootSchemaTitle: "User import",
				output:          filepath.Join(outPath, "user_import.schema.json"),
				extractComments: true,
				doNotReference:  true,
				commentPaths:    []string{"cmd/user"},
			})

			out, err := exec.Command("npx",
				"@adobe/jsonschema2md",
				fmt.Sprintf("--input=%s", outPath),
				fmt.Sprintf("--out=%s", outPath),
				"--schema-extension=schema.json",
				"--example-format=yaml",
				"--header=false",
				"--skip=definedinfact",
				"--skip=typefact",
				"--schema-out=-",
				"--properties=format",
				"--no-readme=true").
				CombinedOutput()

			if err != nil {
				log.Fatal(err)
			}

			log.Println(string(out))

			outPathAbs, _ := filepath.Abs(outPath)
			log.Printf("successfully generated markdown from JSON schema at: %s\n", outPathAbs)
		},
	}

	cmd.Flags().StringVarP(
		&outPath, "out-path",
		"o",
		".generated/docs/import",
		"Output path")

	return cmd
}
