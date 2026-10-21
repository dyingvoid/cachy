package features

import (
	features_commands "github.com/dyingvoid/cachy/internal/features/commands"
	features_conn "github.com/dyingvoid/cachy/internal/features/conn"
)

func (h *ProtocolHandler) executeRepl(
	state *features_conn.State,
	cmd features_commands.ReplConfCmd,
) (Result, error) {
	if cmd.ListeningPort != nil {
		state.Replication.ListeningPort = cmd.ListeningPort
	}
	if cmd.Capa != nil {
		state.Replication.Capa = cmd.Capa
	}

	return StringResult("OK"), nil
}
