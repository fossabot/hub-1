package events

import (
	"time"

	pkgTime "github.com/plgd-dev/hub/pkg/time"
	"google.golang.org/protobuf/proto"
)

const eventTypeResourceRetrievePending = "resourceretrievepending"

func (e *ResourceRetrievePending) Version() uint64 {
	return e.GetEventMetadata().GetVersion()
}

func (e *ResourceRetrievePending) Marshal() ([]byte, error) {
	return proto.Marshal(e)
}

func (e *ResourceRetrievePending) Unmarshal(b []byte) error {
	return proto.Unmarshal(b, e)
}

func (e *ResourceRetrievePending) EventType() string {
	return eventTypeResourceRetrievePending
}

func (e *ResourceRetrievePending) AggregateID() string {
	return e.GetResourceId().ToUUID()
}

func (e *ResourceRetrievePending) GroupID() string {
	return e.GetResourceId().GetDeviceId()
}

func (e *ResourceRetrievePending) IsSnapshot() bool {
	return false
}

func (e *ResourceRetrievePending) Timestamp() time.Time {
	return pkgTime.Unix(0, e.GetEventMetadata().GetTimestamp())
}

func (e *ResourceRetrievePending) CopyData(event *ResourceRetrievePending) {
	e.ResourceId = event.GetResourceId()
	e.ResourceInterface = event.GetResourceInterface()
	e.AuditContext = event.GetAuditContext()
	e.EventMetadata = event.GetEventMetadata()
	e.ValidUntil = event.GetValidUntil()
}

func (e *ResourceRetrievePending) CheckInitialized() bool {
	return e.GetResourceId() != nil &&
		e.GetResourceInterface() != "" &&
		e.GetAuditContext() != nil &&
		e.GetEventMetadata() != nil
}

func (e *ResourceRetrievePending) ValidUntilTime() time.Time {
	return pkgTime.Unix(0, e.GetValidUntil())
}

func (e *ResourceRetrievePending) IsExpired(now time.Time) bool {
	return IsExpired(now, e.ValidUntilTime())
}
