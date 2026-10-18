package features

import "fmt"

type LLenCmd struct {
	Key string
	commandMarker
}

func parseLLenArgs(args []string) (LLenCmd, error) {
	if len(args) != 1 {
		return LLenCmd{}, ErrInvalidArguments
	}

	return LLenCmd{Key: args[0]}, nil
}

func (h *ProtocolHandler) executeLLen(cmd LLenCmd) (Result, error) {
	n, err := h.storage.LLen(cmd.Key)
	if err != nil {
		return nil, err
	}

	return StringResult(fmt.Sprintf("(integer) %d", n)), nil
}
