package commands

type LLenCmd struct {
	Key string
	commandMarker
}

func parseLLenArgs(args []string) (LLenCmd, error) {
	if len(args) != 1 {
		return LLenCmd{}, ErrInvalidArguments
	}

	return LLenCmd{Key: args[0]}, nil
}
