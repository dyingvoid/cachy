package commands

import (
	"errors"
	"reflect"
	"testing"
)

func TestParser_Parse_Get(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{
			name: "simple get",
			in:   "GET mykey",
		},
		{
			name: "get example value",
			in:   "GET value",
		},
		{
			name: "lowercase command",
			in:   "get mykey",
		},
	}

	p := NewProtocolParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := p.Parse(tt.in); err != nil {
				t.Fatalf("Parse(%q) returned unexpected error: %v", tt.in, err)
			}
		})
	}
}

func TestParser_Parse_Set(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{
			name: "set without expiry",
			in:   "SET mykey value",
		},
		{
			name: "set with EX seconds",
			in:   "SET mykey value EX 10",
		},
		{
			name: "set with PX milliseconds",
			in:   "SET mykey value PX 1000",
		},
		{
			name: "lowercase command and expiry flag",
			in:   "set mykey value ex 10",
		},
	}

	p := NewProtocolParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := p.Parse(tt.in); err != nil {
				t.Fatalf("Parse(%q) returned unexpected error: %v", tt.in, err)
			}
		})
	}
}

func TestParser_Parse_LRange(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want LRangeCmd
	}{
		{
			name: "range from zero",
			in:   "LRANGE list_key 0 1",
			want: LRangeCmd{Key: "list_key", Start: 0, Stop: 1},
		},
		{
			name: "range with equal indexes",
			in:   "LRANGE list_key 2 2",
			want: LRangeCmd{Key: "list_key", Start: 2, Stop: 2},
		},
		{
			name: "lowercase command",
			in:   "lrange list_key 2 4",
			want: LRangeCmd{Key: "list_key", Start: 2, Stop: 4},
		},
		{
			name: "negative start and stop",
			in:   "LRANGE list_key -2 -1",
			want: LRangeCmd{Key: "list_key", Start: -2, Stop: -1},
		},
		{
			name: "negative stop",
			in:   "LRANGE list_key 0 -3",
			want: LRangeCmd{Key: "list_key", Start: 0, Stop: -3},
		},
		{
			name: "negative start",
			in:   "LRANGE list_key -1 4",
			want: LRangeCmd{Key: "list_key", Start: -1, Stop: 4},
		},
	}

	p := NewProtocolParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := p.Parse(tt.in)
			if err != nil {
				t.Fatalf("Parse(%q) returned unexpected error: %v", tt.in, err)
			}
			if cmd != tt.want {
				t.Fatalf("Parse(%q) = %+v, want %+v", tt.in, cmd, tt.want)
			}
		})
	}
}

func TestParser_Parse_LPush(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want LPushCmd
	}{
		{
			name: "single value",
			in:   "LPUSH list_key a",
			want: LPushCmd{Key: "list_key", Values: []string{"a"}},
		},
		{
			name: "multiple values",
			in:   "LPUSH list_key a b c",
			want: LPushCmd{Key: "list_key", Values: []string{"a", "b", "c"}},
		},
		{
			name: "lowercase command",
			in:   "lpush list_key a b",
			want: LPushCmd{Key: "list_key", Values: []string{"a", "b"}},
		},
	}

	p := NewProtocolParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := p.Parse(tt.in)
			if err != nil {
				t.Fatalf("Parse(%q) returned unexpected error: %v", tt.in, err)
			}
			if !reflect.DeepEqual(cmd, tt.want) {
				t.Fatalf("Parse(%q) = %+v, want %+v", tt.in, cmd, tt.want)
			}
		})
	}
}

func TestParser_Parse_LLen(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want LLenCmd
	}{
		{
			name: "list key",
			in:   "LLEN list_key",
			want: LLenCmd{Key: "list_key"},
		},
		{
			name: "lowercase command",
			in:   "llen list_key",
			want: LLenCmd{Key: "list_key"},
		},
	}

	p := NewProtocolParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := p.Parse(tt.in)
			if err != nil {
				t.Fatalf("Parse(%q) returned unexpected error: %v", tt.in, err)
			}
			if cmd != tt.want {
				t.Fatalf("Parse(%q) = %+v, want %+v", tt.in, cmd, tt.want)
			}
		})
	}
}

func TestParser_Parse_LPop(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want LPopCmd
	}{
		{
			name: "single count",
			in:   "LPOP list_key 2",
			want: LPopCmd{Key: "list_key", Count: 2},
		},
		{
			name: "zero count",
			in:   "LPOP list_key 0",
			want: LPopCmd{Key: "list_key", Count: 0},
		},
		{
			name: "lowercase command",
			in:   "lpop list_key 1",
			want: LPopCmd{Key: "list_key", Count: 1},
		},
	}

	p := NewProtocolParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := p.Parse(tt.in)
			if err != nil {
				t.Fatalf("Parse(%q) returned unexpected error: %v", tt.in, err)
			}
			if cmd != tt.want {
				t.Fatalf("Parse(%q) = %+v, want %+v", tt.in, cmd, tt.want)
			}
		})
	}
}

func TestParser_Parse_Errors(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want error
	}{
		{
			name: "empty input",
			in:   "   ",
			want: ErrEmptyCommand,
		},
		{
			name: "unknown command",
			in:   "DEL mykey",
			want: ErrUnknownCommand,
		},
		{
			name: "get with no key",
			in:   "GET",
			want: ErrInvalidArguments,
		},
		{
			name: "get with too many args",
			in:   "GET mykey extra",
			want: ErrInvalidArguments,
		},
		{
			name: "set with no value",
			in:   "SET mykey",
			want: ErrInvalidArguments,
		},
		{
			name: "set with unknown expiry flag",
			in:   "SET mykey value XX 10",
			want: ErrInvalidArguments,
		},
		{
			name: "set with non numeric expiry",
			in:   "SET mykey value EX abc",
			want: ErrInvalidArguments,
		},
		{
			name: "set with non positive expiry",
			in:   "SET mykey value EX 0",
			want: ErrInvalidArguments,
		},
		{
			name: "lpush with no values",
			in:   "LPUSH mykey",
			want: ErrInvalidArguments,
		},
		{
			name: "lrange with no indexes",
			in:   "LRANGE mykey",
			want: ErrInvalidArguments,
		},
		{
			name: "lrange with one index",
			in:   "LRANGE mykey 0",
			want: ErrInvalidArguments,
		},
		{
			name: "lrange with too many args",
			in:   "LRANGE mykey 0 1 extra",
			want: ErrInvalidArguments,
		},
		{
			name: "lrange with non numeric start",
			in:   "LRANGE mykey a 1",
			want: ErrInvalidArguments,
		},
		{
			name: "lrange with non numeric stop",
			in:   "LRANGE mykey 0 b",
			want: ErrInvalidArguments,
		},
		{
			name: "llen with no key",
			in:   "LLEN",
			want: ErrInvalidArguments,
		},
		{
			name: "llen with too many args",
			in:   "LLEN mykey extra",
			want: ErrInvalidArguments,
		},
		{
			name: "lpop with no count",
			in:   "LPOP mykey",
			want: ErrInvalidArguments,
		},
		{
			name: "lpop with too many args",
			in:   "LPOP mykey 1 extra",
			want: ErrInvalidArguments,
		},
		{
			name: "lpop with non numeric count",
			in:   "LPOP mykey a",
			want: ErrInvalidArguments,
		},
		{
			name: "lpop with negative count",
			in:   "LPOP mykey -1",
			want: ErrInvalidArguments,
		},
	}

	p := NewProtocolParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := p.Parse(tt.in)
			if !errors.Is(err, tt.want) {
				t.Fatalf("Parse(%q) error = %v, want %v", tt.in, err, tt.want)
			}
		})
	}
}
