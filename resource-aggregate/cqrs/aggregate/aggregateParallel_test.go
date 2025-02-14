package aggregate_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/plgd-dev/hub/pkg/log"
	"github.com/plgd-dev/hub/pkg/net/grpc"
	"github.com/plgd-dev/hub/resource-aggregate/commands"
	"github.com/plgd-dev/hub/resource-aggregate/cqrs/aggregate"
	"github.com/plgd-dev/hub/resource-aggregate/cqrs/eventstore"
	"github.com/plgd-dev/hub/resource-aggregate/cqrs/eventstore/mongodb"
	"github.com/plgd-dev/hub/resource-aggregate/events"
	"github.com/plgd-dev/hub/test/config"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
)

func testNewEventstore(ctx context.Context, t *testing.T) *mongodb.EventStore {
	logger, err := log.NewLogger(log.Config{})
	require.NoError(t, err)
	cfg := config.MakeEventsStoreMongoDBConfig()
	store, err := mongodb.New(
		ctx,
		cfg,
		logger,
	)
	require.NoError(t, err)
	require.NotNil(t, store)

	return store
}

func cleanUpToSnapshot(ctx context.Context, t *testing.T, store *mongodb.EventStore, evs []eventstore.Event) {
	for _, event := range evs {
		if event.IsSnapshot() {
			if err := store.RemoveUpToVersion(ctx, []eventstore.VersionQuery{{GroupID: event.GroupID(), AggregateID: event.AggregateID(), Version: event.Version()}}); err != nil {
				require.NoError(t, err)
			}
			fmt.Printf("snapshot at version %v\n", event.Version())
			break
		}
	}
}

//old 452.969s
//new 474.906s
func Test_parallelRequest(t *testing.T) {
	ctx := context.Background()
	token := config.CreateJwtToken(t, jwt.MapClaims{
		"sub": "test",
	})
	ctx = grpc.CtxWithIncomingToken(ctx, token)
	store := testNewEventstore(ctx, t)
	defer func() {
		err := store.Clear(ctx)
		require.NoError(t, err)
		_ = store.Close(ctx)
	}()

	deviceID := "7397398d-3ae8-4d9a-62d6-511f7b736a60"
	href := "/test/resource/1"

	newAggregate := func(deviceID, href string) *aggregate.Aggregate {
		a, err := aggregate.NewAggregate(deviceID, commands.NewResourceID(deviceID, href).ToUUID(), aggregate.NewDefaultRetryFunc(64), 16, store, func(context.Context) (aggregate.AggregateModel, error) {
			ev := events.NewResourceStateSnapshotTaken()
			ev.ResourceId = commands.NewResourceID(deviceID, href)
			return ev, nil
		}, nil)
		require.NoError(t, err)
		return a
	}

	numParallel := 3
	var wg sync.WaitGroup
	var anyError atomic.Error
	for i := 0; i < numParallel; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100000; j++ {
				if anyError.Load() != nil {
					return
				}
				commandContentChanged := commands.NotifyResourceChangedRequest{
					ResourceId: commands.NewResourceID(deviceID, href),
					Content: &commands.Content{
						Data:        []byte("hello world" + fmt.Sprintf("%v.%v", id, j)),
						ContentType: "text",
					},
					CommandMetadata: &commands.CommandMetadata{
						ConnectionId: uuid.New().String(),
					},
					Status: commands.Status_OK,
				}
				aggr := newAggregate(commandContentChanged.GetResourceId().GetDeviceId(), commandContentChanged.GetResourceId().GetHref())
				events, err := aggr.HandleCommand(ctx, &commandContentChanged)
				if err != nil {
					anyError.Store(err)
					return
				}
				cleanUpToSnapshot(ctx, t, store, events)
			}
		}(i)
	}
	wg.Wait()
	err := anyError.Load()
	require.NoError(t, err)
}
