package config

import (
	"fmt"
	"github.com/go-webauthn/webauthn/protocol"
	webauthnLib "github.com/go-webauthn/webauthn/webauthn"
	"golang.org/x/exp/slices"
	"strings"
	"time"
)

// WebauthnSettings defines the settings for the webauthn authentication mechanism
type WebauthnSettings struct {
	RelyingParty RelyingParty `yaml:"relying_party" json:"relying_party,omitempty" koanf:"relying_party" split_words:"true" jsonschema:"title=relying_party"`
	// Deprecated, use `timeouts` instead.
	Timeout int `yaml:"timeout" json:"timeout,omitempty" koanf:"timeout" jsonschema:"default=60000"`
	// `timeouts` specifies the timeouts for passkey/WebAuthn registration and login.
	Timeouts WebauthnTimeouts `yaml:"timeouts" json:"timeouts,omitempty" koanf:"timeouts" split_words:"true" jsonschema:"title=timeouts"`
	// Deprecated, use `passkey.user_verification` instead
	UserVerification string                `yaml:"user_verification" json:"user_verification,omitempty" koanf:"user_verification" split_words:"true" jsonschema:"default=preferred,enum=required,enum=preferred,enum=discouraged"`
	Handler          *webauthnLib.WebAuthn `jsonschema:"-"`
}

// Validate does not need to validate the config, because the library does this already
func (r *WebauthnSettings) Validate() error {
	validUv := []string{"required", "preferred", "discouraged"}
	if !slices.Contains(validUv, r.UserVerification) {
		return fmt.Errorf("expected user_verification to be one of [%s], got: '%s'", strings.Join(validUv, ", "), r.UserVerification)
	}
	return nil
}

func (r *WebauthnSettings) PostProcess() error {
	requireResidentKey := false

	config := &webauthnLib.Config{
		RPID:                  r.RelyingParty.Id,
		RPDisplayName:         r.RelyingParty.DisplayName,
		RPOrigins:             r.RelyingParty.Origins,
		AttestationPreference: protocol.PreferNoAttestation,
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			RequireResidentKey: &requireResidentKey,
			ResidentKey:        protocol.ResidentKeyRequirementDiscouraged,
			UserVerification:   protocol.VerificationRequired,
		},
		Debug: false,
		Timeouts: webauthnLib.TimeoutsConfig{
			Login: webauthnLib.TimeoutConfig{
				Enforce: true,
				Timeout: time.Duration(r.Timeouts.Login) * time.Millisecond,
			},
			Registration: webauthnLib.TimeoutConfig{
				Enforce: true,
				Timeout: time.Duration(r.Timeouts.Registration) * time.Millisecond,
			},
		},
	}

	handler, err := webauthnLib.New(config)
	if err != nil {
		return err
	}

	r.Handler = handler

	return nil
}

// RelyingParty webauthn settings for your application using hanko.
type RelyingParty struct {
	// `display_name` is the service's name that some WebAuthn Authenticators will display to the user during registration
	// and authentication ceremonies.
	DisplayName string `yaml:"display_name" json:"display_name,omitempty" koanf:"display_name" split_words:"true" jsonschema:"default=Hanko Authentication Service"`
	Icon        string `yaml:"icon" json:"icon,omitempty" koanf:"icon" jsonschema:"-"`
	// `id` is the [effective domain](https://html.spec.whatwg.org/multipage/browsers.html#concept-origin-effective-domain)
	// the passkey/WebAuthn credentials will be bound to.
	Id string `yaml:"id" json:"id,omitempty" koanf:"id" jsonschema:"default=localhost,examples=localhost,example.com,subdomain.example.com"`
	// `origins` is a list of origins for which passkeys/WebAuthn credentials will be accepted by the server. Must
	// include the protocol and can only be the effective domain, or a registrable domain suffix of the effective
	// domain, as specified in the [`id`](#id). Except for `localhost`, the protocol **must** always be `https` for
	// passkeys/WebAuthn to work. IP Addresses will not work.
	//
	// For an Android application the origin must be the base64 url encoded SHA256 fingerprint of the signing
	// certificate.
	Origins []string `yaml:"origins" json:"origins,omitempty" koanf:"origins" jsonschema:"minItems=1,default=http://localhost:8888,examples=android:apk-key-hash:nLSu7wVTbnMOxLgC52f2faTnvCbXQrUn_wF9aCrr-l0,https://login.example.com"`
}

type WebauthnTimeouts struct {
	// `registration` determines the time, in milliseconds, that the client is willing to wait for the credential
	// creation request to the WebAuthn API to complete.
	Registration int `yaml:"registration" json:"registration,omitempty" koanf:"registration" jsonschema:"default=600000"`
	// `login` determines the time, in milliseconds, that the client is willing to wait for the credential
	//  request to the WebAuthn API to complete.
	Login int `yaml:"login" json:"login,omitempty" koanf:"login" jsonschema:"default=600000"`
}
