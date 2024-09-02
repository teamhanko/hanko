package config

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

type Passlink struct {
	// `enabled` determines whether users can authenticate via a link containing a short-living token send by mail.
	Enabled bool `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=false"`
	// `url` is the redirect target URL for passlinks to your frontend.
	// Frontend must be able to handle the passlink token and call the passlink finalize endpoint to complete the authentication.
	// The passlink id (plid) and the token (pltk) are added as query parameters to that URL.
	URL string `yaml:"url" json:"url,omitempty" koanf:"url"`
}

func (p *Passlink) Validate() error {
	if len(strings.TrimSpace(p.URL)) == 0 {
		return errors.New("url must not be empty")
	}
	if url, err := url.Parse(p.URL); err != nil {
		return fmt.Errorf("failed to parse url: %w", err)
	} else if url.Scheme == "" || url.Host == "" {
		return errors.New("url must be a valid URL")
	}
	return nil
}
