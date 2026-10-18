package features_replication

import (
	"errors"
	"reflect"
	"testing"
)

type testMessage struct {
	command string
	args    []string
}

func (m testMessage) Command() string {
	return m.command
}

func (m testMessage) Args() []string {
	return m.args
}

func intPtr(value int) *int {
	return &value
}

func stringPtr(value string) *string {
	return &value
}

func TestTextCodecEncode(t *testing.T) {
	codec := NewTextCodec()

	tests := []struct {
		name string
		msg  Message
		want []byte
	}{
		{
			name: "ping",
			msg:  PingMessage{},
			want: []byte("PING\n"),
		},
		{
			name: "replconf listening port",
			msg:  ReplConfMessage{Port: intPtr(8081)},
			want: []byte("REPLCONF listening-port 8081\n"),
		},
		{
			name: "replconf capa",
			msg:  ReplConfMessage{Capa: stringPtr("psync2")},
			want: []byte("REPLCONF capa psync2\n"),
		},
		{
			name: "replconf port and capa",
			msg: ReplConfMessage{
				Port: intPtr(8081),
				Capa: stringPtr("psync2"),
			},
			want: []byte("REPLCONF listening-port 8081 capa psync2\n"),
		},
		{
			name: "psync",
			msg:  PsyncMessage{ReplicationID: "?", Offset: -1},
			want: []byte("PSYNC ? -1\n"),
		},
		{
			name: "trims command",
			msg:  testMessage{command: "  PSYNC  ", args: []string{"?", "-1"}},
			want: []byte("PSYNC ? -1\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := codec.Encode(tt.msg)
			if err != nil {
				t.Fatalf("Encode returned unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Encode = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTextCodecEncodeErrors(t *testing.T) {
	codec := NewTextCodec()

	tests := []struct {
		name string
		msg  Message
		want error
	}{
		{
			name: "nil message",
			msg:  nil,
			want: ErrEmptyMessage,
		},
		{
			name: "empty command",
			msg:  testMessage{},
			want: ErrEmptyMessage,
		},
		{
			name: "whitespace command",
			msg:  testMessage{command: "   "},
			want: ErrEmptyMessage,
		},
		{
			name: "argument with whitespace",
			msg:  testMessage{command: "SET", args: []string{"key", "two words"}},
			want: ErrInvalidMessage,
		},
		{
			name: "empty argument",
			msg:  testMessage{command: "SET", args: []string{"key", ""}},
			want: ErrInvalidMessage,
		},
		{
			name: "argument with newline",
			msg:  testMessage{command: "SET", args: []string{"key", "value\nPING"}},
			want: ErrInvalidMessage,
		},
		{
			name: "empty replication id",
			msg:  PsyncMessage{ReplicationID: "", Offset: 0},
			want: ErrInvalidMessage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := codec.Encode(tt.msg)
			if !errors.Is(err, tt.want) {
				t.Fatalf("Encode error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestTextCodecDecode(t *testing.T) {
	codec := NewTextCodec()

	tests := []struct {
		name string
		raw  []byte
		want Response
	}{
		{
			name: "plain pong",
			raw:  []byte("PONG\r\n"),
			want: PongResponse{},
		},
		{
			name: "plain ok",
			raw:  []byte("OK\r\n"),
			want: OKResponse{},
		},
		{
			name: "full resync",
			raw:  []byte("FULLRESYNC abc123 0\r\n"),
			want: PsyncResponse{ReplID: "abc123", Offset: 0},
		},
		{
			name: "full resync with negative offset",
			raw:  []byte("FULLRESYNC abc123 -1\r\n"),
			want: PsyncResponse{ReplID: "abc123", Offset: -1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := codec.Decode(tt.raw)
			if err != nil {
				t.Fatalf("Decode returned unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Decode = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestTextCodecDecodeErrors(t *testing.T) {
	codec := NewTextCodec()

	tests := []struct {
		name string
		raw  []byte
		want error
	}{
		{
			name: "whitespace response",
			raw:  []byte(" \r\n"),
			want: ErrEmptyResponse,
		},
		{
			name: "server error",
			raw:  []byte("(error) ERR invalid command\r\n"),
			want: ErrServerError,
		},
		{
			name: "resp simple string",
			raw:  []byte("+OK\r\n"),
			want: ErrUnexpectedResponse,
		},
		{
			name: "resp error",
			raw:  []byte("-ERR invalid command\r\n"),
			want: ErrUnexpectedResponse,
		},
		{
			name: "unknown response",
			raw:  []byte("CONTINUE\r\n"),
			want: ErrUnexpectedResponse,
		},
		{
			name: "psync missing offset",
			raw:  []byte("FULLRESYNC abc123\r\n"),
			want: ErrUnexpectedResponse,
		},
		{
			name: "psync too many fields",
			raw:  []byte("FULLRESYNC abc123 0 extra\r\n"),
			want: ErrUnexpectedResponse,
		},
		{
			name: "psync non numeric offset",
			raw:  []byte("FULLRESYNC abc123 nope\r\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := codec.Decode(tt.raw)
			if err == nil {
				t.Fatal("Decode returned nil error")
			}
			if tt.want != nil && !errors.Is(err, tt.want) {
				t.Fatalf("Decode error = %v, want %v", err, tt.want)
			}
		})
	}
}
