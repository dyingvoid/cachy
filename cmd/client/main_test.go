package main

import (
	"math/big"
	"testing"

	infrastructure_resp3 "github.com/dyingvoid/cachy/internal/infrastructure/resp3"
)

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name  string
		value infrastructure_resp3.Value
		want  string
	}{
		{
			name:  "simple string",
			value: infrastructure_resp3.Value{Kind: infrastructure_resp3.KindSimpleString, String: "PONG"},
			want:  "PONG",
		},
		{
			name:  "bulk string",
			value: infrastructure_resp3.Value{Kind: infrastructure_resp3.KindBulkString, String: "value"},
			want:  "value",
		},
		{
			name:  "simple error",
			value: infrastructure_resp3.Value{Kind: infrastructure_resp3.KindSimpleError, String: "ERR boom"},
			want:  "(error) ERR boom",
		},
		{
			name:  "bulk error",
			value: infrastructure_resp3.Value{Kind: infrastructure_resp3.KindBulkError, String: "ERR boom"},
			want:  "(error) ERR boom",
		},
		{
			name:  "integer",
			value: infrastructure_resp3.Value{Kind: infrastructure_resp3.KindInteger, Integer: -1},
			want:  "-1",
		},
		{
			name:  "boolean",
			value: infrastructure_resp3.Value{Kind: infrastructure_resp3.KindBoolean, Boolean: true},
			want:  "true",
		},
		{
			name:  "double",
			value: infrastructure_resp3.Value{Kind: infrastructure_resp3.KindDouble, Double: 1.5},
			want:  "1.5",
		},
		{
			name:  "big number",
			value: infrastructure_resp3.Value{Kind: infrastructure_resp3.KindBigNumber, BigNumber: big.NewInt(123)},
			want:  "123",
		},
		{
			name:  "null",
			value: infrastructure_resp3.Value{Kind: infrastructure_resp3.KindNull},
			want:  "(nil)",
		},
		{
			name: "verbatim string",
			value: infrastructure_resp3.Value{
				Kind:     infrastructure_resp3.KindVerbatimString,
				Encoding: "txt",
				String:   "hello",
			},
			want: "txt:hello",
		},
		{
			name: "array",
			value: infrastructure_resp3.Value{
				Kind: infrastructure_resp3.KindArray,
				Elements: []infrastructure_resp3.Value{
					{Kind: infrastructure_resp3.KindBulkString, String: "a"},
					{Kind: infrastructure_resp3.KindBulkString, String: "b"},
				},
			},
			want: "[a, b]",
		},
		{
			name: "map",
			value: infrastructure_resp3.Value{
				Kind: infrastructure_resp3.KindMap,
				Pairs: []infrastructure_resp3.Pair{
					{
						Key:   infrastructure_resp3.Value{Kind: infrastructure_resp3.KindBulkString, String: "role"},
						Value: infrastructure_resp3.Value{Kind: infrastructure_resp3.KindBulkString, String: "master"},
					},
				},
			},
			want: "{role => master}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatValue(tt.value); got != tt.want {
				t.Fatalf("formatValue() = %q, want %q", got, tt.want)
			}
		})
	}
}
