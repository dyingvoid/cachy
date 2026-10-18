package core_replication

import (
	"context"
	"fmt"
)

type ReplicationClient struct {
	codec     Codec
	transport ClientTransport
}

func NewReplicationClient(codec Codec, transport ClientTransport) *ReplicationClient {
	return &ReplicationClient{
		codec:     codec,
		transport: transport,
	}
}

func (c *ReplicationClient) Handshake(ctx context.Context) error {
	resp, err := c.send(ctx, PingMessage{})
	if err != nil {
		return err
	}
	if _, ok := resp.(PongResponse); !ok {
		return fmt.Errorf("unexpected PING response: %T", resp)
	}

	port := c.transport.Port()
	capa := "psync2"
	for _, msg := range []Message{
		ReplConfMessage{Port: &port},
		ReplConfMessage{Capa: &capa},
	} {
		resp, err := c.send(ctx, msg)
		if err != nil {
			return err
		}
		if _, ok := resp.(OKResponse); !ok {
			return fmt.Errorf("unexpected REPLCONF response: %T", resp)
		}
	}

	resp, err = c.send(ctx, PsyncMessage{
		ReplicationID: "?",
		Offset:        -1,
	})
	if err != nil {
		return err
	}
	if _, ok := resp.(PsyncResponse); !ok {
		return fmt.Errorf("unexpected PSYNC response: %T", resp)
	}

	return nil
}

func (c *ReplicationClient) send(ctx context.Context, msg Message) (Response, error) {
	payload, err := c.codec.Encode(msg)
	if err != nil {
		return nil, fmt.Errorf("encode message: %w", err)
	}

	raw, err := c.transport.Send(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("send message: %w", err)
	}

	resp, err := c.codec.Decode(raw)
	if err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp, nil
}
