package features

import (
	"errors"
	"reflect"
	"testing"

	core_state "github.com/dyingvoid/cachy/internal/core/state"
	features_commands "github.com/dyingvoid/cachy/internal/features/commands"
	features_conn "github.com/dyingvoid/cachy/internal/features/conn"
	features_replication "github.com/dyingvoid/cachy/internal/features/replication"
)

func TestExecutePsync(t *testing.T) {
	h := &ProtocolHandler{replication: NewReplicationHandler(100)}
	port := 8081
	capa := "psync2"
	state := &features_conn.State{
		Replication: features_replication.State{
			ListeningPort: &port,
			Capa:          &capa,
		},
	}

	res, err := h.executePsync(state, features_commands.PsyncCmd{ReplicationID: "?", Offset: -1})
	if err != nil {
		t.Fatalf("executePsync returned unexpected error: %v", err)
	}
	if len(res) != 3 {
		t.Fatalf("executePsync returned %d results, want 3", len(res))
	}

	want := []Result{
		StringResult("FULLRESYNC " + core_state.ReplicationID() + " 0"),
		FileResult{Len: 0, Content: []byte{}},
	}
	if !reflect.DeepEqual(res[:2], want) {
		t.Fatalf("executePsync = %#v, want %#v", res[:2], want)
	}

	streaming, ok := res[2].(StreamingResult)
	if !ok {
		t.Fatalf("executePsync third result = %T, want StreamingResult", res[2])
	}
	if streaming.Stream == nil {
		t.Fatal("executePsync returned StreamingResult with nil stream")
	}
	defer streaming.Stream.Close()

	if state.Mode != features_conn.ModeReplicationStreaming {
		t.Fatalf("state mode = %v, want %v", state.Mode, features_conn.ModeReplicationStreaming)
	}

	if _, err := h.executePsync(state, features_commands.PsyncCmd{ReplicationID: "?", Offset: -1}); !errors.Is(err, ErrReplicationStateIncomplete) {
		t.Fatalf("second executePsync error = %v, want %v", err, ErrReplicationStateIncomplete)
	}
}
