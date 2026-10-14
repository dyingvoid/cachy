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

func (s *ListStorage) RPush(key string, values ...string) (int, error) {
	sh := s.shards.getShard(key)

	sh.mu.Lock()
	defer sh.mu.Unlock()

	record, ok := sh.data[key]
	if !ok {
		rec := domain.NewListRecord(values...)
		sh.data[key] = rec

		return rec.Len(), nil
	}

	record.PushBack(values...)
	sh.data[key] = record

	return record.Len(), nil
}

func (s *ListStorage) LRange(key string, start, stop int) ([]string, error) {
	sh := s.shards.getShard(key)

	sh.mu.RLock()
	defer sh.mu.RUnlock()

	record, ok := sh.data[key]
	if !ok {
		return []string{}, nil
	}

	return record.Range(start, stop), nil
}
