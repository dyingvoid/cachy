package features

import (
	"github.com/dyingvoid/cachy/internal/config"
	core_state "github.com/dyingvoid/cachy/internal/core/state"
)

type InfoCmd struct {
	commandMarker
}

func parseInfoArgs(args []string) (InfoCmd, error) {
	if len(args) != 0 {
		return InfoCmd{}, ErrInvalidArguments
	}

	return InfoCmd{}, nil
}

func (h *ProtocolHandler) executeInfo(cmd InfoCmd) (Result, error) {
	cfg := config.Get()
	info := NewInfo(
		cfg.Role, cfg.ReplicaOf,
		core_state.ReplicationID(),
		core_state.ReplicationOffset(),
	)

	return MapResult(info), nil
}
