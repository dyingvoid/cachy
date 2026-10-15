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

func NewListRecord(values ...string) ListRecord {
	return ListRecord{List: append([]string{}, values...)}
}

func (lr *ListRecord) PushBack(values ...string) {
	lr.List = append(lr.List, values...)
}

func (lr *ListRecord) PushFront(values ...string) {
	front := make([]string, 0, len(values))
	for i := len(values) - 1; i >= 0; i-- {
		front = append(front, values[i])
	}

	lr.List = append(front, lr.List...)
}

func (lr *ListRecord) Len() int {
	return len(lr.List)
}

func (lr *ListRecord) Range(start, stop int) []string {
	start = lr.normalizeIndex(start)
	stop = lr.normalizeIndex(stop)

	if start >= lr.Len() || start > stop {
		return []string{}
	}

	if stop >= lr.Len() {
		stop = lr.Len() - 1
	}

	return append([]string{}, lr.List[start:stop+1]...)
}

func (lr *ListRecord) normalizeIndex(index int) int {
	if index < 0 {
		index += lr.Len()
	}

	if index < 0 {
		return 0
	}

	return index
}
