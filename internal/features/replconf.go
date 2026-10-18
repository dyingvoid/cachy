package features

import (
	"strconv"

	features_conn "github.com/dyingvoid/cachy/internal/features/conn"
)

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

func (h *ProtocolHandler) executeRepl(
	state *features_conn.State,
	cmd ReplConfCmd,
) (Result, error) {
	if cmd.ListeningPort != nil {
		state.Replication.ListeningPort = cmd.ListeningPort
	}
	if cmd.Capa != nil {
		state.Replication.Capa = cmd.Capa
	}

	return StringResult("OK"), nil
}
