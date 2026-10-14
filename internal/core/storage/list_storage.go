package core_storage

import (
	"github.com/dyingvoid/cachy/internal/domain"
)

type ListStorage struct {
	shards shardCollection[domain.ListRecord]
}

func NewListStorage(cfg Config) *ListStorage {
	s := &ListStorage{
		shards: make([]shard[domain.ListRecord], cfg.ShardCount),
	}
	for i := range s.shards {
		s.shards[i].data = make(map[string]domain.ListRecord)
	}

	return s
}

func (s *ListStorage) RPush(key string, value string) error {
	sh := s.shards.getShard(key)

	sh.mu.Lock()
	defer sh.mu.Unlock()

	record, ok := sh.data[key]
	if !ok {
		sh.data[key] = domain.NewListRecord(value, nil)
		return nil
	}

	record.PushBack(value)
	sh.data[key] = record

	return nil
}
