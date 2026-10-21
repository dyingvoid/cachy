package commands

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
