package pb

import (
	"context"

	"github.com/google/uuid"
	"github.com/plgd-dev/hub/resource-aggregate/commands"
	"google.golang.org/grpc/peer"
)

func (req *GetResourceFromDeviceRequest) ToRACommand(ctx context.Context) (*commands.RetrieveResourceRequest, error) {
	correlationUUID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	connectionID := ""
	peer, ok := peer.FromContext(ctx)
	if ok {
		connectionID = peer.Addr.String()
	}
	href := req.GetResourceId().GetHref()
	if len(href) > 0 && href[0] != '/' {
		href = "/" + href
	}

	return &commands.RetrieveResourceRequest{
		ResourceId:        commands.NewResourceID(req.GetResourceId().GetDeviceId(), href),
		CorrelationId:     correlationUUID.String(),
		TimeToLive:        req.GetTimeToLive(),
		ResourceInterface: req.GetResourceInterface(),
		CommandMetadata: &commands.CommandMetadata{
			ConnectionId: connectionID,
		},
	}, nil
}
