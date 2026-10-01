package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/invopop/jsonschema"
)

const DefaultTenantID = "00000000-0000-0000-0000-000000000001"

type RedisConfig struct {
	// `address` is the address of the redis instance in the form of `host[:port][/database]`.
	Address string `yaml:"address" json:"address" koanf:"address"`
	// `password` is the password for the redis instance.
	Password string `yaml:"password" json:"password,omitempty" koanf:"password"`
}

// DialAddress splits Address into the `host[:port]` that can be passed to
// redis.Dial and the optional database index from the `/database` suffix.
//
// redis.Dial does not understand the suffix (only redis.DialURL does), so
// callers must apply the returned index via redis.DialDatabase. Without this
// an address such as `localhost:6379/9` is not silently ignored: it is passed
// through as a host and the dial fails with `unknown port`.
func (t RedisConfig) DialAddress() (address string, database int, err error) {
	address = t.Address

	slash := strings.LastIndex(address, "/")
	if slash < 0 {
		return address, 0, nil
	}

	suffix := address[slash+1:]
	address = address[:slash]

	// A trailing slash with no index means "unspecified", i.e. database 0.
	if suffix == "" {
		return address, 0, nil
	}

	database, err = strconv.Atoi(suffix)
	if err != nil || database < 0 {
		return "", 0, fmt.Errorf("invalid redis database in address %q: must be a non-negative integer", t.Address)
	}

	return address, database, nil
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
