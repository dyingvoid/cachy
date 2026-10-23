package commands

type InfoCmd struct {
	commandMarker
}

func parseInfoArgs(args []string) (InfoCmd, error) {
	if len(args) != 0 {
		return InfoCmd{}, ErrInvalidArguments
	}

	return InfoCmd{}, nil
}
