package features_replication

import (
	"context"
	"fmt"
	"io"
	"strings"

	features_commands "github.com/dyingvoid/cachy/internal/features/commands"
)

type ClientTransport interface {
	io.ReadWriter
	Port() int
}

type Serializer interface {
	SerializeCommand(features_commands.Command) []byte
}

type Deserializer interface {
	ReadSimpleString() (string, error)
	DiscardBulkString() error
}

type ReplicationClient struct {
	transport    ClientTransport
	serializer   Serializer
	deserializer Deserializer
}

func NewReplicationClient(
	transport ClientTransport,
	serializer Serializer,
	deserializer Deserializer,
) *ReplicationClient {
	return &ReplicationClient{
		transport:    transport,
		serializer:   serializer,
		deserializer: deserializer,
	}
}

func (c *ReplicationClient) Handshake(ctx context.Context) error {
	port := c.transport.Port()
	capa := "psync2"

	steps := []struct {
		cmd  features_commands.Command
		want string
	}{
		{features_commands.PingCmd{}, "PONG"},
		{features_commands.ReplConfCmd{ListeningPort: &port}, "OK"},
		{features_commands.ReplConfCmd{Capa: &capa}, "OK"},
	}
	for _, step := range steps {
		if err := c.confirm(ctx, step.cmd, step.want); err != nil {
			return err
		}
	}

	reply, err := c.exchange(ctx, features_commands.PsyncCmd{
		ReplicationID: "?",
		Offset:        -1,
	})
	if err != nil {
		return err
	}
	if !strings.HasPrefix(reply, "FULLRESYNC ") {
		return fmt.Errorf("unexpected PSYNC reply: %q", reply)
	}

	if err := c.deserializer.DiscardBulkString(); err != nil {
		return fmt.Errorf("read snapshot: %w", err)
	}

	return nil
}

func (c *ReplicationClient) confirm(
	ctx context.Context,
	cmd features_commands.Command,
	want string,
) error {
	reply, err := c.exchange(ctx, cmd)
	if err != nil {
		return err
	}
	if reply != want {
		return fmt.Errorf("unexpected %T reply: %q, want %q", cmd, reply, want)
	}

	return nil
}

func (c *ReplicationClient) exchange(
	ctx context.Context,
	cmd features_commands.Command,
) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("send command: %w", err)
	}

	payload := c.serializer.SerializeCommand(cmd)
	n, err := c.transport.Write(payload)
	if err != nil {
		return "", fmt.Errorf("write command: %w", err)
	}
	if n != len(payload) {
		return "", fmt.Errorf("write command: %w", io.ErrShortWrite)
	}

	reply, err := c.deserializer.ReadSimpleString()
	if err != nil {
		return "", fmt.Errorf("read reply: %w", err)
	}

	return reply, nil
}
