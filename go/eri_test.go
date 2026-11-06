package ern

import (
	"cmp"
	"errors"
	"slices"
	"testing"
	"time"
)

func TestVectorsERI(t *testing.T) {
	cases := loadVectors(t).ERI
	if len(cases) == 0 {
		t.Fatal("no ERI vectors")
	}
	var made []string
	for _, c := range cases {
		got, err := makeERI(c.Prefix, time.UnixMilli(c.UnixMs), c.Random)
		if err != nil {
			t.Fatalf("makeERI(%q, %d, %d): %v", c.Prefix, c.UnixMs, c.Random, err)
		}
		if got != c.ERI {
			t.Errorf("makeERI(%q, %d, %d) = %q, want %q", c.Prefix, c.UnixMs, c.Random, got, c.ERI)
		}
		if !IsERI(got) {
			t.Errorf("IsERI(%q) = false", got)
		}
		at, err := ERITime(got)
		if err != nil || at.UnixMilli() != c.UnixMs {
			t.Errorf("ERITime(%q) = %v, %v, want %d ms", got, at, err, c.UnixMs)
		}
		made = append(made, got[len(c.Prefix)+1:])
	}
	// Ordered by time, then by random part, the bodies must sort as strings.
	order := make([]int, len(cases))
	for i := range order {
		order[i] = i
	}
	slices.SortFunc(order, func(a, b int) int {
		return cmp.Or(cmp.Compare(cases[a].UnixMs, cases[b].UnixMs), cmp.Compare(cases[a].Random, cases[b].Random))
	})
	inTime := make([]string, len(order))
	for i, k := range order {
		inTime[i] = made[k]
	}
	if !slices.IsSorted(inTime) {
		t.Errorf("ERI bodies in time order do not sort as strings: %v", made)
	}
}

func TestNewERI(t *testing.T) {
	before := time.Now().Truncate(time.Millisecond)
	id, err := NewERI("p")
	if err != nil {
		t.Fatal(err)
	}
	after := time.Now()
	if !isPrefixedERI(id, "p") {
		t.Fatalf("NewERI(p) = %q, not a p- ERI", id)
	}
	at, err := ERITime(id)
	if err != nil {
		t.Fatal(err)
	}
	if at.Before(before) || at.After(after) {
		t.Errorf("ERITime(%q) = %v, want between %v and %v", id, at, before, after)
	}
	if _, err := NewProject(acct, org, id); err != nil {
		t.Errorf("a new ERI is refused by the builders: %v", err)
	}
}

func TestNewERIIsRandom(t *testing.T) {
	seen := map[string]bool{}
	for range 1000 {
		id, err := NewERI("u")
		if err != nil {
			t.Fatal(err)
		}
		if seen[id] {
			t.Fatalf("NewERI made %q twice", id)
		}
		seen[id] = true
	}
}

func TestERISortsByTime(t *testing.T) {
	start := time.Date(2026, 11, 6, 9, 0, 0, 0, time.UTC)
	var ids []string
	for i := range 200 {
		// The highest random part first, then the lowest: time must win.
		random := uint32(1<<30 - 1)
		if i%2 == 1 {
			random = 0
		}
		id, err := makeERI("o", start.Add(time.Duration(i*i)*time.Millisecond), random)
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
	}
	if !slices.IsSorted(ids) {
		t.Error("ERIs made one after the other do not sort as strings")
	}
}

func TestMakeERIRefusesBadInput(t *testing.T) {
	now := time.Now()
	cases := map[string]func() (string, error){
		"empty prefix":       func() (string, error) { return makeERI("", now, 0) },
		"prefix too long":    func() (string, error) { return makeERI("abcde", now, 0) },
		"uppercase prefix":   func() (string, error) { return makeERI("P", now, 0) },
		"prefix with a dash": func() (string, error) { return makeERI("p-", now, 0) },
		"before 1970":        func() (string, error) { return makeERI("p", time.UnixMilli(-1), 0) },
		"past 48 bits":       func() (string, error) { return makeERI("p", time.UnixMilli(eriMaxMillis+1), 0) },
		"NewERI bad prefix":  func() (string, error) { return NewERI("Proj") },
	}
	for name, f := range cases {
		if id, err := f(); !errors.Is(err, ErrInvalid) {
			t.Errorf("%s: %q, %v, want ErrInvalid", name, id, err)
		}
	}
}

func TestMakeERIKeepsOnly30RandomBits(t *testing.T) {
	at := time.UnixMilli(1793956800000)
	a, _ := makeERI("p", at, 5)
	b, _ := makeERI("p", at, 5|1<<30|1<<31)
	if a != b {
		t.Errorf("bits above 30 leak into the ERI: %q vs %q", a, b)
	}
}

func TestERITimeRefusesBadInput(t *testing.T) {
	for _, id := range []string{"", "p-01m6r2gsg", "P-01m6r2gsg0000000", "p-01m6r2gsg000000u", "p-8000000000000000"} {
		if _, err := ERITime(id); !errors.Is(err, ErrInvalid) {
			t.Errorf("ERITime(%q): %v, want ErrInvalid", id, err)
		}
	}
}
