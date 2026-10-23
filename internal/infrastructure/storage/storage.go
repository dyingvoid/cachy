package infrastructure_storage

import (
	"log/slog"

	"github.com/dyingvoid/cachy/internal/domain"
)

type MemoryStorage struct {
	shards shardCollection
}

func NewMemoryStorage(log *slog.Logger, cfg Config) *MemoryStorage {
	s := &MemoryStorage{
		shards: make(shardCollection, cfg.ShardCount),
	}
	for i := range s.shards {
		s.shards[i].data = make(map[string]domain.Record)
		go cleanup(log, &s.shards[i], cfg.CleanupInterval, cfg.CleanupBatchSize)
	}

	return s
}

func (s *MemoryStorage) Exists(key string) bool {
	sh := s.shards.getShard(key)

	sh.mu.RLock()
	defer sh.mu.RUnlock()

	record, ok := sh.data[key]

	return ok && !record.Expired()
}

func hasLiveString(record domain.Record) bool {
	stringRecord, ok := record.(domain.StringRecord)

	return ok && !stringRecord.Expired()
}
