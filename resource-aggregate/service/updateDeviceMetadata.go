package service

import (
	"context"
	"fmt"
	"time"

	"github.com/plgd-dev/hub/pkg/log"
	kitNetGrpc "github.com/plgd-dev/hub/pkg/net/grpc"
	"github.com/plgd-dev/hub/resource-aggregate/commands"
	cqrsAggregate "github.com/plgd-dev/hub/resource-aggregate/cqrs/aggregate"
	"github.com/plgd-dev/hub/resource-aggregate/cqrs/eventstore"
	"github.com/plgd-dev/hub/resource-aggregate/events"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func validateUpdateDeviceMetadata(request *commands.UpdateDeviceMetadataRequest) error {
	if err := checkTimeToLive(request.GetTimeToLive()); err != nil {
		return err
	}
	if request.GetDeviceId() == "" {
		return status.Errorf(codes.InvalidArgument, "invalid DeviceId")
	}
	if request.GetStatus() == nil && request.GetShadowSynchronization() == commands.ShadowSynchronization_UNSET {
		return status.Errorf(codes.InvalidArgument, "set.onlineStatus and set.shadowSynchronizationStatus are invalid")
	}

	return nil
}

func (a *aggregate) UpdateDeviceMetadata(ctx context.Context, request *commands.UpdateDeviceMetadataRequest) (events []eventstore.Event, err error) {
	if err = validateUpdateDeviceMetadata(request); err != nil {
		err = fmt.Errorf("invalid update device metadata command: %w", err)
		return
	}

	events, err = a.ag.HandleCommand(ctx, request)
	if err != nil {
		err = fmt.Errorf("unable to process update device metadata command command: %w", err)
		return
	}
	cleanUpToSnapshot(ctx, a, events)

	return
}

func checkTimeToLiveForDefault(defaultTimeToLive time.Duration, reqTimeToLive int64) int64 {
	if defaultTimeToLive == 0 {
		return reqTimeToLive
	}
	if reqTimeToLive != 0 {
		return reqTimeToLive
	}
	return int64(defaultTimeToLive)
}

func (r RequestHandler) UpdateDeviceMetadata(ctx context.Context, request *commands.UpdateDeviceMetadataRequest) (*commands.UpdateDeviceMetadataResponse, error) {
	owner, err := r.validateAccessToDevice(ctx, request.GetDeviceId())
	if err != nil {
		return nil, log.LogAndReturnError(kitNetGrpc.ForwardErrorf(codes.Internal, "cannot validate user access: %v", err))
	}
	request.TimeToLive = checkTimeToLiveForDefault(r.config.Clients.Eventstore.DefaultCommandTimeToLive, request.GetTimeToLive())

	resID := commands.NewResourceID(request.GetDeviceId(), commands.StatusHref)
	aggregate, err := NewAggregate(resID, r.config.Clients.Eventstore.SnapshotThreshold, r.eventstore, DeviceMetadataFactoryModel, cqrsAggregate.NewDefaultRetryFunc(r.config.Clients.Eventstore.ConcurrencyExceptionMaxRetry))
	if err != nil {
		return nil, log.LogAndReturnError(kitNetGrpc.ForwardErrorf(codes.InvalidArgument, "cannot update device('%v') metadata: %v", request.GetDeviceId(), err))
	}

	publishEvents, err := aggregate.UpdateDeviceMetadata(ctx, request)
	if err != nil {
		return nil, log.LogAndReturnError(kitNetGrpc.ForwardErrorf(codes.Internal, "cannot update device('%v') metadata: %v", request.GetDeviceId(), err))
	}

	err = PublishEvents(r.publisher, owner, aggregate.DeviceID(), aggregate.ResourceID(), publishEvents)
	if err != nil {
		log.Errorf("cannot publish device('%v') metadata events: %w", request.GetDeviceId(), err)
	}

	var validUntil int64
	for _, e := range publishEvents {
		if ev, ok := e.(*events.DeviceMetadataUpdatePending); ok {
			validUntil = ev.GetValidUntil()
			break
		}
	}

	return &commands.UpdateDeviceMetadataResponse{
		AuditContext: commands.NewAuditContext(owner, ""),
		ValidUntil:   validUntil,
	}, nil
}
