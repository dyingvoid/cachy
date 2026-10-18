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
	n, err := h.storage.LPush(cmd.Key, cmd.Values...)
	if err != nil {
		return nil, err
	}

	h.replicate(cmd)

	return StringResult(fmt.Sprintf("(integer) %d", n)), nil
}
