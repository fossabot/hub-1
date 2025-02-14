package test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	deviceClient "github.com/plgd-dev/device/client"
	"github.com/plgd-dev/device/client/core"
	"github.com/plgd-dev/device/schema"
	"github.com/plgd-dev/device/schema/acl"
	"github.com/plgd-dev/device/schema/collection"
	"github.com/plgd-dev/device/schema/configuration"
	"github.com/plgd-dev/device/schema/device"
	"github.com/plgd-dev/device/schema/interfaces"
	"github.com/plgd-dev/device/schema/platform"
	"github.com/plgd-dev/device/test/resource/types"
	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/plgd-dev/hub/grpc-gateway/client"
	"github.com/plgd-dev/hub/grpc-gateway/pb"
	"github.com/plgd-dev/hub/resource-aggregate/commands"
	"github.com/plgd-dev/hub/resource-aggregate/events"
	"github.com/plgd-dev/hub/test/config"
	oauthTest "github.com/plgd-dev/hub/test/oauth-server/test"
	"github.com/plgd-dev/kit/v2/codec/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
)

var (
	TestDeviceName string

	TestDevsimResources        []schema.ResourceLink
	TestDevsimBackendResources []schema.ResourceLink
)

const (
	TestResourceSwitchesHref = "/switches"
)

func TestResourceLightInstanceHref(id string) string {
	return "/light/" + id
}

func TestResourceSwitchesInstanceHref(id string) string {
	return TestResourceSwitchesHref + "/" + id
}

func init() {
	TestDeviceName = "devsim-" + MustGetHostname()
	TestDevsimResources = []schema.ResourceLink{
		{
			Href:          platform.ResourceURI,
			ResourceTypes: []string{platform.ResourceType},
			Interfaces:    []string{interfaces.OC_IF_R, interfaces.OC_IF_BASELINE},
			Policy: &schema.Policy{
				BitMask: 3,
			},
		},

		{
			Href:          device.ResourceURI,
			ResourceTypes: []string{types.DEVICE_CLOUD, device.ResourceType},
			Interfaces:    []string{interfaces.OC_IF_R, interfaces.OC_IF_BASELINE},
			Policy: &schema.Policy{
				BitMask: 3,
			},
		},

		{
			Href:          configuration.ResourceURI,
			ResourceTypes: []string{configuration.ResourceType},
			Interfaces:    []string{interfaces.OC_IF_RW, interfaces.OC_IF_BASELINE},
			Policy: &schema.Policy{
				BitMask: 3,
			},
		},

		{
			Href:          TestResourceLightInstanceHref("1"),
			ResourceTypes: []string{types.CORE_LIGHT},
			Interfaces:    []string{interfaces.OC_IF_RW, interfaces.OC_IF_BASELINE},
			Policy: &schema.Policy{
				BitMask: 3,
			},
		},

		{
			Href:          TestResourceSwitchesHref,
			ResourceTypes: []string{collection.ResourceType},
			Interfaces:    []string{interfaces.OC_IF_LL, interfaces.OC_IF_CREATE, interfaces.OC_IF_B, interfaces.OC_IF_BASELINE},
			Policy: &schema.Policy{
				BitMask: 3,
			},
		},
	}
}

func FilterResourceLink(filter func(schema.ResourceLink) bool, links []schema.ResourceLink) []schema.ResourceLink {
	var l []schema.ResourceLink
	for _, link := range links {
		if filter(link) {
			l = append(l, link)
		}
	}
	return l
}

func DefaultSwitchResourceLink(id string) schema.ResourceLink {
	return schema.ResourceLink{
		Href:          TestResourceSwitchesInstanceHref(id),
		ResourceTypes: []string{types.BINARY_SWITCH},
		Interfaces:    []string{interfaces.OC_IF_A, interfaces.OC_IF_BASELINE},
		Policy: &schema.Policy{
			BitMask: schema.BitMask(schema.Discoverable | schema.Observable),
		},
	}
}

func MakeSwitchResourceData(overrides map[string]interface{}) map[string]interface{} {
	data := MakeSwitchResourceDefaultData()
	for k, v := range overrides {
		data[k] = v
	}
	return data
}

func MakeSwitchResourceDefaultData() map[string]interface{} {
	s := DefaultSwitchResourceLink("")
	return map[string]interface{}{
		"if": s.Interfaces,
		"rt": s.ResourceTypes,
		"rep": map[string]interface{}{
			"value": false,
		},
		"p": map[string]interface{}{
			"bm": uint64(s.Policy.BitMask),
		},
	}
}

func AddDeviceSwitchResources(ctx context.Context, t *testing.T, deviceID string, c pb.GrpcGatewayClient, resourceIDs ...string) []schema.ResourceLink {
	toStringArray := func(v interface{}) []string {
		var result []string
		arr, ok := v.([]interface{})
		require.True(t, ok)
		for _, val := range arr {
			str, ok := val.(string)
			require.True(t, ok)
			result = append(result, str)
		}
		return result
	}

	links := make([]schema.ResourceLink, 0, len(resourceIDs))
	for _, resourceID := range resourceIDs {
		req := &pb.CreateResourceRequest{
			ResourceId: commands.NewResourceID(deviceID, TestResourceSwitchesHref),
			Content: &pb.Content{
				ContentType: message.AppOcfCbor.String(),
				Data:        EncodeToCbor(t, MakeSwitchResourceDefaultData()),
			},
		}
		resp, err := c.CreateResource(ctx, req)
		require.NoError(t, err)

		respData, ok := DecodeCbor(t, resp.GetData().GetContent().GetData()).(map[interface{}]interface{})
		require.True(t, ok)

		href, ok := respData["href"].(string)
		require.True(t, ok)
		require.Equal(t, TestResourceSwitchesInstanceHref(resourceID), href)

		resourceTypes := toStringArray(respData["rt"])
		interfaces := toStringArray(respData["if"])

		policy, ok := respData["p"].(map[interface{}]interface{})
		require.True(t, ok)
		bitmask, ok := policy["bm"].(uint64)
		require.True(t, ok)

		links = append(links, schema.ResourceLink{
			Href:          href,
			ResourceTypes: resourceTypes,
			Interfaces:    interfaces,
			Policy: &schema.Policy{
				BitMask: schema.BitMask(bitmask),
			},
		})
	}
	return links
}

func setAccessForCloud(ctx context.Context, t *testing.T, c *deviceClient.Client, deviceID string) {
	cloudSID := config.HubID()
	require.NotEmpty(t, cloudSID)

	d, links, err := c.GetRefDevice(ctx, deviceID)
	require.NoError(t, err)

	defer func() {
		err := d.Release(ctx)
		require.NoError(t, err)
	}()
	p, err := d.Provision(ctx, links)
	require.NoError(t, err)
	defer func() {
		_ = p.Close(ctx)
	}()

	link, err := core.GetResourceLink(links, acl.ResourceURI)
	require.NoError(t, err)

	setAcl := acl.UpdateRequest{
		AccessControlList: []acl.AccessControl{
			{
				Permission: acl.AllPermissions,
				Subject: acl.Subject{
					Subject_Device: &acl.Subject_Device{
						DeviceID: cloudSID,
					},
				},
				Resources: acl.AllResources,
			},
		},
	}

	err = p.UpdateResource(ctx, link, setAcl, nil)
	require.NoError(t, err)
}

func OnboardDevSimForClient(ctx context.Context, t *testing.T, c pb.GrpcGatewayClient, clientId, deviceID, gwHost string, expectedResources []schema.ResourceLink) (string, func()) {
	cloudSID := config.HubID()
	require.NotEmpty(t, cloudSID)
	devClient, err := NewSDKClient()
	require.NoError(t, err)
	defer func() {
		_ = devClient.Close(ctx)
	}()
	deviceID, err = devClient.OwnDevice(ctx, deviceID, deviceClient.WithOTM(deviceClient.OTMType_JustWorks))
	require.NoError(t, err)

	setAccessForCloud(ctx, t, devClient, deviceID)

	code := oauthTest.GetDeviceAuthorizationCode(t, config.OAUTH_SERVER_HOST, clientId, deviceID)

	onboard := func() {
		err = devClient.OnboardDevice(ctx, deviceID, config.DEVICE_PROVIDER, "coaps+tcp://"+gwHost, code, cloudSID)
		require.NoError(t, err)
	}
	if len(expectedResources) > 0 {
		subClient, err := client.New(c).GrpcGatewayClient().SubscribeToEvents(ctx)
		require.NoError(t, err)
		err = subClient.Send(&pb.SubscribeToEvents{
			CorrelationId: "allEvents",
			Action: &pb.SubscribeToEvents_CreateSubscription_{
				CreateSubscription: &pb.SubscribeToEvents_CreateSubscription{},
			},
		})
		require.NoError(t, err)
		ev, err := subClient.Recv()
		require.NoError(t, err)
		expectedEvent := &pb.Event{
			SubscriptionId: ev.SubscriptionId,
			CorrelationId:  "allEvents",
			Type: &pb.Event_OperationProcessed_{
				OperationProcessed: &pb.Event_OperationProcessed{
					ErrorStatus: &pb.Event_OperationProcessed_ErrorStatus{
						Code: pb.Event_OperationProcessed_ErrorStatus_OK,
					},
				},
			},
		}
		CheckProtobufs(t, expectedEvent, ev, RequireToCheckFunc(require.Equal))
		onboard()
		waitForDevice(ctx, t, subClient, deviceID, ev.GetSubscriptionId(), ev.GetCorrelationId(), expectedResources)
		err = subClient.CloseSend()
		require.NoError(t, err)
	} else {
		onboard()
	}

	return deviceID, func() {
		client, err := NewSDKClient()
		require.NoError(t, err)
		err = client.DisownDevice(ctx, deviceID)
		require.NoError(t, err)
		err = client.Close(ctx)
		require.NoError(t, err)
		time.Sleep(time.Second * 2)
	}
}

func OnboardDevSim(ctx context.Context, t *testing.T, c pb.GrpcGatewayClient, deviceID, gwHost string, expectedResources []schema.ResourceLink) (string, func()) {
	return OnboardDevSimForClient(ctx, t, c, config.OAUTH_MANAGER_CLIENT_ID, deviceID, gwHost, expectedResources)
}

func waitForDevice(ctx context.Context, t *testing.T, client pb.GrpcGateway_SubscribeToEventsClient, deviceID, subID, correlationID string, expectedResources []schema.ResourceLink) {
	getID := func(ev *pb.Event) string {
		switch v := ev.GetType().(type) {
		case *pb.Event_DeviceRegistered_:
			return fmt.Sprintf("%T", ev.GetType())
		case *pb.Event_DeviceMetadataUpdated:
			return fmt.Sprintf("%T", ev.GetType())
		case *pb.Event_ResourcePublished:
			return fmt.Sprintf("%T", ev.GetType())
		case *pb.Event_ResourceChanged:
			return fmt.Sprintf("%T:%v", ev.GetType(), v.ResourceChanged.GetResourceId().ToString())
		}
		return ""
	}

	cleanUpEvent := func(ev *pb.Event) {
		switch val := ev.GetType().(type) {
		case *pb.Event_DeviceMetadataUpdated:
			require.NotEmpty(t, val.DeviceMetadataUpdated.GetAuditContext().GetUserId())
			val.DeviceMetadataUpdated.AuditContext = nil
			require.NotZero(t, val.DeviceMetadataUpdated.GetEventMetadata().GetTimestamp())
			val.DeviceMetadataUpdated.EventMetadata = nil
		case *pb.Event_ResourcePublished:
			require.NotEmpty(t, val.ResourcePublished.GetAuditContext().GetUserId())
			val.ResourcePublished.AuditContext = nil
			require.NotZero(t, val.ResourcePublished.GetEventMetadata().GetTimestamp())
			val.ResourcePublished.EventMetadata = nil
			val.ResourcePublished.Resources = CleanUpResourcesArray(val.ResourcePublished.GetResources())
		case *pb.Event_ResourceChanged:
			require.NotEmpty(t, val.ResourceChanged.GetAuditContext().GetUserId())
			val.ResourceChanged.AuditContext = nil
			require.NotZero(t, val.ResourceChanged.GetEventMetadata().GetTimestamp())
			val.ResourceChanged.EventMetadata = nil
			require.NotEmpty(t, val.ResourceChanged.GetContent().GetData())
			val.ResourceChanged.Content = nil
		}
	}

	expectedEvents := map[string]*pb.Event{
		getID(&pb.Event{Type: &pb.Event_DeviceRegistered_{}}): {
			SubscriptionId: subID,
			CorrelationId:  correlationID,
			Type: &pb.Event_DeviceRegistered_{
				DeviceRegistered: &pb.Event_DeviceRegistered{
					DeviceIds: []string{deviceID},
				},
			},
		},
		getID(&pb.Event{Type: &pb.Event_DeviceMetadataUpdated{}}): {
			SubscriptionId: subID,
			CorrelationId:  correlationID,
			Type: &pb.Event_DeviceMetadataUpdated{
				DeviceMetadataUpdated: &events.DeviceMetadataUpdated{
					DeviceId: deviceID,
					Status: &commands.ConnectionStatus{
						Value: commands.ConnectionStatus_ONLINE,
					},
				},
			},
		},
		getID(&pb.Event{Type: &pb.Event_ResourcePublished{}}): {
			SubscriptionId: subID,
			CorrelationId:  correlationID,
			Type: &pb.Event_ResourcePublished{
				ResourcePublished: &events.ResourceLinksPublished{
					DeviceId:  deviceID,
					Resources: ResourceLinksToResources(deviceID, expectedResources),
				},
			},
		},
	}
	for _, r := range expectedResources {
		expectedEvents[getID(&pb.Event{Type: &pb.Event_ResourceChanged{
			ResourceChanged: &events.ResourceChanged{
				ResourceId: commands.NewResourceID(deviceID, r.Href),
			},
		}})] = &pb.Event{
			SubscriptionId: subID,
			CorrelationId:  correlationID,
			Type: &pb.Event_ResourceChanged{
				ResourceChanged: &events.ResourceChanged{
					ResourceId: commands.NewResourceID(deviceID, r.Href),
					Status:     commands.Status_OK,
				},
			},
		}
	}

	for {
		ev, err := client.Recv()
		require.NoError(t, err)

		expectedEvent, ok := expectedEvents[getID(ev)]
		if !ok {
			require.NoError(t, fmt.Errorf("unexpected event %+v", ev))
		}
		cleanUpEvent(ev)
		CheckProtobufs(t, expectedEvent, ev, RequireToCheckFunc(require.Equal))
		delete(expectedEvents, getID(ev))
		if len(expectedEvents) == 0 {
			return
		}
	}
}

func MustGetHostname() string {
	n, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	return n
}

func MustFindDeviceByName(name string) (deviceID string) {
	var err error
	for i := 0; i < 3; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		deviceID, err = FindDeviceByName(ctx, name)
		if err == nil {
			return deviceID
		}
	}
	panic(err)
}

type findDeviceIDByNameHandler struct {
	id     atomic.Value
	name   string
	cancel context.CancelFunc
}

func (h *findDeviceIDByNameHandler) Handle(ctx context.Context, dev *core.Device, deviceLinks schema.ResourceLinks) {
	defer func() {
		err := dev.Close(ctx)
		h.Error(err)
	}()
	l, ok := deviceLinks.GetResourceLink(device.ResourceURI)
	if !ok {
		return
	}
	var d device.Device
	err := dev.GetResource(ctx, l, &d)
	if err != nil {
		return
	}
	if d.Name == h.name {
		h.id.Store(d.ID)
		h.cancel()
	}
}

func (h *findDeviceIDByNameHandler) Error(err error) {}

func FindDeviceByName(ctx context.Context, name string) (deviceID string, _ error) {
	client := core.NewClient()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	h := findDeviceIDByNameHandler{
		name:   name,
		cancel: cancel,
	}

	err := client.GetDevices(ctx, &h)
	if err != nil {
		return "", fmt.Errorf("could not find the device named %s: %w", name, err)
	}
	id, ok := h.id.Load().(string)
	if !ok || id == "" {
		return "", fmt.Errorf("could not find the device named %s: not found", name)
	}
	return id, nil
}

func GetAllBackendResourceLinks() []schema.ResourceLink {
	return append(TestDevsimResources, TestDevsimBackendResources...)
}

func ProtobufToInterface(t *testing.T, val interface{}) interface{} {
	expJSON, err := json.Encode(val)
	require.NoError(t, err)
	var v interface{}
	err = json.Decode(expJSON, &v)
	require.NoError(t, err)
	return v
}

func RequireToCheckFunc(checFunc func(t require.TestingT, expected interface{}, actual interface{}, msgAndArgs ...interface{})) func(t *testing.T, expected interface{}, actual interface{}, msgAndArgs ...interface{}) {
	return func(t *testing.T, expected interface{}, actual interface{}, msgAndArgs ...interface{}) {
		checFunc(t, expected, actual, msgAndArgs)
	}
}

func AssertToCheckFunc(checFunc func(t assert.TestingT, expected interface{}, actual interface{}, msgAndArgs ...interface{}) bool) func(t *testing.T, expected interface{}, actual interface{}, msgAndArgs ...interface{}) {
	return func(t *testing.T, expected interface{}, actual interface{}, msgAndArgs ...interface{}) {
		checFunc(t, expected, actual, msgAndArgs)
	}
}

func CheckProtobufs(t *testing.T, expected interface{}, actual interface{}, checkFunc func(t *testing.T, expected interface{}, actual interface{}, msgAndArgs ...interface{})) {
	v1 := ProtobufToInterface(t, expected)
	v2 := ProtobufToInterface(t, actual)
	checkFunc(t, v1, v2)
}

func NATSSStart(ctx context.Context, t *testing.T) {
	err := exec.CommandContext(ctx, "docker", "start", "nats").Run()
	require.NoError(t, err)
}

func NATSSStop(ctx context.Context, t *testing.T) {
	err := exec.CommandContext(ctx, "docker", "stop", "nats").Run()
	require.NoError(t, err)
}
