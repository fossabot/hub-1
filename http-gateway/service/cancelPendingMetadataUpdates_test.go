package service_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/plgd-dev/hub/grpc-gateway/pb"
	httpgwTest "github.com/plgd-dev/hub/http-gateway/test"
	"github.com/plgd-dev/hub/http-gateway/uri"
	"github.com/plgd-dev/hub/test/config"
	oauthTest "github.com/plgd-dev/hub/test/oauth-server/test"
	pbTest "github.com/plgd-dev/hub/test/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestHandlerCancelPendingMetadataUpdates(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), config.TEST_TIMEOUT)
	defer cancel()

	_, _, devicePendings, shutdown := pbTest.InitPendingEvents(ctx, t)
	defer shutdown()

	require.Equal(t, len(devicePendings), 2)

	type args struct {
		deviceID            string
		correlationIdFilter []string
		accept              string
	}
	tests := []struct {
		name         string
		args         args
		wantErr      bool
		want         *pb.CancelPendingCommandsResponse
		wantHTTPCode int
	}{
		{
			name: "cancel one pending",
			args: args{
				deviceID:            devicePendings[0].DeviceID,
				correlationIdFilter: []string{devicePendings[0].CorrelationID},
				accept:              uri.ApplicationProtoJsonContentType,
			},
			want: &pb.CancelPendingCommandsResponse{
				CorrelationIds: []string{devicePendings[0].CorrelationID},
			},
			wantHTTPCode: http.StatusOK,
		},
		{
			name: "duplicate cancel event",
			args: args{
				deviceID:            devicePendings[0].DeviceID,
				correlationIdFilter: []string{devicePendings[0].CorrelationID},
				accept:              uri.ApplicationProtoJsonContentType,
			},
			wantErr:      true,
			wantHTTPCode: http.StatusNotFound,
		},
		{
			name: "cancel all events",
			args: args{
				deviceID: devicePendings[0].DeviceID,
				accept:   uri.ApplicationProtoJsonContentType,
			},
			want: &pb.CancelPendingCommandsResponse{
				CorrelationIds: []string{devicePendings[1].CorrelationID},
			},
			wantHTTPCode: http.StatusOK,
		},
	}

	token := oauthTest.GetDefaultServiceToken(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rb := httpgwTest.NewRequest(http.MethodDelete, uri.AliasDevicePendingMetadataUpdates, nil).AuthToken(token).Accept(tt.args.accept)
			rb.DeviceId(tt.args.deviceID).AddCorrelantionIdFilter(tt.args.correlationIdFilter)
			v, code, err := doPendingCommand(t, rb.Build())
			assert.Equal(t, tt.wantHTTPCode, code)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			pbTest.CmpCancelPendingCmdResponses(t, tt.want, v)
		})
	}
}
