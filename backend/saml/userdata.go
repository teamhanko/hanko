package saml

import (
	"time"

	saml2 "github.com/russellhaering/gosaml2"
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
	}

	userData.CustomClaimSource = buildCustomClaimSource(attributeMap.Custom, assertionValues)

	return userData
}

// buildCustomClaimSource resolves each AttributeMap.Custom value - the IdP's literal SAML
// attribute Name, never a path - against the assertion. Uses GetAll rather than Get since a
// SAML attribute can carry multiple values (e.g. eduPersonAffiliation); Get would silently
// keep only the first.
func buildCustomClaimSource(mapping map[string]string, values saml2.Values) *thirdparty.CustomClaimSource {
	if len(mapping) == 0 {
		return nil
	}

	attributes := make(map[string]any, len(mapping))
	for _, attributeName := range mapping {
		if all := values.GetAll(attributeName); len(all) > 0 {
			attributes[attributeName] = all
		}
	}

	return &thirdparty.CustomClaimSource{
		Mapping:    mapping,
		Attributes: attributes,
	}
}
