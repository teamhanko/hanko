package config

import (
	"errors"
	"strings"
	"time"
)

type Database struct {
	// `conn_max_lifetime` sets the maximum amount of time a connection may be reused.
	// It must be a (possibly signed) sequence of decimal numbers, each with optional fraction and a unit suffix,
	// such as "300ms", or "2h45m".
	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" json:"conn_max_lifetime,omitempty" koanf:"conn_max_lifetime" jsonschema:"default=1h"`
	// `conn_max_idletime` sets the maximum number of connections in the idle connection pool.
	// It must be a (possibly signed) sequence of decimal numbers, each with optional fraction and a unit suffix,
	// such as "300ms", or "2h45m".
	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idletime" json:"conn_max_idletime,omitempty" koanf:"conn_max_idletime" jsonschema:"default=5m"`
	// `database` determines the name of the database schema to use.
	Database string `yaml:"database" json:"database,omitempty" koanf:"database" jsonschema:"default=hanko"`
	// `dialect` is the name of the database system to use.
	Dialect string `yaml:"dialect" json:"dialect,omitempty" koanf:"dialect" jsonschema:"default=postgres,enum=postgres,enum=mysql,enum=mariadb,enum=cockroach"`
	// `host` is the host the database system is running on.
	Host string `yaml:"host" json:"host,omitempty" koanf:"host" jsonschema:"default=localhost"`
	// `idle_pool` sets the maximum number of connections in the idle connection pool.
	IdlePool int `yaml:"idle_pool" json:"idle_pool,omitempty" koanf:"idle_pool" jsonschema:"default=0"`
	// `password` is the password for the database user to use for connecting to the database.
	Password string `yaml:"password" json:"password,omitempty" koanf:"password" jsonschema:"default=hanko"`
	// `pool` sets the maximum number of open connections to the database.
	Pool int `yaml:"pool" json:"pool,omitempty" koanf:"pool" jsonschema:"default=5"`
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
