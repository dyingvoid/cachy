package domain

import "time"

type StringRecord struct {
	Value    string
	ExpireAt *time.Time
}

func NewStringRecord(value string, expiry *time.Duration) StringRecord {
	r := StringRecord{
		Value: value,
	}

	if expiry != nil {
		deadline := time.Now().Add(*expiry)
		r.ExpireAt = &deadline
	}

	return r
}

func (r StringRecord) Expired() bool {
	if r.ExpireAt == nil {
		return false
	}

	return r.ExpireAt.Before(time.Now())
}

type ListRecord struct {
	List []string
}

func NewListRecord(value string, expiry *time.Duration) ListRecord {
	return ListRecord{List: []string{value}}
}

func (lr *ListRecord) PushBack(value string) {
	lr.List = append(lr.List, value)
}