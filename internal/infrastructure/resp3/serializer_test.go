package resp3

import (
	"errors"
	"math"
	"math/big"
	"testing"
)

func TestSerializer_SimpleString(t *testing.T) {
	s := &Serializer{}

	got, err := s.SimpleString("PONG")
	if err != nil {
		t.Fatalf("SimpleString returned unexpected error: %v", err)
	}
	if want := "+PONG\r\n"; string(got) != want {
		t.Fatalf("SimpleString = %q, want %q", got, want)
	}

	got, err = s.SimpleString("")
	if err != nil {
		t.Fatalf("SimpleString returned unexpected error: %v", err)
	}
	if want := "+\r\n"; string(got) != want {
		t.Fatalf("SimpleString empty = %q, want %q", got, want)
	}
}

func TestSerializer_SimpleStringRejectsCRLF(t *testing.T) {
	s := &Serializer{}

	tests := []struct {
		name string
		in   string
	}{
		{name: "carriage return", in: "a\rb"},
		{name: "line feed", in: "a\nb"},
		{name: "trailing CRLF", in: "a\r\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.SimpleString(tt.in)
			if !errors.Is(err, errContainsCRLF) {
				t.Fatalf("SimpleString(%q) error = %v, want %v", tt.in, err, errContainsCRLF)
			}
			if got != nil {
				t.Fatalf("SimpleString(%q) = %q, want nil", tt.in, got)
			}
		})
	}
}

func TestSerializer_SimpleError(t *testing.T) {
	s := &Serializer{}

	got, err := s.SimpleError(errors.New("ERR boom"))
	if err != nil {
		t.Fatalf("SimpleError returned unexpected error: %v", err)
	}
	if want := "-ERR boom\r\n"; string(got) != want {
		t.Fatalf("SimpleError = %q, want %q", got, want)
	}
}

func TestSerializer_SimpleErrorErrors(t *testing.T) {
	s := &Serializer{}

	tests := []struct {
		name string
		in   error
	}{
		{name: "nil error", in: nil},
		{name: "carriage return", in: errors.New("a\rb")},
		{name: "line feed", in: errors.New("a\nb")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.SimpleError(tt.in)
			if err == nil {
				t.Fatalf("SimpleError(%v) returned nil error", tt.in)
			}
			if got != nil {
				t.Fatalf("SimpleError(%v) = %q, want nil", tt.in, got)
			}
		})
	}
}

func TestSerializer_Integer(t *testing.T) {
	s := &Serializer{}

	tests := []struct {
		name string
		in   int64
		want string
	}{
		{name: "zero", in: 0, want: ":0\r\n"},
		{name: "positive", in: 42, want: ":42\r\n"},
		{name: "negative", in: -42, want: ":-42\r\n"},
		{name: "max int64", in: math.MaxInt64, want: ":9223372036854775807\r\n"},
		{name: "min int64", in: math.MinInt64, want: ":-9223372036854775808\r\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s.Integer(tt.in); string(got) != tt.want {
				t.Fatalf("Integer(%d) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestSerializer_BulkString(t *testing.T) {
	s := &Serializer{}

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "text", in: "hello", want: "$5\r\nhello\r\n"},
		{name: "empty", in: "", want: "$0\r\n\r\n"},
		{name: "multibyte", in: "héllo", want: "$6\r\nhéllo\r\n"},
		{name: "embedded CRLF", in: "a\r\nb", want: "$4\r\na\r\nb\r\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s.BulkString(tt.in); string(got) != tt.want {
				t.Fatalf("BulkString(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestSerializer_Null(t *testing.T) {
	s := &Serializer{}

	if got, want := string(s.Null()), "_\r\n"; got != want {
		t.Fatalf("Null = %q, want %q", got, want)
	}
}

func TestSerializer_Array(t *testing.T) {
	s := &Serializer{}

	tests := []struct {
		name string
		in   [][]byte
		want string
	}{
		{name: "empty", in: nil, want: "*0\r\n"},
		{
			name: "mixed elements",
			in:   [][]byte{s.BulkString("a"), s.Integer(1), s.Boolean(true)},
			want: "*3\r\n$1\r\na\r\n:1\r\n#t\r\n",
		},
		{
			name: "nested array",
			in:   [][]byte{s.Array(s.BulkString("a")), s.Null()},
			want: "*2\r\n*1\r\n$1\r\na\r\n_\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s.Array(tt.in...); string(got) != tt.want {
				t.Fatalf("Array = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSerializer_Boolean(t *testing.T) {
	s := &Serializer{}

	if got, want := string(s.Boolean(true)), "#t\r\n"; got != want {
		t.Fatalf("Boolean(true) = %q, want %q", got, want)
	}
	if got, want := string(s.Boolean(false)), "#f\r\n"; got != want {
		t.Fatalf("Boolean(false) = %q, want %q", got, want)
	}
}

func TestSerializer_Double(t *testing.T) {
	s := &Serializer{}

	tests := []struct {
		name string
		in   float64
		want string
	}{
		{name: "fraction", in: 1.5, want: ",1.5\r\n"},
		{name: "whole number", in: 3, want: ",3\r\n"},
		{name: "zero", in: 0, want: ",0\r\n"},
		{name: "negative", in: -2.25, want: ",-2.25\r\n"},
		{name: "positive infinity", in: math.Inf(1), want: ",inf\r\n"},
		{name: "negative infinity", in: math.Inf(-1), want: ",-inf\r\n"},
		{name: "not a number", in: math.NaN(), want: ",nan\r\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s.Double(tt.in); string(got) != tt.want {
				t.Fatalf("Double(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestSerializer_BigNumber(t *testing.T) {
	s := &Serializer{}

	got, err := s.BigNumber(big.NewInt(123456789))
	if err != nil {
		t.Fatalf("BigNumber returned unexpected error: %v", err)
	}
	if want := "(123456789\r\n"; string(got) != want {
		t.Fatalf("BigNumber = %q, want %q", got, want)
	}

	got, err = s.BigNumber(big.NewInt(-42))
	if err != nil {
		t.Fatalf("BigNumber returned unexpected error: %v", err)
	}
	if want := "(-42\r\n"; string(got) != want {
		t.Fatalf("BigNumber negative = %q, want %q", got, want)
	}
}

func TestSerializer_BigNumberNil(t *testing.T) {
	s := &Serializer{}

	got, err := s.BigNumber(nil)
	if err == nil {
		t.Fatalf("BigNumber(nil) returned nil error")
	}
	if got != nil {
		t.Fatalf("BigNumber(nil) = %q, want nil", got)
	}
}

func TestSerializer_BulkError(t *testing.T) {
	s := &Serializer{}

	got := s.BulkError("ERR oops")
	if want := "!8\r\nERR oops\r\n"; string(got) != want {
		t.Fatalf("BulkError = %q, want %q", got, want)
	}
}

func TestSerializer_VerbatimString(t *testing.T) {
	s := &Serializer{}

	got, err := s.VerbatimString("txt", "hello")
	if err != nil {
		t.Fatalf("VerbatimString returned unexpected error: %v", err)
	}
	if want := "=9\r\ntxt:hello\r\n"; string(got) != want {
		t.Fatalf("VerbatimString = %q, want %q", got, want)
	}

	got, err = s.VerbatimString("mkd", "")
	if err != nil {
		t.Fatalf("VerbatimString returned unexpected error: %v", err)
	}
	if want := "=4\r\nmkd:\r\n"; string(got) != want {
		t.Fatalf("VerbatimString empty = %q, want %q", got, want)
	}
}

func TestSerializer_VerbatimStringRejectsInvalidEncoding(t *testing.T) {
	s := &Serializer{}

	tests := []struct{ name, encoding string }{
		{name: "too short", encoding: "tx"},
		{name: "too long", encoding: "txtx"},
		{name: "empty", encoding: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.VerbatimString(tt.encoding, "hello")
			if err == nil {
				t.Fatalf("VerbatimString(%q) returned nil error", tt.encoding)
			}
			if got != nil {
				t.Fatalf("VerbatimString(%q) = %q, want nil", tt.encoding, got)
			}
		})
	}
}

func TestSerializer_Map(t *testing.T) {
	s := &Serializer{}

	got, err := s.Map(
		KeyValue{Key: s.BulkString("a"), Value: s.Integer(1)},
		KeyValue{Key: s.BulkString("b"), Value: s.Boolean(true)},
	)
	if err != nil {
		t.Fatalf("Map returned unexpected error: %v", err)
	}
	if want := "%2\r\n$1\r\na\r\n:1\r\n$1\r\nb\r\n#t\r\n"; string(got) != want {
		t.Fatalf("Map = %q, want %q", got, want)
	}

	got, err = s.Map()
	if err != nil {
		t.Fatalf("Map returned unexpected error: %v", err)
	}
	if want := "%0\r\n"; string(got) != want {
		t.Fatalf("empty Map = %q, want %q", got, want)
	}
}

func TestSerializer_Attributes(t *testing.T) {
	s := &Serializer{}

	got := s.Attributes(
		KeyValue{Key: s.BulkString("ttl"), Value: s.Integer(3600)},
	)
	if want := "|1\r\n$3\r\nttl\r\n:3600\r\n"; string(got) != want {
		t.Fatalf("Attributes = %q, want %q", got, want)
	}

	if got, want := string(s.Attributes()), "|0\r\n"; got != want {
		t.Fatalf("empty Attributes = %q, want %q", got, want)
	}
}

func TestSerializer_Set(t *testing.T) {
	s := &Serializer{}

	got := s.Set(s.BulkString("a"), s.BulkString("b"))
	if want := "~2\r\n$1\r\na\r\n$1\r\nb\r\n"; string(got) != want {
		t.Fatalf("Set = %q, want %q", got, want)
	}

	if got, want := string(s.Set()), "~0\r\n"; got != want {
		t.Fatalf("empty Set = %q, want %q", got, want)
	}
}

func TestSerializer_Push(t *testing.T) {
	s := &Serializer{}

	got := s.Push(s.BulkString("message"), s.BulkString("payload"))
	if want := ">2\r\n$7\r\nmessage\r\n$7\r\npayload\r\n"; string(got) != want {
		t.Fatalf("Push = %q, want %q", got, want)
	}

	if got, want := string(s.Push()), ">0\r\n"; got != want {
		t.Fatalf("empty Push = %q, want %q", got, want)
	}
}

func TestSerializer_StreamedString(t *testing.T) {
	s := &Serializer{}
	stream := s.StartStreamedString()

	got, err := stream.Chunk([]byte("Hello"))
	if err != nil {
		t.Fatalf("Chunk returned unexpected error: %v", err)
	}
	if want := "$?\r\n;5\r\nHello\r\n"; string(got) != want {
		t.Fatalf("first Chunk = %q, want %q", got, want)
	}

	got, err = stream.Chunk([]byte(" World"))
	if err != nil {
		t.Fatalf("Chunk returned unexpected error: %v", err)
	}
	if want := ";6\r\n World\r\n"; string(got) != want {
		t.Fatalf("second Chunk = %q, want %q", got, want)
	}

	got, err = stream.Close()
	if err != nil {
		t.Fatalf("Close returned unexpected error: %v", err)
	}
	if want := ";0\r\n"; string(got) != want {
		t.Fatalf("Close = %q, want %q", got, want)
	}
}

func TestSerializer_StreamedStringClosedBeforeChunk(t *testing.T) {
	s := &Serializer{}
	stream := s.StartStreamedString()

	got, err := stream.Close()
	if err != nil {
		t.Fatalf("Close returned unexpected error: %v", err)
	}
	if want := "$?\r\n;0\r\n"; string(got) != want {
		t.Fatalf("Close before Chunk = %q, want %q", got, want)
	}
}

func TestSerializer_StreamedStringRejectsAfterClose(t *testing.T) {
	s := &Serializer{}
	stream := s.StartStreamedString()

	if _, err := stream.Chunk([]byte("Hello")); err != nil {
		t.Fatalf("Chunk returned unexpected error: %v", err)
	}
	if _, err := stream.Close(); err != nil {
		t.Fatalf("Close returned unexpected error: %v", err)
	}

	if got, err := stream.Chunk([]byte("more")); err == nil {
		t.Fatalf("Chunk after Close = %q, want error", got)
	}
	if got, err := stream.Close(); err == nil {
		t.Fatalf("second Close = %q, want error", got)
	}
}

func TestSerializer_StreamedArray(t *testing.T) {
	s := &Serializer{}
	stream := s.StartStreamedArray()

	got, err := stream.Element(s.BulkString("a"))
	if err != nil {
		t.Fatalf("Element returned unexpected error: %v", err)
	}
	if want := "*?\r\n$1\r\na\r\n"; string(got) != want {
		t.Fatalf("first Element = %q, want %q", got, want)
	}

	got, err = stream.Element(s.Integer(1))
	if err != nil {
		t.Fatalf("Element returned unexpected error: %v", err)
	}
	if want := ":1\r\n"; string(got) != want {
		t.Fatalf("second Element = %q, want %q", got, want)
	}

	got, err = stream.Close()
	if err != nil {
		t.Fatalf("Close returned unexpected error: %v", err)
	}
	if want := ".\r\n"; string(got) != want {
		t.Fatalf("Close = %q, want %q", got, want)
	}
}

func TestSerializer_StreamedSet(t *testing.T) {
	s := &Serializer{}
	stream := s.StartStreamedSet()

	got, err := stream.Element(s.BulkString("a"))
	if err != nil {
		t.Fatalf("Element returned unexpected error: %v", err)
	}
	if want := "~?\r\n$1\r\na\r\n"; string(got) != want {
		t.Fatalf("Element = %q, want %q", got, want)
	}
}

func TestSerializer_StreamedMap(t *testing.T) {
	s := &Serializer{}
	stream := s.StartStreamedMap()

	got, err := stream.Element(s.BulkString("a"))
	if err != nil {
		t.Fatalf("Element returned unexpected error: %v", err)
	}
	if want := "%?\r\n$1\r\na\r\n"; string(got) != want {
		t.Fatalf("key Element = %q, want %q", got, want)
	}

	got, err = stream.Element(s.Integer(1))
	if err != nil {
		t.Fatalf("Element returned unexpected error: %v", err)
	}
	if want := ":1\r\n"; string(got) != want {
		t.Fatalf("value Element = %q, want %q", got, want)
	}

	got, err = stream.Close()
	if err != nil {
		t.Fatalf("Close returned unexpected error: %v", err)
	}
	if want := ".\r\n"; string(got) != want {
		t.Fatalf("Close = %q, want %q", got, want)
	}
}

func TestSerializer_StreamedAggregateClosedBeforeElement(t *testing.T) {
	s := &Serializer{}

	got, err := s.StartStreamedArray().Close()
	if err != nil {
		t.Fatalf("Close returned unexpected error: %v", err)
	}
	if want := "*?\r\n.\r\n"; string(got) != want {
		t.Fatalf("empty streamed array = %q, want %q", got, want)
	}

	got, err = s.StartStreamedMap().Close()
	if err != nil {
		t.Fatalf("Close returned unexpected error: %v", err)
	}
	if want := "%?\r\n.\r\n"; string(got) != want {
		t.Fatalf("empty streamed map = %q, want %q", got, want)
	}
}

func TestSerializer_StreamedMapRejectsOddElements(t *testing.T) {
	s := &Serializer{}
	stream := s.StartStreamedMap()

	if _, err := stream.Element(s.BulkString("a")); err != nil {
		t.Fatalf("Element returned unexpected error: %v", err)
	}

	got, err := stream.Close()
	if err == nil {
		t.Fatalf("Close with odd elements = %q, want error", got)
	}

	if _, err := stream.Element(s.Integer(1)); err != nil {
		t.Fatalf("Element after failed Close returned unexpected error: %v", err)
	}

	got, err = stream.Close()
	if err != nil {
		t.Fatalf("Close returned unexpected error: %v", err)
	}
	if want := ".\r\n"; string(got) != want {
		t.Fatalf("Close = %q, want %q", got, want)
	}
}

func TestSerializer_StreamedAggregateRejectsAfterClose(t *testing.T) {
	s := &Serializer{}
	stream := s.StartStreamedArray()

	if _, err := stream.Close(); err != nil {
		t.Fatalf("Close returned unexpected error: %v", err)
	}

	if got, err := stream.Element(s.Integer(1)); err == nil {
		t.Fatalf("Element after Close = %q, want error", got)
	}
	if got, err := stream.Close(); err == nil {
		t.Fatalf("second Close = %q, want error", got)
	}
}
