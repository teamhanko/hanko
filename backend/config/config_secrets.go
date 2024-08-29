package config

import (
	"errors"
	"github.com/invopop/jsonschema"
)

type Secrets struct {
	// `keys` are used to en- and decrypt the JWKs which get used to sign the JWTs issued by the API.
	// For every key a JWK is generated, encrypted with the key and persisted in the database.
	//
	// You can use this list for key rotation: add a new key to the beginning of the list and the corresponding
	// JWK will then be used for signing JWTs. All tokens signed with the previous JWK(s) will still
	// be valid until they expire. Removing a key from the list does not remove the corresponding
	// database record. If you remove a key, you also have to remove the database record, otherwise
	// application startup will fail.
	Keys []string `yaml:"keys" json:"keys,omitempty" koanf:"keys" jsonschema:"minItems=1"`
}

func (Secrets) JSONSchemaExtend(schema *jsonschema.Schema) {
	keys, _ := schema.Properties.Get("keys")
	var keysItemsMinLength uint64 = 16
	keys.Items = &jsonschema.Schema{
		Type:      "string",
		Title:     "keys",
		MinLength: &keysItemsMinLength,
	}
}

func (s *Secrets) Validate() error {
	if len(s.Keys) == 0 {
		return errors.New("at least one key must be defined")
	}
	return nil
}
