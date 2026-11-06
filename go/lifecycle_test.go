package ern

import "testing"

func TestVectorsLifecycle(t *testing.T) {
	v := loadVectors(t)
	if len(v.LifecycleStates) != 7 {
		t.Errorf("%d states in the vectors, want 7", len(v.LifecycleStates))
	}
	for _, s := range v.LifecycleStates {
		if !IsState(s) {
			t.Errorf("IsState(%q) = false", s)
		}
	}
	for _, s := range v.BadLifecycleStates {
		if IsState(s) {
			t.Errorf("IsState(%q) = true", s)
		}
	}
}
