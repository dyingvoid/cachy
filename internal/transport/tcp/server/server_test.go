package transport_tcp_server

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

	"github.com/dyingvoid/cachy/internal/features"
	features_commands "github.com/dyingvoid/cachy/internal/features/commands"
	features_conn "github.com/dyingvoid/cachy/internal/features/conn"
)

type eofDeserializer struct{}

func (eofDeserializer) ReadCommand() (features_commands.Command, error) {
	return nil, io.EOF
}

type noopHandler struct{}

func (noopHandler) Handle(
	context.Context, *features_conn.State, features_commands.Command,
) ([]features.Result, error) {
	return nil, nil
}

func TestHandle_ClientClose(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()

	s := &Server{
		handler: noopHandler{},
		newDeserializer: func(io.Reader) Deserializer {
			return eofDeserializer{}
		},
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		s.handle(context.Background(), serverConn)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("handle did not return on io.EOF")
	}
}
