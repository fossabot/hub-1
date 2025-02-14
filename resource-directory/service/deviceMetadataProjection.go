package service

import (
	"context"
	"sync"

	"github.com/plgd-dev/hub/pkg/log"
	"github.com/plgd-dev/hub/resource-aggregate/cqrs/eventstore"
	"github.com/plgd-dev/hub/resource-aggregate/events"
)

type deviceMetadataProjection struct {
	lock sync.Mutex
	data *events.DeviceMetadataSnapshotTaken
}

func NewDeviceMetadataProjection() eventstore.Model {
	return &deviceMetadataProjection{}
}

func (p *deviceMetadataProjection) Clone() *deviceMetadataProjection {
	p.lock.Lock()
	defer p.lock.Unlock()

	data := &events.DeviceMetadataSnapshotTaken{
		DeviceId:              p.data.GetDeviceId(),
		DeviceMetadataUpdated: p.data.GetDeviceMetadataUpdated(),
		UpdatePendings:        p.data.GetUpdatePendings(),
		EventMetadata:         p.data.GetEventMetadata(),
	}

	return &deviceMetadataProjection{
		data: data,
	}
}

func (p *deviceMetadataProjection) EventType() string {
	s := &events.ResourceLinksSnapshotTaken{}
	return s.EventType()
}

func (p *deviceMetadataProjection) Handle(ctx context.Context, iter eventstore.Iter) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	for {
		eu, ok := iter.Next(ctx)
		if !ok {
			break
		}
		log.Debugf("deviceMetadataProjection.Handle deviceID=%v eventype%v version=%v", eu.GroupID(), eu.EventType(), eu.Version())
		if p.data == nil {
			p.data = &events.DeviceMetadataSnapshotTaken{
				DeviceId:      eu.GroupID(),
				EventMetadata: events.MakeEventMeta("", 0, eu.Version()),
			}
		}
		p.data.GetEventMetadata().Version = eu.Version()
		switch eu.EventType() {
		case (&events.DeviceMetadataSnapshotTaken{}).EventType():
			var e events.DeviceMetadataSnapshotTaken
			if err := eu.Unmarshal(&e); err != nil {
				return err
			}
			p.data = &e
		case (&events.DeviceMetadataUpdatePending{}).EventType():
			var e events.DeviceMetadataUpdatePending
			if err := eu.Unmarshal(&e); err != nil {
				return err
			}
			if err := p.data.HandleDeviceMetadataUpdatePending(ctx, &e); err != nil {
				continue
			}
			p.data.DeviceId = e.GetDeviceId()
		case (&events.DeviceMetadataUpdated{}).EventType():
			var e events.DeviceMetadataUpdated
			if err := eu.Unmarshal(&e); err != nil {
				return err
			}
			p.data.DeviceId = e.GetDeviceId()
			if _, err := p.data.HandleDeviceMetadataUpdated(ctx, &e, false); err != nil {
				continue
			}

		}
	}
	return nil
}
