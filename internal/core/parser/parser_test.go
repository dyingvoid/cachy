package core_parser

import (
	"errors"
	"testing"

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
