package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIPConfig_Validate_TrustedProxiesWithDirectExtractor(t *testing.T) {
	cfg := IPConfig{Extractor: IPExtractorDirect, TrustedProxies: []string{"35.191.0.0/16"}}
	assert.NoError(t, cfg.Validate())

	cfg.TrustedProxies = []string{"35.191.0.0"}
	assert.Error(t, cfg.Validate())
}
