package resp3

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dyingvoid/cachy/internal/features"
	features_commands "github.com/dyingvoid/cachy/internal/features/commands"
)

var (
	errContainsCRLF = errors.New("resp: simple string must not contain CR or LF")
)

type Serializer struct {
}

func (s *Serializer) Serialize(res features.Result) ([]byte, error) {
	switch res := res.(type) {
	case features.StringResult:
		return s.SimpleString(string(res))
	case features.BulkStringResult:
		return s.BulkString(string(res)), nil
	case features.IntResult:
		return s.Integer(int64(res)), nil
	case features.ListResult:
		elements := make([][]byte, 0, len(res))
		for _, element := range res {
			elements = append(elements, s.BulkString(element))
		}
		return s.Array(elements...), nil
	case features.MapResult:
		m := res.ToMap()
		keys := make([]string, 0, len(m))
		for key := range m {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		pairs := make([]KeyValue, 0, len(keys))
		for _, key := range keys {
			pairs = append(pairs, KeyValue{
				Key:   s.BulkString(key),
				Value: s.BulkString(m[key]),
			})
		}
		return s.Map(pairs...)
	case features.FileResult:
		return s.BulkString(string(res.Content)), nil
	default:
		return nil, fmt.Errorf("resp: unsupported result type %T", res)
	}
}

func (s *Serializer) SerializeCommand(cmd features_commands.WriteCommand) []byte {
	var fields []string
	switch cmd := cmd.(type) {
	case features_commands.SetCmd:
		fields = []string{"SET", cmd.Key, cmd.Value}
		if cmd.Expiry != nil {
			fields = append(fields, expiryFields(*cmd.Expiry)...)
		}
	case features_commands.LPushCmd:
		fields = make([]string, 0, len(cmd.Values)+2)
		fields = append(fields, "LPUSH", cmd.Key)
		fields = append(fields, cmd.Values...)
	case features_commands.RPushCmd:
		fields = make([]string, 0, len(cmd.Values)+2)
		fields = append(fields, "RPUSH", cmd.Key)
		fields = append(fields, cmd.Values...)
	case features_commands.LPopCmd:
		fields = []string{"LPOP", cmd.Key, strconv.Itoa(cmd.Count)}
	default:
		panic(fmt.Sprintf("resp serializer: unsupported command type %T, value: %#v", cmd, cmd))
	}

	elements := make([][]byte, 0, len(fields))
	for _, field := range fields {
		elements = append(elements, s.BulkString(field))
	}

	return s.Array(elements...)
}

func expiryFields(expiry time.Duration) []string {
	if expiry%time.Second == 0 {
		return []string{"EX", strconv.FormatInt(int64(expiry/time.Second), 10)}
	}

	return []string{"PX", strconv.FormatInt(expiry.Milliseconds(), 10)}
}

func (s *Serializer) SimpleString(str string) ([]byte, error) {
	if strings.ContainsAny(str, "\r\n") {
		return nil, errContainsCRLF
	}
	return fmt.Appendf(nil, "+%s\r\n", str), nil
}

func (s *Serializer) SimpleError(err error) ([]byte, error) {
	if err == nil {
		return nil, errors.New("redis: simple error must not be nil")
	}
	if strings.ContainsAny(err.Error(), "\r\n") {
		return nil, errContainsCRLF
	}
	return fmt.Appendf(nil, "-%s\r\n", err.Error()), nil
}

func (s *Serializer) Integer(n int64) []byte {
	return fmt.Appendf(nil, ":%d\r\n", n)
}

func (s *Serializer) BulkString(str string) []byte {
	return fmt.Appendf(nil, "$%d\r\n%s\r\n", len(str), str)
}

func (s *Serializer) Null() []byte {
	return []byte("_\r\n")
}

func (s *Serializer) Array(elements ...[]byte) []byte {
	buf := fmt.Appendf(nil, "*%d\r\n", len(elements))
	for _, e := range elements {
		buf = append(buf, e...)
	}
	return buf
}

func (s *Serializer) Boolean(b bool) []byte {
	if b {
		return []byte("#t\r\n")
	}
	return []byte("#f\r\n")
}

func (s *Serializer) Double(f float64) []byte {
	switch {
	case math.IsInf(f, 1):
		return []byte(",inf\r\n")
	case math.IsInf(f, -1):
		return []byte(",-inf\r\n")
	case math.IsNaN(f):
		return []byte(",nan\r\n")
	}
	return fmt.Appendf(nil, ",%g\r\n", f)
}

func (s *Serializer) BigNumber(n *big.Int) ([]byte, error) {
	if n == nil {
		return nil, errors.New("redis: big number must not be nil")
	}
	return fmt.Appendf(nil, "(%s\r\n", n.String()), nil
}

func (s *Serializer) BulkError(msg string) []byte {
	return fmt.Appendf(nil, "!%d\r\n%s\r\n", len(msg), msg)
}

func (s *Serializer) VerbatimString(encoding, data string) ([]byte, error) {
	if len(encoding) != 3 {
		return nil, errors.New("redis: verbatim encoding must be exactly 3 bytes")
	}
	return fmt.Appendf(nil, "=%d\r\n%s:%s\r\n", len(encoding)+1+len(data), encoding, data), nil
}

type KeyValue struct {
	Key   []byte
	Value []byte
}

func (s *Serializer) Map(pairs ...KeyValue) ([]byte, error) {
	n := 23
	for _, p := range pairs {
		n += len(p.Key) + len(p.Value)
	}
	buf := fmt.Appendf(make([]byte, 0, n), "%%%d\r\n", len(pairs))
	for _, p := range pairs {
		buf = append(buf, p.Key...)
		buf = append(buf, p.Value...)
	}
	return buf, nil
}

func (s *Serializer) Attributes(pairs ...KeyValue) []byte {
	n := 23
	for _, p := range pairs {
		n += len(p.Key) + len(p.Value)
	}
	buf := fmt.Appendf(make([]byte, 0, n), "|%d\r\n", len(pairs))
	for _, p := range pairs {
		buf = append(buf, p.Key...)
		buf = append(buf, p.Value...)
	}
	return buf
}

func (s *Serializer) Set(elements ...[]byte) []byte {
	n := 23
	for _, e := range elements {
		n += len(e)
	}
	buf := fmt.Appendf(make([]byte, 0, n), "~%d\r\n", len(elements))
	for _, e := range elements {
		buf = append(buf, e...)
	}
	return buf
}

func (s *Serializer) Push(elements ...[]byte) []byte {
	n := 23
	for _, e := range elements {
		n += len(e)
	}
	buf := fmt.Appendf(make([]byte, 0, n), ">%d\r\n", len(elements))
	for _, e := range elements {
		buf = append(buf, e...)
	}
	return buf
}

type StreamedString struct {
	started bool
	closed  bool
}

func (s *Serializer) StartStreamedString() *StreamedString {
	return &StreamedString{}
}

func (ss *StreamedString) Chunk(data []byte) ([]byte, error) {
	if ss.closed {
		return nil, errors.New("redis: streamed string already closed")
	}
	n := len(data) + 24 // ';' + digits + CRLF + data + CRLF
	if !ss.started {
		n += 4 // "$?\r\n"
	}
	buf := make([]byte, 0, n)
	if !ss.started {
		buf = append(buf, '$', '?', '\r', '\n')
		ss.started = true
	}
	buf = append(buf, ';')
	buf = strconv.AppendInt(buf, int64(len(data)), 10)
	buf = append(buf, '\r', '\n')
	buf = append(buf, data...)
	return append(buf, '\r', '\n'), nil
}

func (ss *StreamedString) Close() ([]byte, error) {
	if ss.closed {
		return nil, errors.New("redis: streamed string already closed")
	}
	ss.closed = true
	// Handle Close() called before any Chunk() — the header was never emitted.
	if !ss.started {
		ss.started = true
		return []byte("$?\r\n;0\r\n"), nil
	}
	return []byte(";0\r\n"), nil
}

type StreamedAggregate struct {
	prefix  byte
	started bool
	closed  bool
	count   int
}

func (s *Serializer) StartStreamedArray() *StreamedAggregate {
	return &StreamedAggregate{prefix: '*'}
}

func (s *Serializer) StartStreamedSet() *StreamedAggregate {
	return &StreamedAggregate{prefix: '~'}
}

func (s *Serializer) StartStreamedMap() *StreamedAggregate {
	return &StreamedAggregate{prefix: '%'}
}

func (ss *StreamedAggregate) Element(data []byte) ([]byte, error) {
	if ss.closed {
		return nil, errors.New("redis: streamed aggregate already closed")
	}
	n := len(data)
	if !ss.started {
		n += 4 // "<prefix>?\r\n"
	}
	buf := make([]byte, 0, n)
	if !ss.started {
		buf = append(buf, ss.prefix, '?', '\r', '\n')
		ss.started = true
	}
	buf = append(buf, data...)
	ss.count++
	return buf, nil
}

func (ss *StreamedAggregate) Close() ([]byte, error) {
	if ss.closed {
		return nil, errors.New("redis: streamed aggregate already closed")
	}
	if ss.prefix == '%' && ss.count%2 != 0 {
		return nil, errors.New("redis: streamed map requires an even number of elements (key/value pairs)")
	}
	ss.closed = true
	if !ss.started {
		ss.started = true
		return []byte{ss.prefix, '?', '\r', '\n', '.', '\r', '\n'}, nil
	}
	return []byte(".\r\n"), nil
}
