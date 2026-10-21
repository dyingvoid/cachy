package resp3

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/big"
	"strconv"

	features_commands "github.com/dyingvoid/cachy/internal/features/commands"
)

const (
	maxDepth           = 128
	maxBulkLength      = 512 << 20
	maxAggregateLength = 1 << 20
)

type Kind byte

const (
	KindSimpleString   Kind = '+'
	KindSimpleError    Kind = '-'
	KindInteger        Kind = ':'
	KindBulkString     Kind = '$'
	KindArray          Kind = '*'
	KindNull           Kind = '_'
	KindBoolean        Kind = '#'
	KindDouble         Kind = ','
	KindBigNumber      Kind = '('
	KindBulkError      Kind = '!'
	KindVerbatimString Kind = '='
	KindMap            Kind = '%'
	KindSet            Kind = '~'
	KindPush           Kind = '>'
)

var (
	errUnexpectedType    = errors.New("resp: unexpected type byte")
	errInvalidLength     = errors.New("resp: invalid length")
	errInvalidNull       = errors.New("resp: invalid null")
	errInvalidInteger    = errors.New("resp: invalid integer")
	errInvalidDouble     = errors.New("resp: invalid double")
	errInvalidBoolean    = errors.New("resp: invalid boolean")
	errInvalidBigNumber  = errors.New("resp: invalid big number")
	errInvalidTerminator = errors.New("resp: missing CRLF")
	errInvalidVerbatim   = errors.New("resp: invalid verbatim string")
	errOddMapElements    = errors.New("resp: map requires an even number of elements")
	errTooDeep           = errors.New("resp: nesting too deep")
	errLengthTooLarge    = errors.New("resp: length too large")
	errInvalidCommand    = errors.New("resp: command must be an array of bulk strings")
)

type Pair struct {
	Key   Value
	Value Value
}

type Value struct {
	Kind       Kind
	String     string
	Integer    int64
	Boolean    bool
	Double     float64
	BigNumber  *big.Int
	Encoding   string
	Elements   []Value
	Pairs      []Pair
	Attributes []Pair
	Streamed   bool
}

type Deserializer struct {
	r *bufio.Reader
}

func NewDeserializer(r io.Reader) *Deserializer {
	return &Deserializer{r: bufio.NewReader(r)}
}

func (d *Deserializer) Read() (Value, error) {
	return d.readValue(0)
}

func (d *Deserializer) ReadCommand() (features_commands.Command, error) {
	value, err := d.Read()
	if err != nil {
		return nil, err
	}
	if value.Kind != KindArray {
		return nil, errInvalidCommand
	}

	fields := make([]string, 0, len(value.Elements))
	for _, element := range value.Elements {
		if element.Kind != KindBulkString {
			return nil, errInvalidCommand
		}
		fields = append(fields, element.String)
	}

	cmd, err := features_commands.NewProtocolParser().ParseFields(fields)
	if err != nil {
		return nil, err
	}

	return cmd, nil
}

func (d *Deserializer) readValue(depth int) (Value, error) {
	if depth > maxDepth {
		return Value{}, errTooDeep
	}

	prefix, err := d.r.ReadByte()
	if err != nil {
		return Value{}, err
	}

	var attributes []Pair
	for prefix == '|' {
		pairs, err := d.readAttributes(depth + 1)
		if err != nil {
			return Value{}, err
		}
		attributes = append(attributes, pairs...)

		prefix, err = d.r.ReadByte()
		if err != nil {
			return Value{}, unexpectedEOF(err)
		}
	}

	value, err := d.readValueByPrefix(prefix, depth)
	if err != nil {
		return Value{}, err
	}
	value.Attributes = attributes

	return value, nil
}

func (d *Deserializer) readValueByPrefix(prefix byte, depth int) (Value, error) {
	switch Kind(prefix) {
	case KindSimpleString:
		line, err := d.readLine()
		if err != nil {
			return Value{}, err
		}
		if bytes.ContainsAny(line, "\r\n") {
			return Value{}, errContainsCRLF
		}
		return Value{Kind: KindSimpleString, String: string(line)}, nil

	case KindSimpleError:
		line, err := d.readLine()
		if err != nil {
			return Value{}, err
		}
		if bytes.ContainsAny(line, "\r\n") {
			return Value{}, errContainsCRLF
		}
		return Value{Kind: KindSimpleError, String: string(line)}, nil

	case KindInteger:
		line, err := d.readLine()
		if err != nil {
			return Value{}, err
		}
		n, err := strconv.ParseInt(string(line), 10, 64)
		if err != nil {
			return Value{}, errInvalidInteger
		}
		return Value{Kind: KindInteger, Integer: n}, nil

	case KindBulkString:
		return d.readBulkString()

	case KindArray, KindSet, KindPush, KindMap:
		return d.readAggregate(Kind(prefix), depth)

	case KindNull:
		line, err := d.readLine()
		if err != nil {
			return Value{}, err
		}
		if len(line) != 0 {
			return Value{}, errInvalidNull
		}
		return Value{Kind: KindNull}, nil

	case KindBoolean:
		line, err := d.readLine()
		if err != nil {
			return Value{}, err
		}
		switch string(line) {
		case "t":
			return Value{Kind: KindBoolean, Boolean: true}, nil
		case "f":
			return Value{Kind: KindBoolean, Boolean: false}, nil
		default:
			return Value{}, errInvalidBoolean
		}

	case KindDouble:
		line, err := d.readLine()
		if err != nil {
			return Value{}, err
		}
		f, err := strconv.ParseFloat(string(line), 64)
		if err != nil {
			return Value{}, errInvalidDouble
		}
		return Value{Kind: KindDouble, Double: f}, nil

	case KindBigNumber:
		line, err := d.readLine()
		if err != nil {
			return Value{}, err
		}
		n, ok := new(big.Int).SetString(string(line), 10)
		if !ok {
			return Value{}, errInvalidBigNumber
		}
		return Value{Kind: KindBigNumber, BigNumber: n}, nil

	case KindBulkError:
		data, err := d.readSizedPayload()
		if err != nil {
			return Value{}, err
		}
		return Value{Kind: KindBulkError, String: string(data)}, nil

	case KindVerbatimString:
		data, err := d.readSizedPayload()
		if err != nil {
			return Value{}, err
		}
		if len(data) < 4 || data[3] != ':' {
			return Value{}, errInvalidVerbatim
		}
		return Value{
			Kind:     KindVerbatimString,
			Encoding: string(data[:3]),
			String:   string(data[4:]),
		}, nil

	default:
		return Value{}, fmt.Errorf("%w: %q", errUnexpectedType, prefix)
	}
}

func (d *Deserializer) readBulkString() (Value, error) {
	line, err := d.readLine()
	if err != nil {
		return Value{}, err
	}
	if string(line) == "?" {
		return d.readStreamedString()
	}

	n, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return Value{}, errInvalidLength
	}
	if n == -1 {
		return Value{Kind: KindNull}, nil
	}
	if n < 0 {
		return Value{}, errInvalidLength
	}
	if n > maxBulkLength {
		return Value{}, errLengthTooLarge
	}

	data, err := d.readPayload(int(n))
	if err != nil {
		return Value{}, err
	}

	return Value{Kind: KindBulkString, String: string(data)}, nil
}

func (d *Deserializer) readStreamedString() (Value, error) {
	var data []byte
	for {
		line, err := d.readLine()
		if err != nil {
			return Value{}, err
		}
		if len(line) == 0 || line[0] != ';' {
			return Value{}, errInvalidLength
		}

		n, err := strconv.ParseInt(string(line[1:]), 10, 64)
		if err != nil || n < 0 {
			return Value{}, errInvalidLength
		}
		if n == 0 {
			break
		}
		if n > maxBulkLength {
			return Value{}, errLengthTooLarge
		}

		chunk, err := d.readPayload(int(n))
		if err != nil {
			return Value{}, err
		}
		data = append(data, chunk...)
	}

	return Value{Kind: KindBulkString, String: string(data), Streamed: true}, nil
}

func (d *Deserializer) readAggregate(kind Kind, depth int) (Value, error) {
	line, err := d.readLine()
	if err != nil {
		return Value{}, err
	}
	if string(line) == "?" {
		return d.readStreamedAggregate(kind, depth)
	}

	n, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return Value{}, errInvalidLength
	}
	if n == -1 {
		if kind == KindArray {
			return Value{Kind: KindNull}, nil
		}
		return Value{}, errInvalidLength
	}
	if n < 0 {
		return Value{}, errInvalidLength
	}
	if n > maxAggregateLength {
		return Value{}, errLengthTooLarge
	}

	if kind == KindMap {
		pairs, err := d.readPairs(depth, int(n))
		if err != nil {
			return Value{}, err
		}
		return Value{Kind: KindMap, Pairs: pairs}, nil
	}

	elements := make([]Value, 0, int(n))
	for i := int64(0); i < n; i++ {
		element, err := d.readValue(depth + 1)
		if err != nil {
			return Value{}, aggregateError(err)
		}
		elements = append(elements, element)
	}

	return Value{Kind: kind, Elements: elements}, nil
}

func (d *Deserializer) readStreamedAggregate(kind Kind, depth int) (Value, error) {
	var elements []Value
	for {
		prefix, err := d.r.Peek(1)
		if err != nil {
			return Value{}, unexpectedEOF(err)
		}

		if prefix[0] == '.' {
			line, err := d.readLine()
			if err != nil {
				return Value{}, err
			}
			if string(line) != "." {
				return Value{}, errInvalidTerminator
			}
			break
		}

		element, err := d.readValue(depth + 1)
		if err != nil {
			return Value{}, aggregateError(err)
		}
		elements = append(elements, element)
		if len(elements) > maxAggregateLength {
			return Value{}, errLengthTooLarge
		}
	}

	if kind == KindMap {
		if len(elements)%2 != 0 {
			return Value{}, errOddMapElements
		}

		pairs := make([]Pair, 0, len(elements)/2)
		for i := 0; i < len(elements); i += 2 {
			pairs = append(pairs, Pair{Key: elements[i], Value: elements[i+1]})
		}
		return Value{Kind: KindMap, Pairs: pairs, Streamed: true}, nil
	}

	return Value{Kind: kind, Elements: elements, Streamed: true}, nil
}

func (d *Deserializer) readAttributes(depth int) ([]Pair, error) {
	line, err := d.readLine()
	if err != nil {
		return nil, err
	}

	n, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil || n < 0 {
		return nil, errInvalidLength
	}
	if n > maxAggregateLength {
		return nil, errLengthTooLarge
	}

	return d.readPairs(depth, int(n))
}

func (d *Deserializer) readPairs(depth, n int) ([]Pair, error) {
	pairs := make([]Pair, 0, n)
	for i := 0; i < n; i++ {
		key, err := d.readValue(depth + 1)
		if err != nil {
			return nil, aggregateError(err)
		}

		value, err := d.readValue(depth + 1)
		if err != nil {
			return nil, aggregateError(err)
		}

		pairs = append(pairs, Pair{Key: key, Value: value})
	}

	return pairs, nil
}

func (d *Deserializer) readSizedPayload() ([]byte, error) {
	line, err := d.readLine()
	if err != nil {
		return nil, err
	}

	n, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil || n < 0 {
		return nil, errInvalidLength
	}
	if n > maxBulkLength {
		return nil, errLengthTooLarge
	}

	return d.readPayload(int(n))
}

func (d *Deserializer) readPayload(n int) ([]byte, error) {
	data := make([]byte, n)
	if _, err := io.ReadFull(d.r, data); err != nil {
		return nil, unexpectedEOF(err)
	}
	if err := d.expectCRLF(); err != nil {
		return nil, err
	}

	return data, nil
}

func (d *Deserializer) expectCRLF() error {
	var crlf [2]byte
	if _, err := io.ReadFull(d.r, crlf[:]); err != nil {
		return unexpectedEOF(err)
	}
	if crlf[0] != '\r' || crlf[1] != '\n' {
		return errInvalidTerminator
	}

	return nil
}

func (d *Deserializer) readLine() ([]byte, error) {
	line, err := d.r.ReadBytes('\n')
	if err != nil {
		return nil, unexpectedEOF(err)
	}
	if len(line) < 2 || line[len(line)-2] != '\r' {
		return nil, errInvalidTerminator
	}

	return line[:len(line)-2], nil
}

func unexpectedEOF(err error) error {
	if errors.Is(err, io.EOF) {
		return io.ErrUnexpectedEOF
	}

	return err
}

func aggregateError(err error) error {
	if errors.Is(err, io.EOF) {
		return io.ErrUnexpectedEOF
	}

	return err
}
