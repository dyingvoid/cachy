package core_storage

import (
	"hash/fnv"
	"sync"

	domain_errors "github.com/dyingvoid/cachy/internal/domain/errors"
)

type shard struct {
	mu   sync.RWMutex
	data map[string]string
}

type MemoryStorage struct {
	shards []shard
}

func NewMemoryStorage(cfg Config) *MemoryStorage {
	s := &MemoryStorage{
		shards: make([]shard, cfg.ShardCount),
	}
	for i := range s.shards {
		s.shards[i].data = make(map[string]string)
	}

	return s
}

func (s *MemoryStorage) Set(key, value string) error {
	sh := s.getShard(key)

	sh.mu.Lock()
	sh.data[key] = value
	sh.mu.Unlock()

	return nil
}

func (s *MemoryStorage) Get(key string) (string, error) {
	sh := s.getShard(key)

	sh.mu.Lock()
	value, ok := sh.data[key]
	sh.mu.Unlock()

	if !ok {
		return "", domain_errors.ErrNotFound
	}

	return value, nil
}

func (s *MemoryStorage) getShard(key string) *shard {
	h := fnv.New32a()
	// no actual error returned
	_, _ = h.Write([]byte(key))
	return &s.shards[h.Sum32()%uint32(len(s.shards))]
}
