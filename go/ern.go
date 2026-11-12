// Package ern implements the ESSINGAN Resource Name (ADR-0051, Doc 73-01):
// it parses, builds, validates and matches ERNs. It holds no data and makes
// no network call.
package ern

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// DefaultPartition is the Hub operated by SPITZKOP. A customer's own Management
// Plane uses "essingan-<id>".
const DefaultPartition = "essingan"

// maxLength comes from the Ern schema of the Hub API contract.
const maxLength = 1024

// ErrInvalid is wrapped by every validation error.
var ErrInvalid = errors.New("invalid ERN")

// Pattern is one of the normative forms of Doc 73-01 §4.
type Pattern string

const (
	E1  Pattern = "E1"  // the account
	E2  Pattern = "E2"  // an account-level resource
	E3  Pattern = "E3"  // an organization
	E4  Pattern = "E4"  // an organization-level resource
	E5  Pattern = "E5"  // a project
	E6  Pattern = "E6"  // a project resource on a provider
	E7  Pattern = "E7"  // a logical project resource
	E8  Pattern = "E8"  // a sub-resource
	E9  Pattern = "E9"  // an ESSINGAN-managed object, shared by every account
	E10 Pattern = "E10" // a role session
)

const roleSession = "role-session"

// services is the closed list of Doc 73-01 §5 and ADR-0051 D8.
var services = map[string]bool{
	"account": true, "iam": true, "sts": true, "billing": true, "scu": true,
	"spoke": true, "stif": true, "lithos": true, "arca": true, "veloce": true,
	"specula": true, "sentis": true, "cno": true, "audit": true, "catalogue": true,
}

// IsService reports whether name is a service namespace of that list. Policy
// actions (<service>:<Verb><Type>) use the same namespaces.
func IsService(name string) bool { return services[name] }

// Services returns the service namespaces, sorted.
func Services() []string {
	out := make([]string, 0, len(services))
	for s := range services {
		out = append(out, s)
	}
	slices.Sort(out)
	return out
}

var (
	partitionRe = regexp.MustCompile(`^essingan(-[a-z0-9]+)?$`)
	providerRe  = regexp.MustCompile(`^[a-z][a-z0-9]*$`)
	regionRe    = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	pathRe      = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9/._@+=-]*$`)
	typeRe      = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
	nameRe      = regexp.MustCompile(`^[A-Za-z0-9._@+=-]+$`)
)

// ERN is a parsed ESSINGAN Resource Name:
// ern:<partition>:<service>:<location>:<account>:<resource-path>
type ERN struct {
	Partition string
	Service   string
	Location  string
	Account   string
	Path      string
}

// Parse reads and validates an ERN. A stored ERN never holds a wildcard, so
// Parse refuses them: use Match for policy patterns.
func Parse(s string) (ERN, error) {
	if len(s) > maxLength {
		return ERN{}, invalid("longer than %d characters", maxLength)
	}
	parts := strings.SplitN(s, ":", 6)
	if len(parts) != 6 {
		return ERN{}, invalid("%q does not have 6 components", s)
	}
	if parts[0] != "ern" {
		return ERN{}, invalid("%q does not start with \"ern\"", s)
	}
	e := ERN{Partition: parts[1], Service: parts[2], Location: parts[3], Account: parts[4], Path: parts[5]}
	if err := e.Validate(); err != nil {
		return ERN{}, err
	}
	return e, nil
}

// String returns the text form of the ERN.
func (e ERN) String() string {
	return strings.Join([]string{"ern", e.Partition, e.Service, e.Location, e.Account, e.Path}, ":")
}

// Validate checks every component and the resource path.
func (e ERN) Validate() error {
	_, err := e.classify()
	return err
}

// Pattern returns the form this ERN follows, or "" if it is invalid.
func (e ERN) Pattern() Pattern {
	p, err := e.classify()
	if err != nil {
		return ""
	}
	return p
}

func (e ERN) classify() (Pattern, error) {
	if len(e.String()) > maxLength {
		return "", invalid("longer than %d characters", maxLength)
	}
	if !partitionRe.MatchString(e.Partition) {
		return "", invalid("bad partition %q", e.Partition)
	}
	if !services[e.Service] {
		return "", invalid("unknown service %q", e.Service)
	}
	if err := validateLocation(e.Location); err != nil {
		return "", err
	}
	if e.Account != "" {
		if err := ValidateAccountID(e.Account); err != nil {
			return "", err
		}
	}
	segs, err := splitPath(e.Path)
	if err != nil {
		return "", err
	}
	if e.Account == "" {
		return e.classifyManaged(segs)
	}
	if e.Service == "sts" && isRoleSession(segs) {
		if e.Location != "" {
			return "", invalid("a role session has no location")
		}
		return E10, nil
	}
	if IsERIOf("o", segs[0]) {
		if len(segs) > 1 && IsERIOf("p", segs[1]) {
			return e.classifyInProject(segs[2:])
		}
		return e.classifyInOrganization(segs[1:])
	}
	return e.classifyInAccount(segs)
}

// classifyManaged handles E9, the only form without an account.
func (e ERN) classifyManaged(segs []string) (Pattern, error) {
	if e.Location != "" {
		return "", invalid("an ESSINGAN-managed object has no location")
	}
	if len(segs) < 2 || !typeRe.MatchString(segs[0]) {
		return "", invalid("path %q is not <type>/<name>", e.Path)
	}
	for _, name := range segs[1:] {
		if !nameRe.MatchString(name) {
			return "", invalid("bad name %q", name)
		}
	}
	return E9, nil
}

func (e ERN) classifyInAccount(segs []string) (Pattern, error) {
	if e.Location != "" {
		return "", invalid("an account-level resource has no location")
	}
	if len(segs) == 1 && segs[0] == "account" && e.Service == "account" {
		return E1, nil
	}
	if !arePairs(segs) {
		return "", invalid("path %q is not <type>/<id>", e.Path)
	}
	if len(segs) == 4 {
		return E8, nil
	}
	if e.Service == "account" {
		switch segs[0] {
		case "organization":
			if !IsERIOf("o", segs[1]) {
				return "", invalid("%q is not an organization id", segs[1])
			}
			return E3, nil
		case "project":
			return "", invalid("a project lives in an organization")
		}
	}
	return E2, nil
}

func (e ERN) classifyInOrganization(segs []string) (Pattern, error) {
	if e.Location != "" {
		return "", invalid("an organization-level resource has no location")
	}
	if len(segs) != 2 || !arePairs(segs) {
		return "", invalid("path %q is not o-…/<type>/<id>", e.Path)
	}
	if e.Service == "account" && segs[0] == "project" {
		if !IsERIOf("p", segs[1]) {
			return "", invalid("%q is not a project id", segs[1])
		}
		return E5, nil
	}
	return E4, nil
}

func (e ERN) classifyInProject(segs []string) (Pattern, error) {
	if !arePairs(segs) || len(segs) > 4 {
		return "", invalid("path %q is not o-…/p-…/<type>/<id>", e.Path)
	}
	if len(segs) == 4 {
		return E8, nil
	}
	if e.Location != "" {
		return E6, nil
	}
	return E7, nil
}

// isRoleSession accepts <path>/role-session/<role-name>/<session>, where
// <path> is empty, an organization, or an organization and a project.
func isRoleSession(segs []string) bool {
	switch len(segs) {
	case 3:
	case 4:
		if !IsERIOf("o", segs[0]) {
			return false
		}
	case 5:
		if !IsERIOf("o", segs[0]) || !IsERIOf("p", segs[1]) {
			return false
		}
	default:
		return false
	}
	tail := segs[len(segs)-3:]
	return tail[0] == roleSession && nameRe.MatchString(tail[1]) && nameRe.MatchString(tail[2])
}

// arePairs reports whether segs is a non-empty list of <type>/<eri> pairs.
func arePairs(segs []string) bool {
	if len(segs) == 0 || len(segs)%2 != 0 {
		return false
	}
	for i := 0; i < len(segs); i += 2 {
		if !typeRe.MatchString(segs[i]) || !IsERI(segs[i+1]) {
			return false
		}
	}
	return true
}

func splitPath(path string) ([]string, error) {
	if !pathRe.MatchString(path) {
		return nil, invalid("bad resource path %q", path)
	}
	segs := strings.Split(path, "/")
	for _, s := range segs {
		if s == "" || s == "." || s == ".." {
			return nil, invalid("bad segment in resource path %q", path)
		}
	}
	return segs, nil
}

// validateLocation accepts an empty location or <provider>-<region>. On Lithos
// the region is the ERI of a site (lithos-s-…). It checks the shape only: whether
// the provider is registered is CheckProvider's job.
func validateLocation(loc string) error {
	if loc == "" {
		return nil
	}
	provider, region, ok := strings.Cut(loc, "-")
	if !ok || !providerRe.MatchString(provider) {
		return invalid("bad location %q", loc)
	}
	if provider == "lithos" {
		if !IsERIOf("s", region) {
			return invalid("bad Lithos site in location %q", loc)
		}
		return nil
	}
	if !regionRe.MatchString(region) {
		return invalid("bad region in location %q", loc)
	}
	return nil
}

func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalid, fmt.Sprintf(format, args...))
}
