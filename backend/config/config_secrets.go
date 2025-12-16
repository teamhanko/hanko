package config

import (
	"errors"
	"fmt"

	"github.com/invopop/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type Secrets struct {
	// KeyManagement configures the key management system used for signing JWTs.
	// Supports 'local' (default) which uses the keys defined in the 'keys' field to encrypt a newly generated private
	// RSA key in the database, or 'aws_kms' which uses AWS Key Management Service for key signatures.
	KeyManagement KeyManagement `yaml:"key_management" json:"key_management,omitempty" koanf:"key_management"`
	// `keys` are used to en- and decrypt the JWKs which get used to sign the JWTs issued by the API.
	// For every key a JWK is generated, encrypted with the key and persisted in the database.
	//
	// You can use this list for key rotation: add a new key to the beginning of the list and the corresponding
	// JWK will then be used for signing JWTs. All tokens signed with the previous JWK(s) will still
	// be valid until they expire. Removing a key from the list does not remove the corresponding
	// database record. If you remove a key, you also have to remove the database record, otherwise
	// application startup will fail.
	Keys []string `yaml:"keys" json:"keys,omitempty" koanf:"keys"`
}

func (Secrets) JSONSchemaExtend(schema *jsonschema.Schema) {
	keys, _ := schema.Properties.Get("keys")
	var keysItemsMinLength uint64 = 16
	keys.Items = &jsonschema.Schema{
		Type:      "string",
		Title:     "keys",
		MinLength: &keysItemsMinLength,
	}

	// Require at least one key when key_management.type is "local"
	schema.If = &jsonschema.Schema{
		Properties: func() *orderedmap.OrderedMap[string, *jsonschema.Schema] {
			props := orderedmap.New[string, *jsonschema.Schema]()
			props.Set("key_management", &jsonschema.Schema{
				Properties: func() *orderedmap.OrderedMap[string, *jsonschema.Schema] {
					kmProps := orderedmap.New[string, *jsonschema.Schema]()
					kmProps.Set("type", &jsonschema.Schema{
						Const: "local",
					})
					return kmProps
				}(),
			})
			return props
		}(),
	}
	var minItems uint64 = 1
	schema.Then = &jsonschema.Schema{
		Properties: func() *orderedmap.OrderedMap[string, *jsonschema.Schema] {
			props := orderedmap.New[string, *jsonschema.Schema]()
			props.Set("keys", &jsonschema.Schema{
				MinItems: &minItems,
			})
			return props
		}(),
	}

}

func (s *Secrets) Validate() error {
	if len(s.Keys) == 0 && s.KeyManagement.Type == "local" {
		return errors.New("at least one key must be defined")
	}
	return s.KeyManagement.Validate()
}

type KeyManagement struct {
	Type   KeyManagementStoreType `yaml:"type" json:"type,omitempty" koanf:"type"`
	KeyID  string                 `yaml:"key_id" json:"key_id,omitempty" koanf:"key_id"`
	Region string                 `yaml:"region" json:"region,omitempty" koanf:"region"`
}

func (KeyManagement) JSONSchemaExtend(schema *jsonschema.Schema) {
	typeProperty, _ := schema.Properties.Get("type")
	typeProperty.Enum = []interface{}{KEY_MANAGEMENT_STORE_LOCAL, KEY_MANAGEMENT_STORE_AWS_KMS}

	schema.If = &jsonschema.Schema{
		Properties: func() *orderedmap.OrderedMap[string, *jsonschema.Schema] {
			props := orderedmap.New[string, *jsonschema.Schema]()
			props.Set("type", &jsonschema.Schema{
				Const: "aws_kms",
			})
			return props
		}(),
	}
	schema.Then = &jsonschema.Schema{
		Required: []string{"key_id", "region"},
	}
}

func (k *KeyManagement) Validate() error {
	if k.Type == KEY_MANAGEMENT_STORE_AWS_KMS {
		if k.KeyID == "" {
			return errors.New("key_id is required when key_management.type is aws_kms")
		}
		if k.Region == "" {
			return errors.New("region is required when key_management.type is aws_kms")
		}
	}
	if k.Type != KEY_MANAGEMENT_STORE_LOCAL && k.Type != KEY_MANAGEMENT_STORE_AWS_KMS {
		return fmt.Errorf("unsupported key_management.type: %s", k.Type)
	}
	return nil
}

type KeyManagementStoreType string

var (
	KEY_MANAGEMENT_STORE_AWS_KMS KeyManagementStoreType = "aws_kms"
	KEY_MANAGEMENT_STORE_LOCAL   KeyManagementStoreType = "local"
)
