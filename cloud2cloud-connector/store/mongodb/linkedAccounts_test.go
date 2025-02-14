package mongodb_test

import (
	"context"
	"testing"

	"github.com/plgd-dev/hub/cloud2cloud-connector/store"
	"github.com/plgd-dev/hub/cloud2cloud-connector/test"
	"github.com/plgd-dev/hub/pkg/security/oauth2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStoreInsertLinkedAccount(t *testing.T) {
	testToken := oauth2.Token{
		AccessToken:  "testAccessToken",
		RefreshToken: "testRefreshToken",
	}
	type args struct {
		sub store.LinkedAccount
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid",
			args: args{
				sub: store.LinkedAccount{
					ID:            "testID",
					LinkedCloudID: "testLinkedCloudID",
					UserID:        "userID",
					Data:          store.MakeLinkedAccountData(testToken, testToken),
				},
			},
		},
	}

	s, cleanUpStore := test.NewMongoStore(t)
	defer cleanUpStore()

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.InsertLinkedAccount(ctx, tt.args.sub)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStoreUpdateLinkedAccount(t *testing.T) {
	type args struct {
		sub store.LinkedAccount
	}
	testToken := oauth2.Token{
		AccessToken:  "testAccessToken",
		RefreshToken: "testRefreshToken",
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "invalid ID",
			args: args{
				sub: store.LinkedAccount{
					ID:            "testID1",
					LinkedCloudID: "testLinkedCloudID",
					UserID:        "userID",
					Data:          store.MakeLinkedAccountData(testToken, testToken),
				},
			},
			wantErr: true,
		},
		{
			name: "valid",
			args: args{
				sub: store.LinkedAccount{
					ID:            "testID",
					LinkedCloudID: "testLinkedCloudID",
					UserID:        "userID",
					Data:          store.MakeLinkedAccountData(testToken, testToken),
				},
			},
		},
	}

	s, cleanUpStore := test.NewMongoStore(t)
	defer cleanUpStore()

	ctx := context.Background()
	err := s.InsertLinkedAccount(ctx, store.LinkedAccount{
		ID:            "testID",
		LinkedCloudID: "testLinkedCloudID",
		UserID:        "userID",
		Data:          store.MakeLinkedAccountData(testToken, testToken),
	})
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.UpdateLinkedAccount(ctx, tt.args.sub)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStoreRemoveLinkedAccount(t *testing.T) {
	testToken := oauth2.Token{
		AccessToken:  "testAccessToken",
		RefreshToken: "testRefreshToken",
	}
	type args struct {
		linkedAccountId string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "invalid accountId",
			args: args{
				linkedAccountId: "testNotFound",
			},
			wantErr: true,
		},
		{
			name: "valid",
			args: args{
				linkedAccountId: "testID",
			},
		},
	}

	s, cleanUpStore := test.NewMongoStore(t)
	defer cleanUpStore()

	ctx := context.Background()
	err := s.InsertLinkedAccount(ctx, store.LinkedAccount{
		ID:            "testID",
		LinkedCloudID: "testLinkedCloudID",
		UserID:        "userID",
		Data:          store.MakeLinkedAccountData(testToken, testToken),
	})
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.RemoveLinkedAccount(ctx, tt.args.linkedAccountId)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

type testLinkedAccountHandler struct {
	accs []store.LinkedAccount
}

func (h *testLinkedAccountHandler) Handle(ctx context.Context, iter store.LinkedAccountIter) (err error) {
	for {
		var sub store.LinkedAccount
		if !iter.Next(ctx, &sub) {
			break
		}
		h.accs = append(h.accs, sub)
	}
	return iter.Err()
}

func TestStoreLoadLinkedAccounts(t *testing.T) {
	testToken := oauth2.Token{
		AccessToken:  "testAccessToken",
		RefreshToken: "testRefreshToken",
	}
	linkedAccounts := []store.LinkedAccount{
		{
			ID:            "testID",
			LinkedCloudID: "testLinkedCloudID",
			UserID:        "userID",
			Data:          store.MakeLinkedAccountData(testToken, testToken),
		},
	}

	type args struct {
		query store.Query
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    []store.LinkedAccount
	}{
		{
			name: "all",
			args: args{
				query: store.Query{},
			},
			want: linkedAccounts,
		},
		{
			name: "id",
			args: args{
				query: store.Query{ID: linkedAccounts[0].ID},
			},
			want: []store.LinkedAccount{linkedAccounts[0]},
		},
		{
			name: "not found",
			args: args{
				query: store.Query{ID: "not found"},
			},
		},
	}

	s, cleanUpStore := test.NewMongoStore(t)
	defer cleanUpStore()

	ctx := context.Background()
	for _, a := range linkedAccounts {
		err := s.InsertLinkedAccount(ctx, a)
		require.NoError(t, err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var h testLinkedAccountHandler
			err := s.LoadLinkedAccounts(ctx, tt.args.query, &h)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, h.accs)
			}
		})
	}
}
