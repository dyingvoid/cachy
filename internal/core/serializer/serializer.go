package core_serializer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dyingvoid/cachy/internal/features"
)

type TextSerializer struct{}

func NewTextSerializer() *TextSerializer {
	return &TextSerializer{}
}

func (s *TextSerializer) Serialize(res features.Result) string {
	switch res := res.(type) {
	case features.StringResult:
		return string(res)
	case features.ListResult:
		return strings.Join(res, "\n")
	case features.MapResult:
		return mapToString(res)
	default:
		panic(fmt.Sprintf("text serializer: unsupported result type %T, value: %#v", res, res))
	}
}

func (s *TextSerializer) SerializeError(err error) string {
	return fmt.Sprintf("(error) %s", err)
}

func mapToString(res features.MapResult) string {
	m := res.ToMap()
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteByte('\n')
		}
		fmt.Fprintf(&b, "%s:%s", k, m[k])
	}

	return b.String()
}
