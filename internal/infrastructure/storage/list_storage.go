package infrastructure_storage

import (
	"github.com/dyingvoid/cachy/internal/domain"
)

func (s *MemoryStorage) RPush(key string, values ...string) (int, error) {
	sh := s.shards.getShard(key)

	sh.mu.Lock()
	defer sh.mu.Unlock()

	record, ok := sh.data[key]
	if ok && hasLiveString(record) {
		return 0, domain.ErrWrongType
	}

	list, ok := record.(*domain.ListRecord)
	if !ok {
		list = domain.NewListRecord()
	}

	list.PushBack(values...)
	sh.data[key] = list

	return list.Len(), nil
}

func (s *MemoryStorage) LPush(key string, values ...string) (int, error) {
	sh := s.shards.getShard(key)

	sh.mu.Lock()
	defer sh.mu.Unlock()

	record, ok := sh.data[key]
	if ok && hasLiveString(record) {
		return 0, domain.ErrWrongType
	}

	list, ok := record.(*domain.ListRecord)
	if !ok {
		list = domain.NewListRecord()
	}

	list.PushFront(values...)
	sh.data[key] = list

	return list.Len(), nil
}

func (s *MemoryStorage) LRange(key string, start, stop int) ([]string, error) {
	sh := s.shards.getShard(key)

	sh.mu.RLock()
	defer sh.mu.RUnlock()

	record, ok := sh.data[key]
	if ok && hasLiveString(record) {
		return nil, domain.ErrWrongType
	}

	list, ok := record.(*domain.ListRecord)
	if !ok {
		return []string{}, nil
	}

	return list.Range(start, stop), nil
}

func (s *MemoryStorage) LLen(key string) (int, error) {
	sh := s.shards.getShard(key)

	sh.mu.RLock()
	defer sh.mu.RUnlock()

	record, ok := sh.data[key]
	if ok && hasLiveString(record) {
		return 0, domain.ErrWrongType
	}

	list, ok := record.(*domain.ListRecord)
	if !ok {
		return 0, nil
	}

	return list.Len(), nil
}

func (s *MemoryStorage) LPop(key string, count int) ([]string, error) {
	sh := s.shards.getShard(key)

	sh.mu.Lock()
	defer sh.mu.Unlock()

	record, ok := sh.data[key]
	if ok && hasLiveString(record) {
		return nil, domain.ErrWrongType
	}

	list, ok := record.(*domain.ListRecord)
	if !ok {
		return []string{}, nil
	}

	values := list.PopFront(count)
	if list.Len() == 0 {
		delete(sh.data, key)
	}

	return values, nil
}
