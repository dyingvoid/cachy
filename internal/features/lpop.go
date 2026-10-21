package features

import features_commands "github.com/dyingvoid/cachy/internal/features/commands"

func (h *ProtocolHandler) executeLPop(cmd features_commands.LPopCmd) (Result, error) {
	values, err := h.storage.LPop(cmd.Key, cmd.Count)
	if err != nil {
		return nil, err
	}

	h.replicate(cmd)

	return ListResult(values), nil
}
