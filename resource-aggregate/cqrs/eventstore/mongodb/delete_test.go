package mongodb_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/plgd-dev/hub/pkg/log"
	"github.com/plgd-dev/hub/resource-aggregate/cqrs/eventstore"
	"github.com/plgd-dev/hub/resource-aggregate/cqrs/eventstore/mongodb"
	"github.com/plgd-dev/hub/resource-aggregate/cqrs/eventstore/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func addEventsForDeleteToDB(t *testing.T, ctx context.Context, store *mongodb.EventStore) int {
	const eventCount = 2000
	const deviceCount = 10
	const resourceCount = 100
	var resourceVersion [resourceCount]uint64
	var resourceTimestamp [resourceCount]int64
	var resourceEvents [resourceCount][]eventstore.Event
	for i := 0; i < eventCount; i++ {
		deviceIndex := i % deviceCount
		resourceIndex := i % resourceCount
		if i < resourceCount {
			resourceTimestamp[i] = int64((eventCount / resourceCount) * i)
		}

		resourceEvents[resourceIndex] = append(resourceEvents[resourceIndex], test.MockEvent{
			VersionI:     resourceVersion[resourceIndex],
			EventTypeI:   "testType",
			IsSnapshotI:  false,
			AggregateIDI: "resource" + strconv.Itoa(resourceIndex),
			GroupIDI:     "device" + strconv.Itoa(deviceIndex),
			TimestampI:   1 + resourceTimestamp[resourceIndex],
		})

		resourceVersion[resourceIndex]++
		resourceTimestamp[resourceIndex]++
	}

	for _, v := range resourceEvents {
		saveStatus, err := store.Save(ctx, v...)
		require.NoError(t, err)
		require.Equal(t, eventstore.Ok, saveStatus)
	}

	return eventCount
}

func TestEventStore_Delete(t *testing.T) {
	logger, err := log.NewLogger(log.Config{})
	require.NoError(t, err)

	ctx := context.Background()
	store, err := NewTestEventStore(ctx, logger)
	assert.NoError(t, err)
	assert.NotNil(t, store)
	defer func() {
		t.Log("clearing db")
		err := store.Clear(ctx)
		require.NoError(t, err)
		_ = store.Close(ctx)
	}()

	type args struct {
		query []eventstore.DeleteQuery
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Invalid query",
			args: args{
				query: []eventstore.DeleteQuery{},
			},
			wantErr: true,
		},
		{
			name: "Invalid groupID",
			args: args{
				query: []eventstore.DeleteQuery{{
					GroupID: "badId",
				}},
			},
			wantErr: false,
		},
		{
			name: "Invalid and valid groupID",
			args: args{
				query: []eventstore.DeleteQuery{{
					GroupID: "badId",
				}, {
					GroupID: "device1",
				}},
			},
			wantErr: false,
		},
		{
			name: "Delete single device",
			args: args{
				query: []eventstore.DeleteQuery{{
					GroupID: "device5",
				}},
			},
			wantErr: false,
		},
		{
			name: "Delete multiple devices",
			args: args{
				query: []eventstore.DeleteQuery{{
					GroupID: "device2",
				}, {
					GroupID: "device3",
				}, {
					GroupID: "device5",
				}, {
					GroupID: "device7",
				}},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addEventsForDeleteToDB(t, ctx, store)
			defer func() {
				err = store.ClearCollections(ctx)
				require.NoError(t, err)
			}()

			err := store.Delete(ctx, tt.args.query)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// get all events after deletion
			handler := test.NewMockEventHandler()
			err = store.GetEvents(ctx, []eventstore.GetEventsQuery{{}}, 0, handler)
			require.NoError(t, err)
			// no documents with deleted group id should remain
			for _, q := range tt.args.query {
				require.False(t, handler.ContainsGroupID(q.GroupID))
			}
		})
	}
}
