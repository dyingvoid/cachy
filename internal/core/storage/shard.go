package core_storage

import (
	"hash/fnv"
	"log/slog"
	"sync"
	"time"

	"github.com/dyingvoid/cachy/internal/domain"
)

type shard[T domain.StringRecord | domain.ListRecord] struct {
	mu   sync.RWMutex
	data map[string]T
}

type shardCollection[T domain.StringRecord | domain.ListRecord] []shard[T]

func (s shardCollection[T]) getShard(key string) *shard[T] {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))

	idx := h.Sum32() % uint32(len(s))
	return &s[idx]
}

func cleanup(
	log *slog.Logger,
	shard *shard[domain.StringRecord],
	interval time.Duration,
	batchSize int,
) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		log.Debug("starting shard cleanup")

		start := time.Now()
		shard.mu.Lock()
		count := 0

		for key, record := range shard.data {
			if record.Expired() {
				delete(shard.data, key)
				count++
				if count >= batchSize {
					break
				}
			}
		}

		shard.mu.Unlock()
		log.Debug(
			"shard cleaned up",
			slog.Int("records", count),
			slog.Duration("elapsed", time.Since(start)),
		)
	}
}
