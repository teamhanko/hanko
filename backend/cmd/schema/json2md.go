package schema

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/invopop/jsonschema"
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/config"
	"log"
	"os"
	"os/exec"
)

func NewJson2MdCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "json2md",
		Short: "Generate markdown from JSONSchema",
		Run: func(cmd *cobra.Command, args []string) {
			r := new(jsonschema.Reflector)
			r.DoNotReference = true
			if err := r.AddGoComments("github.com/teamhanko/hanko/backend", "./config"); err != nil {
				log.Fatal(err)
			}

			if err := r.AddGoComments("github.com/teamhanko/hanko/backend", "./ee"); err != nil {
				log.Fatal(err)
			}

			s := r.Reflect(&config.Config{})
			s.Title = "Config"

			data, err := json.MarshalIndent(s, "", "  ")
			if err != nil {
				log.Fatal(err)
			}

			outPath := "./docs/.generated/config"
			if _, err := os.Stat(outPath); errors.Is(err, os.ErrNotExist) {
				err := os.MkdirAll(outPath, 0750)
				if err != nil {
					log.Fatal(err)
				}
			}

			err = os.WriteFile(fmt.Sprintf("%s/hanko.config.json", outPath), data, 0600)
			if err != nil {
				log.Fatal(err)
			}

			out, err := exec.Command("npx",
				"@adobe/jsonschema2md",
				"--input=docs/.generated/config",
				"--out=docs/.generated/config/md",
				"--schema-extension=config.json",
				"--example-format=yaml",
				"--header=false",
				"--skip=definedinfact",
				"--skip=typesection",
				"--schema-out=-",
				"--properties=format",
				"--no-readme=true").
				CombinedOutput()

			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(string(out))
		},
	}

	return cmd
}
