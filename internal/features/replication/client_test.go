package features_replication

import (
	"context"
	"errors"
	"io"
	"reflect"
	"testing"

	"github.com/dyingvoid/cachy/internal/features"
	features_commands "github.com/dyingvoid/cachy/internal/features/commands"
	features_conn "github.com/dyingvoid/cachy/internal/features/conn"
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
	commands    []features_commands.Command
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

func (f *fakeDeserializer) ReadCommand() (features_commands.Command, error) {
	if len(f.commands) == 0 {
		return nil, io.EOF
	}

	cmd := f.commands[0]
	f.commands = f.commands[1:]
	return cmd, nil
}

func (f *fakeDeserializer) DiscardBulkString() error {
	return f.snapshotErr
}

type fakeHandler struct {
	commands []features_commands.Command
	err      error
}

func (f *fakeHandler) Handle(
	_ context.Context,
	_ *features_conn.State,
	cmd features_commands.Command,
) ([]features.Result, error) {
	f.commands = append(f.commands, cmd)
	if f.err != nil {
		return nil, f.err
	}

	return nil, nil
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

func TestReplicationClientListen(t *testing.T) {
	commands := []features_commands.Command{
		features_commands.SetCmd{Key: "key", Value: "value"},
		features_commands.RPushCmd{Key: "list", Values: []string{"a", "b"}},
	}
	handler := &fakeHandler{}
	client := NewReplicationClient(
		&fakeTransport{port: 8082},
		&fakeSerializer{},
		&fakeDeserializer{commands: commands},
	)

	if err := client.Listen(context.Background(), handler); !errors.Is(err, io.EOF) {
		t.Fatalf("Listen error = %v, want %v", err, io.EOF)
	}
	if !reflect.DeepEqual(handler.commands, commands) {
		t.Fatalf("Listen applied %#v, want %#v", handler.commands, commands)
	}
}

func TestReplicationClientListenHandlerError(t *testing.T) {
	handlerErr := errors.New("apply failed")
	client := NewReplicationClient(
		&fakeTransport{port: 8082},
		&fakeSerializer{},
		&fakeDeserializer{
			commands: []features_commands.Command{
				features_commands.SetCmd{Key: "key", Value: "value"},
			},
		},
	)

	err := client.Listen(context.Background(), &fakeHandler{err: handlerErr})
	if !errors.Is(err, handlerErr) {
		t.Fatalf("Listen error = %v, want %v", err, handlerErr)
	}
}

func TestReplicationClientListenCanceledContext(t *testing.T) {
	client := NewReplicationClient(
		&fakeTransport{port: 8082},
		&fakeSerializer{},
		&fakeDeserializer{},
	)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := client.Listen(ctx, &fakeHandler{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("Listen error = %v, want %v", err, context.Canceled)
	}
}
