package program_test

import (
	"testing"

	"manage2/internal/program"
)

func TestNewSet(t *testing.T) {
	s := program.NewSet[int]()
	if len(s) != 0 {
		t.Errorf("expected empty set, got: %d", len(s))
	}
}

func TestSetAdd(t *testing.T) {
	s := program.NewSet[int]()
	if s.Contains(1) {
		t.Error("expected set to not contain index 1")
	}
	s.Add(1)
	if len(s) != 1 {
		t.Errorf("expected set to contain 1 element, got: %d", len(s))
	}
	if !s.Contains(1) {
		t.Error("expected set to contain index 1")
	}
}
