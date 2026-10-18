package features

import "sync"

type ReplicationHandler struct {
	mu       sync.RWMutex
	streams  map[*ReplicationStream]struct{}
	capacity int
}

type ReplicationStream struct {
	handler   *ReplicationHandler
	commands  chan ReplicationCommand
	done      chan struct{}
	mu        sync.Mutex
	closed    bool
	closeOnce sync.Once
}

func NewReplicationHandler(capacity int) *ReplicationHandler {
	return &ReplicationHandler{
		streams:  make(map[*ReplicationStream]struct{}),
		capacity: capacity,
	}
}

func (h *ReplicationHandler) Subscribe() *ReplicationStream {
	stream := &ReplicationStream{
		handler:  h,
		commands: make(chan ReplicationCommand, h.capacity),
		done:     make(chan struct{}),
	}

	h.mu.Lock()
	h.streams[stream] = struct{}{}
	h.mu.Unlock()

	return stream
}

func (h *ReplicationHandler) Publish(cmd ReplicationCommand) {
	h.mu.RLock()
	streams := make([]*ReplicationStream, 0, len(h.streams))
	for stream := range h.streams {
		streams = append(streams, stream)
	}
	h.mu.RUnlock()

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

func (s *ReplicationStream) Commands() <-chan ReplicationCommand {
	return s.commands
}

func (s *ReplicationStream) Close() {
	s.handler.unsubscribe(s)
}

func (s *ReplicationStream) send(cmd ReplicationCommand) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}

	select {
	case s.commands <- cmd:
	case <-s.done:
	}
}

func (s *ReplicationStream) close() {
	s.closeOnce.Do(func() {
		close(s.done)
	})

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}

	s.closed = true
	close(s.commands)
}
