package core_serializer

import (
	"fmt"

	"golang.org/x/exp/constraints"
)

type RespSerializer struct {
}

func (r *RespSerializer) OK(s string) []byte {
	return []byte("+OK\r\n")
}

func (r *RespSerializer) Error(err error) []byte {
	return fmt.Append(nil, "-ERR ", err.Error(), "\r\n")
}

func (r *RespSerializer) Int[T constraints.Integer](n T) []byte {
	return fmt.Appendf(nil, ":%d\r\n", n)
}

func (r *RespSerializer) BulkString(str string) []byte {
	return fmt.Appendf(nil, "$%d\r\n%s\r\n", len(str), str)
}

func (r *RespSerializer) NilableBulkString(str *string) []byte {
	if str == nil {
		return fmt.Append(nil, "$-1\r\n")
	}

	return r.BulkString(*str)
}

func (r *RespSerializer) Array(elements ...[]byte) []byte {
	out := fmt.Appendf(nil, "*%d\r\n", len(elements))
	for _, element := range elements {
		out = append(out, element...)
	}

	return out
}
