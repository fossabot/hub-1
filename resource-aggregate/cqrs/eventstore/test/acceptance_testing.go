package test

import (
	"context"
	"testing"
	"time"

	"github.com/plgd-dev/hub/resource-aggregate/cqrs/eventstore"
	"github.com/stretchr/testify/require"
)

// AcceptanceTest is the acceptance test that all implementations of EventStore
// should pass. It should manually be called from a test case in each
// implementation:
//
//   func TestEventStore(t *testing.T) {
//       ctx := context.Background() // Or other when testing namespaces.
//       store := NewEventStore()
//       eventstore.AcceptanceTest(t, ctx, store)
//   }
//

func getEvents(fromVersion uint64, num uint64, firstEventSnapshot bool, groupID string, aggregateID string, timestamp int64) []eventstore.Event {
	e := []eventstore.Event{
		MockEvent{
			VersionI:     fromVersion,
			EventTypeI:   "test0",
			AggregateIDI: aggregateID,
			GroupIDI:     groupID,
			IsSnapshotI:  firstEventSnapshot,
			TimestampI:   timestamp,
		},
	}
	for i := uint64(1); i < num; i++ {
		e = append(e, MockEvent{
			VersionI:     fromVersion + i,
			EventTypeI:   "test0",
			AggregateIDI: aggregateID,
			GroupIDI:     groupID,
			TimestampI:   timestamp + int64(i),
		})
	}
	return e
}

type eventsFilter func(eventstore.Event) bool

func filterEvents(events []eventstore.Event, filter eventsFilter) []eventstore.Event {
	newEvents := make([]eventstore.Event, 0, len(events))
	for _, v := range events {
		if filter(v) {
			newEvents = append(newEvents, v)
		}
	}
	return newEvents
}

const aggregateID1 = "aggregateID1"
const aggregateID2 = "aggregateID2"
const aggregateID3 = "aggregateID3"
const aggregateID4 = "aggregateID4"

const groupID1 = "deviceId1"
const groupID2 = "deviceId2"
const groupID3 = "deviceId3"

func GetEventsTest(t *testing.T, ctx context.Context, store eventstore.EventStore) {
	t.Log("testing GetEvents")

	const timestamp1 = int64(0)
	const timestamp2 = int64(20)
	const timestamp3 = int64(40)
	const timestamp4 = int64(60)

	t.Log("insert events")
	groupID1Events := getEvents(0, 5, false, groupID1, aggregateID1, timestamp1)
	saveStatus, err := store.Save(ctx, groupID1Events...)
	require.NoError(t, err)
	require.Equal(t, eventstore.Ok, saveStatus)

	groupID2AggID2Events := getEvents(0, 5, false, groupID2, aggregateID2, timestamp2)
	saveStatus, err = store.Save(ctx, groupID2AggID2Events...)
	require.NoError(t, err)
	require.Equal(t, eventstore.Ok, saveStatus)
	groupID2AggID3Events := getEvents(0, 5, false, groupID2, aggregateID3, timestamp3)
	saveStatus, err = store.Save(ctx, groupID2AggID3Events...)
	require.NoError(t, err)
	require.Equal(t, eventstore.Ok, saveStatus)

	groupID3Events := getEvents(0, 5, false, groupID3, aggregateID4, timestamp4)
	saveStatus, err = store.Save(ctx, groupID3Events...)
	require.NoError(t, err)
	require.Equal(t, eventstore.Ok, saveStatus)

	groupID2Events := groupID2AggID2Events
	groupID2Events = append(groupID2Events, groupID2AggID3Events...)
	allEvents := append(groupID1Events, groupID2Events...)
	allEvents = append(allEvents, groupID3Events...)

	t.Log("get all events")
	saveEh := NewMockEventHandler()
	err = store.GetEvents(ctx, []eventstore.GetEventsQuery{{}}, 0, saveEh)
	require.NoError(t, err)
	require.True(t, saveEh.Equals(allEvents))

	t.Logf("get groupid %v and %v events", groupID1, groupID2)
	saveEh = NewMockEventHandler()
	err = store.GetEvents(ctx, []eventstore.GetEventsQuery{{GroupID: groupID1}, {GroupID: groupID2}}, 0, saveEh)
	require.NoError(t, err)
	events := groupID1Events
	events = append(events, groupID2Events...)
	require.True(t, saveEh.Equals(events))

	t.Logf("get aggregateid %v events", aggregateID2)
	saveEh = NewMockEventHandler()
	err = store.GetEvents(ctx, []eventstore.GetEventsQuery{{AggregateID: aggregateID2}}, 0, saveEh)
	require.NoError(t, err)
	require.True(t, saveEh.Equals(groupID2AggID2Events))

	t.Logf("get groupid %v and aggregateid %v events", groupID1, aggregateID4)
	saveEh = NewMockEventHandler()
	err = store.GetEvents(ctx, []eventstore.GetEventsQuery{{GroupID: groupID1}, {GroupID: groupID3, AggregateID: aggregateID4}}, 0, saveEh)
	require.NoError(t, err)
	events = groupID1Events
	events = append(events, groupID3Events...)
	require.True(t, saveEh.Equals(events))

	timestamp := timestamp4 - 1
	t.Logf("get events with timestamp > %v", timestamp)
	saveEh = NewMockEventHandler()
	err = store.GetEvents(ctx, []eventstore.GetEventsQuery{{}}, timestamp, saveEh)
	require.NoError(t, err)
	require.True(t, saveEh.Equals(groupID3Events))

	timestamp = timestamp3 + 2
	t.Logf("get groupid (%v, %v) events with timestamp > %v", groupID2, groupID3, timestamp)
	saveEh = NewMockEventHandler()
	err = store.GetEvents(ctx, []eventstore.GetEventsQuery{{GroupID: groupID2}, {GroupID: groupID3}}, timestamp, saveEh)
	require.NoError(t, err)
	events = filterEvents(allEvents, func(e eventstore.Event) bool {
		return e.Timestamp().UnixNano() > timestamp
	})
	require.True(t, saveEh.Equals(events))

	timestamp = timestamp2 - 1
	t.Logf("get aggregateid (%v, %v) events with timestamp > %v", aggregateID3, aggregateID4, timestamp)
	saveEh = NewMockEventHandler()
	err = store.GetEvents(ctx, []eventstore.GetEventsQuery{{AggregateID: aggregateID3}, {AggregateID: aggregateID4}}, timestamp, saveEh)
	require.NoError(t, err)
	events = groupID2AggID3Events
	events = append(events, groupID3Events...)
	require.True(t, saveEh.Equals(events))
}

func emptySaveFailTest(t *testing.T, ctx context.Context, store eventstore.EventStore) {
	t.Log("try save no events")
	saveStatus, err := store.Save(ctx, nil)
	require.Error(t, err)
	require.Equal(t, eventstore.Fail, saveStatus)
}

func invalidTimpestampFailTest(t *testing.T, ctx context.Context, store eventstore.EventStore) {
	t.Log("try save descreasing timestamp")
	timestamp := time.Date(2021, time.April, 1, 13, 37, 00, 0, time.UTC).UnixNano()
	events := getEvents(0, 2, false, groupID1, aggregateID1, timestamp)
	mockEvent := events[1].(MockEvent)
	mockEvent.TimestampI = timestamp - 1
	events[1] = mockEvent
	saveStatus, err := store.Save(ctx, events...)
	require.Error(t, err)
	require.Equal(t, eventstore.Fail, saveStatus)
}

func AcceptanceTest(t *testing.T, ctx context.Context, store eventstore.EventStore) {
	type Path struct {
		GroupID     string
		AggregateID string
	}

	aggregateID1Path := Path{
		AggregateID: aggregateID1,
		GroupID:     groupID1,
	}
	aggregateID2Path := Path{
		AggregateID: aggregateID2,
		GroupID:     groupID1,
	}
	aggregateID3Path := Path{
		AggregateID: aggregateID3,
		GroupID:     groupID2,
	}
	aggregateID4Path := Path{
		AggregateID: aggregateID4,
		GroupID:     groupID3,
	}

	timestamp := time.Date(2021, time.April, 1, 13, 37, 00, 0, time.UTC).UnixNano()

	emptySaveFailTest(t, ctx, store)
	invalidTimpestampFailTest(t, ctx, store)

	t.Log("save event, VersionI 0")
	saveStatus, err := store.Save(ctx, getEvents(0, 6, false, aggregateID1Path.GroupID, aggregateID1Path.AggregateID, timestamp)[0])
	require.NoError(t, err)
	require.Equal(t, eventstore.Ok, saveStatus)

	t.Log("save event, VersionI 1")
	saveStatus, err = store.Save(ctx, getEvents(0, 6, false, aggregateID1Path.GroupID, aggregateID1Path.AggregateID, timestamp)[1])
	require.NoError(t, err)
	require.Equal(t, eventstore.Ok, saveStatus)

	t.Log("try to save same event VersionI 1 twice")
	saveStatus, err = store.Save(ctx, getEvents(0, 6, false, aggregateID1Path.GroupID, aggregateID1Path.AggregateID, timestamp)[1])
	require.NoError(t, err)
	require.Equal(t, eventstore.ConcurrencyException, saveStatus)

	t.Log("save event, VersionI 2")
	saveStatus, err = store.Save(ctx, getEvents(0, 6, false, aggregateID1Path.GroupID, aggregateID1Path.AggregateID, timestamp)[2])
	require.NoError(t, err)
	require.Equal(t, eventstore.Ok, saveStatus)

	t.Log("save multiple events, VersionI 3, 4 and 5")
	saveStatus, err = store.Save(ctx, getEvents(0, 6, false, aggregateID1Path.GroupID, aggregateID1Path.AggregateID, timestamp)[3:6]...)
	require.NoError(t, err)
	require.Equal(t, eventstore.Ok, saveStatus)

	t.Log("save event for another aggregate")
	saveStatus, err = store.Save(ctx, getEvents(0, 6, false, aggregateID2Path.GroupID, aggregateID2Path.AggregateID, timestamp)[0:4]...)
	require.NoError(t, err)
	require.Equal(t, eventstore.Ok, saveStatus)

	t.Log("save events and then save snapshot with events")
	saveStatus, err = store.Save(ctx, getEvents(0, 3, false, aggregateID4Path.GroupID, aggregateID4Path.AggregateID, timestamp)...)
	require.NoError(t, err)
	require.Equal(t, eventstore.Ok, saveStatus)

	saveStatus, err = store.Save(ctx, getEvents(3, 4, true, aggregateID4Path.GroupID, aggregateID4Path.AggregateID, timestamp)...)
	require.NoError(t, err)
	require.Equal(t, eventstore.Ok, saveStatus)

	t.Log("load events from snapshot")
	saveEh := NewMockEventHandler()
	err = store.LoadFromSnapshot(ctx, []eventstore.SnapshotQuery{
		{
			GroupID:     aggregateID4Path.GroupID,
			AggregateID: aggregateID4Path.AggregateID,
		},
	}, saveEh)
	require.NoError(t, err)
	require.Equal(t, getEvents(3, 4, true, aggregateID4Path.GroupID, aggregateID4Path.AggregateID, timestamp), saveEh.events[aggregateID4Path.GroupID][aggregateID4Path.AggregateID])

	t.Log("test if need snapshot occurs from save")
	bigEv := getEvents(7, 1, false, aggregateID4Path.GroupID, aggregateID4Path.AggregateID, timestamp)[0].(MockEvent)
	bigEv.DataI = make([]byte, 7*1024*1024)

	saveStatus, err = store.Save(ctx, bigEv)
	require.NoError(t, err)
	require.Equal(t, eventstore.Ok, saveStatus)

	bigEv.VersionI++
	saveStatus, err = store.Save(ctx, bigEv)
	require.NoError(t, err)
	require.Equal(t, eventstore.Ok, saveStatus)

	bigEv.VersionI++
	saveStatus, err = store.Save(ctx, bigEv)
	require.NoError(t, err)
	require.Equal(t, eventstore.SnapshotRequired, saveStatus)

	bigEv.IsSnapshotI = true
	saveStatus, err = store.Save(ctx, bigEv)
	require.NoError(t, err)
	require.Equal(t, eventstore.Ok, saveStatus)
	exp := []eventstore.Event{bigEv}

	bigEv.VersionI++
	bigEv.IsSnapshotI = false
	saveStatus, err = store.Save(ctx, bigEv)
	require.NoError(t, err)
	require.Equal(t, eventstore.Ok, saveStatus)
	exp = append(exp, bigEv)

	saveEh1 := NewMockEventHandler()
	err = store.LoadFromSnapshot(ctx, []eventstore.SnapshotQuery{
		{
			GroupID:     aggregateID4Path.GroupID,
			AggregateID: aggregateID4Path.AggregateID,
		},
	}, saveEh1)
	require.NoError(t, err)
	require.Equal(t, exp, saveEh1.events[aggregateID4Path.GroupID][aggregateID4Path.AggregateID])

	t.Log("load events for non-existing aggregate")
	eh1 := NewMockEventHandler()
	err = store.LoadFromSnapshot(ctx, []eventstore.SnapshotQuery{{GroupID: "notExist"}}, eh1)
	require.NoError(t, err)
	require.Equal(t, 0, len(eh1.events))

	t.Log("load events")
	eh2 := NewMockEventHandler()
	err = store.LoadFromSnapshot(ctx, []eventstore.SnapshotQuery{
		{
			GroupID:     aggregateID1Path.GroupID,
			AggregateID: aggregateID1Path.AggregateID,
		},
	}, eh2)
	require.NoError(t, err)
	require.Equal(t, getEvents(0, 6, false, aggregateID1Path.GroupID, aggregateID1Path.AggregateID, timestamp)[:6], eh2.events[aggregateID1Path.GroupID][aggregateID1Path.AggregateID])

	t.Log("load events from version")
	eh3 := NewMockEventHandler()
	err = store.LoadFromVersion(ctx, []eventstore.VersionQuery{
		{
			GroupID:     aggregateID1Path.GroupID,
			AggregateID: aggregateID1Path.AggregateID,
			Version:     getEvents(0, 6, false, aggregateID1Path.GroupID, aggregateID1Path.AggregateID, timestamp)[2].Version(),
		},
	}, eh3)
	require.NoError(t, err)
	require.Equal(t, getEvents(0, 6, false, aggregateID1Path.GroupID, aggregateID1Path.AggregateID, timestamp)[2:6], eh3.events[aggregateID1Path.GroupID][aggregateID1Path.AggregateID])

	t.Log("load multiple aggregates by all queries")
	eh4 := NewMockEventHandler()
	err = store.LoadFromVersion(ctx, []eventstore.VersionQuery{
		{
			GroupID:     aggregateID1Path.GroupID,
			AggregateID: aggregateID1Path.AggregateID,
		},
		{
			GroupID:     aggregateID2Path.GroupID,
			AggregateID: aggregateID2Path.AggregateID,
		},
	}, eh4)
	require.NoError(t, err)
	require.Equal(t, getEvents(0, 6, false, aggregateID1Path.GroupID, aggregateID1Path.AggregateID, timestamp)[0:6], eh4.events[aggregateID1Path.GroupID][aggregateID1Path.AggregateID])
	require.Equal(t, getEvents(0, 6, false, aggregateID2Path.GroupID, aggregateID2Path.AggregateID, timestamp)[0:4], eh4.events[aggregateID2Path.GroupID][aggregateID2Path.AggregateID])

	t.Log("load multiple aggregates by groupId")
	eh5 := NewMockEventHandler()
	err = store.LoadFromSnapshot(ctx, []eventstore.SnapshotQuery{
		{
			GroupID: aggregateID1Path.GroupID,
		},
	}, eh5)
	require.NoError(t, err)
	require.Equal(t,
		getEvents(0, 6, false, aggregateID1Path.GroupID, aggregateID1Path.AggregateID, timestamp)[0:6],
		eh5.events[aggregateID1Path.GroupID][aggregateID1Path.AggregateID])
	require.Equal(t,
		getEvents(0, 6, false, aggregateID2Path.GroupID, aggregateID2Path.AggregateID, timestamp)[0:4],
		eh5.events[aggregateID2Path.GroupID][aggregateID2Path.AggregateID])

	t.Log("load multiple aggregates by all")
	eh6 := NewMockEventHandler()
	saveStatus, err = store.Save(ctx, getEvents(0, 6, false, aggregateID3Path.GroupID, aggregateID3Path.AggregateID, timestamp)[0])
	require.NoError(t, err)
	require.Equal(t, eventstore.Ok, saveStatus)
	err = store.LoadFromSnapshot(ctx, []eventstore.SnapshotQuery{{GroupID: aggregateID1Path.GroupID}, {GroupID: aggregateID2Path.GroupID}, {GroupID: aggregateID3Path.GroupID}}, eh6)
	require.NoError(t, err)
	require.Equal(t, getEvents(0, 6, false, aggregateID1Path.GroupID, aggregateID1Path.AggregateID, timestamp)[0:6],
		eh6.events[aggregateID1Path.GroupID][aggregateID1Path.AggregateID])
	require.Equal(t,
		getEvents(0, 6, false, aggregateID2Path.GroupID, aggregateID2Path.AggregateID, timestamp)[0:4],
		eh6.events[aggregateID2Path.GroupID][aggregateID2Path.AggregateID])
	require.Equal(t, []eventstore.Event{
		getEvents(0, 6, false, aggregateID3Path.GroupID, aggregateID3Path.AggregateID, timestamp)[0],
	}, eh6.events[aggregateID3Path.GroupID][aggregateID3Path.AggregateID])

	t.Log("load events up to version")
	eh7 := NewMockEventHandler()
	err = store.LoadUpToVersion(ctx, []eventstore.VersionQuery{
		{
			GroupID:     aggregateID1Path.GroupID,
			AggregateID: aggregateID1Path.AggregateID,
			Version:     getEvents(0, 6, false, aggregateID1Path.GroupID, aggregateID1Path.AggregateID, timestamp)[5].Version(),
		},
	}, eh7)
	require.NoError(t, err)
	require.Equal(t, getEvents(0, 6, false, aggregateID1Path.GroupID, aggregateID1Path.AggregateID, timestamp)[0:5], eh7.events[aggregateID1Path.GroupID][aggregateID1Path.AggregateID])

	t.Log("load events up to version")
	eh8 := NewMockEventHandler()
	err = store.LoadUpToVersion(ctx, []eventstore.VersionQuery{
		{
			GroupID:     aggregateID1Path.GroupID,
			AggregateID: aggregateID1Path.AggregateID,
			Version:     getEvents(0, 6, false, aggregateID1Path.GroupID, aggregateID1Path.AggregateID, timestamp)[0].Version(),
		},
	}, eh8)
	require.NoError(t, err)
	require.Equal(t, 0, len(eh8.events[aggregateID1Path.GroupID][aggregateID1Path.AggregateID]))

	t.Log("load events up to version without version specified")
	eh9 := NewMockEventHandler()
	err = store.LoadUpToVersion(ctx, []eventstore.VersionQuery{
		{
			GroupID:     aggregateID1Path.GroupID,
			AggregateID: aggregateID1Path.AggregateID,
		},
	}, eh9)
	require.NoError(t, err)
	require.Equal(t, 0, len(eh9.events[aggregateID1Path.GroupID][aggregateID1Path.AggregateID]))

	t.Log("test projection all")
	model := NewMockEventHandler()
	p := eventstore.NewProjection(store, func(context.Context, string, string) (eventstore.Model, error) { return model, nil }, nil)

	err = p.Project(ctx, []eventstore.SnapshotQuery{{GroupID: aggregateID1Path.GroupID}, {GroupID: aggregateID2Path.GroupID}, {GroupID: aggregateID3Path.GroupID}})
	require.NoError(t, err)
	require.Equal(t, getEvents(0, 6, false, aggregateID1Path.GroupID, aggregateID1Path.AggregateID, timestamp)[0:6], model.events[aggregateID1Path.GroupID][aggregateID1Path.AggregateID])
	require.Equal(t, getEvents(0, 6, false, aggregateID2Path.GroupID, aggregateID2Path.AggregateID, timestamp)[0:4], model.events[aggregateID2Path.GroupID][aggregateID2Path.AggregateID])
	require.Equal(t, []eventstore.Event{
		getEvents(0, 6, false, aggregateID3Path.GroupID, aggregateID3Path.AggregateID, timestamp)[0],
	}, model.events[aggregateID3Path.GroupID][aggregateID3Path.AggregateID])

	t.Log("test projection group")
	model1 := NewMockEventHandler()
	p = eventstore.NewProjection(store, func(context.Context, string, string) (eventstore.Model, error) { return model1, nil }, nil)

	err = p.Project(ctx, []eventstore.SnapshotQuery{{GroupID: aggregateID1Path.GroupID}})
	require.NoError(t, err)
	require.Equal(t, getEvents(0, 6, false, aggregateID1Path.GroupID, aggregateID1Path.AggregateID, timestamp)[0:6], model1.events[aggregateID1Path.GroupID][aggregateID1Path.AggregateID])
	require.Equal(t, getEvents(0, 6, false, aggregateID2Path.GroupID, aggregateID2Path.AggregateID, timestamp)[0:4], model1.events[aggregateID2Path.GroupID][aggregateID2Path.AggregateID])

	t.Log("test projection aggregate")
	model2 := NewMockEventHandler()
	p = eventstore.NewProjection(store, func(context.Context, string, string) (eventstore.Model, error) { return model2, nil }, nil)

	err = p.Project(ctx, []eventstore.SnapshotQuery{
		{
			GroupID:     aggregateID2Path.GroupID,
			AggregateID: aggregateID2Path.AggregateID,
		},
	})
	require.NoError(t, err)
	require.Equal(t, getEvents(0, 6, false, aggregateID2Path.GroupID, aggregateID2Path.AggregateID, timestamp)[0:4], model2.events[aggregateID2Path.GroupID][aggregateID2Path.AggregateID])

	t.Log("remove events up to version")
	versionToRemove := 3
	err = store.RemoveUpToVersion(ctx, []eventstore.VersionQuery{
		{
			GroupID:     aggregateID1Path.GroupID,
			AggregateID: aggregateID1Path.AggregateID,
			Version:     getEvents(0, 6, false, aggregateID1Path.GroupID, aggregateID1Path.AggregateID, timestamp)[versionToRemove].Version(),
		},
	})
	require.NoError(t, err)

	eh10 := NewMockEventHandler()
	err = store.LoadFromVersion(ctx, []eventstore.VersionQuery{
		{
			GroupID:     aggregateID1Path.GroupID,
			AggregateID: aggregateID1Path.AggregateID,
		},
	}, eh10)
	require.NoError(t, err)
	require.Equal(t, getEvents(0, 6, false, aggregateID1Path.GroupID, aggregateID1Path.AggregateID, timestamp)[versionToRemove:6], eh10.events[aggregateID1Path.GroupID][aggregateID1Path.AggregateID])
}
