package resp3

import (
	"bytes"
	"errors"
	"io"
	"math"
	"math/big"
	"strconv"
	"strings"
	"testing"
)

func TestDeserializer_RoundTrip(t *testing.T) {
	s := &Serializer{}
	must := mustBytes(t)

	streamedString := s.StartStreamedString()
	streamedStringBuf := append([]byte{}, must(streamedString.Chunk([]byte("Hello")))...)
	streamedStringBuf = append(streamedStringBuf, must(streamedString.Chunk([]byte(" World")))...)
	streamedStringBuf = append(streamedStringBuf, must(streamedString.Close())...)

	streamedArray := s.StartStreamedArray()
	streamedArrayBuf := append([]byte{}, must(streamedArray.Element(s.BulkString("a")))...)
	streamedArrayBuf = append(streamedArrayBuf, must(streamedArray.Element(s.Integer(1)))...)
	streamedArrayBuf = append(streamedArrayBuf, must(streamedArray.Close())...)

	streamedSet := s.StartStreamedSet()
	streamedSetBuf := append([]byte{}, must(streamedSet.Element(s.BulkString("a")))...)
	streamedSetBuf = append(streamedSetBuf, must(streamedSet.Close())...)

	streamedMap := s.StartStreamedMap()
	streamedMapBuf := append([]byte{}, must(streamedMap.Element(s.BulkString("a")))...)
	streamedMapBuf = append(streamedMapBuf, must(streamedMap.Element(s.Integer(1)))...)
	streamedMapBuf = append(streamedMapBuf, must(streamedMap.Close())...)

	attributes := s.Attributes(KeyValue{Key: s.BulkString("ttl"), Value: s.Integer(3600)})
	attributedValue := append(append([]byte{}, attributes...), must(s.SimpleString("OK"))...)

	tests := []struct {
		name string
		in   []byte
		want Value
	}{
		{
			name: "simple string",
			in:   must(s.SimpleString("OK")),
			want: Value{Kind: KindSimpleString, String: "OK"},
		},
		{
			name: "simple error",
			in:   must(s.SimpleError(errors.New("ERR boom"))),
			want: Value{Kind: KindSimpleError, String: "ERR boom"},
		},
		{
			name: "integer",
			in:   s.Integer(-42),
			want: Value{Kind: KindInteger, Integer: -42},
		},
		{
			name: "bulk string",
			in:   s.BulkString("hello"),
			want: Value{Kind: KindBulkString, String: "hello"},
		},
		{
			name: "empty bulk string",
			in:   s.BulkString(""),
			want: Value{Kind: KindBulkString, String: ""},
		},
		{
			name: "bulk string with embedded CRLF",
			in:   s.BulkString("a\r\nb"),
			want: Value{Kind: KindBulkString, String: "a\r\nb"},
		},
		{
			name: "null",
			in:   s.Null(),
			want: Value{Kind: KindNull},
		},
		{
			name: "array",
			in:   s.Array(s.BulkString("a"), s.Integer(1)),
			want: Value{
				Kind: KindArray,
				Elements: []Value{
					{Kind: KindBulkString, String: "a"},
					{Kind: KindInteger, Integer: 1},
				},
			},
		},
		{
			name: "empty array",
			in:   s.Array(),
			want: Value{Kind: KindArray},
		},
		{
			name: "nested array",
			in:   s.Array(s.Array(s.Integer(1), s.Integer(2)), s.Null()),
			want: Value{
				Kind: KindArray,
				Elements: []Value{
					{
						Kind: KindArray,
						Elements: []Value{
							{Kind: KindInteger, Integer: 1},
							{Kind: KindInteger, Integer: 2},
						},
					},
					{Kind: KindNull},
				},
			},
		},
		{
			name: "boolean true",
			in:   s.Boolean(true),
			want: Value{Kind: KindBoolean, Boolean: true},
		},
		{
			name: "boolean false",
			in:   s.Boolean(false),
			want: Value{Kind: KindBoolean, Boolean: false},
		},
		{
			name: "double",
			in:   s.Double(1.5),
			want: Value{Kind: KindDouble, Double: 1.5},
		},
		{
			name: "double positive infinity",
			in:   s.Double(math.Inf(1)),
			want: Value{Kind: KindDouble, Double: math.Inf(1)},
		},
		{
			name: "double negative infinity",
			in:   s.Double(math.Inf(-1)),
			want: Value{Kind: KindDouble, Double: math.Inf(-1)},
		},
		{
			name: "double not a number",
			in:   s.Double(math.NaN()),
			want: Value{Kind: KindDouble, Double: math.NaN()},
		},
		{
			name: "big number",
			in:   must(s.BigNumber(big.NewInt(123456789))),
			want: Value{Kind: KindBigNumber, BigNumber: big.NewInt(123456789)},
		},
		{
			name: "bulk error",
			in:   s.BulkError("ERR oops"),
			want: Value{Kind: KindBulkError, String: "ERR oops"},
		},
		{
			name: "verbatim string",
			in:   must(s.VerbatimString("txt", "hello")),
			want: Value{Kind: KindVerbatimString, Encoding: "txt", String: "hello"},
		},
		{
			name: "map",
			in: must(s.Map(
				KeyValue{Key: s.BulkString("a"), Value: s.Integer(1)},
				KeyValue{Key: s.BulkString("b"), Value: s.Boolean(true)},
			)),
			want: Value{
				Kind: KindMap,
				Pairs: []Pair{
					{
						Key:   Value{Kind: KindBulkString, String: "a"},
						Value: Value{Kind: KindInteger, Integer: 1},
					},
					{
						Key:   Value{Kind: KindBulkString, String: "b"},
						Value: Value{Kind: KindBoolean, Boolean: true},
					},
				},
			},
		},
		{
			name: "empty map",
			in:   must(s.Map()),
			want: Value{Kind: KindMap},
		},
		{
			name: "set",
			in:   s.Set(s.BulkString("a"), s.BulkString("b")),
			want: Value{
				Kind: KindSet,
				Elements: []Value{
					{Kind: KindBulkString, String: "a"},
					{Kind: KindBulkString, String: "b"},
				},
			},
		},
		{
			name: "push",
			in:   s.Push(s.BulkString("message"), s.BulkString("payload")),
			want: Value{
				Kind: KindPush,
				Elements: []Value{
					{Kind: KindBulkString, String: "message"},
					{Kind: KindBulkString, String: "payload"},
				},
			},
		},
		{
			name: "streamed string",
			in:   streamedStringBuf,
			want: Value{Kind: KindBulkString, String: "Hello World", Streamed: true},
		},
		{
			name: "streamed array",
			in:   streamedArrayBuf,
			want: Value{
				Kind: KindArray,
				Elements: []Value{
					{Kind: KindBulkString, String: "a"},
					{Kind: KindInteger, Integer: 1},
				},
				Streamed: true,
			},
		},
		{
			name: "streamed set",
			in:   streamedSetBuf,
			want: Value{
				Kind:     KindSet,
				Elements: []Value{{Kind: KindBulkString, String: "a"}},
				Streamed: true,
			},
		},
		{
			name: "streamed map",
			in:   streamedMapBuf,
			want: Value{
				Kind: KindMap,
				Pairs: []Pair{
					{
						Key:   Value{Kind: KindBulkString, String: "a"},
						Value: Value{Kind: KindInteger, Integer: 1},
					},
				},
				Streamed: true,
			},
		},
		{
			name: "attributes on simple string",
			in:   attributedValue,
			want: Value{
				Kind:   KindSimpleString,
				String: "OK",
				Attributes: []Pair{
					{
						Key:   Value{Kind: KindBulkString, String: "ttl"},
						Value: Value{Kind: KindInteger, Integer: 3600},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDeserializer(bytes.NewReader(tt.in))
			got, err := d.Read()
			if err != nil {
				t.Fatalf("Read(%q) returned unexpected error: %v", tt.in, err)
			}
			assertValue(t, got, tt.want)
		})
	}
}

func TestDeserializer_LegacyNull(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{name: "null type", in: "_\r\n"},
		{name: "null bulk string", in: "$-1\r\n"},
		{name: "null array", in: "*-1\r\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDeserializer(strings.NewReader(tt.in))
			got, err := d.Read()
			if err != nil {
				t.Fatalf("Read(%q) returned unexpected error: %v", tt.in, err)
			}
			assertValue(t, got, Value{Kind: KindNull})
		})
	}
}

func TestDeserializer_MultipleReads(t *testing.T) {
	d := NewDeserializer(strings.NewReader("+OK\r\n:42\r\n"))

	got, err := d.Read()
	if err != nil {
		t.Fatalf("first Read returned unexpected error: %v", err)
	}
	assertValue(t, got, Value{Kind: KindSimpleString, String: "OK"})

	got, err = d.Read()
	if err != nil {
		t.Fatalf("second Read returned unexpected error: %v", err)
	}
	assertValue(t, got, Value{Kind: KindInteger, Integer: 42})

	if _, err := d.Read(); !errors.Is(err, io.EOF) {
		t.Fatalf("third Read error = %v, want %v", err, io.EOF)
	}
}

func TestDeserializer_MultipleAttributeBlocks(t *testing.T) {
	in := "|1\r\n+ttl\r\n:3600\r\n|1\r\n+flag\r\n#t\r\n*2\r\n:1\r\n:2\r\n"
	want := Value{
		Kind: KindArray,
		Elements: []Value{
			{Kind: KindInteger, Integer: 1},
			{Kind: KindInteger, Integer: 2},
		},
		Attributes: []Pair{
			{
				Key:   Value{Kind: KindSimpleString, String: "ttl"},
				Value: Value{Kind: KindInteger, Integer: 3600},
			},
			{
				Key:   Value{Kind: KindSimpleString, String: "flag"},
				Value: Value{Kind: KindBoolean, Boolean: true},
			},
		},
	}

	d := NewDeserializer(strings.NewReader(in))
	got, err := d.Read()
	if err != nil {
		t.Fatalf("Read returned unexpected error: %v", err)
	}
	assertValue(t, got, want)
}

func TestDeserializer_AttributesOnNestedValue(t *testing.T) {
	in := "*2\r\n|1\r\n+k\r\n+v\r\n:5\r\n:6\r\n"
	want := Value{
		Kind: KindArray,
		Elements: []Value{
			{
				Kind:    KindInteger,
				Integer: 5,
				Attributes: []Pair{
					{
						Key:   Value{Kind: KindSimpleString, String: "k"},
						Value: Value{Kind: KindSimpleString, String: "v"},
					},
				},
			},
			{Kind: KindInteger, Integer: 6},
		},
	}

	d := NewDeserializer(strings.NewReader(in))
	got, err := d.Read()
	if err != nil {
		t.Fatalf("Read returned unexpected error: %v", err)
	}
	assertValue(t, got, want)
}

func TestDeserializer_EmptyStreamedValues(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want Value
	}{
		{
			name: "streamed string",
			in:   "$?\r\n;0\r\n",
			want: Value{Kind: KindBulkString, String: "", Streamed: true},
		},
		{
			name: "streamed array",
			in:   "*?\r\n.\r\n",
			want: Value{Kind: KindArray, Streamed: true},
		},
		{
			name: "streamed set",
			in:   "~?\r\n.\r\n",
			want: Value{Kind: KindSet, Streamed: true},
		},
		{
			name: "streamed map",
			in:   "%?\r\n.\r\n",
			want: Value{Kind: KindMap, Streamed: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDeserializer(strings.NewReader(tt.in))
			got, err := d.Read()
			if err != nil {
				t.Fatalf("Read(%q) returned unexpected error: %v", tt.in, err)
			}
			assertValue(t, got, tt.want)
		})
	}
}

func TestDeserializer_Errors(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want error
	}{
		{name: "empty input", in: "", want: io.EOF},
		{name: "truncated simple string", in: "+OK", want: io.ErrUnexpectedEOF},
		{name: "invalid simple string terminator", in: "+OK\n", want: errInvalidTerminator},
		{name: "simple string with embedded CR", in: "+a\rb\r\n", want: errContainsCRLF},
		{name: "invalid integer", in: ":abc\r\n", want: errInvalidInteger},
		{name: "invalid bulk length", in: "$abc\r\n", want: errInvalidLength},
		{name: "negative bulk length", in: "$-2\r\n", want: errInvalidLength},
		{
			name: "bulk length too large",
			in:   "$" + strconv.Itoa(maxBulkLength+1) + "\r\n",
			want: errLengthTooLarge,
		},
		{name: "truncated bulk payload", in: "$5\r\nab", want: io.ErrUnexpectedEOF},
		{name: "invalid bulk terminator", in: "$3\r\nabcXY", want: errInvalidTerminator},
		{name: "invalid array length", in: "*abc\r\n", want: errInvalidLength},
		{name: "negative map length", in: "%-2\r\n", want: errInvalidLength},
		{name: "negative set length", in: "~-1\r\n", want: errInvalidLength},
		{
			name: "array length too large",
			in:   "*" + strconv.Itoa(maxAggregateLength+1) + "\r\n",
			want: errLengthTooLarge,
		},
		{name: "invalid null", in: "_x\r\n", want: errInvalidNull},
		{name: "invalid boolean", in: "#x\r\n", want: errInvalidBoolean},
		{name: "invalid double", in: ",abc\r\n", want: errInvalidDouble},
		{name: "invalid big number", in: "(abc\r\n", want: errInvalidBigNumber},
		{name: "invalid big number fraction", in: "(1.5\r\n", want: errInvalidBigNumber},
		{name: "truncated bulk error", in: "!5\r\nabc", want: io.ErrUnexpectedEOF},
		{name: "invalid verbatim string", in: "=3\r\ntxt\r\n", want: errInvalidVerbatim},
		{name: "verbatim string missing separator", in: "=4\r\nabcX\r\n", want: errInvalidVerbatim},
		{name: "unexpected type byte", in: "@foo\r\n", want: errUnexpectedType},
		{name: "attributes without value", in: "|1\r\n+a\r\n+b\r\n", want: io.ErrUnexpectedEOF},
		{name: "truncated attributes", in: "|1\r\n+a\r\n", want: io.ErrUnexpectedEOF},
		{name: "truncated array", in: "*2\r\n:1\r\n", want: io.ErrUnexpectedEOF},
		{name: "truncated streamed string", in: "$?\r\n;5\r\nHello\r\n", want: io.ErrUnexpectedEOF},
		{name: "invalid streamed string chunk", in: "$?\r\n;x\r\n", want: errInvalidLength},
		{name: "truncated streamed array", in: "*?\r\n:1\r\n", want: io.ErrUnexpectedEOF},
		{
			name: "invalid streamed aggregate terminator",
			in:   "*?\r\n:1\r\n.bad\r\n",
			want: errInvalidTerminator,
		},
		{name: "streamed map with odd elements", in: "%?\r\n+a\r\n.\r\n", want: errOddMapElements},
		{
			name: "nesting too deep",
			in:   strings.Repeat("*1\r\n", maxDepth+2) + "+x\r\n",
			want: errTooDeep,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDeserializer(strings.NewReader(tt.in))
			got, err := d.Read()
			if !errors.Is(err, tt.want) {
				t.Fatalf("Read(%q) error = %v, want %v", tt.in, err, tt.want)
			}
			if got.Kind != 0 {
				t.Fatalf("Read(%q) = %+v, want zero Value", tt.in, got)
			}
		})
	}
}

func mustBytes(t *testing.T) func([]byte, error) []byte {
	t.Helper()
	return func(b []byte, err error) []byte {
		t.Helper()
		if err != nil {
			t.Fatalf("serialize returned unexpected error: %v", err)
		}
		return b
	}
}

func assertValue(t *testing.T, got, want Value) {
	t.Helper()
	if !valuesEqual(got, want) {
		t.Fatalf("value = %+v, want %+v", got, want)
	}
}

func valuesEqual(got, want Value) bool {
	if got.Kind != want.Kind ||
		got.String != want.String ||
		got.Integer != want.Integer ||
		got.Boolean != want.Boolean ||
		!floatsEqual(got.Double, want.Double) ||
		!bigNumbersEqual(got.BigNumber, want.BigNumber) ||
		got.Encoding != want.Encoding ||
		got.Streamed != want.Streamed {
		return false
	}

	if !valuesEqualSlice(got.Elements, want.Elements) ||
		!pairsEqual(got.Pairs, want.Pairs) ||
		!pairsEqual(got.Attributes, want.Attributes) {
		return false
	}

	return true
}

func valuesEqualSlice(got, want []Value) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if !valuesEqual(got[i], want[i]) {
			return false
		}
	}
	return true
}

func pairsEqual(got, want []Pair) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if !valuesEqual(got[i].Key, want[i].Key) ||
			!valuesEqual(got[i].Value, want[i].Value) {
			return false
		}
	}
	return true
}

func floatsEqual(got, want float64) bool {
	if math.IsNaN(got) && math.IsNaN(want) {
		return true
	}
	return got == want
}

func bigNumbersEqual(got, want *big.Int) bool {
	if got == nil || want == nil {
		return got == nil && want == nil
	}
	return got.Cmp(want) == 0
}
