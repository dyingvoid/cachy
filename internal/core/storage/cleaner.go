package core_storage

import "time"

func cleanup(
	shard *shard,
	interval time.Duration,
	batchSize int,
) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
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
	}
}
