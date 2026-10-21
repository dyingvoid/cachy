package features

import features_commands "github.com/dyingvoid/cachy/internal/features/commands"

func (h *ProtocolHandler) executeLLen(cmd features_commands.LLenCmd) (Result, error) {
	n, err := h.storage.LLen(cmd.Key)
	if err != nil {
		return nil, err
	}

	return IntResult(n), nil
}
