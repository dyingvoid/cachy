package server

import (
	"fmt"
	"strings"

	"github.com/dyingvoid/cachy/internal/domain"
)

type TextSerializer struct{}

func NewTextSerializer() *TextSerializer {
	return &TextSerializer{}
}

func (s *TextSerializer) Serialize(res domain.Result) string {
	switch res := res.(type) {
	case domain.StringResult:
		return string(res)
	case domain.ListResult:
		return strings.Join(res, "\n")
	default:
		return ""
	}
}

func (s *TextSerializer) SerializeError(err error) string {
	return fmt.Sprintf("(error) %s", err)
}
