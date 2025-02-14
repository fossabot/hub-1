package service

import (
	"context"
	"fmt"

	"github.com/plgd-dev/hub/pkg/log"
	kitNetGrpc "github.com/plgd-dev/hub/pkg/net/grpc"
	"github.com/plgd-dev/hub/resource-aggregate/commands"
	cqrsAggregate "github.com/plgd-dev/hub/resource-aggregate/cqrs/aggregate"
	"github.com/plgd-dev/hub/resource-aggregate/cqrs/eventbus"
	"github.com/plgd-dev/hub/resource-aggregate/cqrs/eventstore"
	"github.com/plgd-dev/hub/resource-aggregate/cqrs/utils"
	raEvents "github.com/plgd-dev/hub/resource-aggregate/events"
	"google.golang.org/grpc/codes"
)

type getOwnerDevicesFunc = func(ctx context.Context, owner string, deviceIDs []string) ([]string, error)

//RequestHandler for handling incoming request
type RequestHandler struct {
	UnimplementedResourceAggregateServer
	config              Config
	eventstore          EventStore
	publisher           eventbus.Publisher
	getOwnerDevicesFunc getOwnerDevicesFunc
}

//NewRequestHandler factory for new RequestHandler
func NewRequestHandler(config Config, eventstore EventStore, publisher eventbus.Publisher, getOwnerDevicesFunc getOwnerDevicesFunc) *RequestHandler {
	return &RequestHandler{
		config:              config,
		eventstore:          eventstore,
		publisher:           publisher,
		getOwnerDevicesFunc: getOwnerDevicesFunc,
	}
}

func PublishEvents(publisher eventbus.Publisher, owner, deviceId, resourceId string, events []eventbus.Event) error {
	var errors []error
	for _, event := range events {
		// timeout si driven by flusherTimeout.
		err := publisher.Publish(context.Background(), utils.GetPublishSubject(owner, event), deviceId, resourceId, event)
		if err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("cannot publish events: %v", errors)
	}
	return nil
}

// Check if device with given ID belongs to given owner
func (r RequestHandler) isUserDevice(ctx context.Context, owner string, deviceID string) (bool, error) {
	deviceIds, err := r.getOwnerDevicesFunc(ctx, owner, []string{deviceID})
	if err != nil {
		return false, err
	}
	return len(deviceIds) == 1, nil
}

func (r RequestHandler) validateAccessToDevice(ctx context.Context, deviceID string) (string, error) {
	owner, err := kitNetGrpc.OwnerFromTokenMD(ctx, r.config.APIs.GRPC.Authorization.OwnerClaim)
	if err != nil {
		return "", kitNetGrpc.ForwardErrorf(codes.InvalidArgument, "invalid owner: %v", err)
	}
	ok, err := r.isUserDevice(ctx, owner, deviceID)
	if err != nil {
		return "", kitNetGrpc.ForwardErrorf(codes.Internal, "cannot validate: %v", err)
	}
	if !ok {
		return "", kitNetGrpc.ForwardErrorf(codes.PermissionDenied, "access denied")
	}
	return owner, nil
}

// Return owner and list of owned devices from the input slices.
//
// Function iterates over input slice of device IDs and returns owner name, and the intersection
// of the input device IDs with owned devices.
func (r RequestHandler) getOwnedDevices(ctx context.Context, deviceIDs []string) (string, []string, error) {
	owner, err := kitNetGrpc.OwnerFromTokenMD(ctx, r.config.APIs.GRPC.Authorization.OwnerClaim)
	if err != nil {
		return "", nil, kitNetGrpc.ForwardErrorf(codes.InvalidArgument, "invalid owner: %v", err)
	}

	ownedDevices, err := r.getOwnerDevicesFunc(ctx, owner, deviceIDs)
	if err != nil {
		return "", nil, kitNetGrpc.ForwardErrorf(codes.InvalidArgument, "cannot validate: %v", err)
	}
	return owner, ownedDevices, nil
}

func (r RequestHandler) PublishResourceLinks(ctx context.Context, request *commands.PublishResourceLinksRequest) (*commands.PublishResourceLinksResponse, error) {
	owner, err := r.validateAccessToDevice(ctx, request.GetDeviceId())
	if err != nil {
		return nil, log.LogAndReturnError(kitNetGrpc.ForwardErrorf(codes.Internal, "cannot validate user access: %v", err))
	}

	resID := commands.NewResourceID(request.DeviceId, commands.ResourceLinksHref)
	aggregate, err := NewAggregate(resID, r.config.Clients.Eventstore.SnapshotThreshold, r.eventstore, ResourceLinksFactoryModel, cqrsAggregate.NewDefaultRetryFunc(r.config.Clients.Eventstore.ConcurrencyExceptionMaxRetry))
	if err != nil {
		return nil, log.LogAndReturnError(kitNetGrpc.ForwardErrorf(codes.InvalidArgument, "cannot publish resource links: %v", err))
	}

	events, err := aggregate.PublishResourceLinks(ctx, request)
	if err != nil {
		return nil, log.LogAndReturnError(kitNetGrpc.ForwardErrorf(codes.Internal, "cannot publish resource links: %v", err))
	}

	err = PublishEvents(r.publisher, owner, aggregate.DeviceID(), aggregate.ResourceID(), events)
	if err != nil {
		log.Errorf("cannot publish resource links published events: %v", err)
	}
	auditContext := commands.NewAuditContext(owner, "")
	return newPublishResourceLinksResponse(events, aggregate.DeviceID(), auditContext), nil
}

func newPublishResourceLinksResponse(events []eventstore.Event, deviceID string, auditContext *commands.AuditContext) *commands.PublishResourceLinksResponse {
	for _, event := range events {
		if rlp, ok := event.(*raEvents.ResourceLinksPublished); ok {
			return &commands.PublishResourceLinksResponse{
				AuditContext:       auditContext,
				PublishedResources: rlp.Resources,
				DeviceId:           deviceID,
			}
		}
	}
	return &commands.PublishResourceLinksResponse{
		AuditContext: auditContext,
		DeviceId:     deviceID,
	}
}

func (r RequestHandler) UnpublishResourceLinks(ctx context.Context, request *commands.UnpublishResourceLinksRequest) (*commands.UnpublishResourceLinksResponse, error) {
	owner, err := r.validateAccessToDevice(ctx, request.GetDeviceId())
	if err != nil {
		return nil, log.LogAndReturnError(kitNetGrpc.ForwardErrorf(codes.Internal, "cannot validate user access: %v", err))
	}

	resID := commands.NewResourceID(request.DeviceId, commands.ResourceLinksHref)
	aggregate, err := NewAggregate(resID, r.config.Clients.Eventstore.SnapshotThreshold, r.eventstore, ResourceLinksFactoryModel, cqrsAggregate.NewDefaultRetryFunc(r.config.Clients.Eventstore.ConcurrencyExceptionMaxRetry))
	if err != nil {
		return nil, log.LogAndReturnError(kitNetGrpc.ForwardErrorf(codes.InvalidArgument, "cannot unpublish resource links: %v", err))
	}

	events, err := aggregate.UnpublishResourceLinks(ctx, request)
	if err != nil {
		return nil, log.LogAndReturnError(kitNetGrpc.ForwardErrorf(codes.Internal, "cannot unpublish resource links: %v", err))
	}

	err = PublishEvents(r.publisher, owner, aggregate.DeviceID(), aggregate.ResourceID(), events)
	if err != nil {
		log.Errorf("cannot publish resource links unpublished events: %v", err)
	}
	auditContext := commands.NewAuditContext(owner, "")
	return newUnpublishResourceLinksResponse(events, aggregate.DeviceID(), auditContext), nil
}

func newUnpublishResourceLinksResponse(events []eventstore.Event, deviceID string, auditContext *commands.AuditContext) *commands.UnpublishResourceLinksResponse {
	for _, event := range events {
		if rlu, ok := event.(*raEvents.ResourceLinksUnpublished); ok {
			return &commands.UnpublishResourceLinksResponse{
				AuditContext:     auditContext,
				UnpublishedHrefs: rlu.Hrefs,
				DeviceId:         deviceID,
			}
		}
	}
	return &commands.UnpublishResourceLinksResponse{
		AuditContext: auditContext,
		DeviceId:     deviceID,
	}
}

func (r RequestHandler) NotifyResourceChanged(ctx context.Context, request *commands.NotifyResourceChangedRequest) (*commands.NotifyResourceChangedResponse, error) {
	owner, err := r.validateAccessToDevice(ctx, request.GetResourceId().GetDeviceId())
	if err != nil {
		return nil, log.LogAndReturnError(kitNetGrpc.ForwardErrorf(codes.Internal, "cannot validate user access: %v", err))
	}
	aggregate, err := NewAggregate(request.ResourceId, r.config.Clients.Eventstore.SnapshotThreshold, r.eventstore, ResourceStateFactoryModel, cqrsAggregate.NewDefaultRetryFunc(r.config.Clients.Eventstore.ConcurrencyExceptionMaxRetry))
	if err != nil {
		return nil, log.LogAndReturnError(kitNetGrpc.ForwardErrorf(codes.InvalidArgument, "cannot notify about resource content change: %v", err))
	}

	events, err := aggregate.NotifyResourceChanged(ctx, request)
	if err != nil {
		return nil, log.LogAndReturnError(kitNetGrpc.ForwardErrorf(codes.Internal, "cannot notify about resource content change: %v", err))
	}

	err = PublishEvents(r.publisher, owner, aggregate.DeviceID(), aggregate.ResourceID(), events)
	if err != nil {
		log.Errorf("cannot publish resource content changed notification events: %v", err)
	}
	auditContext := commands.NewAuditContext(owner, "")
	return &commands.NotifyResourceChangedResponse{
		AuditContext: auditContext,
	}, nil
}

func (r RequestHandler) UpdateResource(ctx context.Context, request *commands.UpdateResourceRequest) (*commands.UpdateResourceResponse, error) {
	owner, err := r.validateAccessToDevice(ctx, request.GetResourceId().GetDeviceId())
	if err != nil {
		return nil, log.LogAndReturnError(kitNetGrpc.ForwardErrorf(codes.Internal, "cannot validate user access: %v", err))
	}
	request.TimeToLive = checkTimeToLiveForDefault(r.config.Clients.Eventstore.DefaultCommandTimeToLive, request.GetTimeToLive())

	aggregate, err := NewAggregate(request.ResourceId, r.config.Clients.Eventstore.SnapshotThreshold, r.eventstore, ResourceStateFactoryModel, cqrsAggregate.NewDefaultRetryFunc(r.config.Clients.Eventstore.ConcurrencyExceptionMaxRetry))
	if err != nil {
		return nil, log.LogAndReturnError(kitNetGrpc.ForwardErrorf(codes.InvalidArgument, "cannot update resource content: %v", err))
	}

	events, err := aggregate.UpdateResource(ctx, request)
	if err != nil {
		return nil, log.LogAndReturnError(kitNetGrpc.ForwardErrorf(codes.Internal, "cannot update resource content: %v", err))
	}

	err = PublishEvents(r.publisher, owner, aggregate.DeviceID(), aggregate.ResourceID(), events)
	if err != nil {
		log.Errorf("cannot publish resource content update events: %v", err)
	}

	var validUntil int64
	for _, e := range events {
		if ev, ok := e.(*raEvents.ResourceUpdatePending); ok {
			validUntil = ev.GetValidUntil()
			break
		}
	}

	auditContext := commands.NewAuditContext(owner, request.GetCorrelationId())
	return &commands.UpdateResourceResponse{
		AuditContext: auditContext,
		ValidUntil:   validUntil,
	}, nil
}

func (r RequestHandler) ConfirmResourceUpdate(ctx context.Context, request *commands.ConfirmResourceUpdateRequest) (*commands.ConfirmResourceUpdateResponse, error) {
	owner, err := r.validateAccessToDevice(ctx, request.GetResourceId().GetDeviceId())
	if err != nil {
		return nil, log.LogAndReturnError(kitNetGrpc.ForwardErrorf(codes.Internal, "cannot validate user access: %v", err))
	}
	aggregate, err := NewAggregate(request.ResourceId, r.config.Clients.Eventstore.SnapshotThreshold, r.eventstore, ResourceStateFactoryModel, cqrsAggregate.NewDefaultRetryFunc(r.config.Clients.Eventstore.ConcurrencyExceptionMaxRetry))
	if err != nil {
		return nil, log.LogAndReturnError(kitNetGrpc.ForwardErrorf(codes.InvalidArgument, "cannot confirm resource content update: %v", err))
	}

	events, err := aggregate.ConfirmResourceUpdate(ctx, request)
	if err != nil {
		return nil, log.LogAndReturnError(kitNetGrpc.ForwardErrorf(codes.Internal, "cannot confirm resource content update: %v", err))
	}

	err = PublishEvents(r.publisher, owner, aggregate.DeviceID(), aggregate.ResourceID(), events)
	if err != nil {
		log.Errorf("cannot publish resource content update confirmation events: %v", err)
	}
	auditContext := commands.NewAuditContext(owner, request.GetCorrelationId())
	return &commands.ConfirmResourceUpdateResponse{
		AuditContext: auditContext,
	}, nil
}

func (r RequestHandler) RetrieveResource(ctx context.Context, request *commands.RetrieveResourceRequest) (*commands.RetrieveResourceResponse, error) {
	owner, err := r.validateAccessToDevice(ctx, request.GetResourceId().GetDeviceId())
	if err != nil {
		return nil, log.LogAndReturnError(kitNetGrpc.ForwardErrorf(codes.Internal, "cannot validate user access: %v", err))
	}
	request.TimeToLive = checkTimeToLiveForDefault(r.config.Clients.Eventstore.DefaultCommandTimeToLive, request.GetTimeToLive())

	aggregate, err := NewAggregate(request.ResourceId, r.config.Clients.Eventstore.SnapshotThreshold, r.eventstore, ResourceStateFactoryModel, cqrsAggregate.NewDefaultRetryFunc(r.config.Clients.Eventstore.ConcurrencyExceptionMaxRetry))
	if err != nil {
		return nil, log.LogAndReturnError(kitNetGrpc.ForwardErrorf(codes.InvalidArgument, "cannot retrieve resource content: %v", err))
	}

	events, err := aggregate.RetrieveResource(ctx, request)
	if err != nil {
		return nil, log.LogAndReturnError(kitNetGrpc.ForwardErrorf(codes.Internal, "cannot retrieve resource content: %v", err))
	}

	err = PublishEvents(r.publisher, owner, aggregate.DeviceID(), aggregate.ResourceID(), events)
	if err != nil {
		log.Errorf("cannot publish resource content retrieve events: %v", err)
	}

	var validUntil int64
	for _, e := range events {
		if ev, ok := e.(*raEvents.ResourceRetrievePending); ok {
			validUntil = ev.GetValidUntil()
			break
		}
	}

	auditContext := commands.NewAuditContext(owner, request.GetCorrelationId())
	return &commands.RetrieveResourceResponse{
		AuditContext: auditContext,
		ValidUntil:   validUntil,
	}, nil
}

func (r RequestHandler) ConfirmResourceRetrieve(ctx context.Context, request *commands.ConfirmResourceRetrieveRequest) (*commands.ConfirmResourceRetrieveResponse, error) {
	owner, err := r.validateAccessToDevice(ctx, request.GetResourceId().GetDeviceId())
	if err != nil {
		return nil, log.LogAndReturnError(kitNetGrpc.ForwardErrorf(codes.Internal, "cannot validate user access: %v", err))
	}
	aggregate, err := NewAggregate(request.ResourceId, r.config.Clients.Eventstore.SnapshotThreshold, r.eventstore, ResourceStateFactoryModel, cqrsAggregate.NewDefaultRetryFunc(r.config.Clients.Eventstore.ConcurrencyExceptionMaxRetry))
	if err != nil {
		return nil, log.LogAndReturnError(kitNetGrpc.ForwardErrorf(codes.InvalidArgument, "ccannot confirm resource content retrieve: %v", err))
	}

	events, err := aggregate.ConfirmResourceRetrieve(ctx, request)
	if err != nil {
		return nil, log.LogAndReturnError(kitNetGrpc.ForwardErrorf(codes.Internal, "cannot confirm resource content retrieve: %v", err))
	}

	err = PublishEvents(r.publisher, owner, aggregate.DeviceID(), aggregate.ResourceID(), events)
	if err != nil {
		log.Errorf("cannot publish resource content retrieve confirmation events: %v", err)
	}

	auditContext := commands.NewAuditContext(owner, request.GetCorrelationId())
	return &commands.ConfirmResourceRetrieveResponse{
		AuditContext: auditContext,
	}, nil
}

func (r RequestHandler) DeleteResource(ctx context.Context, request *commands.DeleteResourceRequest) (*commands.DeleteResourceResponse, error) {
	owner, err := r.validateAccessToDevice(ctx, request.GetResourceId().GetDeviceId())
	if err != nil {
		return nil, log.LogAndReturnError(kitNetGrpc.ForwardErrorf(codes.Internal, "cannot validate user access: %v", err))
	}
	request.TimeToLive = checkTimeToLiveForDefault(r.config.Clients.Eventstore.DefaultCommandTimeToLive, request.GetTimeToLive())

	aggregate, err := NewAggregate(request.ResourceId, r.config.Clients.Eventstore.SnapshotThreshold, r.eventstore, ResourceStateFactoryModel, cqrsAggregate.NewDefaultRetryFunc(r.config.Clients.Eventstore.ConcurrencyExceptionMaxRetry))
	if err != nil {
		return nil, log.LogAndReturnError(kitNetGrpc.ForwardErrorf(codes.InvalidArgument, "cannot delete resource: %v", err))
	}

	events, err := aggregate.DeleteResource(ctx, request)
	if err != nil {
		return nil, log.LogAndReturnError(kitNetGrpc.ForwardErrorf(codes.Internal, "cannot delete resource: %v", err))
	}

	err = PublishEvents(r.publisher, owner, aggregate.DeviceID(), aggregate.ResourceID(), events)
	if err != nil {
		log.Errorf("cannot publish delete resource events: %v", err)
	}

	var validUntil int64
	for _, e := range events {
		if ev, ok := e.(*raEvents.ResourceDeletePending); ok {
			validUntil = ev.GetValidUntil()
			break
		}
	}

	auditContext := commands.NewAuditContext(owner, request.GetCorrelationId())
	return &commands.DeleteResourceResponse{
		AuditContext: auditContext,
		ValidUntil:   validUntil,
	}, nil
}

func (r RequestHandler) ConfirmResourceDelete(ctx context.Context, request *commands.ConfirmResourceDeleteRequest) (*commands.ConfirmResourceDeleteResponse, error) {
	owner, err := r.validateAccessToDevice(ctx, request.GetResourceId().GetDeviceId())
	if err != nil {
		return nil, log.LogAndReturnError(kitNetGrpc.ForwardErrorf(codes.Internal, "cannot validate user access: %v", err))
	}

	aggregate, err := NewAggregate(request.ResourceId, r.config.Clients.Eventstore.SnapshotThreshold, r.eventstore, ResourceStateFactoryModel, cqrsAggregate.NewDefaultRetryFunc(r.config.Clients.Eventstore.ConcurrencyExceptionMaxRetry))
	if err != nil {
		return nil, log.LogAndReturnError(kitNetGrpc.ForwardErrorf(codes.InvalidArgument, "cannot confirm resource deletion: %v", err))
	}

	events, err := aggregate.ConfirmResourceDelete(ctx, request)
	if err != nil {
		return nil, log.LogAndReturnError(kitNetGrpc.ForwardErrorf(codes.Internal, "cannot confirm resource deletion: %v", err))
	}

	err = PublishEvents(r.publisher, owner, aggregate.DeviceID(), aggregate.ResourceID(), events)
	if err != nil {
		log.Errorf("cannot publish resource delete confirmation events: %v", err)
	}

	auditContext := commands.NewAuditContext(owner, request.GetCorrelationId())
	return &commands.ConfirmResourceDeleteResponse{
		AuditContext: auditContext,
	}, nil
}

func (r RequestHandler) CreateResource(ctx context.Context, request *commands.CreateResourceRequest) (*commands.CreateResourceResponse, error) {
	owner, err := r.validateAccessToDevice(ctx, request.GetResourceId().GetDeviceId())
	if err != nil {
		return nil, log.LogAndReturnError(kitNetGrpc.ForwardErrorf(codes.Internal, "cannot validate user access: %v", err))
	}
	request.TimeToLive = checkTimeToLiveForDefault(r.config.Clients.Eventstore.DefaultCommandTimeToLive, request.GetTimeToLive())

	aggregate, err := NewAggregate(request.ResourceId, r.config.Clients.Eventstore.SnapshotThreshold, r.eventstore, ResourceStateFactoryModel, cqrsAggregate.NewDefaultRetryFunc(r.config.Clients.Eventstore.ConcurrencyExceptionMaxRetry))
	if err != nil {
		return nil, log.LogAndReturnError(kitNetGrpc.ForwardErrorf(codes.InvalidArgument, "cannot create resource: %v", err))
	}

	events, err := aggregate.CreateResource(ctx, request)
	if err != nil {
		return nil, log.LogAndReturnError(kitNetGrpc.ForwardErrorf(codes.Internal, "cannot create resource: %v", err))
	}

	err = PublishEvents(r.publisher, owner, aggregate.DeviceID(), aggregate.ResourceID(), events)
	if err != nil {
		log.Errorf("cannot publish resource create events: %v", err)
	}

	var validUntil int64
	for _, e := range events {
		if ev, ok := e.(*raEvents.ResourceCreatePending); ok {
			validUntil = ev.GetValidUntil()
			break
		}
	}

	auditContext := commands.NewAuditContext(owner, request.GetCorrelationId())
	return &commands.CreateResourceResponse{
		AuditContext: auditContext,
		ValidUntil:   validUntil,
	}, nil
}

func (r RequestHandler) ConfirmResourceCreate(ctx context.Context, request *commands.ConfirmResourceCreateRequest) (*commands.ConfirmResourceCreateResponse, error) {
	owner, err := r.validateAccessToDevice(ctx, request.GetResourceId().GetDeviceId())
	if err != nil {
		return nil, log.LogAndReturnError(kitNetGrpc.ForwardErrorf(codes.Internal, "cannot validate user access: %v", err))
	}

	aggregate, err := NewAggregate(request.ResourceId, r.config.Clients.Eventstore.SnapshotThreshold, r.eventstore, ResourceStateFactoryModel, cqrsAggregate.NewDefaultRetryFunc(r.config.Clients.Eventstore.ConcurrencyExceptionMaxRetry))
	if err != nil {
		return nil, log.LogAndReturnError(kitNetGrpc.ForwardErrorf(codes.InvalidArgument, "cannot confirm resource creation: %v", err))
	}

	events, err := aggregate.ConfirmResourceCreate(ctx, request)
	if err != nil {
		return nil, log.LogAndReturnError(kitNetGrpc.ForwardErrorf(codes.Internal, "cannot confirm resource creation: %v", err))
	}

	err = PublishEvents(r.publisher, owner, aggregate.DeviceID(), aggregate.ResourceID(), events)
	if err != nil {
		log.Errorf("cannot publish resource create confirmation events: %v", err)
	}
	auditContext := commands.NewAuditContext(owner, request.GetCorrelationId())
	return &commands.ConfirmResourceCreateResponse{
		AuditContext: auditContext,
	}, nil
}
