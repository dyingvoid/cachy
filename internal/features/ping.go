package features

type PingCmd struct {
	commandMarker
}

func parsePingArgs(args []string) (PingCmd, error) {
	if len(args) != 0 {
		return PingCmd{}, ErrInvalidArguments
	}

	return PingCmd{}, nil
}
