package infrastructure_storage

import (
	"github.com/dyingvoid/cachy/internal/domain"
)

func (s *MemoryStorage) Set(key string, value domain.StringRecord) error {
	ssh := s.strings.getShard(key)
	lsh := s.lists.getShard(key)

	ssh.mu.Lock()
	defer ssh.mu.Unlock()

	lsh.mu.RLock()
	defer lsh.mu.RUnlock()

	if _, ok := lsh.data[key]; ok {
		return domain.ErrWrongType
	}

	ssh.data[key] = value

	return nil
}

func (s *MemoryStorage) Get(key string) (domain.StringRecord, error) {
	ssh := s.strings.getShard(key)
	lsh := s.lists.getShard(key)

	ssh.mu.RLock()
	defer ssh.mu.RUnlock()

	lsh.mu.RLock()
	defer lsh.mu.RUnlock()

	if _, ok := lsh.data[key]; ok {
		return domain.StringRecord{}, domain.ErrWrongType
	}

	value, ok := ssh.data[key]
	if !ok || value.Expired() {
		return domain.StringRecord{}, domain.ErrNotFound
	}

	return value, nil
}

func (s *MemoryStorage) Delete(key string) error {
	ssh := s.strings.getShard(key)
	lsh := s.lists.getShard(key)

	ssh.mu.Lock()
	defer ssh.mu.Unlock()

	lsh.mu.RLock()
	defer lsh.mu.RUnlock()

	if _, ok := lsh.data[key]; ok {
		return domain.ErrWrongType
	}

	if _, ok := ssh.data[key]; !ok {
		return domain.ErrNotFound
	}

	delete(ssh.data, key)

	return nil
}
