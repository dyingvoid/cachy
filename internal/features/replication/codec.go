package features_replication

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	ErrEmptyMessage       = errors.New("empty message")
	ErrInvalidMessage     = errors.New("invalid message")
	ErrEmptyResponse      = errors.New("empty response")
	ErrServerError        = errors.New("server error")
	ErrUnexpectedResponse = errors.New("unexpected response")
)

type TextCodec struct{}

func NewTextCodec() *TextCodec {
	return &TextCodec{}
}

func (c *TextCodec) Encode(msg Message) ([]byte, error) {
	if msg == nil {
		return nil, fmt.Errorf("encode message: %w", ErrEmptyMessage)
	}

	command := strings.TrimSpace(msg.Command())
	if command == "" {
		return nil, fmt.Errorf("encode message: %w", ErrEmptyMessage)
	}

	args := msg.Args()
	fields := make([]string, 0, len(args)+1)
	fields = append(fields, command)
	fields = append(fields, args...)

	for _, field := range fields {
		if field == "" || strings.ContainsAny(field, " \t\r\n") {
			return nil, fmt.Errorf("encode field %q: %w", field, ErrInvalidMessage)
		}
	}

	return []byte(strings.Join(fields, " ") + "\n"), nil
}

func (c *TextCodec) Decode(raw []byte) (Response, error) {
	response := strings.TrimSpace(string(raw))
	if response == "" {
		return nil, fmt.Errorf("decode response: %w", ErrEmptyResponse)
	}

	const errorPrefix = "(error)"
	if strings.HasPrefix(response, errorPrefix) {
		return nil, fmt.Errorf(
			"decode response: %w: %s",
			ErrServerError,
			strings.TrimSpace(strings.TrimPrefix(response, errorPrefix)),
		)
	}

	switch {
	case strings.EqualFold(response, "PONG"):
		return PongResponse{}, nil
	case strings.EqualFold(response, "OK"):
		return OKResponse{}, nil
	case strings.HasPrefix(strings.ToUpper(response), "FULLRESYNC "):
		return decodePsync(response)
	default:
		return nil, fmt.Errorf(
			"decode response %q: %w",
			response,
			ErrUnexpectedResponse,
		)
	}
}

func decodePsync(response string) (Response, error) {
	fields := strings.Fields(response)
	if len(fields) != 3 || !strings.EqualFold(fields[0], "FULLRESYNC") {
		return nil, fmt.Errorf(
			"decode psync response %q: %w",
			response,
			ErrUnexpectedResponse,
		)
	}

	offset, err := strconv.ParseInt(fields[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf(
			"decode psync offset %q: %w",
			fields[2],
			err,
		)
	}

	return PsyncResponse{
		ReplID: fields[1],
		Offset: offset,
	}, nil
}
