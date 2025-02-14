package events_test

import (
	"testing"

	"github.com/plgd-dev/hub/resource-aggregate/commands"
	"github.com/plgd-dev/hub/resource-aggregate/events"
	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

var testEventResourceCreatePending events.ResourceCreatePending = events.ResourceCreatePending{
	ResourceId: &commands.ResourceId{
		DeviceId: "dev1",
		Href:     "/dev1",
	},
	Content: &commands.Content{
		Data:              []byte{'t', 'e', 'x', 't'},
		ContentType:       "text",
		CoapContentFormat: int32(message.TextPlain),
	},
	AuditContext: &commands.AuditContext{
		UserId:        "501",
		CorrelationId: "1",
	},
	EventMetadata: &events.EventMetadata{
		Version:      42,
		Timestamp:    12345,
		ConnectionId: "con1",
		Sequence:     1,
	},
}

func TestResourceCreatePending_CopyData(t *testing.T) {
	type args struct {
		event *events.ResourceCreatePending
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Identity",
			args: args{
				event: &testEventResourceCreatePending,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var e events.ResourceCreatePending
			e.CopyData(tt.args.event)
			require.True(t, proto.Equal(tt.args.event, &e))
		})
	}
}

func TestResourceCreatePending_CheckInitialized(t *testing.T) {
	type args struct {
		event *events.ResourceCreatePending
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Uninitialized",
			args: args{
				event: &events.ResourceCreatePending{},
			},
			want: false,
		},
		{
			name: "Initialized",
			args: args{
				event: &testEventResourceCreatePending,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.args.event.CheckInitialized())
		})
	}
}
