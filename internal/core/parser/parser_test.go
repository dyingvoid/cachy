package core_parser

import (
	"errors"
	"testing"

	"github.com/dyingvoid/cachy/internal/domain"
	domain_errors "github.com/dyingvoid/cachy/internal/domain/errors"
)

func TestParser_Parse_Get(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{
			name: "simple get",
			in:   "GET mykey",
		},
		{
			name: "get example value",
			in:   "GET value",
		},
		{
			name: "lowercase command",
			in:   "get mykey",
		},
	}

	p := NewParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := p.Parse(tt.in); err != nil {
				t.Fatalf("Parse(%q) returned unexpected error: %v", tt.in, err)
			}
		})
	}
}

func TestParser_Parse_Set(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{
			name: "set without expiry",
			in:   "SET mykey value",
		},
		{
			name: "set with EX seconds",
			in:   "SET mykey value EX 10",
		},
		{
			name: "set with PX milliseconds",
			in:   "SET mykey value PX 1000",
		},
		{
			name: "lowercase command and expiry flag",
			in:   "set mykey value ex 10",
		},
	}

	p := NewParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := p.Parse(tt.in); err != nil {
				t.Fatalf("Parse(%q) returned unexpected error: %v", tt.in, err)
			}
		})
	}
}

func TestParser_Parse_LRange(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want domain.LRangeCmd
	}{
		{
			name: "range from zero",
			in:   "LRANGE list_key 0 1",
			want: domain.LRangeCmd{Key: "list_key", Start: 0, Stop: 1},
		},
		{
			name: "range with equal indexes",
			in:   "LRANGE list_key 2 2",
			want: domain.LRangeCmd{Key: "list_key", Start: 2, Stop: 2},
		},
		{
			name: "lowercase command",
			in:   "lrange list_key 2 4",
			want: domain.LRangeCmd{Key: "list_key", Start: 2, Stop: 4},
		},
		{
			name: "negative start and stop",
			in:   "LRANGE list_key -2 -1",
			want: domain.LRangeCmd{Key: "list_key", Start: -2, Stop: -1},
		},
		{
			name: "negative stop",
			in:   "LRANGE list_key 0 -3",
			want: domain.LRangeCmd{Key: "list_key", Start: 0, Stop: -3},
		},
		{
			name: "negative start",
			in:   "LRANGE list_key -1 4",
			want: domain.LRangeCmd{Key: "list_key", Start: -1, Stop: 4},
		},
	}

	p := NewParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := p.Parse(tt.in)
			if err != nil {
				t.Fatalf("Parse(%q) returned unexpected error: %v", tt.in, err)
			}
			if cmd != tt.want {
				t.Fatalf("Parse(%q) = %+v, want %+v", tt.in, cmd, tt.want)
			}
		})
	}
}

func TestParser_Parse_Errors(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want error
	}{
		{
			name: "empty input",
			in:   "   ",
			want: domain_errors.ErrEmptyCommand,
		},
		{
			name: "unknown command",
			in:   "DEL mykey",
			want: domain_errors.ErrUnknownCommand,
		},
		{
			name: "get with no key",
			in:   "GET",
			want: domain_errors.ErrInvalidArguments,
		},
		{
			name: "get with too many args",
			in:   "GET mykey extra",
			want: domain_errors.ErrInvalidArguments,
		},
		{
			name: "set with no value",
			in:   "SET mykey",
			want: domain_errors.ErrInvalidArguments,
		},
		{
			name: "set with unknown expiry flag",
			in:   "SET mykey value XX 10",
			want: domain_errors.ErrInvalidArguments,
		},
		{
			name: "set with non numeric expiry",
			in:   "SET mykey value EX abc",
			want: domain_errors.ErrInvalidArguments,
		},
		{
			name: "set with non positive expiry",
			in:   "SET mykey value EX 0",
			want: domain_errors.ErrInvalidArguments,
		},
		{
			name: "lrange with no indexes",
			in:   "LRANGE mykey",
			want: domain_errors.ErrInvalidArguments,
		},
		{
			name: "lrange with one index",
			in:   "LRANGE mykey 0",
			want: domain_errors.ErrInvalidArguments,
		},
		{
			name: "lrange with too many args",
			in:   "LRANGE mykey 0 1 extra",
			want: domain_errors.ErrInvalidArguments,
		},
		{
			name: "lrange with non numeric start",
			in:   "LRANGE mykey a 1",
			want: domain_errors.ErrInvalidArguments,
		},
		{
			name: "lrange with non numeric stop",
			in:   "LRANGE mykey 0 b",
			want: domain_errors.ErrInvalidArguments,
		},
	}

	p := NewParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := p.Parse(tt.in)
			if !errors.Is(err, tt.want) {
				t.Fatalf("Parse(%q) error = %v, want %v", tt.in, err, tt.want)
			}
		})
	}
}
