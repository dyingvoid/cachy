package features

import (
	"github.com/dyingvoid/cachy/internal/config"
	core_state "github.com/dyingvoid/cachy/internal/core/state"
	features_commands "github.com/dyingvoid/cachy/internal/features/commands"
)

func (h *ProtocolHandler) executeInfo(cmd features_commands.InfoCmd) (Result, error) {
	cfg := config.Get()
	info := NewInfo(
		cfg.Role, cfg.ReplicaOf,
		core_state.ReplicationID(),
		core_state.ReplicationOffset(),
	)

	return MapResult(info), nil
}
