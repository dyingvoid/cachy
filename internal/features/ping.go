package features

import features_commands "github.com/dyingvoid/cachy/internal/features/commands"

func (h *ProtocolHandler) executePing(cmd features_commands.PingCmd) (StringResult, error) {
	return StringResult("PONG"), nil
}
