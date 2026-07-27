package saml

import (
	"time"

	"github.com/fatih/structs"
	saml2 "github.com/russellhaering/gosaml2"
	samlConfig "github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/thirdparty"
)

// ExtractUserData extracts user data from SAML assertion using attribute mapping
func ExtractUserData(
	assertionInfo *saml2.AssertionInfo,
	providerConfig *ProviderConfig,
	audienceURI string,
) *thirdparty.UserData {
	firstAssertion := assertionInfo.Assertions[0]
	assertionValues := assertionInfo.Values
	attributeMap := &providerConfig.AttributeMap

	// Extract email
	emailAddress := assertionValues.Get(attributeMap.Email)

	email := thirdparty.Email{
		Email:    emailAddress,
		Verified: assertionValues.Get(attributeMap.EmailVerified) == "true",
		Primary:  true,
	}

	// Parse expiration time
	expiresIn, _ := time.Parse(time.RFC3339, firstAssertion.Conditions.NotOnOrAfter)

	// GetProvider user data
	userData := &thirdparty.UserData{}
	userData.Emails = append(userData.Emails, email)

	userData.Metadata = &thirdparty.Claims{
		Issuer:            firstAssertion.Issuer.Value,
		Subject:           firstAssertion.Subject.NameID.Value,
		Aud:               audienceURI,
		Iat:               float64(assertionInfo.AuthnInstant.Unix()),
		Exp:               float64(expiresIn.Unix()),
		Name:              assertionValues.Get(attributeMap.Name),
		FamilyName:        assertionValues.Get(attributeMap.FamilyName),
		GivenName:         assertionValues.Get(attributeMap.GivenName),
		MiddleName:        assertionValues.Get(attributeMap.MiddleName),
		NickName:          assertionValues.Get(attributeMap.NickName),
		PreferredUsername: assertionValues.Get(attributeMap.PreferredUsername),
		Profile:           assertionValues.Get(attributeMap.Profile),
		Picture:           assertionValues.Get(attributeMap.Picture),
		Website:           assertionValues.Get(attributeMap.Website),
		Gender:            assertionValues.Get(attributeMap.Gender),
		Birthdate:         assertionValues.Get(attributeMap.Birthdate),
		ZoneInfo:          assertionValues.Get(attributeMap.ZoneInfo),
		Locale:            assertionValues.Get(attributeMap.Locale),
		UpdatedAt:         assertionValues.Get(attributeMap.UpdatedAt),
		Email:             emailAddress,
		EmailVerified:     email.Verified || providerConfig.SkipEmailVerification,
		Phone:             assertionValues.Get(attributeMap.Phone),
		PhoneVerified:     assertionValues.Get(attributeMap.PhoneVerified) != "",
		CustomClaims:      mapCustomClaims(assertionInfo.Values, attributeMap),
	}

	return userData
}

// mapCustomClaims extracts custom claims that are not in the standard attribute map
func mapCustomClaims(values saml2.Values, attributeMap *samlConfig.AttributeMap) map[string]interface{} {
	customAttributes := make(map[string]interface{})

	// Get all standard attribute names
	s := structs.New(attributeMap)
	standardAttributes := make(map[string]bool)
	for _, field := range s.Fields() {
		if attrValue, ok := field.Value().(string); ok && attrValue != "" {
			standardAttributes[attrValue] = true
		}
	}

	// Extract any attributes not in the standard map
	for name := range values {
		if !standardAttributes[name] {
			customAttributes[name] = values.Get(name)
		}
	}

	return customAttributes
}
