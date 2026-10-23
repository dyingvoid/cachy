package features_replication

import (
	"sync"

	features_commands "github.com/dyingvoid/cachy/internal/features/commands"
)

type ReplicationHandler struct {
	mu      sync.RWMutex
	streams map[*ReplicationStream]struct{}
	cfg     Config
}

func NewReplicationHandler(cfg Config) *ReplicationHandler {
	return &ReplicationHandler{
		streams: make(map[*ReplicationStream]struct{}),
		cfg:     cfg,
	}
}

func (h *ReplicationHandler) Subscribe() *ReplicationStream {
	stream := &ReplicationStream{
		handler:  h,
		commands: make(chan features_commands.WriteCommand, h.cfg.ReplicationBuffer),
		done:     make(chan struct{}),
	}

	h.mu.Lock()
	h.streams[stream] = struct{}{}
	h.mu.Unlock()

	return stream
}

func (h *ReplicationHandler) Publish(cmd features_commands.WriteCommand) {
	h.mu.RLock()
	streams := make([]*ReplicationStream, 0, len(h.streams))
	for stream := range h.streams {
		streams = append(streams, stream)
	}
	h.mu.RUnlock()

	// TODO: writing to slow chan will block everyone else
	for _, stream := range streams {
		stream.send(cmd)
	}
}

func (h *ReplicationHandler) unsubscribe(stream *ReplicationStream) {
	h.mu.Lock()
	_, ok := h.streams[stream]
	if ok {
		delete(h.streams, stream)
	}
	h.mu.Unlock()

	if ok {
		stream.close()
	}
}
