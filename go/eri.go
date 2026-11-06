package ern

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"regexp"
	"strings"
	"time"
)

// crockford is the lowercase Crockford base32 alphabet (no i, l, o or u). It
// is in ASCII order, so ERIs sort as strings in the order they were made.
const crockford = "0123456789abcdefghjkmnpqrstvwxyz"

const (
	eriTimeChars   = 10 // the ULID timestamp: 48 bits of Unix milliseconds
	eriRandomChars = 6  // 30 random bits
	eriMaxMillis   = 1<<48 - 1
)

var (
	// eriRe is <type-prefix>-<16 chars>, in lowercase Crockford base32. The
	// prefix length comes from the Hub API contract.
	eriRe       = regexp.MustCompile(`^[a-z]{1,4}-[0-9a-hjkmnp-tv-z]{16}$`)
	eriPrefixRe = regexp.MustCompile(`^[a-z]{1,4}$`)
	accountIDRe = regexp.MustCompile(`^[0-9]{12}$`)
)

// IsERI reports whether s is a well-formed ESSINGAN Resource Identifier, such
// as p-01jcz7k2m9q4x8n2.
func IsERI(s string) bool {
	return eriRe.MatchString(s)
}

// NewERI returns a new ERI with the given type prefix, such as "p" for a
// project. The 16 characters are the 10 time characters of a ULID followed by
// 6 random ones, so ERIs sort by creation time. Two ERIs made in the same
// millisecond are in random order. Callers keep a unique index: 30 random
// bits make a clash unlikely, not impossible.
func NewERI(prefix string) (string, error) {
	var b [4]byte
	_, _ = rand.Read(b[:]) // crypto/rand.Read never returns an error
	return makeERI(prefix, time.Now(), binary.BigEndian.Uint32(b[:]))
}

// DeriveERI returns the ERI of an object that already exists elsewhere under
// another stable id (seed), created at createdAt: the same inputs always give
// the same ERI, so nothing has to be stored to name it. The 6 last characters
// come from SHA-256(prefix, NUL, seed) instead of chance. A zero or pre-1970
// createdAt counts as the Unix epoch. Same formula as essingan-iam #129.
func DeriveERI(prefix string, createdAt time.Time, seed string) (string, error) {
	if createdAt.Before(time.Unix(0, 0)) {
		createdAt = time.Unix(0, 0)
	}
	sum := sha256.Sum256([]byte(prefix + "\x00" + seed))
	return makeERI(prefix, createdAt, binary.BigEndian.Uint32(sum[:4])>>2)
}

// IsERIOf reports whether s is an ERI with the given prefix, such as "p".
func IsERIOf(prefix, s string) bool {
	return IsERI(s) && strings.HasPrefix(s, prefix+"-")
}

// ERITime returns the creation time held in an ERI, to the millisecond.
func ERITime(id string) (time.Time, error) {
	if !IsERI(id) {
		return time.Time{}, invalid("bad ERI %q", id)
	}
	body := id[strings.IndexByte(id, '-')+1:]
	if body[0] > '7' {
		return time.Time{}, invalid("ERI %q holds a time past the 48-bit limit", id)
	}
	var ms int64
	for i := range eriTimeChars {
		ms = ms<<5 | int64(strings.IndexByte(crockford, body[i]))
	}
	return time.UnixMilli(ms).UTC(), nil
}

func makeERI(prefix string, t time.Time, random uint32) (string, error) {
	if !eriPrefixRe.MatchString(prefix) {
		return "", invalid("bad ERI prefix %q", prefix)
	}
	ms := t.UnixMilli()
	if ms < 0 || ms > eriMaxMillis {
		return "", invalid("time %s is outside the ERI range", t)
	}
	var b strings.Builder
	b.Grow(len(prefix) + 1 + eriTimeChars + eriRandomChars)
	b.WriteString(prefix)
	b.WriteByte('-')
	b.WriteString(base32(uint64(ms), eriTimeChars))
	b.WriteString(base32(uint64(random), eriRandomChars)) // keeps the low 30 bits
	return b.String(), nil
}

// base32 writes the low 5*width bits of n on exactly width Crockford
// characters, most significant first.
func base32(n uint64, width int) string {
	out := make([]byte, width)
	for i := width - 1; i >= 0; i-- {
		out[i] = crockford[n&31]
		n >>= 5
	}
	return string(out)
}

// ValidateAccountID checks a 12-digit account ID.
func ValidateAccountID(id string) error {
	if !accountIDRe.MatchString(id) {
		return invalid("bad account ID %q", id)
	}
	return nil
}

// DisplayAccountID formats an account ID for people, as 4820-1937-5521. It
// returns "" for anything else than 12 digits.
func DisplayAccountID(id string) string {
	if ValidateAccountID(id) != nil {
		return ""
	}
	return id[:4] + "-" + id[4:8] + "-" + id[8:]
}
