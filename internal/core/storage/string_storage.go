package core_storage

import (
	"log/slog"

	"github.com/dyingvoid/cachy/internal/domain"
	"github.com/dyingvoid/cachy/internal/features"
)

type StringStorage struct {
	log    *slog.Logger
	shards shardCollection[domain.StringRecord]
}

func NewMemoryStorage(log *slog.Logger, cfg Config) *StringStorage {
	s := &StringStorage{
		log:    log,
		shards: make([]shard[domain.StringRecord], cfg.ShardCount),
	}
	for i := range s.shards {
		s.shards[i].data = make(map[string]domain.StringRecord)
		go cleanup(log, &s.shards[i], cfg.CleanupInterval, cfg.CleanupBatchSize)
	}

	return s
}

func (s *StringStorage) Set(key string, value domain.StringRecord) error {
	sh := s.shards.getShard(key)

	sh.mu.Lock()
	sh.data[key] = value
	sh.mu.Unlock()

	return nil
}

func (s *StringStorage) Get(key string) (domain.StringRecord, error) {
	sh := s.shards.getShard(key)

	sh.mu.RLock()
	value, ok := sh.data[key]
	sh.mu.RUnlock()

	if !ok {
		return domain.StringRecord{}, features.ErrNotFound
	}

	if value.Expired() {
		sh.mu.Lock()
		defer sh.mu.Unlock()

		value, ok = sh.data[key]
		if !ok {
			return domain.StringRecord{}, features.ErrNotFound
		}
		if value.Expired() {
			delete(sh.data, key)
			return domain.StringRecord{}, features.ErrNotFound
		}
	}

	return value, nil
}

func (s *StringStorage) Delete(key string) error {
	sh := s.shards.getShard(key)
	sh.mu.Lock()
	defer sh.mu.Unlock()

	_, ok := sh.data[key]
	if !ok {
		return features.ErrNotFound
	}

	delete(sh.data, key)

	return nil
}

func (s *StringStorage) Exists(key string) bool {
	sh := s.shards.getShard(key)

	sh.mu.RLock()
	value, ok := sh.data[key]
	sh.mu.RUnlock()

	return ok && !value.Expired()
}
