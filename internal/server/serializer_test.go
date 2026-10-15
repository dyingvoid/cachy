package server

import (
	"testing"

	domain_errors "github.com/dyingvoid/cachy/internal/domain/errors"
)

func TestTextSerializer_SerializeError(t *testing.T) {
	s := NewTextSerializer()

	got := s.SerializeError(domain_errors.ErrWrongType)
	want := "(error) WRONGTYPE Operation against a key holding the wrong kind of value"
	if got != want {
		t.Fatalf("SerializeError = %q, want %q", got, want)
	}
}
