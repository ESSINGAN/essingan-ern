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
//
// An invalid pattern or name returns false and an error: a caller must treat it
// as no match.
func Match(pattern, name string) (bool, error) {
	e, err := Parse(name)
	if err != nil {
		return false, err
	}
	if len(pattern) > maxLength || !patternRe.MatchString(pattern) {
		return false, invalid("bad pattern %q", pattern)
	}
	if pattern == "*" {
		return true, nil
	}
	parts := strings.SplitN(pattern, ":", 6)
	if len(parts) != 6 || strings.Contains(parts[5], ":") || parts[5] == "" {
		return false, invalid("pattern %q does not have 6 components", pattern)
	}
	segs := strings.Split(parts[5], "/")
	for _, s := range segs {
		if s == "" {
			return false, invalid("empty segment in pattern %q", pattern)
		}
	}

	names := strings.Split(e.Path, "/")
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
