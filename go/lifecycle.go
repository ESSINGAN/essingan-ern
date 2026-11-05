package ern

// State is the lifecycle vocabulary shared by every service (ADR-0057, Doc 73-12
// §3, and the LifecycleState schema of the Hub API contract). A service may add
// sub-states inside its own resource, never new top-level states.
type State string

const (
	Pending  State = "PENDING"
	Creating State = "CREATING"
	Active   State = "ACTIVE"
	Updating State = "UPDATING"
	Deleting State = "DELETING"
	Deleted  State = "DELETED"
	Failed   State = "FAILED"
)

// IsState reports whether s belongs to the shared vocabulary.
func IsState(s string) bool {
	switch State(s) {
	case Pending, Creating, Active, Updating, Deleting, Deleted, Failed:
		return true
	}
	return false
}
