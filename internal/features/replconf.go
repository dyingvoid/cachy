package features

import "strconv"

type ReplConfCmd struct {
	ListeningPort *int
	Capa          *string
	commandMarker
}

func parseReplConfArgs(args []string) (ReplConfCmd, error) {
	if len(args) != 2 {
		return ReplConfCmd{}, ErrInvalidArguments
	}

	switch args[0] {
	case "listening-port":
		port, err := strconv.Atoi(args[1])
		if err != nil {
			return ReplConfCmd{}, ErrInvalidArguments
		}
		return ReplConfCmd{ListeningPort: &port}, nil

	case "capa":
		return ReplConfCmd{Capa: &args[1]}, nil

	default:
		return ReplConfCmd{}, ErrInvalidArguments
	}
}
