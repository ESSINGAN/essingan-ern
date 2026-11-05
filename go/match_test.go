package ern

import (
	"errors"
	"strings"
	"testing"
)

func TestVectorsMatch(t *testing.T) {
	for _, c := range loadVectors(t).Match {
		got, err := Match(c.Pattern, c.ERN)
		if err != nil {
			t.Errorf("Match(%q, %q): %v", c.Pattern, c.ERN, err)
			continue
		}
		if got != c.Matches {
			t.Errorf("Match(%q, %q) = %v, want %v (%s)", c.Pattern, c.ERN, got, c.Matches, c.Why)
		}
	}
}

func TestVectorsBadPatterns(t *testing.T) {
	const name = "ern:essingan:iam:::policy/SCUBuyer"
	for _, c := range loadVectors(t).BadPatterns {
		got, err := Match(c.Pattern, name)
		if got || !errors.Is(err, ErrInvalid) {
			t.Errorf("Match(%q) = %v, %v, want false and ErrInvalid (%s)", c.Pattern, got, err, c.Why)
		}
	}
}

func TestMatchRefusesAnInvalidName(t *testing.T) {
	got, err := Match("*", "ern:essingan:iam::482019375521:user/*")
	if got || !errors.Is(err, ErrInvalid) {
		t.Errorf("Match = %v, %v, want false and ErrInvalid", got, err)
	}
}

// A hostile pattern must not make Match slow: this one blows up a backtracking
// matcher.
func TestMatchHostilePattern(t *testing.T) {
	const prefix = "ern:essingan:catalogue:::artefact/"
	name := prefix + strings.Repeat("a", 900)
	pattern := prefix + strings.Repeat("a*", 200) + "b"
	got, err := Match(pattern, name)
	if err != nil || got {
		t.Errorf("Match = %v, %v, want false", got, err)
	}
}

func TestMatchRefusesAnOverlongPattern(t *testing.T) {
	const prefix = "ern:essingan:catalogue:::artefact/"
	pattern := prefix + strings.Repeat("a", maxLength-len(prefix)) + "*"
	if _, err := Match(pattern, "ern:essingan:iam:::policy/SCUBuyer"); !errors.Is(err, ErrInvalid) {
		t.Errorf("a pattern over the limit: %v, want ErrInvalid", err)
	}
}
