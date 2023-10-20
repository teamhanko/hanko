package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/invopop/jsonschema"
	"github.com/teamhanko/hanko/backend/cmd/user"
	"github.com/teamhanko/hanko/backend/config"
)

func main() {
	if err := generateSchema("./config", "./json_schema/hanko.config.json", &config.Config{}); err != nil {
		panic(err)
	}
	if err := generateSchema("./cmd/user", "./json_schema/hanko.user_import.json", &user.ImportOrExportList{}); err != nil {
		panic(err)
	}

}

func generateSchema(codePath, filePath string, structure interface{}) error {
	r := new(jsonschema.Reflector)

	if err := r.AddGoComments("github.com/teamhanko/hanko/backend", codePath); err != nil {
		panic(err.Error())
	}

	s := r.Reflect(structure)

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		panic(err.Error())
	}

	err = os.WriteFile(filePath, data, 0600)
	if err != nil {
		fmt.Println(err)
	}
	return nil
}
