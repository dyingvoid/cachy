package core_replication

import (
	"context"
	"fmt"
	"strconv"
	"strings"
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
	resp, err := c.sendMsg(ctx, Message{Command: "PING"})
	if err != nil {
		return err
	}
	if resp != "PONG" {
		return fmt.Errorf("unexpected PING response: %+v", resp)
	}

	for _, args := range [][]string{
		{"listening-port", strconv.Itoa(c.transport.Port())},
		{"capa", "psync2"},
	} {
		resp, err := c.sendMsg(ctx, Message{
			Command: "REPLCONF",
			Args:    args,
		})
		if err != nil {
			return err
		}
		if resp != "+OK" {
			return fmt.Errorf("unexpected REPLCONF response: %+v", resp)
		}
	}

	resp, err = c.sendMsg(ctx, Message{
		Command: "PSYNC",
		Args:    []string{"?", "-1"},
	})
	if err != nil {
		return err
	}

	fields := strings.Fields(resp)
	if len(fields) != 3 || fields[0] != "FULLRESYNC" {
		return fmt.Errorf("unexpected psync response: %+v", resp)
	}
	return nil
}

func (c *ReplicationClient) sendMsg(ctx context.Context, msg Message) (string, error) {
	var zero string
	payload, err := c.codec.Encode(msg)
	if err != nil {
		return zero, fmt.Errorf("encode message: %w", err)
	}

	raw, err := c.transport.Send(ctx, payload)
	if err != nil {
		return zero, fmt.Errorf("send message: %w", err)
	}

	resp, err := c.codec.Decode(raw)
	if err != nil {
		return zero, fmt.Errorf("decode response: %w", err)
	}

	return resp, nil
}
