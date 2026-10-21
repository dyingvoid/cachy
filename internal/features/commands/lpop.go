package commands

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
