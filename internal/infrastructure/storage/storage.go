package infrastructure_storage

import (
	"log/slog"

	"github.com/dyingvoid/cachy/internal/domain"
)

type MemoryStorage struct {
	strings shardCollection[domain.StringRecord]
	lists   shardCollection[domain.ListRecord]
}

func NewMemoryStorage(log *slog.Logger, cfg Config) *MemoryStorage {
	s := &MemoryStorage{
		strings: make([]shard[domain.StringRecord], cfg.ShardCount),
		lists:   make([]shard[domain.ListRecord], cfg.ShardCount),
	}
	for i := range s.strings {
		s.strings[i].data = make(map[string]domain.StringRecord)
		go cleanup(log, &s.strings[i], cfg.CleanupInterval, cfg.CleanupBatchSize)
	}
	for i := range s.lists {
		s.lists[i].data = make(map[string]domain.ListRecord)
	}

	return s
}

func (s *MemoryStorage) Exists(key string) bool {
	ssh := s.strings.getShard(key)
	lsh := s.lists.getShard(key)

	ssh.mu.RLock()
	defer ssh.mu.RUnlock()

	lsh.mu.RLock()
	defer lsh.mu.RUnlock()

	if _, ok := lsh.data[key]; ok {
		return true
	}

	return hasLiveString(ssh.data, key)
}

func hasLiveString(data map[string]domain.StringRecord, key string) bool {
	record, ok := data[key]

	return ok && !record.Expired()
}
