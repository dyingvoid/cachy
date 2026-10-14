package server

import (
	"strings"

	"github.com/dyingvoid/cachy/internal/domain"
)

type TextResultSerializer struct{}

func NewTextResultSerializer() *TextResultSerializer {
	return &TextResultSerializer{}
}

func (s *TextResultSerializer) Serialize(res domain.Result) string {
	switch res := res.(type) {
	case domain.StringResult:
		return string(res)
	case domain.ListResult:
		return strings.Join(res, "\n")
	default:
		return ""
	}
}
