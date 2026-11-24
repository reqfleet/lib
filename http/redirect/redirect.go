// Derived from reqfleet/shibuya; modified for lib.
package redirect

import (
	"net/url"
	"strings"
)

// IsValidRedirectURL validates that a redirect URL is safe to use.
//
// It ensures the URL is a local absolute-path reference (same origin) and helps
// prevent open redirect vulnerabilities.
func IsValidRedirectURL(redirectURL string) bool {
	if redirectURL == "" {
		return false
	}
	// Prevent header injection and other control character issues.
	if strings.IndexFunc(redirectURL, func(r rune) bool {
		return r < 0x20 || r == 0x7f
	}) != -1 {
		return false
	}

	u, err := url.Parse(redirectURL)
	if err != nil {
		return false
	}
	// Only allow local absolute-path references (same origin). Reject anything
	// with a scheme/host/userinfo.
	if u.IsAbs() || u.Scheme != "" || u.Host != "" || u.User != nil || u.Opaque != "" {
		return false
	}
	if !strings.HasPrefix(u.Path, "/") {
		return false
	}
	// Reject protocol-relative and backslash variants.
	if strings.HasPrefix(redirectURL, "//") || strings.Contains(redirectURL, "\\") {
		return false
	}
	return true
}
