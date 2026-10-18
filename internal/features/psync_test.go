package features

import (
	"errors"
	"reflect"
	"testing"

	core_state "github.com/dyingvoid/cachy/internal/core/state"
)

func TestParser_Parse_Psync(t *testing.T) {
	p := NewProtocolParser()

	cmd, err := p.Parse("psync ? -1")
	if err != nil {
		t.Fatalf("Parse(%q) returned unexpected error: %v", "psync ? -1", err)
	}

	want := PsyncCmd{ReplicationID: "?", Offset: -1}
	if !reflect.DeepEqual(cmd, want) {
		t.Fatalf("Parse(%q) = %+v, want %+v", "psync ? -1", cmd, want)
	}
}

func TestParser_Parse_PsyncErrors(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{name: "no arguments", in: "PSYNC"},
		{name: "missing offset", in: "PSYNC ?"},
		{name: "too many arguments", in: "PSYNC ? -1 extra"},
		{name: "non numeric offset", in: "PSYNC ? abc"},
	}

	p := NewProtocolParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := p.Parse(tt.in)
			if !errors.Is(err, ErrInvalidArguments) {
				t.Fatalf("Parse(%q) error = %v, want %v", tt.in, err, ErrInvalidArguments)
			}
		})
	}
}

func TestExecutePsync(t *testing.T) {
	h := &ProtocolHandler{}

	res, err := h.executePsync(PsyncCmd{ReplicationID: "?", Offset: -1})
	if err != nil {
		t.Fatalf("executePsync returned unexpected error: %v", err)
	}

	want := StringResult("+FULLRESYNC " + core_state.ReplicationID() + " 0")
	if res != want {
		t.Fatalf("executePsync = %q, want %q", res, want)
	}
}
