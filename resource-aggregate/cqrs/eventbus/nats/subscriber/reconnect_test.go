package subscriber_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/plgd-dev/hub/pkg/log"
	"github.com/plgd-dev/hub/resource-aggregate/cqrs/eventbus/nats/publisher"
	"github.com/plgd-dev/hub/resource-aggregate/cqrs/eventbus/nats/subscriber"
	natsTest "github.com/plgd-dev/hub/resource-aggregate/cqrs/eventbus/nats/test"
	"github.com/plgd-dev/hub/test"
	"github.com/plgd-dev/hub/test/config"
	"github.com/stretchr/testify/require"
)

func TestSubscriberReconnect(t *testing.T) {
	topics := []string{"test_subscriber_topic0" + uuid.Must(uuid.NewRandom()).String(), "test_subscriber_topic1" + uuid.Must(uuid.NewRandom()).String()}

	timeout := time.Second * 30

	logger, err := log.NewLogger(log.Config{})
	require.NoError(t, err)

	naPubClient, pub, err := natsTest.NewClientAndPublisher(config.MakePublisherConfig(), logger, publisher.WithMarshaler(json.Marshal))
	require.NoError(t, err)
	require.NotNil(t, pub)
	defer func() {
		pub.Close()
		naPubClient.Close()
	}()

	naSubClient, subscriber, err := natsTest.NewClientAndSubscriber(config.MakeSubscriberConfig(),
		logger,
		subscriber.WithGoPool(func(f func()) error { go f(); return nil }),
		subscriber.WithUnmarshaler(json.Unmarshal))
	require.NoError(t, err)
	require.NotNil(t, subscriber)
	defer func() {
		subscriber.Close()
		naSubClient.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Add handlers and observers.
	t.Log("Subscribe to first topic")
	m0, _ := testNewSubscription(t, ctx, subscriber, "sub-0", topics[0:1])

	AggregateID1 := "aggregateID1"
	aggregateID1Path := Path{
		AggregateId: AggregateID1,
		GroupId:     "deviceId",
	}

	eventsToPublish := []mockEvent{
		{
			EventTypeI:   "test0",
			AggregateIDI: AggregateID1,
		},
		{
			VersionI:     1,
			EventTypeI:   "test1",
			AggregateIDI: AggregateID1,
		},
	}

	err = pub.Publish(ctx, topics[0:1], aggregateID1Path.GroupId, aggregateID1Path.AggregateId, eventsToPublish[0])
	require.NoError(t, err)

	event0, err := m0.waitForEvent(timeout)
	require.NoError(t, err)
	require.Equal(t, eventsToPublish[0], event0)

	ch := make(chan bool)
	reconnectID := subscriber.AddReconnectFunc(func() {
		ch <- true
	})
	defer subscriber.RemoveReconnectFunc(reconnectID)

	test.NATSSStop(ctx, t)
	test.NATSSStart(ctx, t)

	select {
	case <-ch:
	case <-ctx.Done():
		require.NoError(t, fmt.Errorf("Timeout"))
	}
	naClient1, pub1, err := natsTest.NewClientAndPublisher(config.MakePublisherConfig(), logger, publisher.WithMarshaler(json.Marshal))
	require.NoError(t, err)
	require.NotNil(t, pub1)
	defer func() {
		pub1.Close()
		naClient1.Close()
	}()
	err = pub1.Publish(ctx, topics[0:1], aggregateID1Path.GroupId, aggregateID1Path.AggregateId, eventsToPublish[1])
	require.NoError(t, err)
	event0, err = m0.waitForEvent(timeout)
	require.NoError(t, err)
	require.Equal(t, eventsToPublish[1], event0)
}
