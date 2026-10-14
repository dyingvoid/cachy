package core_storage

import (
	"log/slog"
	"time"
)

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
