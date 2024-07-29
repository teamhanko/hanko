package schema

import (
	"encoding/json"
	"github.com/invopop/jsonschema"
	"github.com/spf13/cobra"
	"os"
)

func NewGenerateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "generate",
		Short: "Generate JSON Schema",
		Long:  ``,
	}
}

func RegisterCommands(parent *cobra.Command) {
	cmd := NewGenerateCommand()
	cmd.AddCommand(NewGenerateConfigCommand())
	cmd.AddCommand(NewGenerateImportCommand())
	parent.AddCommand(cmd)
}

func generateSchema(structure interface{}, outPath string, extractComments, doNotReference bool, commentPaths ...string) error {
	r := new(jsonschema.Reflector)

	if doNotReference {
		r.DoNotReference = true
	}

	if extractComments {
		for _, path := range commentPaths {
			if err := r.AddGoComments("github.com/teamhanko/hanko/backend", path); err != nil {
				return err
			}
		}
	}
	s := r.Reflect(structure)

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(outPath, data, 0600)
	if err != nil {
		return err
	}
	return nil
}
