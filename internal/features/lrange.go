package features

import features_commands "github.com/dyingvoid/cachy/internal/features/commands"

func (h *ProtocolHandler) executeLRange(cmd features_commands.LRangeCmd) (Result, error) {
	values, err := h.storage.LRange(cmd.Key, cmd.Start, cmd.Stop)
	if err != nil {
		return nil, err
	}

	return ListResult(values), nil
}
