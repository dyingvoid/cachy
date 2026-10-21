package resp3

import (
	"errors"
	"io"
	"strings"
	"testing"
)

func TestDeserializer_ReadSimpleString(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
		err  error
	}{
		{name: "simple string", in: "+PONG\r\n", want: "PONG"},
		{name: "unexpected kind", in: "$4\r\nPONG\r\n", err: errUnexpectedKind},
		{name: "truncated", in: "+PONG", err: io.ErrUnexpectedEOF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewDeserializer(strings.NewReader(tt.in)).ReadSimpleString()
			if tt.err != nil {
				if !errors.Is(err, tt.err) {
					t.Fatalf("ReadSimpleString() error = %v, want %v", err, tt.err)
				}
				return
			}
			if err != nil {
				t.Fatalf("ReadSimpleString() returned unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("ReadSimpleString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDeserializer_DiscardBulkString(t *testing.T) {
	tests := []struct {
		name string
		in   string
		err  error
	}{
		{name: "bulk string", in: "$3\r\nabc\r\n"},
		{name: "empty bulk string", in: "$0\r\n\r\n"},
		{name: "unexpected kind", in: "+OK\r\n", err: errUnexpectedKind},
		{name: "truncated", in: "$3\r\nab", err: io.ErrUnexpectedEOF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewDeserializer(strings.NewReader(tt.in)).DiscardBulkString()
			if tt.err != nil {
				if !errors.Is(err, tt.err) {
					t.Fatalf("DiscardBulkString() error = %v, want %v", err, tt.err)
				}
				return
			}
			if err != nil {
				t.Fatalf("DiscardBulkString() returned unexpected error: %v", err)
			}
		})
	}
}
