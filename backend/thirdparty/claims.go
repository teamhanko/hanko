package thirdparty

import (
	"errors"

	zeroLogger "github.com/rs/zerolog/log"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
	"github.com/teamhanko/hanko/backend/v2/utils"
)

type ClaimsAddress struct {
	Formatted  string `json:"formatted,omitempty" mapstructure:"formatted,omitempty"`
	Locality   string `json:"locality,omitempty" mapstructure:"locality,omitempty"`
	PostalCode string `json:"postal_code,omitempty" mapstructure:"postal_code,omitempty"`
	Region     string `json:"region,omitempty" mapstructure:"region,omitempty"`
	Street     string `json:"street_address,omitempty" mapstructure:"street_address,omitempty"`
}

type Claims struct {
	// Reserved claims
	Issuer  string  `json:"iss,omitempty" mapstructure:"iss,omitempty"`
	Subject string  `json:"sub,omitempty" mapstructure:"sub,omitempty"`
	Aud     string  `json:"aud,omitempty" mapstructure:"aud,omitempty"`
	Iat     float64 `json:"iat,omitempty" mapstructure:"iat,omitempty"`
	Exp     float64 `json:"exp,omitempty" mapstructure:"exp,omitempty"`

	// Default profile claims
	Address           *ClaimsAddress `json:"address,omitempty" mapstructure:"address,omitempty"`
	Birthdate         string         `json:"birthdate,omitempty" mapstructure:"birthdate,omitempty"`
	Email             string         `json:"email,omitempty" mapstructure:"email,omitempty"`
	EmailVerified     bool           `json:"email_verified,omitempty" mapstructure:"email_verified,omitempty"`
	FamilyName        string         `json:"family_name,omitempty" mapstructure:"family_name,omitempty"`
	Gender            string         `json:"gender,omitempty" mapstructure:"gender,omitempty"`
	GivenName         string         `json:"given_name,omitempty" mapstructure:"given_name,omitempty"`
	Locale            string         `json:"locale,omitempty" mapstructure:"locale,omitempty"`
	MiddleName        string         `json:"middle_name,omitempty" mapstructure:"middle_name,omitempty"`
	Name              string         `json:"name,omitempty" mapstructure:"name,omitempty"`
	NickName          string         `json:"nickname,omitempty" mapstructure:"nickname,omitempty"`
	Phone             string         `json:"phone,omitempty" mapstructure:"phone,omitempty"`
	PhoneVerified     bool           `json:"phone_verified,omitempty" mapstructure:"phone_verified,omitempty"`
	Picture           string         `json:"picture,omitempty" mapstructure:"picture,omitempty"`
	PreferredUsername string         `json:"preferred_username,omitempty" mapstructure:"preferred_username,omitempty"`
	Profile           string         `json:"profile,omitempty" mapstructure:"profile,omitempty"`
	UpdatedAt         string         `json:"updated_at,omitempty" mapstructure:"updated_at,omitempty"`
	Website           string         `json:"website,omitempty" mapstructure:"website,omitempty"`
	ZoneInfo          string         `json:"zoneinfo,omitempty" mapstructure:"zoneinfo,omitempty"`

	// Custom profile claims that are oidc specific
	CustomClaims map[string]interface{} `json:"custom_claims,omitempty" mapstructure:"custom_claims,remain,omitempty"`
}

type claimWarning struct {
	Field  string
	Reason string
	Value  string
}

func (c *Claims) ProviderProfile() (models.ProviderProfile, []claimWarning) {
	if c == nil {
		return models.ProviderProfile{}, nil
	}

	profile := models.ProviderProfile{
		Name:       c.Name,
		GivenName:  c.GivenName,
		FamilyName: c.FamilyName,
		Picture:    c.Picture,
	}

	var warnings []claimWarning

	if profile.Picture != "" {
		if err := utils.ValidatePictureURL(profile.Picture); err != nil {
			reason := "invalid"
			var perr utils.PictureURLError
			if errors.As(err, &perr) {
				reason = perr.Reason
			}

			warnings = append(warnings, claimWarning{
				Field:  "picture",
				Reason: reason,
				Value:  profile.Picture,
			})
			profile.Picture = ""
		}
	}

	return profile, warnings
}

func (c *Claims) ProviderProfileWithLogging(operation string, providerID string) models.ProviderProfile {
	profile, warnings := c.ProviderProfile()
	logInvalidClaimWarnings(operation, providerID, warnings)
	return profile
}

func logInvalidClaimWarnings(operation string, providerID string, warnings []claimWarning) {
	if len(warnings) == 0 {
		return
	}

	l := zeroLogger.With().
		Str("component", "thirdparty").
		Str("operation", operation).
		Str("provider_id", providerID).
		Logger()

	for _, w := range warnings {
		l.Warn().
			Str("claim", w.Field).
			Str("validation_reason", w.Reason).
			Msg("ignored invalid claim while syncing provider profile")
	}
}
