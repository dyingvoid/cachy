package features

import "strconv"

type LRangeCmd struct {
	Key   string
	Start int
	Stop  int
	commandMarker
}

func parseLRangeArgs(args []string) (LRangeCmd, error) {
	if len(args) != 3 {
		return LRangeCmd{}, ErrInvalidArguments
	}

	start, err := strconv.Atoi(args[1])
	if err != nil {
		return LRangeCmd{}, ErrInvalidArguments
	}

	stop, err := strconv.Atoi(args[2])
	if err != nil {
		return LRangeCmd{}, ErrInvalidArguments
	}

	return LRangeCmd{Key: args[0], Start: start, Stop: stop}, nil
}

func (h *ProtocolHandler) executeLRange(cmd LRangeCmd) (Result, error) {
	if err := h.ensureListKey(cmd.Key); err != nil {
		return nil, err
	}

	values, err := h.listStorage.LRange(cmd.Key, cmd.Start, cmd.Stop)
	if err != nil {
		return nil, err
	}

	return ListResult(values), nil
}
