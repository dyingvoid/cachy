package core_storage

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/dyingvoid/cachy/internal/domain"
	domain_errors "github.com/dyingvoid/cachy/internal/domain/errors"
)

const longCleanupInterval = time.Hour

func newTestStorage(t *testing.T, cfg Config) *MemoryStorage {
	t.Helper()

	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	return NewMemoryStorage(log, cfg)
}

func testConfig() Config {
	return Config{
		ShardCount:       8,
		CleanupInterval:  longCleanupInterval,
		CleanupBatchSize: 100,
	}
}

func ptrTime(t time.Time) *time.Time {
	return &t
}

func expiredRecord(value string) domain.Record {
	return domain.Record{
		Value:    value,
		ExpireAt: ptrTime(time.Now().Add(-time.Second)),
	}
}

func shardContains(s *MemoryStorage, key string) bool {
	sh := s.getShard(key)

	sh.mu.RLock()
	defer sh.mu.RUnlock()

	_, ok := sh.data[key]

	return ok
}

func TestMemoryStorage_SetGet(t *testing.T) {
	s := newTestStorage(t, testConfig())

	record := domain.NewRecord("value", nil)

	if err := s.Set("key", record); err != nil {
		t.Fatalf("Set() returned unexpected error: %v", err)
	}

	got, err := s.Get("key")
	if err != nil {
		t.Fatalf("Get() returned unexpected error: %v", err)
	}

	if got.Value != record.Value {
		t.Fatalf("Get() value = %q, want %q", got.Value, record.Value)
	}

	if got.ExpireAt != nil {
		t.Fatalf("Get() expireAt = %v, want nil", got.ExpireAt)
	}
}

func TestMemoryStorage_SetGet_ManyKeys(t *testing.T) {
	s := newTestStorage(t, testConfig())

	const keys = 100

	for i := range keys {
		key := fmt.Sprintf("key-%d", i)
		if err := s.Set(key, domain.NewRecord(key, nil)); err != nil {
			t.Fatalf("Set(%q) returned unexpected error: %v", key, err)
		}
	}

	for i := range keys {
		key := fmt.Sprintf("key-%d", i)

		got, err := s.Get(key)
		if err != nil {
			t.Fatalf("Get(%q) returned unexpected error: %v", key, err)
		}

		if got.Value != key {
			t.Fatalf("Get(%q) value = %q, want %q", key, got.Value, key)
		}
	}
}

func TestMemoryStorage_Set_Overwrites(t *testing.T) {
	s := newTestStorage(t, testConfig())

	if err := s.Set("key", domain.NewRecord("first", nil)); err != nil {
		t.Fatalf("Set() returned unexpected error: %v", err)
	}
	if err := s.Set("key", domain.NewRecord("second", nil)); err != nil {
		t.Fatalf("Set() returned unexpected error: %v", err)
	}

	got, err := s.Get("key")
	if err != nil {
		t.Fatalf("Get() returned unexpected error: %v", err)
	}

	if got.Value != "second" {
		t.Fatalf("Get() value = %q, want %q", got.Value, "second")
	}
}

func TestMemoryStorage_Get_NotFound(t *testing.T) {
	s := newTestStorage(t, testConfig())

	if _, err := s.Get("missing"); !errors.Is(err, domain_errors.ErrNotFound) {
		t.Fatalf("Get() error = %v, want %v", err, domain_errors.ErrNotFound)
	}
}

func TestMemoryStorage_Delete(t *testing.T) {
	s := newTestStorage(t, testConfig())

	if err := s.Set("key", domain.NewRecord("value", nil)); err != nil {
		t.Fatalf("Set() returned unexpected error: %v", err)
	}

	if err := s.Delete("key"); err != nil {
		t.Fatalf("Delete() returned unexpected error: %v", err)
	}

	if _, err := s.Get("key"); !errors.Is(err, domain_errors.ErrNotFound) {
		t.Fatalf("Get() after Delete() error = %v, want %v", err, domain_errors.ErrNotFound)
	}

	if err := s.Delete("key"); !errors.Is(err, domain_errors.ErrNotFound) {
		t.Fatalf("Delete() on missing key error = %v, want %v", err, domain_errors.ErrNotFound)
	}
}

func TestMemoryStorage_Get_Expired(t *testing.T) {
	s := newTestStorage(t, testConfig())

	if err := s.Set("key", expiredRecord("value")); err != nil {
		t.Fatalf("Set() returned unexpected error: %v", err)
	}

	if _, err := s.Get("key"); !errors.Is(err, domain_errors.ErrNotFound) {
		t.Fatalf("Get() of expired record error = %v, want %v", err, domain_errors.ErrNotFound)
	}

	if shardContains(s, "key") {
		t.Fatal("Get() of expired record did not delete it from the shard")
	}
}

func TestMemoryStorage_Get_NotExpired(t *testing.T) {
	s := newTestStorage(t, testConfig())

	record := domain.Record{
		Value:    "value",
		ExpireAt: ptrTime(time.Now().Add(time.Minute)),
	}

	if err := s.Set("key", record); err != nil {
		t.Fatalf("Set() returned unexpected error: %v", err)
	}

	got, err := s.Get("key")
	if err != nil {
		t.Fatalf("Get() returned unexpected error: %v", err)
	}

	if got.Value != record.Value {
		t.Fatalf("Get() value = %q, want %q", got.Value, record.Value)
	}
}

func TestMemoryStorage_GetShard(t *testing.T) {
	s := newTestStorage(t, testConfig())

	const key = "key"

	first := s.getShard(key)

	if second := s.getShard(key); first != second {
		t.Fatal("getShard() is not deterministic for the same key")
	}

	for i := range 100 {
		sh := s.getShard(fmt.Sprintf("key-%d", i))

		found := false
		for j := range s.shards {
			if sh == &s.shards[j] {
				found = true
				break
			}
		}

		if !found {
			t.Fatalf("getShard(%q) returned a shard outside the storage", fmt.Sprintf("key-%d", i))
		}
	}
}

func TestMemoryStorage_Cleanup_RemovesExpiredOnly(t *testing.T) {
	cfg := testConfig()
	cfg.CleanupInterval = 5 * time.Millisecond

	s := newTestStorage(t, cfg)

	if err := s.Set("expired", expiredRecord("expired")); err != nil {
		t.Fatalf("Set() returned unexpected error: %v", err)
	}
	if err := s.Set("live", domain.NewRecord("live", nil)); err != nil {
		t.Fatalf("Set() returned unexpected error: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for shardContains(s, "expired") {
		if time.Now().After(deadline) {
			t.Fatal("background cleanup did not remove the expired record")
		}

		time.Sleep(time.Millisecond)
	}

	if !shardContains(s, "live") {
		t.Fatal("background cleanup removed a record that has not expired")
	}
}

func TestMemoryStorage_ConcurrentAccess(t *testing.T) {
	s := newTestStorage(t, testConfig())

	const (
		goroutines = 32
		iterations = 50
	)

	var wg sync.WaitGroup

	for g := range goroutines {
		wg.Add(1)

		go func(g int) {
			defer wg.Done()

			for i := range iterations {
				key := fmt.Sprintf("key-%d", g%8)

				_ = s.Set(key, domain.NewRecord(fmt.Sprintf("value-%d", i), nil))
				_, _ = s.Get(key)
				_ = s.Delete(key)
			}
		}(g)
	}

	wg.Wait()
}
