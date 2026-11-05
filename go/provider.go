package ern

import "strings"

// ProviderRegistry answers whether a provider is registered. The list lives in
// the catalogue (ADR-0051 D7, as amended by D63), never in code: adding a
// provider must not need a new version of this library.
type ProviderRegistry interface {
	HasProvider(name string) bool
}

// ProviderSet is a ProviderRegistry over a list the caller has already read
// from the catalogue.
type ProviderSet map[string]bool

// NewProviderSet returns a ProviderSet holding names.
func NewProviderSet(names ...string) ProviderSet {
	set := make(ProviderSet, len(names))
	for _, name := range names {
		set[name] = true
	}
	return set
}

// HasProvider implements ProviderRegistry.
func (s ProviderSet) HasProvider(name string) bool {
	return s[name]
}

// Provider returns the provider of e's location, or "" for a logical resource.
func (e ERN) Provider() string {
	provider, _, _ := strings.Cut(e.Location, "-")
	return provider
}

// CheckProvider refuses an ERN whose provider is not in the registry. Parse and
// Validate check the shape only, because they have no registry to ask. A logical
// resource has no provider, so it always passes when a registry is given.
func (e ERN) CheckProvider(registry ProviderRegistry) error {
	if registry == nil {
		return invalid("no provider registry to check %s against", e)
	}
	provider := e.Provider()
	if provider != "" && !registry.HasProvider(provider) {
		return invalid("provider %q is not registered", provider)
	}
	return nil
}
