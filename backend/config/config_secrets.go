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
	// Type specifies the key management system to use. Supported values are 'local' (default) for local key storage
	// or 'aws_kms' for AWS Key Management Service.
	// When using 'aws_kms,' the AWS credentials must be set using the standard AWS credential chain (in order of precedence).
	// 1. Environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_SESSION_TOKEN)
	// 2. Shared credentials file (~/.aws/credentials)
	// 3. Shared config file (~/.aws/config)
	// 4. IAM role for Amazon EC2 (via instance metadata service)
	// 5. IAM role for Amazon ECS (via container credentials)
	// 6. IAM role for Amazon EKS (via service account token)
	Type KeyManagementStoreType `yaml:"type" json:"type,omitempty" koanf:"type"`
	// KeyID is the AWS KMS key identifier (ARN or alias) used for signing operations.
	// Required when Type is 'aws_kms'.
	KeyID string `yaml:"key_id" json:"key_id,omitempty" koanf:"key_id"`
	// Region is the AWS region where the KMS key is located.
	// Required when Type is 'aws_kms'.
	Region string `yaml:"region" json:"region,omitempty" koanf:"region"`
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
