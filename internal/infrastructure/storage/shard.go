package infrastructure_storage

import (
	"hash/fnv"
	"log/slog"
	"sync"
	"time"

	"github.com/dyingvoid/cachy/internal/domain"
)

type shard struct {
	mu   sync.RWMutex
	data map[string]domain.Record
}

type shardCollection []shard

func (s shardCollection) getShard(key string) *shard {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))

	idx := h.Sum32() % uint32(len(s))
	return &s[idx]
}

func cleanup(
	log *slog.Logger,
	shard *shard,
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
