package model

import "testing"

func TestNewID(t *testing.T) {
	id := NewID()
	if len(id) != 8 {
		t.Errorf("expected 8-char ID, got %d chars: %q", len(id), id)
	}

	// IDs should be unique
	id2 := NewID()
	if id == id2 {
		t.Errorf("expected unique IDs, got same: %q", id)
	}
}
