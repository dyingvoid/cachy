package core_replication

import "context"

type Message struct {
	ID   string
	Data []byte
}

type Codec interface {
	Encode(msg Message) ([]byte, error)
}

type Transport interface {
	Send(ctx context.Context, payload []byte) error
	Close() error
}

type ReplicationClient struct {
	coded     Codec
	transport Transport
}

func (rc *ReplicationClient) Handshake() {

}
