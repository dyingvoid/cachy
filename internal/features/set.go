package features

import (
	"github.com/dyingvoid/cachy/internal/domain"
	features_commands "github.com/dyingvoid/cachy/internal/features/commands"
)

func (h *ProtocolHandler) executeSet(cmd features_commands.SetCmd) (Result, error) {
	r := domain.NewStringRecord(cmd.Value, cmd.Expiry)
	if err := h.storage.Set(cmd.Key, r); err != nil {
		return nil, err
	}

	h.replicate(cmd)

	return StringResult("OK"), nil
}
