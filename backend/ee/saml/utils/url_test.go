package utils

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamhanko/hanko/backend/ee/saml/config"
	"testing"
)

func TestUriUtils_IsAllowedRedirect(t *testing.T) {
	// given
	domain := "https://mateki.de"
	cfg := config.Saml{
		AllowedRedirectURLS: []string{
			domain,
		},
	}
	err := cfg.PostProcess()
	require.NoError(t, err)

	isAllowed := IsAllowedRedirect(cfg, domain)
	assert.True(t, isAllowed)
}

func TestUriUtils_IsAllowedRedirect_With_Slash(t *testing.T) {
	// given
	domain := "https://mateki.de"
	cfg := config.Saml{
		AllowedRedirectURLS: []string{
			domain,
		},
	}
	err := cfg.PostProcess()
	require.NoError(t, err)

	isAllowed := IsAllowedRedirect(cfg, fmt.Sprintf("%s/", domain))
	assert.True(t, isAllowed)
}

func TestUriUtils_EmptyRedirectIsNotAllowed(t *testing.T) {
	// given
	domain := "https://mateki.de"
	cfg := config.Saml{
		AllowedRedirectURLS: []string{
			domain,
		},
	}
	err := cfg.PostProcess()
	require.NoError(t, err)

	isAllowed := IsAllowedRedirect(cfg, "")
	assert.False(t, isAllowed)
}

func TestUriUtils_WrongRedirectIsNotAllowed(t *testing.T) {
	// given
	domain := "https://mateki.de"
	cfg := config.Saml{
		AllowedRedirectURLS: []string{
			domain,
		},
	}
	err := cfg.PostProcess()
	require.NoError(t, err)

	isAllowed := IsAllowedRedirect(cfg, "http://localhost")
	assert.False(t, isAllowed)
}
