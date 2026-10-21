package features

import features_commands "github.com/dyingvoid/cachy/internal/features/commands"

func (h *ProtocolHandler) executeLPush(cmd features_commands.LPushCmd) (Result, error) {
	n, err := h.storage.LPush(cmd.Key, cmd.Values...)
	if err != nil {
		return nil, err
	}

	h.replicate(cmd)

	return IntResult(n), nil
}
