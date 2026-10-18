package infrastructure_storage

import (
	"errors"
	"testing"

	"github.com/dyingvoid/cachy/internal/domain"
)

func TestMemoryStorage_StringKeyRejectsListCommands(t *testing.T) {
	s := newTestStorage()

	if err := s.Set("key", domain.NewStringRecord("value", nil)); err != nil {
		t.Fatalf("Set returned unexpected error: %v", err)
	}

	tests := []struct {
		name string
		call func() error
	}{
		{
			name: "RPush",
			call: func() error {
				_, err := s.RPush("key", "a")
				return err
			},
		},
		{
			name: "LPush",
			call: func() error {
				_, err := s.LPush("key", "a")
				return err
			},
		},
		{
			name: "LRange",
			call: func() error {
				_, err := s.LRange("key", 0, -1)
				return err
			},
		},
		{
			name: "LLen",
			call: func() error {
				_, err := s.LLen("key")
				return err
			},
		},
		{
			name: "LPop",
			call: func() error {
				_, err := s.LPop("key", 1)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(); !errors.Is(err, domain.ErrWrongType) {
				t.Fatalf("%s returned error %v, want %v", tt.name, err, domain.ErrWrongType)
			}
		})
	}
}

func TestMemoryStorage_ListKeyRejectsStringCommands(t *testing.T) {
	s := newTestStorage()

	if _, err := s.RPush("key", "a"); err != nil {
		t.Fatalf("RPush returned unexpected error: %v", err)
	}

	if err := s.Set("key", domain.NewStringRecord("value", nil)); !errors.Is(err, domain.ErrWrongType) {
		t.Fatalf("Set returned error %v, want %v", err, domain.ErrWrongType)
	}

	if _, err := s.Get("key"); !errors.Is(err, domain.ErrWrongType) {
		t.Fatalf("Get returned error %v, want %v", err, domain.ErrWrongType)
	}

	if err := s.Delete("key"); !errors.Is(err, domain.ErrWrongType) {
		t.Fatalf("Delete returned error %v, want %v", err, domain.ErrWrongType)
	}
}
