package ern

import (
	"regexp"
	"strings"
)

// patternRe is the ErnPattern schema of the Hub API contract.
var patternRe = regexp.MustCompile(`^(\*|ern:essingan[A-Za-z0-9:/._@+=*?-]*)$`)

// Match reports whether name, a stored ERN, matches a policy pattern. "*" alone
// matches everything. Elsewhere, "*" matches any characters and "?" exactly one,
// within one component or path segment, so they never cross ":" or "/".
// "**", only as the whole last segment of the resource path, matches one or
// more whole segments: what lives under the path before it, as a trailing
// "/**" in .gitignore (ADR-0051 D10).
//
// An invalid pattern or name returns false and an error: a caller must treat it
// as no match.
func Match(pattern, name string) (bool, error) {
	e, err := Parse(name)
	if err != nil {
		return false, err
	}
	if err := ValidatePattern(pattern); err != nil {
		return false, err
	}
	if pattern == "*" {
		return true, nil
	}
	parts := strings.SplitN(pattern, ":", 6)
	segs := strings.Split(parts[5], "/")

	names := strings.Split(e.Path, "/")
	if last := len(segs) - 1; segs[last] == "**" {
		// The segments before ** must match, and at least one must follow.
		segs = segs[:last]
		if len(names) <= len(segs) {
			return false, nil
		}
		names = names[:len(segs)]
	}
	if len(segs) != len(names) {
		return false, nil
	}
	if !glob(parts[1], e.Partition) || !glob(parts[2], e.Service) ||
		!glob(parts[3], e.Location) || !glob(parts[4], e.Account) {
		return false, nil
	}
	for i := range segs {
		if !glob(segs[i], names[i]) {
			return false, nil
		}
	}
	return true, nil
}

// ValidatePattern checks a policy pattern the way Match reads it: "*" alone,
// or six components whose resource path has no empty segment, with "**" only
// as its whole last segment. A policy store calls it before saving a rule, so
// a rule Match would refuse is never stored.
func ValidatePattern(pattern string) error {
	if len(pattern) > maxLength || !patternRe.MatchString(pattern) {
		return invalid("bad pattern %q", pattern)
	}
	if pattern == "*" {
		return nil
	}
	parts := strings.SplitN(pattern, ":", 6)
	if len(parts) != 6 || strings.Contains(parts[5], ":") || parts[5] == "" {
		return invalid("pattern %q does not have 6 components", pattern)
	}
	if strings.Contains(strings.Join(parts[:5], ":"), "**") {
		return invalid("** outside the resource path in pattern %q", pattern)
	}
	segs := strings.Split(parts[5], "/")
	for i, s := range segs {
		if s == "" {
			return invalid("empty segment in pattern %q", pattern)
		}
		if strings.Contains(s, "**") && (s != "**" || i != len(segs)-1) {
			return invalid("** is allowed only as the last segment, in pattern %q", pattern)
		}
	}
	return nil
}

// glob matches one component or segment, in time proportional to
// len(pattern)*len(s), whatever the pattern.
func glob(pattern, s string) bool {
	p, i := 0, 0
	star, retry := -1, 0
	for i < len(s) {
		switch {
		case p < len(pattern) && pattern[p] == '*':
			star, retry = p, i
			p++
		case p < len(pattern) && (pattern[p] == '?' || pattern[p] == s[i]):
			p++
			i++
		case star >= 0:
			retry++
			p, i = star+1, retry
		default:
			return false
		}
	}
	for p < len(pattern) && pattern[p] == '*' {
		p++
	}
	return p == len(pattern)
}
