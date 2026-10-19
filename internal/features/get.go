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
	record, err := h.storage.Get(cmd.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %w", err)
	}

	return BulkStringResult(record.Value), nil
}
