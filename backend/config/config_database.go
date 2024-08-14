package config

import (
	"errors"
	"strings"
)

type Database struct {
	// `database` determines the name of the database schema to use.
	Database string `yaml:"database" json:"database,omitempty" koanf:"database" jsonschema:"default=hanko"`
	// `dialect` is the name of the database system to use.
	Dialect string `yaml:"dialect" json:"dialect,omitempty" koanf:"dialect" jsonschema:"default=postgres,enum=postgres,enum=mysql,enum=mariadb,enum=cockroach"`
	// `host` is the host the database system is running on.
	Host string `yaml:"host" json:"host,omitempty" koanf:"host" jsonschema:"default=localhost"`
	// `password` is the password for the database user to use for connecting to the database.
	Password string `yaml:"password" json:"password,omitempty" koanf:"password" jsonschema:"default=hanko"`
	// `port` is the port the database system is running on.
	Port string `yaml:"port" json:"port,omitempty" koanf:"port" jsonschema:"default=5432"`
	// `url` is a datasource connection string. It can be used instead of the rest of the database configuration
	// options. If this `url` is set then it is prioritized, i.e. the rest of the options, if set, have no effect.
	//
	// Schema: `dialect://username:password@host:port/database`
	Url string `yaml:"url" json:"url,omitempty" koanf:"url" jsonschema:"example=postgres://hanko:hanko@localhost:5432/hanko"`
	// `user` is the database user to use for connecting to the database.
	User string `yaml:"user" json:"user,omitempty" koanf:"user" jsonschema:"default=hanko"`
}

func (d *Database) Validate() error {
	if len(strings.TrimSpace(d.Url)) > 0 {
		return nil
	}
	if len(strings.TrimSpace(d.Database)) == 0 {
		return errors.New("database must not be empty")
	}
	if len(strings.TrimSpace(d.User)) == 0 {
		return errors.New("user must not be empty")
	}
	if len(strings.TrimSpace(d.Host)) == 0 {
		return errors.New("host must not be empty")
	}
	if len(strings.TrimSpace(d.Port)) == 0 {
		return errors.New("port must not be empty")
	}
	if len(strings.TrimSpace(d.Dialect)) == 0 {
		return errors.New("dialect must not be empty")
	}
	return nil
}
