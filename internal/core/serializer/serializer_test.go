package core_serializer

import (
	"testing"
	"time"

	"github.com/dyingvoid/cachy/internal/features"
	features_commands "github.com/dyingvoid/cachy/internal/features/commands"
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

func TestTextSerializerSerializeCommand(t *testing.T) {
	seconds := 10 * time.Second
	milliseconds := 1500 * time.Millisecond

	tests := []struct {
		name    string
		command features_commands.WriteCommand
		want    string
	}{
		{
			name:    "set",
			command: features_commands.SetCmd{Key: "key", Value: "value"},
			want:    "SET key value",
		},
		{
			name:    "set with seconds",
			command: features_commands.SetCmd{Key: "key", Value: "value", Expiry: &seconds},
			want:    "SET key value EX 10",
		},
		{
			name:    "set with milliseconds",
			command: features_commands.SetCmd{Key: "key", Value: "value", Expiry: &milliseconds},
			want:    "SET key value PX 1500",
		},
		{
			name:    "rpush",
			command: features_commands.RPushCmd{Key: "key", Values: []string{"a", "b"}},
			want:    "RPUSH key a b",
		},
		{
			name:    "lpush",
			command: features_commands.LPushCmd{Key: "key", Values: []string{"a", "b"}},
			want:    "LPUSH key a b",
		},
		{
			name:    "lpop",
			command: features_commands.LPopCmd{Key: "key", Count: 2},
			want:    "LPOP key 2",
		},
	}

	s := NewTextSerializer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s.SerializeCommand(tt.command); got != tt.want {
				t.Fatalf("SerializeCommand() = %q, want %q", got, tt.want)
			}
		})
	}
}
