package features_replication

import "io"

type Codec interface {
	Encode(msg Message) ([]byte, error)
	Decode(raw []byte) (Response, error)
}

type ClientTransport interface {
	io.ReadWriter
	Port() int
	Close() error
}
