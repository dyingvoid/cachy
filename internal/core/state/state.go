package core_state

import (
	"crypto/rand"
	"encoding/hex"
	"sync/atomic"
)

func init() {
	id, err := genID()
	if err != nil {
		panic(err)
	}

	replicationID.Store(&id)
}

var (
	replicationID     atomic.Pointer[string]
	replicationOffset atomic.Int64
)

func ReplicationID() string {
	p := replicationID.Load()
	if p == nil {
		return ""
	}
	return *p
}

func ReplicationOffset() int64 {
	return replicationOffset.Load()
}

func IncReplicationOffset(delta int64) int64 {
	return replicationOffset.Add(delta)
}

func genID() (string, error) {
	b := make([]byte, 40)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}
