package core_storage

import (
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/dyingvoid/cachy/internal/domain"
)

func testCleanupLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func shardHas(sh *shard, key string) bool {
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	_, ok := sh.data[key]

	return ok
}

func shardLen(sh *shard) int {
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	return len(sh.data)
}

func waitFor(timeout time.Duration, cond func() bool) bool {
	deadline := time.Now().Add(timeout)

	for !cond() {
		if time.Now().After(deadline) {
			return false
		}

		time.Sleep(time.Millisecond)
	}

	return true
}

func TestCleanup_RemovesExpiredOnly(t *testing.T) {
	sh := &shard{data: map[string]domain.Record{
		"expired": expiredRecord("expired"),
		"live":    domain.NewRecord("live", nil),
	}}

	go cleanup(testCleanupLogger(), sh, 5*time.Millisecond, 100)

	if !waitFor(2*time.Second, func() bool { return !shardHas(sh, "expired") }) {
		t.Fatal("cleanup did not remove the expired record")
	}

	if !shardHas(sh, "live") {
		t.Fatal("cleanup removed a record that has not expired")
	}
}

func TestCleanup_RespectsBatchSize(t *testing.T) {
	const records = 5

	data := make(map[string]domain.Record, records)
	for i := range records {
		data[fmt.Sprintf("key-%d", i)] = expiredRecord("value")
	}

	sh := &shard{data: data}

	go cleanup(testCleanupLogger(), sh, 100*time.Millisecond, 1)

	if !waitFor(time.Second, func() bool { return shardLen(sh) < records }) {
		t.Fatal("cleanup did not remove any expired records")
	}

	time.Sleep(20 * time.Millisecond)

	if got := shardLen(sh); got != records-1 {
		t.Fatalf("cleanup removed %d records in one tick, want 1", records-got)
	}
}

func TestCleanup_RemovesAllExpiredOverTime(t *testing.T) {
	const records = 5

	data := make(map[string]domain.Record, records)
	for i := range records {
		data[fmt.Sprintf("key-%d", i)] = expiredRecord("value")
	}

	sh := &shard{data: data}

	go cleanup(testCleanupLogger(), sh, 5*time.Millisecond, 2)

	if !waitFor(2*time.Second, func() bool { return shardLen(sh) == 0 }) {
		t.Fatalf("cleanup left %d expired records, want 0", shardLen(sh))
	}
}

func TestCleanup_EmptyShard(t *testing.T) {
	sh := &shard{data: map[string]domain.Record{}}

	go cleanup(testCleanupLogger(), sh, 5*time.Millisecond, 10)

	time.Sleep(30 * time.Millisecond)

	if got := shardLen(sh); got != 0 {
		t.Fatalf("cleanup left %d records in an empty shard, want 0", got)
	}
}
