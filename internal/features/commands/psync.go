package commands

import "strconv"

type PsyncCmd struct {
	ReplicationID string
	Offset        int64
	commandMarker
}

func parsePsyncArgs(args []string) (PsyncCmd, error) {
	if len(args) != 2 {
		return PsyncCmd{}, ErrInvalidArguments
	}

	offset, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return PsyncCmd{}, ErrInvalidArguments
	}

	return PsyncCmd{
		ReplicationID: args[0],
		Offset:        offset,
	}, nil
}
