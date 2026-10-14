package domain

import "time"

type Record struct {
	Value    string
	ExpireAt *time.Time
}

func NewRecord(value string, expiry *time.Duration) Record {
	r := Record{
		Value: value,
	}

	if expiry != nil {
		deadline := time.Now().Add(*expiry)
		r.ExpireAt = &deadline
	}

	return r
}
