package thirdparty

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gobwas/glob"
	"github.com/teamhanko/hanko/backend/v3/config"
)

func IsAllowedRedirect(config config.ThirdParty, redirectTo string) bool {
	if redirectTo == "" {
		return false
	}

	// Keep the existing behavior: configured redirect URLs must not have a
	// trailing slash, and incoming redirect URLs are normalized by trimming one.
	redirectTo = strings.TrimSuffix(redirectTo, "/")

	// The user-controlled redirect target must be a real absolute URL. Unlike
	// configured allowlist entries, this value is not a glob pattern. It is the
	// actual destination the browser may be redirected to, so it must parse.
	redirectURL, err := url.Parse(redirectTo)
	if err != nil {
		return false
	}

	// Reject relative URLs, protocol-relative URLs, malformed URLs, and URLs
	// without a hostname. This prevents values such as:
	//
	//   /foo
	//   //evil.com
	//   https:///broken
	//
	// from being accepted as redirect destinations.
	if !redirectURL.IsAbs() || redirectURL.Hostname() == "" {
		return false
	}

	// Only allow normal browser-navigation schemes. This prevents redirects to
	// dangerous or unexpected schemes such as:
	//
	//   javascript:alert(1)
	//   data:text/html,...
	//   file:///etc/passwd
	if redirectURL.Scheme != "http" && redirectURL.Scheme != "https" {
		return false
	}

	for allowedRedirectPattern, pattern := range config.AllowedRedirectURLMap {
		// Keep legacy semantics: the configured allowlist entry is still used
		// as a glob against the full redirect URL string. This preserves support
		// for existing patterns such as:
		//
		//   http://localhost:8888**
		//   https://*.example.com/**
		//
		// However, a raw string glob alone is not safe for host validation,
		// because it can confuse "trusted host as prefix" with "trusted host as
		// actual destination". Therefore, a glob match is necessary but no longer
		// sufficient.
		if !pattern.Match(redirectTo) {
			continue
		}

		// After the legacy glob matched, perform an additional host-boundary
		// check. This prevents host-collision bypasses such as:
		//
		//   allowed pattern: http://127.0.0.1**
		//   attacker URL:    http://127.0.0.1.evil.com
		//
		// The glob may match as a string, but the parsed host is not actually
		// 127.0.0.1, so it must be rejected.
		if matchesAllowedRedirectHostBoundary(allowedRedirectPattern, redirectURL) {
			return true
		}
	}

	return false
}

// matchesAllowedRedirectHostBoundary verifies that the parsed destination host
// of redirectURL is compatible with the host part of the configured allowlist
// pattern.
//
// This function deliberately does NOT parse allowedRedirectPattern as a URL.
// Existing configurations may contain glob wildcards in places where net/url
// would reject them, e.g.:
//
//	http://localhost:8888**
//
// Instead, we extract the scheme and authority/host part textually, then
// delegate to matchesHostPatternSafely, which applies conservative host
// matching rules: exact literal hosts require an exact match; patterns whose
// wildcards are all pinned down by literal text (e.g. "*.example.com",
// "foo.*.bar.com", "192.168.*.*") are matched as a full glob; and patterns
// with an unanchored trailing wildcard (e.g. "127.0.0.1**") fall back to an
// exact match on the static text before the first wildcard, so that
// "127.0.0.1**" does not allow "127.0.0.1.evil.com".
//
// The original glob match is still responsible for path/query matching and
// broader legacy pattern compatibility. This helper only adds a safe boundary
// check around scheme, host, and port.
func matchesAllowedRedirectHostBoundary(allowedRedirectPattern string, redirectURL *url.URL) bool {
	allowedScheme, allowedAuthorityPattern, ok := extractSchemeAndAuthorityPattern(allowedRedirectPattern)
	if !ok {
		return false
	}

	// If the allowlist pattern contains a scheme, the destination URL must use
	// the same scheme. This prevents an http-only pattern from allowing https,
	// and vice versa. Keeping this strict avoids surprising behavior.
	if allowedScheme != "" && !strings.EqualFold(allowedScheme, redirectURL.Scheme) {
		return false
	}

	actualHost := normalizeHost(redirectURL.Hostname())
	actualPort := redirectURL.Port()

	allowedHostPattern, allowedPortPattern := splitAuthorityPattern(allowedAuthorityPattern)
	allowedHostPattern = normalizeHost(allowedHostPattern)

	if !matchesHostPatternSafely(allowedHostPattern, actualHost) {
		return false
	}

	// If the allowlist pattern contains a static port, require the destination
	// URL to use exactly that port.
	//
	// Example:
	//
	//   allowed: http://localhost:8888**
	//
	// allows:
	//
	//   http://localhost:8888/foo
	//
	// but rejects:
	//
	//   http://localhost:9999/foo
	//
	// If the configured port itself contains a wildcard, we only use the static
	// prefix before the wildcard. This preserves compatibility while still
	// preventing obvious host/port confusion.
	if allowedPortPattern != "" && actualPort != allowedPortPattern {
		return false
	}

	return true
}

// extractSchemeAndAuthorityPattern extracts the scheme and authority part from
// a configured redirect glob pattern without parsing it as a URL.
//
// It accepts patterns like:
//
//	http://localhost:8888**
//	https://*.example.com/**
//
// and returns:
//
//	scheme:    "http"
//	authority: "localhost:8888**"
//
// or:
//
//	scheme:    "https"
//	authority: "*.example.com"
//
// The authority is everything between "://" and the first "/", "?", or "#".
func extractSchemeAndAuthorityPattern(pattern string) (scheme string, authorityPattern string, ok bool) {
	schemeSeparatorIndex := strings.Index(pattern, "://")
	if schemeSeparatorIndex < 0 {
		return "", "", false
	}

	scheme = pattern[:schemeSeparatorIndex]
	rest := pattern[schemeSeparatorIndex+len("://"):]

	authorityEndIndex := strings.IndexAny(rest, "/?#")
	if authorityEndIndex >= 0 {
		authorityPattern = rest[:authorityEndIndex]
	} else {
		authorityPattern = rest
	}

	if scheme == "" || authorityPattern == "" {
		return "", "", false
	}

	return scheme, authorityPattern, true
}

// splitAuthorityPattern separates the configured authority pattern into host
// and port parts.
//
// Examples:
//
//	"localhost:8888**"      -> host "localhost",      port "8888"
//	"127.0.0.1**"           -> host "127.0.0.1**",    port ""
//	"*.example.com:443"     -> host "*.example.com",  port "443"
//	"example.com"           -> host "example.com",    port ""
//	"[::1]:8888**"          -> host "::1",            port "8888"
//
// The returned port is the static prefix before any glob wildcard, because an
// existing configuration may have a pattern such as ":8888**".
func splitAuthorityPattern(authorityPattern string) (hostPattern string, portPattern string) {
	// IPv6 literals in URLs are written as [::1] or [2001:db8::1]. Handle those
	// specially so the colons inside the address are not mistaken for a port
	// separator.
	if strings.HasPrefix(authorityPattern, "[") {
		closingBracketIndex := strings.Index(authorityPattern, "]")
		if closingBracketIndex < 0 {
			return authorityPattern, ""
		}

		hostPattern = authorityPattern[1:closingBracketIndex]
		rest := authorityPattern[closingBracketIndex+1:]

		if strings.HasPrefix(rest, ":") {
			portPattern = staticPrefixBeforeWildcard(rest[1:])
		}

		return hostPattern, portPattern
	}

	// For non-IPv6 authorities, the last colon separates host and port.
	// This is intentionally simple because configured patterns are legacy glob
	// strings, not necessarily valid URLs.
	colonIndex := strings.LastIndex(authorityPattern, ":")
	if colonIndex < 0 {
		return authorityPattern, ""
	}

	hostPattern = authorityPattern[:colonIndex]
	portPattern = staticPrefixBeforeWildcard(authorityPattern[colonIndex+1:])

	return hostPattern, portPattern
}

// matchesHostPatternSafely decides whether the actual parsed destination host
// is allowed by the configured host pattern.
//
// A host pattern is handled using one of three rules:
//
//  1. No wildcard characters at all: the actual host must match exactly.
//
//  2. The pattern's last "**" is not pinned down by literal text after it
//     (nothing follows, or only further wildcard characters follow, e.g.
//     "127.0.0.1**", "127.0.0.1**[a-z]", "127.0.0.1**.*"). Such a "**"
//     conveys no fixed, unspoofable boundary, so falling back to a full glob
//     match would let an attacker smuggle in their own apex domain:
//
//     allowed pattern: 127.0.0.1**
//     attacker host:   127.0.0.1.evil.com
//
//     Instead we require an exact match on the static text before the first
//     wildcard character. (A bare trailing "**" with nothing after it, e.g.
//     "*.example.com**", is first stripped and re-evaluated as the pattern
//     that remains, so this fallback only actually triggers when a further
//     wildcard character trails the "**" itself.)
//
//  3. Otherwise every wildcard -- including any "**" -- is anchored by fixed
//     literal text, so the pattern is compiled and matched as a full glob
//     (with "." as the only separator). This handles subdomain wildcards
//     ("*.example.com", "**.example.com"), bounded mid-host wildcards
//     ("foo.*.bar.com", "foo-*.bar.com", "192.168.*.*"), and anchored
//     mid-pattern super-globs ("foo.**.bar.com"), while still rejecting
//     "example.com", "example.com.evil.com", and "badexample.com" against
//     "*.example.com".
func matchesHostPatternSafely(allowedHostPattern string, actualHost string) bool {
	if allowedHostPattern == "" || actualHost == "" {
		return false
	}

	// Fast path: a pattern with no wildcard characters is a plain literal
	// host and must match exactly. (This is subsumed by the glob.Compile
	// branch below too, but is kept as a cheap allocation-free fast path for
	// the common case of a plain configured host.)
	if !strings.ContainsAny(allowedHostPattern, "*?[") {
		return actualHost == allowedHostPattern
	}

	if lastSuperGlobIndex := strings.LastIndex(allowedHostPattern, "**"); lastSuperGlobIndex >= 0 {
		tail := allowedHostPattern[lastSuperGlobIndex+len("**"):]

		if tail == "" {
			// A trailing "**" with nothing after it places no requirement on
			// what may follow -- it is purely decorative legacy shorthand
			// (e.g. "127.0.0.1**", but also "*.example.com**"). Strip it and
			// re-evaluate the remaining pattern on its own terms, so a
			// pattern like "*.example.com**" reduces to the well-understood
			// "*.example.com" wildcard-subdomain pattern instead of being
			// needlessly forced through the unbounded-super-glob fallback
			// below (which would make it unmatchable, since its static
			// prefix is empty).
			return matchesHostPatternSafely(allowedHostPattern[:lastSuperGlobIndex], actualHost)
		}

		if strings.ContainsAny(tail, "*?[") {
			// The "**" is followed by further wildcard characters, so its
			// right-hand boundary still isn't pinned down by a fixed literal
			// suffix (e.g. "127.0.0.1**[a-z]" only requires some trailing
			// a-z character; "127.0.0.1**.*" only requires some final
			// label). Fall back to the conservative rule that protects
			// against host-collision bypasses: allowed "127.0.0.1**[a-z]"
			// must not match "127.0.0.1.evil.com" just because it happens to
			// end in a lowercase letter.
			allowedHostPrefix := normalizeHost(staticPrefixBeforeWildcard(allowedHostPattern))
			if allowedHostPrefix == "" {
				return false
			}

			return actualHost == allowedHostPrefix
		}
	}

	// Every wildcard in the pattern -- including any "**" -- is anchored by
	// literal text that pins down where it must stop matching (e.g. the
	// ".example.com" suffix in "*.example.com", or the "foo." prefix and
	// ".bar.com" suffix in "foo.*.bar.com"). Compiling and matching the full
	// host pattern (hosts never contain "/", so only "." is a separator)
	// reproduces gobwas/glob's documented semantics and forces the actual
	// host to align with the pattern's fixed structure, so it cannot be used
	// to smuggle in an attacker-controlled apex domain.
	hostGlob, err := glob.Compile(allowedHostPattern, '.')
	if err != nil {
		return false
	}

	return hostGlob.Match(actualHost)
}

// staticPrefixBeforeWildcard returns the part of a configured glob segment that
// appears before the first wildcard character.
//
// This lets us derive the stable host or port prefix from legacy patterns such
// as:
//
//	localhost**
//	8888**
//
// without requiring the entire allowlist entry to be a valid URL.
func staticPrefixBeforeWildcard(value string) string {
	wildcardIndex := strings.IndexAny(value, "*?[")
	if wildcardIndex < 0 {
		return value
	}

	return value[:wildcardIndex]
}

// normalizeHost canonicalizes hostnames for comparison.
//
// It lowercases hostnames because DNS names are case-insensitive, and removes a
// trailing dot so that:
//
//	example.com
//
// and:
//
//	example.com.
//
// compare the same.
func normalizeHost(host string) string {
	return strings.TrimSuffix(strings.ToLower(host), ".")
}

func GetErrorUrl(redirectTo string, err error) string {
	var redirectUrl string
	switch v := err.(type) {
	case *ThirdPartyError:
		redirectUrl = fmt.Sprintf("%s?%s", redirectTo, v.Query())
	default:
		u := url.Values{}
		u.Add("error", ErrorCodeServerError)
		u.Add("error_description", "an internal error has occurred")
		redirectUrl = fmt.Sprintf("%s?%s", redirectTo, u.Encode())
	}
	return redirectUrl
}
