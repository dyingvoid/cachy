package commands

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
