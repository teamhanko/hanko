package schema

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/config"
	"log"
	"os/exec"
	"path/filepath"
)

func NewMarkdownConfigCommand() *cobra.Command {
	var (
		outPath string
	)

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Generate markdown for the backend config",
		Run: func(cmd *cobra.Command, args []string) {
			err := generateSchema(generateSchemaParams{
				structure:       &config.Config{},
				rootSchemaTitle: "Config",
				output:          filepath.Join(outPath, "config.schema.json"),
				extractComments: true,
				doNotReference:  true,
				commentPaths:    []string{"config", "ee"},
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
		".generated/docs/config",
		"Output path")

	return cmd
}
