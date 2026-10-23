package features_conn

type Mode uint8

const (
	ModeDefault Mode = iota
	ModeReplicationStreaming
)

type State struct {
	Mode          Mode
	ListeningPort *int
	Capa          *string
	Offset        int64
	ReplicationID string
}

func NewState() *State {
	return &State{Mode: ModeDefault}
}

func (s *State) TryStartReplication() bool {
	if s.Mode == ModeReplicationStreaming || !s.ready() {
		return false
	}

	s.Mode = ModeReplicationStreaming
	return true
}

func (s *State) ready() bool {
	return s.ListeningPort != nil && s.Capa != nil && s.ReplicationID != ""
}
