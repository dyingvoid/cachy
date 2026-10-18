package core_replication

import (
	"context"
	"errors"
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

func (f *fakeTransport) Send(context.Context, []byte) ([]byte, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.raw, nil
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
	transport := &fakeTransport{port: 8082}
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
			transport: &fakeTransport{port: 8082},
		},
		{
			name:      "unexpected replconf response",
			responses: []Response{PongResponse{}, PongResponse{}},
			transport: &fakeTransport{port: 8082},
		},
		{
			name:      "unexpected psync response",
			responses: []Response{PongResponse{}, OKResponse{}, OKResponse{}, OKResponse{}},
			transport: &fakeTransport{port: 8082},
		},
		{
			name:      "transport error",
			transport: &fakeTransport{port: 8082, err: errors.New("transport failed")},
		},
		{
			name:      "decode error",
			responses: []Response{PongResponse{}},
			transport: &fakeTransport{port: 8082},
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
