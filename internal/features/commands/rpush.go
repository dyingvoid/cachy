package commands

type RPushCmd struct {
	Key    string
	Values []string
	commandMarker
}

func parseRPushArgs(args []string) (RPushCmd, error) {
	if len(args) < 2 {
		return RPushCmd{}, ErrInvalidArguments
	}

	return RPushCmd{Key: args[0], Values: args[1:]}, nil
}
