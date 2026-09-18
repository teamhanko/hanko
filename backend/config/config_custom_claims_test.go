package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCustomClaimDefinitions_ValidateMapping_Success(t *testing.T) {
	defs := customClaimDefsForTest()

	err := defs.ValidateMapping(map[string]string{"matriculation_number": "urn:oid:mat_nr"})

	assert.NoError(t, err)
}

func TestCustomClaimDefinitions_ValidateMapping_UndeclaredClaim(t *testing.T) {
	defs := customClaimDefsForTest()

	err := defs.ValidateMapping(map[string]string{"not_a_real_claim": "urn:oid:mat_nr"})

	assert.Error(t, err)
}

func customClaimDefsForTest() CustomClaimDefinitions {
	return CustomClaimDefinitions{
		"matriculation_number": {Name: "matriculation_number", Type: CustomClaimTypeString},
	}
}

func TestConfig_ValidateCrossConfig_RejectsUndeclaredSamlAttributeMapCustom(t *testing.T) {
	cfg := Config{
		ApplicationConfig: ApplicationConfig{SecretKeys: []string{"abcdefghijklmnop"}},
		TenantConfig: TenantConfig{
			CustomClaims: CustomClaims{Definitions: customClaimDefsForTest()},
			Saml: Saml{
				IdentityProviders: []IdentityProvider{
					{
						Name: "uni-a",
						AttributeMap: AttributeMap{
							Custom: map[string]string{"not_a_real_claim": "urn:oid:mat_nr"},
						},
					},
				},
			},
		},
	}

	err := cfg.ValidateCrossConfig()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "uni-a")
}

func TestConfig_ValidateCrossConfig_RejectsUndeclaredCustomProviderClaimMapping(t *testing.T) {
	cfg := Config{
		ApplicationConfig: ApplicationConfig{SecretKeys: []string{"abcdefghijklmnop"}},
		TenantConfig: TenantConfig{
			CustomClaims: CustomClaims{Definitions: customClaimDefsForTest()},
			ThirdParty: ThirdParty{
				CustomProviders: CustomThirdPartyProviders{
					"myprovider": CustomThirdPartyProvider{
						CustomClaimMapping: map[string]string{"not_a_real_claim": "some_claim"},
					},
				},
			},
		},
	}

	err := cfg.ValidateCrossConfig()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "myprovider")
}

func TestConfig_ValidateCrossConfig_AllowsDeclaredMappings(t *testing.T) {
	cfg := Config{
		ApplicationConfig: ApplicationConfig{SecretKeys: []string{"abcdefghijklmnop"}},
		TenantConfig: TenantConfig{
			CustomClaims: CustomClaims{Definitions: customClaimDefsForTest()},
			Saml: Saml{
				IdentityProviders: []IdentityProvider{
					{
						Name: "uni-a",
						AttributeMap: AttributeMap{
							Custom: map[string]string{"matriculation_number": "urn:oid:mat_nr"},
						},
					},
				},
			},
			ThirdParty: ThirdParty{
				CustomProviders: CustomThirdPartyProviders{
					"myprovider": CustomThirdPartyProvider{
						CustomClaimMapping: map[string]string{"matriculation_number": "some_claim"},
					},
				},
			},
		},
	}

	err := cfg.ValidateCrossConfig()

	assert.NoError(t, err)
}

func TestCustomClaims_PostProcess_CopiesKeyVerbatimIntoName(t *testing.T) {
	claims := CustomClaims{
		Definitions: CustomClaimDefinitions{
			"Matriculation_Number": CustomClaimDefinition{Type: CustomClaimTypeString},
		},
	}

	err := claims.PostProcess()

	assert.NoError(t, err)
	definition, ok := claims.Definitions["Matriculation_Number"]
	assert.True(t, ok, "key must not be case-normalized")
	assert.Equal(t, "Matriculation_Number", definition.Name)
}

func TestCustomClaims_Validate_Success(t *testing.T) {
	claims := CustomClaims{
		Definitions: CustomClaimDefinitions{
			"matriculation_number": {Name: "matriculation_number", Type: CustomClaimTypeString},
			"is_staff":             {Name: "is_staff", Type: CustomClaimTypeBoolean},
			"affiliation":          {Name: "affiliation", Type: CustomClaimTypeStringList},
			"age":                  {Name: "age", Type: CustomClaimTypeNumber},
		},
	}

	assert.NoError(t, claims.Validate())
}

func TestCustomClaims_Validate_NilDefinitions(t *testing.T) {
	claims := CustomClaims{}

	assert.NoError(t, claims.Validate())
}

func TestCustomClaimDefinition_Validate_MixedCaseNameAllowed(t *testing.T) {
	definition := CustomClaimDefinition{Name: "mYcLaIm", Type: CustomClaimTypeString}

	assert.NoError(t, definition.Validate(), "case is the admin's choice, not restricted")
}

func TestCustomClaimDefinition_Validate_BadName(t *testing.T) {
	cases := []string{
		"1_number", // must start with a letter
		"has-dash", // dash not allowed
		"a.b",      // gjson path separator - would break UserJWT.CustomClaims lookups
		"a|b",      // gjson union operator
		"",         // empty
		strings.Repeat("a", 65),
	}

	for _, name := range cases {
		definition := CustomClaimDefinition{Name: name, Type: CustomClaimTypeString}
		err := definition.Validate()
		assert.Error(t, err, "expected %q to be rejected", name)
	}
}

func TestCustomClaimDefinition_Validate_ReservedName(t *testing.T) {
	for name := range reservedCustomClaimNames {
		definition := CustomClaimDefinition{Name: name, Type: CustomClaimTypeString}
		err := definition.Validate()
		assert.Error(t, err, "expected reserved name %q to be rejected", name)
	}
}

func TestCustomClaimDefinition_Validate_ReservedName_CaseInsensitive(t *testing.T) {
	definition := CustomClaimDefinition{Name: "Email", Type: CustomClaimTypeString}

	err := definition.Validate()

	assert.Error(t, err, "reserved-name check must not be bypassable by casing")
}

func TestCustomClaimDefinition_Validate_UnknownType(t *testing.T) {
	definition := CustomClaimDefinition{Name: "role", Type: "object"}

	err := definition.Validate()

	assert.Error(t, err)
}

func TestCustomClaims_Validate_TooManyDefinitions(t *testing.T) {
	definitions := make(CustomClaimDefinitions, maxCustomClaimDefinitions+1)
	for i := 0; i < maxCustomClaimDefinitions+1; i++ {
		name := "claim_" + string(rune('a'+i))
		definitions[name] = CustomClaimDefinition{Name: name, Type: CustomClaimTypeString}
	}
	claims := CustomClaims{Definitions: definitions}

	err := claims.Validate()

	assert.Error(t, err)
}
