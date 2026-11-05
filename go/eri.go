package ern

import (
	"regexp"
	"strings"
)

var (
	// eriRe is <type-prefix>-<16 chars>, in lowercase Crockford base32 (no i, l,
	// o or u). The prefix length comes from the Hub API contract.
	eriRe       = regexp.MustCompile(`^[a-z]{1,4}-[0-9a-hjkmnp-tv-z]{16}$`)
	accountIDRe = regexp.MustCompile(`^[0-9]{12}$`)
)

// IsERI reports whether s is a well-formed ESSINGAN Resource Identifier, such
// as p-01jcz7k2m9q4x8n2. It does not generate ERIs.
func IsERI(s string) bool {
	return eriRe.MatchString(s)
}

// ValidateAccountID checks a 12-digit account ID.
func ValidateAccountID(id string) error {
	if !accountIDRe.MatchString(id) {
		return invalid("bad account ID %q", id)
	}
	return nil
}

func isPrefixedERI(s, prefix string) bool {
	return IsERI(s) && strings.HasPrefix(s, prefix+"-")
}
