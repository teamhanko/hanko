package config

type Config struct {
	Server         Server
	Webauthn       WebauthnSettings
	Passlink       Passlink
	Logging        Logging
	PrivateApiKeys ApiKeys
}

func Default() *Config {
	return &Config{
		Server: Server{
			Public: ServerSettings{
				Adress: ":8000",
			},
			Private: ServerSettings{
				Adress: ":8001",
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
		Passlink:       Passlink{
			Email:               Email{},
			Limit:               Limit{},
			AllowedRedirectUrls: nil,
			DefaultRedirectUrl:  "",
			Smtp:                SMTP{},
		},
		Logging:        Logging{
			Level:  "info",
			Format: "",
		},
		PrivateApiKeys: map[string]string{"apiKeyId": "apiKey"},
	}
}

type Server struct {
	Public       ServerSettings
	Private      ServerSettings
	ExternalHost string
}

type ServerSettings struct {
	// The Adress to listen on in the form of host:port
	// See net.Dial for details of the address format.
	Adress string
}

type WebauthnSettings struct {
	RelyingParty RelyingParty
	Timeouts     Timeouts
}

type ApiKeys map[string]string

func (c ApiKeys) Contains(secret string) bool {
	for _, v := range c {
		if v == secret {
			return true
		}
	}
	return false
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
