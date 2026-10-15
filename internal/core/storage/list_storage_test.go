package core_storage

import (
	"reflect"
	"testing"
)

func newTestListStorage() *ListStorage {
	return NewListStorage(Config{ShardCount: 4})
}

func TestListStorage_LPush(t *testing.T) {
	s := newTestListStorage()

	n, err := s.LPush("list_key", "a", "b", "c")
	if err != nil {
		t.Fatalf("LPush returned unexpected error: %v", err)
	}
	if n != 3 {
		t.Fatalf("LPush returned %d, want 3", n)
	}

	got, err := s.LRange("list_key", 0, -1)
	if err != nil {
		t.Fatalf("LRange returned unexpected error: %v", err)
	}
	if want := []string{"c", "b", "a"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("LRange after LPush = %v, want %v", got, want)
	}

	n, err = s.LPush("list_key", "d", "e")
	if err != nil {
		t.Fatalf("LPush returned unexpected error: %v", err)
	}
	if n != 5 {
		t.Fatalf("LPush returned %d, want 5", n)
	}

	got, err = s.LRange("list_key", 0, -1)
	if err != nil {
		t.Fatalf("LRange returned unexpected error: %v", err)
	}
	if want := []string{"e", "d", "c", "b", "a"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("LRange after second LPush = %v, want %v", got, want)
	}
}

func TestListStorage_LLen(t *testing.T) {
	s := newTestListStorage()

	n, err := s.LLen("missing")
	if err != nil {
		t.Fatalf("LLen returned unexpected error: %v", err)
	}
	if n != 0 {
		t.Fatalf("LLen on missing key = %d, want 0", n)
	}

	if _, err := s.RPush("list_key", "a", "b", "c", "d"); err != nil {
		t.Fatalf("RPush returned unexpected error: %v", err)
	}

	n, err = s.LLen("list_key")
	if err != nil {
		t.Fatalf("LLen returned unexpected error: %v", err)
	}
	if n != 4 {
		t.Fatalf("LLen = %d, want 4", n)
	}

	if _, err := s.LPush("list_key", "e"); err != nil {
		t.Fatalf("LPush returned unexpected error: %v", err)
	}

	n, err = s.LLen("list_key")
	if err != nil {
		t.Fatalf("LLen returned unexpected error: %v", err)
	}
	if n != 5 {
		t.Fatalf("LLen after LPush = %d, want 5", n)
	}
}

func TestListStorage_LRange(t *testing.T) {
	s := newTestListStorage()
	if _, err := s.RPush("list_key", "a", "b", "c", "d", "e"); err != nil {
		t.Fatalf("RPush returned unexpected error: %v", err)
	}

	tests := []struct {
		name  string
		key   string
		start int
		stop  int
		want  []string
	}{
		{
			name:  "first two",
			key:   "list_key",
			start: 0,
			stop:  1,
			want:  []string{"a", "b"},
		},
		{
			name:  "middle",
			key:   "list_key",
			start: 2,
			stop:  4,
			want:  []string{"c", "d", "e"},
		},
		{
			name:  "single element",
			key:   "list_key",
			start: 3,
			stop:  3,
			want:  []string{"d"},
		},
		{
			name:  "stop past the end",
			key:   "list_key",
			start: 3,
			stop:  100,
			want:  []string{"d", "e"},
		},
		{
			name:  "start at length",
			key:   "list_key",
			start: 5,
			stop:  6,
			want:  []string{},
		},
		{
			name:  "start past the end",
			key:   "list_key",
			start: 100,
			stop:  200,
			want:  []string{},
		},
		{
			name:  "start greater than stop",
			key:   "list_key",
			start: 3,
			stop:  1,
			want:  []string{},
		},
		{
			name:  "negative start and stop",
			key:   "list_key",
			start: -2,
			stop:  -1,
			want:  []string{"d", "e"},
		},
		{
			name:  "whole list with negative stop",
			key:   "list_key",
			start: 0,
			stop:  -1,
			want:  []string{"a", "b", "c", "d", "e"},
		},
		{
			name:  "all except last two",
			key:   "list_key",
			start: 0,
			stop:  -3,
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "negative start",
			key:   "list_key",
			start: -3,
			stop:  4,
			want:  []string{"c", "d", "e"},
		},
		{
			name:  "negative start out of range",
			key:   "list_key",
			start: -6,
			stop:  2,
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "negative stop out of range",
			key:   "list_key",
			start: 0,
			stop:  -6,
			want:  []string{"a"},
		},
		{
			name:  "both negative out of range",
			key:   "list_key",
			start: -6,
			stop:  -6,
			want:  []string{"a"},
		},
		{
			name:  "negative indexes on missing key",
			key:   "missing",
			start: -2,
			stop:  -1,
			want:  []string{},
		},
		{
			name:  "missing key",
			key:   "missing",
			start: 0,
			stop:  1,
			want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.LRange(tt.key, tt.start, tt.stop)
			if err != nil {
				t.Fatalf("LRange returned unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("LRange(%q, %d, %d) = %v, want %v", tt.key, tt.start, tt.stop, got, tt.want)
			}
		})
	}
}
