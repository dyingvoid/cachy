package infrastructure_storage

import (
	"github.com/dyingvoid/cachy/internal/domain"
)

func (s *MemoryStorage) RPush(key string, values ...string) (int, error) {
	ssh := s.strings.getShard(key)
	lsh := s.lists.getShard(key)

	ssh.mu.RLock()
	defer ssh.mu.RUnlock()

	lsh.mu.Lock()
	defer lsh.mu.Unlock()

	if hasLiveString(ssh.data, key) {
		return 0, domain.ErrWrongType
	}

	record, ok := lsh.data[key]
	if !ok {
		record = domain.NewListRecord()
	}

	record.PushBack(values...)
	lsh.data[key] = record

	return record.Len(), nil
}

func (s *MemoryStorage) LPush(key string, values ...string) (int, error) {
	ssh := s.strings.getShard(key)
	lsh := s.lists.getShard(key)

	ssh.mu.RLock()
	defer ssh.mu.RUnlock()

	lsh.mu.Lock()
	defer lsh.mu.Unlock()

	if hasLiveString(ssh.data, key) {
		return 0, domain.ErrWrongType
	}

	record, ok := lsh.data[key]
	if !ok {
		record = domain.NewListRecord()
	}

	record.PushFront(values...)
	lsh.data[key] = record

	return record.Len(), nil
}

func (s *MemoryStorage) LRange(key string, start, stop int) ([]string, error) {
	ssh := s.strings.getShard(key)
	lsh := s.lists.getShard(key)

	ssh.mu.RLock()
	defer ssh.mu.RUnlock()

	lsh.mu.RLock()
	defer lsh.mu.RUnlock()

	if hasLiveString(ssh.data, key) {
		return nil, domain.ErrWrongType
	}

	record, ok := lsh.data[key]
	if !ok {
		return []string{}, nil
	}

	return record.Range(start, stop), nil
}

func (s *MemoryStorage) LLen(key string) (int, error) {
	ssh := s.strings.getShard(key)
	lsh := s.lists.getShard(key)

	ssh.mu.RLock()
	defer ssh.mu.RUnlock()

	lsh.mu.RLock()
	defer lsh.mu.RUnlock()

	if hasLiveString(ssh.data, key) {
		return 0, domain.ErrWrongType
	}

	record, ok := lsh.data[key]
	if !ok {
		return 0, nil
	}

	return record.Len(), nil
}

func (s *MemoryStorage) LPop(key string, count int) ([]string, error) {
	ssh := s.strings.getShard(key)
	lsh := s.lists.getShard(key)

	ssh.mu.RLock()
	defer ssh.mu.RUnlock()

	lsh.mu.Lock()
	defer lsh.mu.Unlock()

	if hasLiveString(ssh.data, key) {
		return nil, domain.ErrWrongType
	}

	record, ok := lsh.data[key]
	if !ok {
		return []string{}, nil
	}

	values := record.PopFront(count)
	if record.Len() == 0 {
		delete(lsh.data, key)
	} else {
		lsh.data[key] = record
	}

	return values, nil
}
