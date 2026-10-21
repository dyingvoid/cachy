package commands

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
