package config

import "github.com/invopop/jsonschema"

const DefaultTenantID = "00000000-0000-0000-0000-000000000001"

type RedisConfig struct {
	// `address` is the address of the redis instance in the form of `host[:port][/database]`.
	Address string `yaml:"address" json:"address" koanf:"address"`
	// `password` is the password for the redis instance.
	Password string `yaml:"password" json:"password,omitempty" koanf:"password"`
}

func (t RedisConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	password, _ := schema.Properties.Get("password")
	schema.Properties.Set("password", &jsonschema.Schema{
		Description: password.Description,
		AnyOf: []*jsonschema.Schema{
			{Type: "string"},
			{Type: "null"},
		},
	})
}
