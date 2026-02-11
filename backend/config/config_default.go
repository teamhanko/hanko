package config

import "time"

func DefaultConfig() *Config {
	return &Config{
		ConvertLegacyConfig:                  false,
		ConvertLegacyServerSideSessionConfig: true,
		Service: Service{
			Name: "Hanko Authentication Service",
		},
		Secrets: Secrets{
			Keys: []string{"abcedfghijklmnopqrstuvwxyz"},
			KeyManagement: KeyManagement{
				Type: "local",
			},
		},
		Server: Server{
			Public: ServerSettings{
				Address: ":8000",
				Cors: Cors{
					AllowOrigins:                []string{"http://localhost:8888"},
					UnsafeWildcardOriginAllowed: false,
				},
			},
			Admin: ServerSettings{
				Address: ":8001",
			},
		},
		Webauthn: WebauthnSettings{
			RelyingParty: RelyingParty{
				Id:          "localhost",
				DisplayName: "Hanko Authentication Service",
				Origins:     []string{"http://localhost:8888"},
			},
			UserVerification: "preferred",
			Timeout:          600000,
			Timeouts: WebauthnTimeouts{
				Registration: 600000,
				Login:        600000,
			},
		},
		SecurityNotifications: SecurityNotifications{
			Notifications: SecurityNotificationTypes{
				EmailCreate: SecurityNotificationConfiguration{
					Enabled: true,
				},
				EmailDelete: SecurityNotificationConfiguration{
					Enabled: true,
				},
				PasswordUpdate: SecurityNotificationConfiguration{
					Enabled: true,
				},
				PasskeyCreate: SecurityNotificationConfiguration{
					Enabled: true,
				},
				PrimaryEmailUpdate: SecurityNotificationConfiguration{
					Enabled: true,
				},
				MFACreate: SecurityNotificationConfiguration{
					Enabled: true,
				},
				MFADelete: SecurityNotificationConfiguration{
					Enabled: true,
				},
			},
		},
		Smtp: SMTP{
			Host: "localhost",
			Port: "465",
		},
		EmailDelivery: EmailDelivery{
			Enabled: true,
			SMTP: SMTP{
				Host: "localhost",
				Port: "465",
			},
			FromAddress: "noreply@hanko.io",
			FromName:    "Hanko",
		},
		Passcode: Passcode{
			TTL: 300,
		},
		Password: Password{
			Enabled:               true,
			Optional:              false,
			AcquireOnRegistration: "always",
			AcquireOnLogin:        "never",
			Recovery:              true,
			MinLength:             8,
		},
		Database: Database{
			Database: "hanko",
			User:     "hanko",
			Password: "hanko",
			Port:     "5432",
			Dialect:  "postgres",
			Host:     "localhost",
		},
		Session: Session{
			AllowRevocation:  true,
			AcquireIPAddress: true,
			AcquireUserAgent: true,
			Lifespan:         "12h",
			Cookie: Cookie{
				HttpOnly:  true,
				Retention: "persistent",
				SameSite:  "strict",
				Secure:    true,
			},
			Limit:         5,
			ShowOnProfile: true,
		},
		AuditLog: AuditLog{
			ConsoleOutput: AuditLogConsole{
				Enabled:      true,
				OutputStream: OutputStreamStdOut,
			},
			Mask:      true,
			Retention: "720h",
		},
		Emails: Emails{
			RequireVerification: true,
			MaxNumOfAddresses:   5,
		},
		RateLimiter: RateLimiter{
			Enabled: true,
			Store:   RATE_LIMITER_STORE_IN_MEMORY,
			OTPLimits: RateLimits{
				Tokens:   3,
				Interval: 1 * time.Minute,
			},
			PasswordLimits: RateLimits{
				Tokens:   5,
				Interval: 1 * time.Minute,
			},
			PasscodeLimits: RateLimits{
				Tokens:   3,
				Interval: 1 * time.Minute,
			},
			TokenLimits: RateLimits{
				Tokens:   3,
				Interval: 1 * time.Minute,
			},
		},
		Account: Account{
			AllowDeletion: false,
			AllowSignup:   true,
		},
		ThirdParty: ThirdParty{
			Providers: ThirdPartyProviders{
				Apple: ThirdPartyProvider{
					DisplayName:  "Apple",
					AllowLinking: true,
					ID:           "apple",
				},
				Discord: ThirdPartyProvider{
					DisplayName:  "Discord",
					AllowLinking: true,
					ID:           "discord",
				},
				LinkedIn: ThirdPartyProvider{
					DisplayName:  "LinkedIn",
					AllowLinking: true,
					ID:           "linkedin",
				},
				Microsoft: ThirdPartyProvider{
					DisplayName:  "Microsoft",
					AllowLinking: true,
					ID:           "microsoft",
				},
				GitHub: ThirdPartyProvider{
					DisplayName:  "GitHub",
					AllowLinking: true,
					ID:           "github",
				},
				Google: ThirdPartyProvider{
					DisplayName:  "Google",
					AllowLinking: true,
					ID:           "google",
				},
				Facebook: ThirdPartyProvider{
					DisplayName:  "Facebook",
					AllowLinking: true,
					ID:           "facebook",
				},
			},
		},
		Passkey: Passkey{
			Enabled:               true,
			Optional:              true,
			AcquireOnRegistration: "always",
			AcquireOnLogin:        "always",
			UserVerification:      "preferred",
			AttestationPreference: "direct",
			Limit:                 10,
		},
		Email: Email{
			Enabled:               true,
			Optional:              false,
			AcquireOnRegistration: true,
			AcquireOnLogin:        true,
			RequireVerification:   true,
			Limit:                 5,
			UseAsLoginIdentifier:  true,
			MaxLength:             120,
			UseForAuthentication:  true,
			PasscodeTtl:           300,
			PasscodeCharset:       PasscodeCharsetNumeric,
		},
		Username: Username{
			Enabled:               false,
			Optional:              true,
			AcquireOnRegistration: true,
			AcquireOnLogin:        true,
			UseAsLoginIdentifier:  true,
			MinLength:             3,
			MaxLength:             32,
		},
		MFA: MFA{
			AcquireOnLogin:               false,
			AcquireOnRegistration:        true,
			DeviceTrustCookieName:        "hanko-device-token",
			DeviceTrustDuration:          30 * 24 * time.Hour, // 30 days
			DeviceTrustMaxUsersPerDevice: 20,
			DeviceTrustPolicy:            "prompt",
			Enabled:                      true,
			Optional:                     true,
			SecurityKeys: SecurityKeys{
				AttestationPreference:   "direct",
				AuthenticatorAttachment: "cross-platform",
				Enabled:                 true,
				Limit:                   10,
				UserVerification:        "discouraged",
			},
			TOTP: TOTP{
				Enabled: true,
			},
		},
		Privacy: Privacy{
			ShowAccountExistenceHints:  false,
			OnlyShowActualLoginMethods: false,
		},
		Debug: false,
		FlowLocker: FlowLocker{
			Enabled: true,
			Store:   FLOW_LOCKER_STORE_IN_MEMORY,
		},
		MultiTenant: DefaultMultiTenantConfig(),
	}
}
