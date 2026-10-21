package features_replication

import (
	"context"
	"errors"
	"reflect"
	"testing"

	features_commands "github.com/dyingvoid/cachy/internal/features/commands"
)

type fakeTransport struct {
	port int
	err  error
}

func (f *fakeTransport) Read([]byte) (int, error) {
	return 0, errors.New("unexpected read")
}

func (f *fakeTransport) Write(p []byte) (int, error) {
	if f.err != nil {
		return 0, f.err
	}

	return len(p), nil
}

func (f *fakeTransport) Port() int {
	return f.port
}

type fakeSerializer struct {
	commands []features_commands.Command
}

func (f *fakeSerializer) SerializeCommand(cmd features_commands.Command) []byte {
	f.commands = append(f.commands, cmd)
	return []byte("command")
}

type fakeDeserializer struct {
	replies     []string
	snapshotErr error
	readErr     error
}

func (f *fakeDeserializer) ReadSimpleString() (string, error) {
	if f.readErr != nil {
		return "", f.readErr
	}
	if len(f.replies) == 0 {
		return "", errors.New("no replies left")
	}

	reply := f.replies[0]
	f.replies = f.replies[1:]
	return reply, nil
}

func (f *fakeDeserializer) DiscardBulkString() error {
	return f.snapshotErr
}

func TestReplicationClientHandshake(t *testing.T) {
	serializer := &fakeSerializer{}
	deserializer := &fakeDeserializer{
		replies: []string{"PONG", "OK", "OK", "FULLRESYNC abc123 0"},
	}
	client := NewReplicationClient(&fakeTransport{port: 8082}, serializer, deserializer)

	if err := client.Handshake(context.Background()); err != nil {
		t.Fatalf("Handshake returned unexpected error: %v", err)
	}

	port := 8082
	capa := "psync2"
	want := []features_commands.Command{
		features_commands.PingCmd{},
		features_commands.ReplConfCmd{ListeningPort: &port},
		features_commands.ReplConfCmd{Capa: &capa},
		features_commands.PsyncCmd{ReplicationID: "?", Offset: -1},
	}
	if !reflect.DeepEqual(serializer.commands, want) {
		t.Fatalf("Handshake sent %#v, want %#v", serializer.commands, want)
	}
}

func TestReplicationClientHandshakeErrors(t *testing.T) {
	tests := []struct {
		name         string
		replies      []string
		readErr      error
		snapshotErr  error
		transportErr error
	}{
		{
			name:    "unexpected ping reply",
			replies: []string{"OK"},
		},
		{
			name:    "unexpected replconf reply",
			replies: []string{"PONG", "PONG"},
		},
		{
			name:    "unexpected psync reply",
			replies: []string{"PONG", "OK", "OK", "OK"},
		},
		{
			name:    "read error",
			replies: []string{"PONG"},
			readErr: errors.New("read failed"),
		},
		{
			name:        "snapshot error",
			replies:     []string{"PONG", "OK", "OK", "FULLRESYNC abc123 0"},
			snapshotErr: errors.New("snapshot failed"),
		},
		{
			name:         "transport error",
			transportErr: errors.New("transport failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewReplicationClient(
				&fakeTransport{port: 8082, err: tt.transportErr},
				&fakeSerializer{},
				&fakeDeserializer{
					replies:     tt.replies,
					readErr:     tt.readErr,
					snapshotErr: tt.snapshotErr,
				},
			)

			if err := client.Handshake(context.Background()); err == nil {
				t.Fatal("Handshake returned nil error")
			}
		})
	}
}

func TestReplicationClientHandshakeCanceledContext(t *testing.T) {
	client := NewReplicationClient(
		&fakeTransport{port: 8082},
		&fakeSerializer{},
		&fakeDeserializer{},
	)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := client.Handshake(ctx); err == nil {
		t.Fatal("Handshake returned nil error")
	}
}
