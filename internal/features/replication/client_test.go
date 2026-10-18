package features_replication

import (
	"context"
	"errors"
	"io"
	"reflect"
	"testing"
)

type fakeCodec struct {
	encoded   []Message
	responses []Response
	err       error
}

func (f *fakeCodec) Encode(msg Message) ([]byte, error) {
	f.encoded = append(f.encoded, msg)
	return []byte(msg.Command()), nil
}

func (f *fakeCodec) Decode([]byte) (Response, error) {
	if f.err != nil {
		return nil, f.err
	}
	if len(f.responses) == 0 {
		return nil, errors.New("no responses")
	}

	resp := f.responses[0]
	f.responses = f.responses[1:]
	return resp, nil
}

type fakeTransport struct {
	port int
	raw  []byte
	err  error
}

func (f *fakeTransport) Read(p []byte) (int, error) {
	if f.err != nil {
		return 0, f.err
	}
	if len(f.raw) == 0 {
		return 0, io.EOF
	}

	n := copy(p, f.raw)
	f.raw = f.raw[n:]

	return n, nil
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

func (f *fakeTransport) Close() error {
	return nil
}

func TestReplicationClientHandshake(t *testing.T) {
	codec := &fakeCodec{
		responses: []Response{
			PongResponse{},
			OKResponse{},
			OKResponse{},
			PsyncResponse{ReplID: "abc123", Offset: 0},
		},
	}
	transport := &fakeTransport{
		port: 8082,
		raw:  []byte("PONG\nOK\nOK\nFULLRESYNC abc123 0\n$0\r\n\n"),
	}
	client := NewReplicationClient(codec, transport)

	if err := client.Handshake(context.Background()); err != nil {
		t.Fatalf("Handshake returned unexpected error: %v", err)
	}

	port := 8082
	capa := "psync2"
	want := []Message{
		PingMessage{},
		ReplConfMessage{Port: &port},
		ReplConfMessage{Capa: &capa},
		PsyncMessage{ReplicationID: "?", Offset: -1},
	}
	if !reflect.DeepEqual(codec.encoded, want) {
		t.Fatalf("Handshake encoded %+v, want %+v", codec.encoded, want)
	}
}

func TestReplicationClientHandshakeWithFile(t *testing.T) {
	codec := &fakeCodec{
		responses: []Response{
			PongResponse{},
			OKResponse{},
			OKResponse{},
			PsyncResponse{ReplID: "abc123", Offset: 0},
		},
	}
	transport := &fakeTransport{
		port: 8082,
		raw:  []byte("PONG\nOK\nOK\nFULLRESYNC abc123 0\n$3\r\nabc\n"),
	}
	client := NewReplicationClient(codec, transport)

	if err := client.Handshake(context.Background()); err != nil {
		t.Fatalf("Handshake returned unexpected error: %v", err)
	}
}

func TestReplicationClientHandshakeErrors(t *testing.T) {
	tests := []struct {
		name      string
		responses []Response
		transport *fakeTransport
		codecErr  error
	}{
		{
			name:      "unexpected ping response",
			responses: []Response{OKResponse{}},
			transport: &fakeTransport{port: 8082, raw: []byte("OK\n")},
		},
		{
			name:      "unexpected replconf response",
			responses: []Response{PongResponse{}, PongResponse{}},
			transport: &fakeTransport{port: 8082, raw: []byte("PONG\nPONG\n")},
		},
		{
			name:      "unexpected psync response",
			responses: []Response{PongResponse{}, OKResponse{}, OKResponse{}, OKResponse{}},
			transport: &fakeTransport{port: 8082, raw: []byte("PONG\nOK\nOK\nOK\n")},
		},
		{
			name:      "unexpected file header",
			responses: []Response{PongResponse{}, OKResponse{}, OKResponse{}, PsyncResponse{ReplID: "abc123", Offset: 0}},
			transport: &fakeTransport{port: 8082, raw: []byte("PONG\nOK\nOK\nFULLRESYNC abc123 0\nOK\n")},
		},
		{
			name:      "invalid file length",
			responses: []Response{PongResponse{}, OKResponse{}, OKResponse{}, PsyncResponse{ReplID: "abc123", Offset: 0}},
			transport: &fakeTransport{port: 8082, raw: []byte("PONG\nOK\nOK\nFULLRESYNC abc123 0\n$nope\r\n")},
		},
		{
			name:      "transport error",
			transport: &fakeTransport{port: 8082, err: errors.New("transport failed")},
		},
		{
			name:      "decode error",
			responses: []Response{PongResponse{}},
			transport: &fakeTransport{port: 8082, raw: []byte("PONG\n")},
			codecErr:  errors.New("decode failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codec := &fakeCodec{
				responses: tt.responses,
				err:       tt.codecErr,
			}
			client := NewReplicationClient(codec, tt.transport)

			if err := client.Handshake(context.Background()); err == nil {
				t.Fatal("Handshake returned nil error")
			}
		})
	}
}
