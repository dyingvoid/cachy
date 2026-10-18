package features_conn

import features_replication "github.com/dyingvoid/cachy/internal/features/replication"

type Mode uint8

const (
	ModeDefault Mode = iota
	ModeReplicationStreaming
)

type State struct {
	Mode        Mode
	Replication features_replication.State
}

func NewState() *State {
	return &State{Mode: ModeDefault}
}

func (s *State) TryStartReplication() bool {
	if s.Mode == ModeReplicationStreaming || !s.Replication.Ready() {
		return false
	}

	s.Mode = ModeReplicationStreaming
	return true
}
