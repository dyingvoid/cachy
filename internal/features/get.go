package features

import "fmt"

type GetCmd struct {
	Key string
	commandMarker
}

func parseGetArgs(args []string) (GetCmd, error) {
	if len(args) != 1 {
		return GetCmd{}, ErrInvalidArguments
	}

	return GetCmd{Key: args[0]}, nil
}

func (h *ProtocolHandler) executeGet(cmd GetCmd) (Result, error) {
	if err := h.ensureStringKey(cmd.Key); err != nil {
		return nil, err
	}

	record, err := h.stringStorage.Get(cmd.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %w", err)
	}

	return StringResult(record.Value), nil
}
