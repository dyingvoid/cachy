package domain

import "strconv"

type Result interface {
	isResult()
}

type StringResult string

type ListResult []string

func (ListResult) isResult() {}

func (StringResult) isResult() {}

type MapResult interface {
	Result
	ToMap() map[string]string
}

type Info struct {
	Role              string
	ReplicaOf         *string
	ReplicationID     string
	ReplicationOffset int64
}

func NewInfo(
	role string, replicaOf *string,
	replicationID string, replicationOffset int64,
) Info {
	return Info{
		Role:              role,
		ReplicaOf:         replicaOf,
		ReplicationID:     replicationID,
		ReplicationOffset: replicationOffset,
	}
}

func (i Info) ToMap() map[string]string {
	m := map[string]string{
		"role":               i.Role,
		"replication_id":     i.ReplicationID,
		"replication_offset": strconv.FormatInt(i.ReplicationOffset, 10),
	}
	if i.ReplicaOf != nil {
		m["replica_of"] = *i.ReplicaOf
	}

	return m
}

func (i Info) isResult() {}
