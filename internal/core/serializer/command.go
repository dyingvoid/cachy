package core_serializer

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dyingvoid/cachy/internal/features"
)

func (s *TextSerializer) SerializeCommand(cmd features.WriteCommand) string {
	switch cmd := cmd.(type) {
	case features.SetCmd:
		// TODO: serialize function
		return serializeSet(cmd)
	case features.LPushCmd:
		return serializePush("LPUSH", cmd.Key, cmd.Values)
	case features.RPushCmd:
		return serializePush("RPUSH", cmd.Key, cmd.Values)
	case features.LPopCmd:
		// TODO: use serialize function
		return fmt.Sprintf("LPOP %s %d", cmd.Key, cmd.Count)
	default:
		panic(fmt.Sprintf("text serializer: unsupported command type %T, value: %#v", cmd, cmd))
	}
}

func serializeSet(cmd features.SetCmd) string {
	fields := []string{"SET", cmd.Key, cmd.Value}
	if cmd.Expiry != nil {
		fields = append(fields, expiryFields(*cmd.Expiry)...)
	}

	return strings.Join(fields, " ")
}

func expiryFields(expiry time.Duration) []string {
	if expiry%time.Second == 0 {
		return []string{"EX", strconv.FormatInt(int64(expiry/time.Second), 10)}
	}

	return []string{"PX", strconv.FormatInt(expiry.Milliseconds(), 10)}
}

func serializePush(name, key string, values []string) string {
	fields := make([]string, 0, len(values)+2)
	fields = append(fields, name, key)
	fields = append(fields, values...)

	return strings.Join(fields, " ")
}
