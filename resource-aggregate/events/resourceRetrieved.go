package events

import (
	"time"

	pkgTime "github.com/plgd-dev/hub/pkg/time"
	commands "github.com/plgd-dev/hub/resource-aggregate/commands"
	"google.golang.org/protobuf/proto"
)

const eventTypeResourceRetrieved = "resourceretrieved"

func (e *ResourceRetrieved) Version() uint64 {
	return e.GetEventMetadata().GetVersion()
}

func (e *ResourceRetrieved) Marshal() ([]byte, error) {
	return proto.Marshal(e)
}

func (e *ResourceRetrieved) Unmarshal(b []byte) error {
	return proto.Unmarshal(b, e)
}

func (e *ResourceRetrieved) EventType() string {
	return eventTypeResourceRetrieved
}

func (e *ResourceRetrieved) AggregateID() string {
	return e.GetResourceId().ToUUID()
}

func (e *ResourceRetrieved) GroupID() string {
	return e.GetResourceId().GetDeviceId()
}

func (e *ResourceRetrieved) IsSnapshot() bool {
	return false
}

func (e *ResourceRetrieved) Timestamp() time.Time {
	return pkgTime.Unix(0, e.GetEventMetadata().GetTimestamp())
}

func (e *ResourceRetrieved) CopyData(event *ResourceRetrieved) {
	e.ResourceId = event.GetResourceId()
	e.Content = event.GetContent()
	e.AuditContext = event.GetAuditContext()
	e.EventMetadata = event.GetEventMetadata()
	e.Status = event.GetStatus()
}

func (e *ResourceRetrieved) CheckInitialized() bool {
	return e.GetResourceId() != nil &&
		e.GetContent() != nil &&
		e.GetAuditContext() != nil &&
		e.GetEventMetadata() != nil &&
		e.GetStatus() != commands.Status(0)
}
