package events

import (
	"context"
	"fmt"
	"time"

	"github.com/plgd-dev/hub/pkg/net/grpc"
	pkgTime "github.com/plgd-dev/hub/pkg/time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/plgd-dev/hub/resource-aggregate/commands"
	"github.com/plgd-dev/hub/resource-aggregate/cqrs/aggregate"
	"github.com/plgd-dev/hub/resource-aggregate/cqrs/eventstore"
)

const eventTypeResourceLinksSnapshotTaken = "resourcelinkssnapshottaken"

func (e *ResourceLinksSnapshotTaken) AggregateID() string {
	return commands.MakeLinksResourceUUID(e.GetDeviceId())
}

func (e *ResourceLinksSnapshotTaken) GroupID() string {
	return e.GetDeviceId()
}

func (e *ResourceLinksSnapshotTaken) Version() uint64 {
	return e.GetEventMetadata().GetVersion()
}

func (e *ResourceLinksSnapshotTaken) Marshal() ([]byte, error) {
	return proto.Marshal(e)
}

func (e *ResourceLinksSnapshotTaken) Unmarshal(b []byte) error {
	return proto.Unmarshal(b, e)
}

func (e *ResourceLinksSnapshotTaken) EventType() string {
	return eventTypeResourceLinksSnapshotTaken
}

func (e *ResourceLinksSnapshotTaken) IsSnapshot() bool {
	return true
}

func (e *ResourceLinksSnapshotTaken) Timestamp() time.Time {
	return pkgTime.Unix(0, e.GetEventMetadata().GetTimestamp())
}

func (e *ResourceLinksSnapshotTaken) CopyData(event *ResourceLinksSnapshotTaken) {
	e.Resources = event.GetResources()
	e.DeviceId = event.GetDeviceId()
	e.EventMetadata = event.GetEventMetadata()
	e.AuditContext = event.GetAuditContext()
}

func (e *ResourceLinksSnapshotTaken) CheckInitialized() bool {
	return e.GetResources() != nil &&
		e.GetDeviceId() != "" &&
		e.GetAuditContext() != nil &&
		e.GetEventMetadata() != nil
}

// Examine published resources by the ResourceLinksPublished, compare it with cached resources and
// return array of new or changed resources.
func (e *ResourceLinksSnapshotTaken) GetNewPublishedLinks(pub *ResourceLinksPublished) []*commands.Resource {

	if e.GetResources() == nil {
		return pub.GetResources()
	}

	published := make([]*commands.Resource, 0, len(pub.GetResources()))

	for _, resPub := range pub.GetResources() {
		resSnap, ok := e.Resources[resPub.Href]
		if !ok || !EqualResource(resPub, resSnap) {
			published = append(published, resPub)
		}
	}

	return published
}

func (e *ResourceLinksSnapshotTaken) HandleEventResourceLinksPublished(ctx context.Context, pub *ResourceLinksPublished) ([]*commands.Resource, error) {
	published := e.GetNewPublishedLinks(pub)

	for _, res := range published {
		if e.GetResources() == nil {
			e.Resources = make(map[string]*commands.Resource)
		}
		e.GetResources()[res.GetHref()] = res
	}
	e.DeviceId = pub.GetDeviceId()
	e.EventMetadata = pub.GetEventMetadata()
	e.AuditContext = pub.GetAuditContext()
	return published, nil
}

func (e *ResourceLinksSnapshotTaken) HandleEventResourceLinksUnpublished(ctx context.Context, upub *ResourceLinksUnpublished) ([]string, error) {
	var unpublished []string
	if len(upub.GetHrefs()) == 0 {
		unpublished = make([]string, 0, len(e.GetResources()))
		for href := range e.GetResources() {
			unpublished = append(unpublished, href)
		}
		e.Resources = make(map[string]*commands.Resource)
	} else {
		unpublished = make([]string, 0, len(upub.GetHrefs()))
		if len(e.GetResources()) > 0 {
			for _, href := range upub.GetHrefs() {
				if _, present := e.GetResources()[href]; present {
					unpublished = append(unpublished, href)
					delete(e.GetResources(), href)
				}
			}
		}
	}
	e.EventMetadata = upub.GetEventMetadata()
	e.AuditContext = upub.GetAuditContext()
	return unpublished, nil
}

func (e *ResourceLinksSnapshotTaken) HandleEventResourceLinksSnapshotTaken(ctx context.Context, s *ResourceLinksSnapshotTaken) {
	e.CopyData(s)
}

func (e *ResourceLinksSnapshotTaken) Handle(ctx context.Context, iter eventstore.Iter) error {
	for {
		eu, ok := iter.Next(ctx)
		if !ok {
			break
		}
		if eu.EventType() == "" {
			return status.Errorf(codes.Internal, "cannot determine type of event")
		}
		switch eu.EventType() {
		case (&ResourceLinksSnapshotTaken{}).EventType():
			var s ResourceLinksSnapshotTaken
			if err := eu.Unmarshal(&s); err != nil {
				return status.Errorf(codes.Internal, "%v", err)
			}
			e.HandleEventResourceLinksSnapshotTaken(ctx, &s)
		case (&ResourceLinksPublished{}).EventType():
			var s ResourceLinksPublished
			if err := eu.Unmarshal(&s); err != nil {
				return status.Errorf(codes.Internal, "%v", err)
			}
			_, _ = e.HandleEventResourceLinksPublished(ctx, &s)
		case (&ResourceLinksUnpublished{}).EventType():
			var s ResourceLinksUnpublished
			if err := eu.Unmarshal(&s); err != nil {
				return status.Errorf(codes.Internal, "%v", err)
			}
			_, _ = e.HandleEventResourceLinksUnpublished(ctx, &s)
		}
	}
	return iter.Err()
}

func (e *ResourceLinksSnapshotTaken) HandleCommand(ctx context.Context, cmd aggregate.Command, newVersion uint64) ([]eventstore.Event, error) {
	userID, err := grpc.SubjectFromTokenMD(ctx)
	if err != nil {
		return nil, err
	}
	switch req := cmd.(type) {
	case *commands.PublishResourceLinksRequest:
		if req.GetCommandMetadata() == nil {
			return nil, status.Errorf(codes.InvalidArgument, errInvalidCommandMetadata)
		}

		em := MakeEventMeta(req.GetCommandMetadata().GetConnectionId(), req.GetCommandMetadata().GetSequence(), newVersion)
		ac := commands.NewAuditContext(userID, "")

		rlp := ResourceLinksPublished{
			Resources:     req.GetResources(),
			DeviceId:      req.GetDeviceId(),
			AuditContext:  ac,
			EventMetadata: em,
		}
		published, err := e.HandleEventResourceLinksPublished(ctx, &rlp)
		if err != nil {
			return nil, err
		}
		if len(published) == 0 {
			return nil, nil
		}
		rlp.Resources = published
		return []eventstore.Event{&rlp}, nil
	case *commands.UnpublishResourceLinksRequest:
		if newVersion == 0 {
			return nil, status.Errorf(codes.NotFound, errInvalidVersion)
		}
		if req.CommandMetadata == nil {
			return nil, status.Errorf(codes.InvalidArgument, errInvalidCommandMetadata)
		}

		em := MakeEventMeta(req.GetCommandMetadata().GetConnectionId(), req.GetCommandMetadata().GetSequence(), newVersion)
		ac := commands.NewAuditContext(userID, "")
		rlu := ResourceLinksUnpublished{
			Hrefs:         req.GetHrefs(),
			DeviceId:      req.GetDeviceId(),
			AuditContext:  ac,
			EventMetadata: em,
		}
		unpublished, err := e.HandleEventResourceLinksUnpublished(ctx, &rlu)
		if err != nil {
			return nil, err
		}
		if len(unpublished) == 0 {
			return nil, nil
		}
		rlu.Hrefs = unpublished
		return []eventstore.Event{&rlu}, nil
	}

	return nil, fmt.Errorf("unknown command (%T)", cmd)
}

func (e *ResourceLinksSnapshotTaken) TakeSnapshot(version uint64) (eventstore.Event, bool) {
	// we need to return as new event because `e` is a pointer,
	// otherwise ResourceLinksSnapshotTaken.Handle override version/resource of snapshot which will be fired to eventbus
	resources := make(map[string]*commands.Resource)
	for key, resource := range e.GetResources() {
		resources[key] = resource
	}
	return &ResourceLinksSnapshotTaken{
		DeviceId:      e.GetDeviceId(),
		EventMetadata: MakeEventMeta(e.GetEventMetadata().GetConnectionId(), e.GetEventMetadata().GetSequence(), version),
		Resources:     resources,
		AuditContext:  e.GetAuditContext(),
	}, true
}

func NewResourceLinksSnapshotTaken() *ResourceLinksSnapshotTaken {

	return &ResourceLinksSnapshotTaken{
		Resources:     make(map[string]*commands.Resource),
		EventMetadata: &EventMetadata{},
	}
}

func (e *ResourceLinksSnapshotTaken) ToResourceLinksPublished() *ResourceLinksPublished {
	resources := make([]*commands.Resource, 0, len(e.GetResources()))
	for _, r := range e.GetResources() {
		resources = append(resources, r)
	}

	return &ResourceLinksPublished{
		DeviceId:      e.GetDeviceId(),
		EventMetadata: e.GetEventMetadata(),
		Resources:     resources,
		AuditContext:  e.GetAuditContext(),
	}
}
