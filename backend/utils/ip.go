package utils

import (
	"fmt"
	"net"

	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/v3/config"
)

func ConfigureIPExtractor(e *echo.Echo, cfg config.IPConfig) error {
	switch cfg.Extractor {
	case "", config.IPExtractorDirect:
		e.IPExtractor = echo.ExtractIPDirect()
		return nil

	case config.IPExtractorXForwardedFor:
		trustOptions, err := buildTrustOptions(cfg.TrustedProxies)
		if err != nil {
			return err
		}

		e.IPExtractor = echo.ExtractIPFromXFFHeader(trustOptions...)
		return nil

	case config.IPExtractorXRealIP:
		trustOptions, err := buildTrustOptions(cfg.TrustedProxies)
		if err != nil {
			return err
		}

		e.IPExtractor = echo.ExtractIPFromRealIPHeader(trustOptions...)
		return nil

	default:
		return fmt.Errorf("unsupported IP extractor %q", cfg.Extractor)
	}
}

func buildTrustOptions(trustedProxies []string) ([]echo.TrustOption, error) {
	if len(trustedProxies) == 0 {
		return nil, fmt.Errorf("trusted_proxies must be configured when using a header-based IP extractor")
	}

	trustOptions := make([]echo.TrustOption, 0, len(trustedProxies))

	for _, cidr := range trustedProxies {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, fmt.Errorf("invalid trusted proxy CIDR %q: %w", cidr, err)
		}

		trustOptions = append(trustOptions, echo.TrustIPRange(ipNet))
	}

	return trustOptions, nil
}
