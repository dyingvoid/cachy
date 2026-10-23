package features_replication

import (
	"reflect"
	"testing"
	"time"

	features_commands "github.com/dyingvoid/cachy/internal/features/commands"
)

func TestReplicationHandlerPublish(t *testing.T) {
	h := NewReplicationHandler(Config{ReplicationBuffer: 100})
	first := h.Subscribe()
	second := h.Subscribe()
	defer first.Close()
	defer second.Close()

	cmd := features_commands.SetCmd{Key: "key", Value: "value"}
	h.Publish(cmd)

	for i, stream := range []*ReplicationStream{first, second} {
		select {
		case got := <-stream.Chan():
			if !reflect.DeepEqual(got, cmd) {
				t.Fatalf("stream %d command = %#v, want %#v", i, got, cmd)
			}
		default:
			t.Fatalf("stream %d received no command", i)
		}
	}
}

func TestReplicationStreamClose(t *testing.T) {
	h := NewReplicationHandler(Config{ReplicationBuffer: 100})
	stream := h.Subscribe()

	stream.Close()

	if _, ok := <-stream.Chan(); ok {
		t.Fatal("commands channel is open after Close")
	}

	h.Publish(features_commands.SetCmd{Key: "key", Value: "value"})
	stream.Close()
}

func TestReplicationStreamCloseUnblocksPublish(t *testing.T) {
	h := NewReplicationHandler(Config{ReplicationBuffer: 1})
	stream := h.Subscribe()
	h.Publish(features_commands.SetCmd{Key: "first", Value: "value"})

	published := make(chan struct{})
	go func() {
		h.Publish(features_commands.SetCmd{Key: "second", Value: "value"})
		close(published)
	}()

	stream.Close()

	select {
	case <-published:
	case <-time.After(time.Second):
		t.Fatal("Publish did not return after Close")
	}
}
