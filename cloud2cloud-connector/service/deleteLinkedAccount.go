package service

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/plgd-dev/hub/cloud2cloud-connector/store"
	"github.com/plgd-dev/kit/v2/log"
)

func cancelLinkedAccountDevicesSubscription(ctx context.Context, cloud store.LinkedCloud, linkedAccount *LinkedAccountData, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := cancelDevicesSubscription(ctx, linkedAccount.linkedAccount, cloud, linkedAccount.subscription.ID)
		if err != nil {
			log.Error(err)
		}
	}()
}

func cancelLinkedAccountDeviceSubscription(ctx context.Context, cloud store.LinkedCloud, linkedAccount *LinkedAccountData, device *DeviceData, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := cancelDeviceSubscription(ctx, linkedAccount.linkedAccount, cloud, device.subscription.DeviceID, device.subscription.ID)
		if err != nil {
			log.Error(err)
		}
	}()
}

func cancelLinkedAccountResourceSubscription(ctx context.Context, cloud store.LinkedCloud, linkedAccount *LinkedAccountData, resource *ResourceData, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := cancelResourceSubscription(ctx, linkedAccount.linkedAccount, cloud, resource.subscription.DeviceID, resource.subscription.Href, resource.subscription.ID); err != nil {
			log.Error(err)
		}
	}()
}

func cancelLinkedAccountSubscription(ctx context.Context, cloud store.LinkedCloud, linkedAccount *LinkedAccountData) {
	var wg sync.WaitGroup
	if linkedAccount.isSubscribed {
		cancelLinkedAccountDevicesSubscription(ctx, cloud, linkedAccount, &wg)
	}
	linkedAccount.devices.Range(func(_, deviceI interface{}) bool {
		device := deviceI.(*DeviceData)
		if device.isSubscribed {
			cancelLinkedAccountDeviceSubscription(ctx, cloud, linkedAccount, device, &wg)
		}
		device.resources.Range(func(_, resourceI interface{}) bool {
			resource := resourceI.(*ResourceData)
			if resource.isSubscribed {
				cancelLinkedAccountResourceSubscription(ctx, cloud, linkedAccount, resource, &wg)
			}
			return true
		})
		return true
	})

	wg.Wait()
}

func (rh *RequestHandler) deleteLinkedAccount(w http.ResponseWriter, r *http.Request) (int, error) {
	vars := mux.Vars(r)
	cloudID := vars[cloudIDKey]
	accountID := vars[accountIDKey]

	linkedAccount, err := rh.store.PullOutLinkedAccount(r.Context(), cloudID, accountID)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("cannot load linked account: %w", err)
	}
	cloud, ok := rh.store.LoadCloud(cloudID)
	if !ok {
		return http.StatusOK, nil
	}
	cancelLinkedAccountSubscription(r.Context(), cloud, linkedAccount)

	return http.StatusOK, nil
}

func (rh *RequestHandler) DeleteLinkedAccount(w http.ResponseWriter, r *http.Request) {
	statusCode, err := rh.deleteLinkedAccount(w, r)
	if err != nil {
		logAndWriteErrorResponse(fmt.Errorf("cannot delete linked accounts: %w", err), statusCode, w)
	}
}
