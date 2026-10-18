package features

import (
	"errors"
	"reflect"
	"testing"

	core_state "github.com/dyingvoid/cachy/internal/core/state"
	features_conn "github.com/dyingvoid/cachy/internal/features/conn"
	features_replication "github.com/dyingvoid/cachy/internal/features/replication"
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
	port := 8081
	capa := "psync2"
	state := &features_conn.State{
		Replication: features_replication.State{
			ListeningPort: &port,
			Capa:          &capa,
		},
	}

	res, err := h.executePsync(state, PsyncCmd{ReplicationID: "?", Offset: -1})
	if err != nil {
		t.Fatalf("executePsync returned unexpected error: %v", err)
	}

	want := []Result{
		StringResult("FULLRESYNC " + core_state.ReplicationID() + " 0"),
		FileResult{Len: 0, Content: []byte{}},
		StreamingResult{},
	}
	if !reflect.DeepEqual(res, want) {
		t.Fatalf("executePsync = %#v, want %#v", res, want)
	}
	if state.Mode != features_conn.ModeReplicationStreaming {
		t.Fatalf("state mode = %v, want %v", state.Mode, features_conn.ModeReplicationStreaming)
	}
}
