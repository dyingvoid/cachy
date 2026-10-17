package core_replication

import "context"

type Codec interface {
	Encode(msg Message) ([]byte, error)
	Decode([]byte) (string, error)
	DecodePsync([]byte) (PsyncResponse, error)
}

type ClientTransport interface {
	Send(ctx context.Context, payload []byte) ([]byte, error)
	Port() int
	Close() error
}
