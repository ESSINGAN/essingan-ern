package ern

import (
	"strings"
	"testing"
)

// The seeds are every name and pattern of the shared vectors.

// FuzzParse checks that Parse never panics, and that a name it accepts
// prints back exactly and parses again to the same value.
func FuzzParse(f *testing.F) {
	v := loadVectors(f)
	for _, c := range v.Valid {
		f.Add(c.ERN)
	}
	for _, c := range v.Invalid {
		f.Add(c.ERN)
	}
	f.Fuzz(func(t *testing.T, s string) {
		e, err := Parse(s)
		if err != nil {
			return
		}
		if e.String() != s {
			t.Fatalf("Parse(%q).String() = %q", s, e.String())
		}
		again, err := Parse(e.String())
		if err != nil || again != e {
			t.Fatalf("Parse(%q) again = %+v, %v", s, again, err)
		}
		if e.Pattern() == "" {
			t.Fatalf("Parse(%q) accepted a name with no pattern", s)
		}
	})
}

// FuzzMatch checks that Match never panics, never says yes to an invalid
// name, and that a valid name always matches itself and "*".
func FuzzMatch(f *testing.F) {
	v := loadVectors(f)
	for _, c := range v.Match {
		f.Add(c.Pattern, c.ERN)
	}
	for _, c := range v.BadPatterns {
		f.Add(c.Pattern, "ern:essingan:iam:::policy/SCUBuyer")
	}
	f.Fuzz(func(t *testing.T, pattern, name string) {
		ok, err := Match(pattern, name)
		if err != nil && ok {
			t.Fatalf("Match(%q, %q) = true with an error: %v", pattern, name, err)
		}
		if _, perr := Parse(name); perr != nil {
			if ok {
				t.Fatalf("Match(%q, %q) = true for an invalid name", pattern, name)
			}
			return
		}
		if self, err := Match(name, name); err != nil || !self {
			t.Fatalf("Match(name, name) = %v, %v for %q", self, err, name)
		}
		if all, err := Match("*", name); err != nil || !all {
			t.Fatalf(`Match("*", %q) = %v, %v`, name, all, err)
		}
		if ok && strings.Contains(pattern, "**") && !strings.HasSuffix(pattern, "/**") && !strings.HasSuffix(pattern, ":**") {
			t.Fatalf("Match(%q, %q) = true with ** not at the end", pattern, name)
		}
	})
}
