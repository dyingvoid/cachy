package core_replication

import "strconv"

type Message interface {
	Command() string
	Args() []string
}

type PingMessage struct{}

func (PingMessage) Command() string {
	return "PING"
}

func (PingMessage) Args() []string {
	return nil
}

type ReplConfMessage struct {
	Port *int
	Capa *string
}

func (ReplConfMessage) Command() string {
	return "REPLCONF"
}

func (m ReplConfMessage) Args() []string {
	args := make([]string, 0, 2)
	if m.Port != nil {
		args = append(args, "listening-port", strconv.Itoa(*m.Port))
	}
	if m.Capa != nil {
		args = append(args, "capa", *m.Capa)
	}

	return args
}

type PsyncMessage struct {
	ReplicationID string
	Offset        int64
}

func (PsyncMessage) Command() string {
	return "PSYNC"
}

func (m PsyncMessage) Args() []string {
	return []string{m.ReplicationID, strconv.FormatInt(m.Offset, 10)}
}
