package features

import "fmt"

type RPushCmd struct {
	Key    string
	Values []string
	commandMarker
}

func parseRPushArgs(args []string) (RPushCmd, error) {
	if len(args) < 2 {
		return RPushCmd{}, ErrInvalidArguments
	}

	return RPushCmd{Key: args[0], Values: args[1:]}, nil
}

func (h *ProtocolHandler) executeRPush(cmd RPushCmd) (Result, error) {
	if err := h.ensureListKey(cmd.Key); err != nil {
		return nil, err
	}

	n, err := h.listStorage.RPush(cmd.Key, cmd.Values...)
	if err != nil {
		return nil, err
	}

	return StringResult(fmt.Sprintf("(integer) %d", n)), nil
}
