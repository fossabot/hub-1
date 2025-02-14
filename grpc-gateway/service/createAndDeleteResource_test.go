package service_test

import (
	"context"
	"crypto/tls"
	"testing"

	"github.com/plgd-dev/device/schema"
	"github.com/plgd-dev/device/schema/collection"
	"github.com/plgd-dev/device/schema/interfaces"
	"github.com/plgd-dev/device/test/resource/types"
	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/plgd-dev/hub/grpc-gateway/client"
	"github.com/plgd-dev/hub/grpc-gateway/pb"
	"github.com/plgd-dev/hub/pkg/log"
	kitNetGrpc "github.com/plgd-dev/hub/pkg/net/grpc"
	"github.com/plgd-dev/hub/resource-aggregate/commands"
	"github.com/plgd-dev/hub/resource-aggregate/events"
	"github.com/plgd-dev/hub/test"
	"github.com/plgd-dev/hub/test/config"
	oauthService "github.com/plgd-dev/hub/test/oauth-server/service"
	oauthTest "github.com/plgd-dev/hub/test/oauth-server/test"
	pbTest "github.com/plgd-dev/hub/test/pb"
	"github.com/plgd-dev/hub/test/service"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func subscribeToAllEvents(t *testing.T, ctx context.Context, c pb.GrpcGatewayClient, correlationId string) (pb.GrpcGateway_SubscribeToEventsClient, string) {
	subClient, err := client.New(c).GrpcGatewayClient().SubscribeToEvents(ctx)
	require.NoError(t, err)
	err = subClient.Send(&pb.SubscribeToEvents{
		CorrelationId: correlationId,
		Action: &pb.SubscribeToEvents_CreateSubscription_{
			CreateSubscription: &pb.SubscribeToEvents_CreateSubscription{},
		},
	})
	require.NoError(t, err)
	ev, err := subClient.Recv()
	require.NoError(t, err)
	expectedEvent := &pb.Event{
		SubscriptionId: ev.GetSubscriptionId(),
		CorrelationId:  correlationId,
		Type:           pbTest.OperationProcessedOK(),
	}
	test.CheckProtobufs(t, expectedEvent, ev, test.RequireToCheckFunc(require.Equal))
	return subClient, ev.GetSubscriptionId()
}

func createSwitchResource(t *testing.T, ctx context.Context, c pb.GrpcGatewayClient, deviceID, switchID string) {
	got, err := c.CreateResource(ctx, &pb.CreateResourceRequest{
		ResourceId: commands.NewResourceID(deviceID, test.TestResourceSwitchesHref),
		Content: &pb.Content{
			ContentType: message.AppOcfCbor.String(),
			Data:        test.EncodeToCbor(t, test.MakeSwitchResourceDefaultData()),
		},
	})
	require.NoError(t, err)
	switchData := pbTest.MakeCreateLightResourceResponseData(switchID)
	want := pbTest.MakeResourceCreated(t, deviceID, test.TestResourceSwitchesHref, switchData)
	pbTest.CmpResourceCreated(t, want, got.GetData())
}

func deleteSwitchResource(t *testing.T, ctx context.Context, c pb.GrpcGatewayClient, deviceID, switchID string) {
	got, err := c.DeleteResource(ctx, &pb.DeleteResourceRequest{
		ResourceId: commands.NewResourceID(deviceID, test.TestResourceSwitchesInstanceHref(switchID)),
	})
	require.NoError(t, err)
	want := pbTest.MakeResourceDeleted(t, deviceID, test.TestResourceSwitchesInstanceHref(switchID))
	pbTest.CmpResourceDeleted(t, want, got.GetData())
}

func createSwitchResourceExpectedEvents(t *testing.T, deviceID, subID, correlationID, switchID string) map[string]*pb.Event {
	cpEvent := &pb.Event{
		SubscriptionId: subID,
		CorrelationId:  correlationID,
		Type: &pb.Event_ResourceCreatePending{
			ResourceCreatePending: pbTest.MakeResourceCreatePending(t, deviceID, test.TestResourceSwitchesHref,
				test.MakeSwitchResourceDefaultData()),
		},
	}

	rchangedEvent := &pb.Event{
		SubscriptionId: subID,
		CorrelationId:  correlationID,
		Type: &pb.Event_ResourceChanged{
			ResourceChanged: pbTest.MakeResourceChanged(t, deviceID, test.TestResourceSwitchesHref, map[string]interface{}{
				"if": []interface{}{interfaces.OC_IF_LL, interfaces.OC_IF_CREATE, interfaces.OC_IF_B, interfaces.OC_IF_BASELINE},
				"links": []interface{}{
					map[string]interface{}{
						"href": test.TestResourceSwitchesInstanceHref(switchID),
						"if":   []string{interfaces.OC_IF_A, interfaces.OC_IF_BASELINE},
						"p": map[string]interface{}{
							"bm": uint64(schema.Discoverable | schema.Observable),
						},
						"rel": []interface{}{
							"hosts",
						},
						"rt": []interface{}{types.BINARY_SWITCH},
					},
				},
				"rt":                        []interface{}{collection.ResourceType},
				"rts":                       []interface{}{types.BINARY_SWITCH},
				"rts-m":                     []interface{}{types.BINARY_SWITCH},
				"x.org.openconnectivity.bl": uint64(94),
			}),
		},
	}

	rcreatedEvent := &pb.Event{
		SubscriptionId: subID,
		CorrelationId:  correlationID,
		Type: &pb.Event_ResourceCreated{
			ResourceCreated: pbTest.MakeResourceCreated(t, deviceID, test.TestResourceSwitchesHref,
				test.MakeSwitchResourceData(map[string]interface{}{
					"href": test.TestResourceSwitchesInstanceHref(switchID),
					"rep": map[string]interface{}{
						"if":    []string{interfaces.OC_IF_A, interfaces.OC_IF_BASELINE},
						"rt":    []string{types.BINARY_SWITCH},
						"value": false,
					},
				}),
			),
		},
	}

	rpublishedEvent := &pb.Event{
		SubscriptionId: subID,
		CorrelationId:  correlationID,
		Type: &pb.Event_ResourcePublished{
			ResourcePublished: &events.ResourceLinksPublished{
				DeviceId: deviceID,
				Resources: []*commands.Resource{
					{
						Href:          test.TestResourceSwitchesInstanceHref(switchID),
						DeviceId:      deviceID,
						ResourceTypes: []string{types.BINARY_SWITCH},
						Interfaces:    []string{interfaces.OC_IF_A, interfaces.OC_IF_BASELINE},
						Policy: &commands.Policy{
							BitFlags: int32(schema.Discoverable | schema.Observable),
						},
					},
				},
				AuditContext: commands.NewAuditContext(oauthService.DeviceUserID, ""),
			},
		},
	}

	rchangedEvent2 := &pb.Event{
		SubscriptionId: subID,
		CorrelationId:  correlationID,
		Type: &pb.Event_ResourceChanged{
			ResourceChanged: pbTest.MakeResourceChanged(t, deviceID, test.TestResourceSwitchesInstanceHref(switchID), map[string]interface{}{
				"if":    []string{interfaces.OC_IF_A, interfaces.OC_IF_BASELINE},
				"rt":    []string{types.BINARY_SWITCH},
				"value": false,
			}),
		},
	}

	return map[string]*pb.Event{
		pbTest.GetEventID(cpEvent):         cpEvent,
		pbTest.GetEventID(rchangedEvent):   rchangedEvent,
		pbTest.GetEventID(rcreatedEvent):   rcreatedEvent,
		pbTest.GetEventID(rpublishedEvent): rpublishedEvent,
		pbTest.GetEventID(rchangedEvent2):  rchangedEvent2,
	}
}

func deleteSwitchResourceExpectedEvents(t *testing.T, deviceID, subID, correlationID, switchID string) map[string]*pb.Event {
	deletePending := &pb.Event{
		SubscriptionId: subID,
		CorrelationId:  correlationID,
		Type: &pb.Event_ResourceDeletePending{
			ResourceDeletePending: &events.ResourceDeletePending{
				ResourceId:   commands.NewResourceID(deviceID, test.TestResourceSwitchesInstanceHref(switchID)),
				AuditContext: commands.NewAuditContext(oauthService.DeviceUserID, ""),
			},
		},
	}

	deleted := &pb.Event{
		SubscriptionId: subID,
		CorrelationId:  correlationID,
		Type: &pb.Event_ResourceDeleted{
			ResourceDeleted: pbTest.MakeResourceDeleted(t, deviceID, test.TestResourceSwitchesInstanceHref(switchID)),
		},
	}

	unpublished := &pb.Event{
		SubscriptionId: subID,
		CorrelationId:  correlationID,
		Type: &pb.Event_ResourceUnpublished{
			ResourceUnpublished: &events.ResourceLinksUnpublished{
				DeviceId:     deviceID,
				Hrefs:        []string{test.TestResourceSwitchesInstanceHref(switchID)},
				AuditContext: commands.NewAuditContext(oauthService.DeviceUserID, ""),
			},
		},
	}

	changed := &pb.Event{
		SubscriptionId: subID,
		CorrelationId:  correlationID,
		Type: &pb.Event_ResourceChanged{
			ResourceChanged: pbTest.MakeResourceChanged(t, deviceID, test.TestResourceSwitchesHref, map[string]interface{}{
				"if":                        []interface{}{interfaces.OC_IF_LL, interfaces.OC_IF_CREATE, interfaces.OC_IF_B, interfaces.OC_IF_BASELINE},
				"links":                     []interface{}{},
				"rt":                        []interface{}{collection.ResourceType},
				"rts":                       []interface{}{types.BINARY_SWITCH},
				"rts-m":                     []interface{}{types.BINARY_SWITCH},
				"x.org.openconnectivity.bl": uint64(94),
			}),
		},
	}

	return map[string]*pb.Event{
		pbTest.GetEventID(deletePending): deletePending,
		pbTest.GetEventID(deleted):       deleted,
		pbTest.GetEventID(unpublished):   unpublished,
		pbTest.GetEventID(changed):       changed,
	}
}

func validateEvents(t *testing.T, subClient pb.GrpcGateway_SubscribeToEventsClient, expectedEvents map[string]*pb.Event) {
	for {
		ev, err := subClient.Recv()
		if kitNetGrpc.IsContextDeadlineExceeded(err) {
			require.Failf(t, "missing events", "expected events not received: %+v", expectedEvents)
		}
		require.NoError(t, err)

		eventId := pbTest.GetEventID(ev)
		expected, ok := expectedEvents[eventId]
		if !ok {
			require.Failf(t, "unexpected event", "invalid event: %+v", ev)
		}
		pbTest.CmpEvent(t, expected, ev)
		delete(expectedEvents, eventId)
		if len(expectedEvents) == 0 {
			break
		}
	}
}

func TestCreateAndDeleteResource(t *testing.T) {
	deviceID := test.MustFindDeviceByName(test.TestDeviceName)
	ctx, cancel := context.WithTimeout(context.Background(), config.TEST_TIMEOUT)
	defer cancel()

	tearDown := service.SetUp(ctx, t)
	defer tearDown()
	log.Setup(log.Config{Debug: true})
	ctx = kitNetGrpc.CtxWithToken(ctx, oauthTest.GetDefaultServiceToken(t))

	conn, err := grpc.Dial(config.GRPC_HOST, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
		RootCAs: test.GetRootCertificatePool(t),
	})))
	require.NoError(t, err)
	c := pb.NewGrpcGatewayClient(conn)

	_, shutdownDevSim := test.OnboardDevSim(ctx, t, c, deviceID, config.GW_HOST, test.GetAllBackendResourceLinks())
	defer shutdownDevSim()

	const correlationID = "allEvents"
	subClient, subID := subscribeToAllEvents(t, ctx, c, correlationID)
	const switchID = "1"

	for i := 0; i < 5; i++ {
		createSwitchResource(t, ctx, c, deviceID, switchID)
		expectedCreateEvents := createSwitchResourceExpectedEvents(t, deviceID, subID, correlationID, switchID)
		validateEvents(t, subClient, expectedCreateEvents)

		deleteSwitchResource(t, ctx, c, deviceID, switchID)
		expectedDeleteEvents := deleteSwitchResourceExpectedEvents(t, deviceID, subID, correlationID, switchID)
		validateEvents(t, subClient, expectedDeleteEvents)
	}
}
