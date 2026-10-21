package features

import features_commands "github.com/dyingvoid/cachy/internal/features/commands"

func (h *ProtocolHandler) executeRPush(cmd features_commands.RPushCmd) (Result, error) {
	n, err := h.storage.RPush(cmd.Key, cmd.Values...)
	if err != nil {
		return nil, err
	}

	h.replicate(cmd)

	return IntResult(n), nil
}
