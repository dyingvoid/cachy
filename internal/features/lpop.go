package features

import "strconv"

type LPopCmd struct {
	Key   string
	Count int
	commandMarker
}

func parseLPopArgs(args []string) (LPopCmd, error) {
	if len(args) != 2 {
		return LPopCmd{}, ErrInvalidArguments
	}

	count, err := strconv.Atoi(args[1])
	if err != nil || count < 0 {
		return LPopCmd{}, ErrInvalidArguments
	}

	return LPopCmd{Key: args[0], Count: count}, nil
}

func (h *ProtocolHandler) executeLPop(cmd LPopCmd) (Result, error) {
	if err := h.ensureListKey(cmd.Key); err != nil {
		return nil, err
	}

	values, err := h.listStorage.LPop(cmd.Key, cmd.Count)
	if err != nil {
		return nil, err
	}

	return ListResult(values), nil
}
