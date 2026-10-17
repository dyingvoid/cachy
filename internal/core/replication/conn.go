package core_replication

type ConnMode uint8

const (
	ConnModeDefault ConnMode = iota
	ConnModeReplicationStreaming
)

type PsyncResponse struct {
	ReplID string
	Offset int64
}