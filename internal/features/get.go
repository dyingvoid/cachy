package features

import (
	"fmt"

	features_commands "github.com/dyingvoid/cachy/internal/features/commands"
)

func (h *ProtocolHandler) executeGet(cmd features_commands.GetCmd) (Result, error) {
	record, err := h.storage.Get(cmd.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %w", err)
	}

	return BulkStringResult(record.Value), nil
}
