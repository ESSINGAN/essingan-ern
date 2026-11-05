package ern

import (
	"encoding/json"
	"errors"
	"os"
	"regexp"
	"strings"
	"testing"
)

// contractErnRe is the Ern schema of hub-api.openapi.yaml: everything the
// library accepts must also pass the contract.
var contractErnRe = regexp.MustCompile(`^ern:essingan(-[a-z0-9]+)?:[a-z]+:[a-z0-9-]*:([0-9]{12})?:[A-Za-z0-9][A-Za-z0-9/._@+=-]*$`)

type vectors struct {
	Valid []struct {
		ERN     string  `json:"ern"`
		Pattern Pattern `json:"pattern"`
	} `json:"valid"`
	Parts []struct {
		ERN       string `json:"ern"`
		Partition string `json:"partition"`
		Service   string `json:"service"`
		Location  string `json:"location"`
		Account   string `json:"account"`
		Path      string `json:"path"`
	} `json:"parts"`
	Invalid []struct {
		ERN string `json:"ern"`
		Why string `json:"why"`
	} `json:"invalid"`
	Match []struct {
		Pattern string `json:"pattern"`
		ERN     string `json:"ern"`
		Matches bool   `json:"matches"`
		Why     string `json:"why"`
	} `json:"match"`
	BadPatterns []struct {
		Pattern string `json:"pattern"`
		Why     string `json:"why"`
	} `json:"bad_patterns"`
	Registry struct {
		Providers []string `json:"providers"`
		Checks    []struct {
			ERN        string `json:"ern"`
			Registered bool   `json:"registered"`
			Why        string `json:"why"`
		} `json:"checks"`
	} `json:"registry"`
	LifecycleStates    []string `json:"lifecycle_states"`
	BadLifecycleStates []string `json:"bad_lifecycle_states"`
}

func loadVectors(t *testing.T) vectors {
	t.Helper()
	data, err := os.ReadFile("testdata/vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var v vectors
	if err := json.Unmarshal(data, &v); err != nil {
		t.Fatal(err)
	}
	return v
}

func TestVectorsValid(t *testing.T) {
	seen := map[Pattern]bool{}
	for _, c := range loadVectors(t).Valid {
		e, err := Parse(c.ERN)
		if err != nil {
			t.Errorf("Parse(%q): %v", c.ERN, err)
			continue
		}
		if got := e.Pattern(); got != c.Pattern {
			t.Errorf("Pattern(%q) = %s, want %s", c.ERN, got, c.Pattern)
		}
		if got := e.String(); got != c.ERN {
			t.Errorf("String() = %q, want %q", got, c.ERN)
		}
		if !contractErnRe.MatchString(c.ERN) {
			t.Errorf("%q fails the contract's schema", c.ERN)
		}
		seen[c.Pattern] = true
	}
	for _, p := range []Pattern{E1, E2, E3, E4, E5, E6, E7, E8, E9, E10} {
		if !seen[p] {
			t.Errorf("no valid vector for %s", p)
		}
	}
}

func TestVectorsParts(t *testing.T) {
	for _, c := range loadVectors(t).Parts {
		e, err := Parse(c.ERN)
		if err != nil {
			t.Errorf("Parse(%q): %v", c.ERN, err)
			continue
		}
		want := ERN{c.Partition, c.Service, c.Location, c.Account, c.Path}
		if e != want {
			t.Errorf("Parse(%q) = %+v, want %+v", c.ERN, e, want)
		}
	}
}

func TestVectorsInvalid(t *testing.T) {
	for _, c := range loadVectors(t).Invalid {
		_, err := Parse(c.ERN)
		if !errors.Is(err, ErrInvalid) {
			t.Errorf("Parse(%q) = %v, want ErrInvalid (%s)", c.ERN, err, c.Why)
		}
	}
}

func TestLengthLimit(t *testing.T) {
	const prefix = "ern:essingan:catalogue:::artefact/"
	atLimit := prefix + strings.Repeat("a", maxLength-len(prefix))
	if _, err := Parse(atLimit); err != nil {
		t.Errorf("an ERN of %d characters: %v", maxLength, err)
	}
	if _, err := Parse(atLimit + "a"); !errors.Is(err, ErrInvalid) {
		t.Errorf("an ERN of %d characters: %v, want ErrInvalid", maxLength+1, err)
	}
}

func TestIsERI(t *testing.T) {
	for s, want := range map[string]bool{
		"p-01jcz7k2m9q4x8n2":     true,
		"spk-00000000000000mp":   true,
		"cred-01jeh2m4p6r8t3v5":  true,
		"p-01jcz7k2m9q4x8n":      false,
		"p-01jcz7k2m9q4x8n22":    false,
		"P-01jcz7k2m9q4x8n2":     false,
		"prefx-01jcz7k2m9q4x8n2": false,
		"p-01jcz7k2m9q4x8nu":     false,
		"p_01jcz7k2m9q4x8n2":     false,
		"":                       false,
	} {
		if got := IsERI(s); got != want {
			t.Errorf("IsERI(%q) = %v, want %v", s, got, want)
		}
	}
}

func TestVectorsRegistry(t *testing.T) {
	v := loadVectors(t)
	registry := NewProviderSet(v.Registry.Providers...)
	for _, c := range v.Registry.Checks {
		e, err := Parse(c.ERN)
		if err != nil {
			t.Errorf("Parse(%q): %v", c.ERN, err)
			continue
		}
		err = e.CheckProvider(registry)
		if c.Registered && err != nil {
			t.Errorf("CheckProvider(%q): %v", c.ERN, err)
		}
		if !c.Registered && !errors.Is(err, ErrInvalid) {
			t.Errorf("CheckProvider(%q) = %v, want ErrInvalid (%s)", c.ERN, err, c.Why)
		}
	}
}

func TestCheckProviderWithoutRegistry(t *testing.T) {
	e, err := Parse("ern:essingan:arca:hetzner-fsn1:482019375521:o-01jd2v8k4m7q9x3a/p-01jf3w9n5r8t2y6b/instance/db-01jf4x2n5r8t2y6c")
	if err != nil {
		t.Fatal(err)
	}
	if err := e.CheckProvider(nil); !errors.Is(err, ErrInvalid) {
		t.Errorf("CheckProvider(nil) = %v, want ErrInvalid", err)
	}
}

func TestProvider(t *testing.T) {
	for location, want := range map[string]string{
		"":                  "",
		"scaleway-fr-par":   "scaleway",
		"lithos-s-01jf4x2n": "lithos",
	} {
		if got := (ERN{Location: location}).Provider(); got != want {
			t.Errorf("Provider(%q) = %q, want %q", location, got, want)
		}
	}
}
