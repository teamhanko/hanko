package config

import (
	"fmt"
	"github.com/spf13/viper"
	"path/filepath"
	"runtime"
	"strings"
)

type Config struct {
	Server   Server
	Webauthn WebauthnSettings
	Passlink Passlink
	Logging  Logging
	Database Database
}

// Load loads config from given file or default places
func Load(cfgFile *string) *Config {
	if cfgFile != nil && *cfgFile != "" {
		// Use given config file
		viper.SetConfigFile(*cfgFile)
	} else {
		// Use config file from default places
		// Get base path of binary call
		_, b, _, _ := runtime.Caller(0)
		basePath := filepath.Dir(b)

		viper.SetConfigType("yaml")
		viper.AddConfigPath(basePath)
		viper.AddConfigPath("/etc/config")
		viper.AddConfigPath("/etc/secrets")
		viper.AddConfigPath("./config")
		viper.SetConfigName("hanko-config")
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
	c := defaultConfig()
	err := viper.Unmarshal(c)
	if err != nil {
		panic(fmt.Sprintf("unable to decode config into struct, %v", err))
	}

	return c
}

func defaultConfig() *Config {
	return &Config{
		Server: Server{
			Public: ServerSettings{
				Address: ":8000",
			},
			Private: ServerSettings{
				Address: ":8001",
			},
			ExternalHost: "",
		},
		Webauthn: WebauthnSettings{
			RelyingParty: RelyingParty{
				Id:          "localhost",
				DisplayName: "Hanko GmbH",
				Icon:        "https://hanko.io/logo.png",
				Origins:     []string{"http://localhost:3000"},
			},
			Timeouts: Timeouts{
				Authentication: 60000,
				Registration:   60000,
			},
		},
		Passlink: Passlink{
			Email:               Email{},
			Limit:               Limit{},
			AllowedRedirectUrls: nil,
			DefaultRedirectUrl:  "",
			Smtp:                SMTP{},
		},
		Logging: Logging{
			Level:  "info",
			Format: "",
		},
		Database: Database{
			Database: "hanko",
			User:     "postgres",
			Password: "postgres",
			Host:     "localhost",
			Port:     "5432",
			Dialect:  "postgres",
		},
	}
}

type Server struct {
	Public       ServerSettings
	Private      ServerSettings
	ExternalHost string
}

type ServerSettings struct {
	// The Address to listen on in the form of host:port
	// See net.Dial for details of the address format.
	Address string
}

type WebauthnSettings struct {
	RelyingParty RelyingParty
	Timeouts     Timeouts
}

type RelyingParty struct {
	Id          string
	DisplayName string
	Icon        string
	Origins     []string
}

type Timeouts struct {
	Authentication int
	Registration   int
}

type SMTP struct {
	Host     string
	Port     string
	User     string
	Password string
}

type Email struct {
	Interval            string
	From                string
	Customization       *Customization
	CustomTemplatesPath string
}

type Limit struct {
	Tokens        uint64
	Interval      string
	SweepInterval string
	SweepMinTTL   string
}

type Customization struct {
	BrandColor   *string
	BorderRadius *int
}

type Passlink struct {
	Email               Email
	Limit               Limit
	AllowedRedirectUrls []string
	DefaultRedirectUrl  string
	Smtp                SMTP
}

type Logging struct {
	Level  string
	Format string
}

// Database connection settings
type Database struct {
	Database string `json:"database"`
	User     string `json:"user"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Dialect  string `json:"dialect"`
}
