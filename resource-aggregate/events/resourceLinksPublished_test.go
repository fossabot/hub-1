package events_test

import (
	"testing"

	commands "github.com/plgd-dev/hub/resource-aggregate/commands"
	"github.com/plgd-dev/hub/resource-aggregate/events"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

var testEventResourceLinksPublished events.ResourceLinksPublished = events.ResourceLinksPublished{
	Resources: []*commands.Resource{
		{
			Href:                  "/res1",
			DeviceId:              "dev1",
			ResourceTypes:         []string{"type1", "type2"},
			Interfaces:            []string{"if1", "if2"},
			Anchor:                "anchor1",
			Title:                 "Resource1",
			SupportedContentTypes: []string{"stype1", "stype2"},
			ValidUntil:            123,
			Policy: &commands.Policy{
				BitFlags: 42,
			},
			EndpointInformations: []*commands.EndpointInformation{
				{
					Endpoint: "ep1",
					Priority: 1,
				},
			},
		},
	},
	DeviceId: "dev1",
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

func TestResourceLinksPublished_CopyData(t *testing.T) {
	type args struct {
		event *events.ResourceLinksPublished
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Identity",
			args: args{
				event: &testEventResourceLinksPublished,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var e events.ResourceLinksPublished
			e.CopyData(tt.args.event)
			require.True(t, proto.Equal(tt.args.event, &e))
		})
	}
}

func TestResourceLinksPublished_CheckInitialized(t *testing.T) {
	type args struct {
		event *events.ResourceDeletePending
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Uninitialized",
			args: args{
				event: &events.ResourceDeletePending{},
			},
			want: false,
		},
		{
			name: "Initialized",
			args: args{
				event: &testEventResourceDeletePending,
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
