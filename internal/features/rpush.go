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
	n, err := h.storage.RPush(cmd.Key, cmd.Values...)
	if err != nil {
		return nil, err
	}

	h.replicate(cmd)

	return StringResult(fmt.Sprintf("(integer) %d", n)), nil
}
