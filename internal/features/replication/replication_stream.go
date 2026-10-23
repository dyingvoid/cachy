package features_replication

import (
	"sync"

	features_commands "github.com/dyingvoid/cachy/internal/features/commands"
)

type ReplicationStream struct {
	handler   *ReplicationHandler
	commands  chan features_commands.WriteCommand
	done      chan struct{}
	mu        sync.Mutex
	closed    bool
	closeOnce sync.Once
}

func (s *ReplicationStream) Chan() <-chan features_commands.WriteCommand {
	return s.commands
}

func (s *ReplicationStream) Close() {
	s.handler.unsubscribe(s)
}

func (s *ReplicationStream) send(cmd features_commands.WriteCommand) {
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