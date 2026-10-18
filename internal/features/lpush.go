package features

import "fmt"

type LPushCmd struct {
	Key    string
	Values []string
	commandMarker
}

func parseLPushArgs(args []string) (LPushCmd, error) {
	if len(args) < 2 {
		return LPushCmd{}, ErrInvalidArguments
	}

	return LPushCmd{Key: args[0], Values: args[1:]}, nil
}

func (h *ProtocolHandler) executeLPush(cmd LPushCmd) (Result, error) {
	if err := h.ensureListKey(cmd.Key); err != nil {
		return nil, err
	}

	n, err := h.listStorage.LPush(cmd.Key, cmd.Values...)
	if err != nil {
		return nil, err
	}

	return StringResult(fmt.Sprintf("(integer) %d", n)), nil
}
