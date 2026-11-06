package ern

import (
	"math/rand"
	"strings"
	"testing"
)

// refGlob is the textbook dynamic programme: slow, but obviously right.
func refGlob(p, s string) bool {
	m := make([][]bool, len(p)+1)
	for i := range m {
		m[i] = make([]bool, len(s)+1)
	}
	m[len(p)][len(s)] = true
	for i := len(p) - 1; i >= 0; i-- {
		for j := len(s); j >= 0; j-- {
			switch {
			case p[i] == '*':
				m[i][j] = m[i+1][j] || (j < len(s) && m[i][j+1])
			case j < len(s) && (p[i] == '?' || p[i] == s[j]):
				m[i][j] = m[i+1][j+1]
			}
		}
	}
	return m[0][0]
}

// The linear glob must agree with the reference on random words (idea from
// essingan-iam #129).
func TestGlobAgreesWithTheReference(t *testing.T) {
	r := rand.New(rand.NewSource(7))
	word := func(alphabet string, n int) string {
		var b strings.Builder
		for range r.Intn(n + 1) {
			b.WriteByte(alphabet[r.Intn(len(alphabet))])
		}
		return b.String()
	}
	for range 200000 {
		p, s := word("ab*?", 8), word("ab", 10)
		if glob(p, s) != refGlob(p, s) {
			t.Fatalf("glob(%q, %q) = %v, reference %v", p, s, glob(p, s), refGlob(p, s))
		}
	}
}
