package features_replication

type State struct {
	ListeningPort *int
	Capa          *string
	Offset        int64
	ReplicationID string
}

func (s State) Ready() bool {
	return s.ListeningPort != nil && s.Capa != nil && s.ReplicationID != ""
}
