package infrastructure_storage

import (
	"github.com/dyingvoid/cachy/internal/domain"
)

func (s *MemoryStorage) Set(key string, value domain.StringRecord) error {
	sh := s.shards.getShard(key)

	sh.mu.Lock()
	defer sh.mu.Unlock()

	if record, ok := sh.data[key]; ok {
		if _, isString := record.(domain.StringRecord); !isString {
			return domain.ErrWrongType
		}
	}

	sh.data[key] = value

	return nil
}

func (s *MemoryStorage) Get(key string) (domain.StringRecord, error) {
	sh := s.shards.getShard(key)

	sh.mu.RLock()
	defer sh.mu.RUnlock()

	record, ok := sh.data[key]
	if !ok {
		return domain.StringRecord{}, domain.ErrNotFound
	}

	value, ok := record.(domain.StringRecord)
	if !ok {
		return domain.StringRecord{}, domain.ErrWrongType
	}

	if value.Expired() {
		return domain.StringRecord{}, domain.ErrNotFound
	}

	return value, nil
}

func (s *MemoryStorage) Delete(key string) error {
	sh := s.shards.getShard(key)

	sh.mu.Lock()
	defer sh.mu.Unlock()

	record, ok := sh.data[key]
	if !ok {
		return domain.ErrNotFound
	}

	if _, isString := record.(domain.StringRecord); !isString {
		return domain.ErrWrongType
	}

	delete(sh.data, key)

	return nil
}
