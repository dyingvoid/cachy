package features_replication

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type ReplicationClient struct {
	codec     Codec
	transport ClientTransport
	reader    *bufio.Reader
}

func NewReplicationClient(codec Codec, transport ClientTransport) *ReplicationClient {
	return &ReplicationClient{
		codec:     codec,
		transport: transport,
		reader:    bufio.NewReader(transport),
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

	if err := c.receiveFile(); err != nil {
		return err
	}

	return nil
}

func (c *ReplicationClient) receiveFile() error {
	header, err := c.reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("read file header: %w", err)
	}

	header = strings.TrimSpace(header)
	if !strings.HasPrefix(header, "$") {
		return fmt.Errorf(
			"read file header %q: %w",
			header,
			ErrUnexpectedResponse,
		)
	}

	size, err := strconv.Atoi(strings.TrimPrefix(header, "$"))
	if err != nil || size < 0 {
		return fmt.Errorf(
			"parse file length %q: %w",
			header,
			ErrUnexpectedResponse,
		)
	}

	content := make([]byte, size)
	if _, err := io.ReadFull(c.reader, content); err != nil {
		return fmt.Errorf("read file content: %w", err)
	}

	// TODO: later, maybe no need for this
	if _, err := c.reader.ReadByte(); err != nil {
		return fmt.Errorf("read file terminator: %w", err)
	}

	return nil
}

func (c *ReplicationClient) send(ctx context.Context, msg Message) (Response, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("send message: %w", err)
	}

	payload, err := c.codec.Encode(msg)
	if err != nil {
		return nil, fmt.Errorf("encode message: %w", err)
	}

	n, err := c.transport.Write(payload)
	if err != nil {
		return nil, fmt.Errorf("write message: %w", err)
	}
	if n != len(payload) {
		return nil, fmt.Errorf("write message: %w", io.ErrShortWrite)
	}

	raw, err := c.reader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	resp, err := c.codec.Decode(raw)
	if err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp, nil
}
