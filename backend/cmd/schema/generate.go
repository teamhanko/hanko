package schema

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/invopop/jsonschema"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
)

func NewGenerateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate JSON Schema",
		Long:  ``,
	}

	cmd.AddCommand(NewGenerateImportCommand())
	cmd.AddCommand(NewGenerateConfigCommand())

	return cmd
}

type generateSchemaParams struct {
	structure       interface{}
	rootSchemaTitle string
	output          string
	doNotReference  bool
	extractComments bool
	commentPaths    []string
}

func generateSchema(params generateSchemaParams) error {
	r := new(jsonschema.Reflector)

	if params.doNotReference {
		r.DoNotReference = true
	}

	if params.extractComments {
		for _, path := range params.commentPaths {
			if err := r.AddGoComments("github.com/teamhanko/hanko/backend", path); err != nil {
				return err
			}
		}
	}

	s := r.Reflect(params.structure)

	if params.rootSchemaTitle != "" {
		s.Title = params.rootSchemaTitle
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	outPath, outFile := filepath.Split(params.output)

	if outPath != "" {
		_, err = os.Stat(outPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				log.Printf("directory %s does not exist, creating directory", outPath)

				mkDirErr := os.MkdirAll(outPath, 0750)

				if mkDirErr != nil {
					return fmt.Errorf("could not create directory %s: %w", outPath, mkDirErr)
				}
			} else {
				return fmt.Errorf("could not get file info: %w", err)
			}
		}
	}

	var destination string
	if outFile == "" {
		log.Println("no output file given, using default: output.schema.json")
		destination, err = filepath.Abs(filepath.Join(outPath, "output.schema.json"))
	} else {
		destination, err = filepath.Abs(filepath.Join(outPath, outFile))
	}

	if err != nil {
		return fmt.Errorf("could not determine file destination: %w", err)
	}

	err = os.WriteFile(destination, data, 0600)
	if err != nil {
		return fmt.Errorf("could not write file to destination %s: %w", destination, err)
	}

	log.Printf("schema generated successfully at: %s\n", destination)

	return nil
}
