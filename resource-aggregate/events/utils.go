package events

import (
	"time"

	commands "github.com/plgd-dev/hub/resource-aggregate/commands"
)

//MakeEventMeta for creating EventMetadata for event.
func MakeEventMeta(connectionId string, sequence, version uint64) *EventMetadata {
	return &EventMetadata{
		ConnectionId: connectionId,
		Sequence:     sequence,
		Version:      version,
		Timestamp:    time.Now().UnixNano(),
	}
}

func EqualResource(x, y *commands.Resource) bool {
	return x.DeviceId == y.DeviceId &&
		EqualStringSlice(x.ResourceTypes, y.ResourceTypes) &&
		EqualStringSlice(x.Interfaces, y.Interfaces) &&
		x.Anchor == y.Anchor &&
		x.Title == y.Title &&
		EqualStringSlice(x.SupportedContentTypes, y.SupportedContentTypes) &&
		x.ValidUntil == y.ValidUntil &&
		((x.Policy == nil && y.Policy == nil) ||
			(x.Policy != nil && y.Policy != nil && x.Policy.BitFlags == y.Policy.BitFlags))
}

func EqualStringSlice(x, y []string) bool {
	if len(x) != len(y) {
		return false
	}
	for i := range x {
		if x[i] != y[i] {
			return false
		}
	}
	return true
}

func IsExpired(now time.Time, validUntil time.Time) bool {
	if validUntil.IsZero() {
		return false
	}
	return now.After(validUntil)
}
