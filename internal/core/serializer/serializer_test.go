package core_serializer

import (
	"testing"

	"github.com/dyingvoid/cachy/internal/features"
)

func TestTextSerializerSerialize(t *testing.T) {
	tests := []struct {
		name   string
		result features.Result
		want   string
	}{
		{
			name:   "string",
			result: features.StringResult("PONG"),
			want:   "PONG",
		},
		{
			name:   "empty file",
			result: features.FileResult{Len: 0, Content: []byte{}},
			want:   "$0\r\n",
		},
		{
			name:   "file with content",
			result: features.FileResult{Len: 3, Content: []byte("abc")},
			want:   "$3\r\nabc",
		},
	}

	s := NewTextSerializer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s.Serialize(tt.result); got != tt.want {
				t.Fatalf("Serialize() = %q, want %q", got, tt.want)
			}
		})
	}
}
