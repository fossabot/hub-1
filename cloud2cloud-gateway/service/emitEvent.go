package service

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	netHttp "net/http"
	"strconv"
	"time"

	"github.com/plgd-dev/hub/cloud2cloud-connector/events"
	"github.com/plgd-dev/hub/cloud2cloud-gateway/store"
	"github.com/plgd-dev/hub/pkg/log"
	cmClient "github.com/plgd-dev/hub/pkg/security/certManager/client"
)

type incrementSubscriptionSequenceNumberFunc func(ctx context.Context) (uint64, error)
type emitEventFunc func(ctx context.Context, eventType events.EventType, s store.Subscription, incrementSubscriptionSequenceNumber incrementSubscriptionSequenceNumberFunc, rep interface{}) (remove bool, err error)

func makeEmitEventRequestBody(accept []string, rep interface{}) ([]byte, string, error) {
	encoder, contentType, err := getEncoder(accept)
	if err != nil {
		return nil, "", fmt.Errorf("cannot get encoder: %w", err)
	}
	body, err := encoder(rep)
	if err != nil {
		return nil, "", fmt.Errorf("cannot encode data to body: %w", err)
	}
	return body, contentType, nil
}

func makeEmitEventRequest(ctx context.Context, eventType events.EventType, s store.Subscription, seqNum uint64, rep interface{}) (*netHttp.Request, error) {
	r, w := io.Pipe()
	req, err := netHttp.NewRequestWithContext(ctx, netHttp.MethodPost, s.URL, r)
	if err != nil {
		return nil, fmt.Errorf("cannot create post request: %w", err)
	}
	timestamp := time.Now()
	req.Header.Set(events.EventTypeKey, string(eventType))
	req.Header.Set(events.SubscriptionIDKey, s.ID)
	req.Header.Set(events.SequenceNumberKey, strconv.FormatUint(seqNum, 10))
	req.Header.Set(events.CorrelationIDKey, s.CorrelationID)
	req.Header.Set(events.EventTimestampKey, strconv.FormatInt(timestamp.Unix(), 10))

	var body []byte
	if rep != nil {
		var contentType string
		body, contentType, err = makeEmitEventRequestBody(s.Accept, rep)
		if err != nil {
			return nil, fmt.Errorf("cannot create post request body: %w", err)
		}
		req.Header.Set(events.ContentTypeKey, contentType)
	}

	if len(body) > 0 {
		go func() {
			defer func() {
				if err := w.Close(); err != nil {
					log.Errorf("failed to close write pipe: %w", err)
				}
			}()
			_, err := w.Write(body)
			if err != nil {
				log.Errorf("cannot write data to client: %w", err)
			}
		}()
	} else {
		if err := w.Close(); err != nil {
			log.Errorf("failed to close write pipe: %w", err)
		}
	}

	req.Header.Set(events.EventSignatureKey, events.CalculateEventSignature(
		s.SigningSecret,
		req.Header.Get(events.ContentTypeKey),
		eventType,
		req.Header.Get(events.SubscriptionIDKey),
		seqNum,
		timestamp,
		body,
	))
	req.Header.Set("Connection", "close")
	req.Close = true

	return req, nil
}

func createEmitEventFunc(cfg cmClient.Config, timeout time.Duration, logger log.Logger) (emitEventFunc, func(), error) {
	certManager, err := cmClient.New(cfg, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create cert manager: %w", err)
	}
	closeFunc := func() {
		certManager.Close()
	}
	trans := netHttp.DefaultTransport.(*netHttp.Transport).Clone()
	trans.TLSClientConfig = certManager.GetTLSConfig()
	client := netHttp.Client{
		Transport: trans,
	}
	emitFunc := func(ctx context.Context, eventType events.EventType, s store.Subscription, incrementSubscriptionSequenceNumber incrementSubscriptionSequenceNumberFunc, rep interface{}) (remove bool, err error) {
		log.Debugf("emitEvent: %v: %+v", eventType, s)
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		seqNum, err := incrementSubscriptionSequenceNumber(ctx)
		if err != nil {
			return false, fmt.Errorf("cannot increment sequence number: %w", err)
		}

		req, err := makeEmitEventRequest(ctx, eventType, s, seqNum, rep)
		if err != nil {
			return false, fmt.Errorf("cannot create request: %w", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			return false, fmt.Errorf("cannot post: %w", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				log.Errorf("failed to close response body stream: %w", err)
			}
		}()
		if resp.StatusCode != netHttp.StatusOK {
			errBody, _ := ioutil.ReadAll(resp.Body)
			return resp.StatusCode == netHttp.StatusGone, fmt.Errorf("%v: unexpected statusCode %v: body: '%v'", s.URL, resp.StatusCode, string(errBody))
		}
		return eventType == events.EventType_SubscriptionCanceled, nil
	}

	return emitFunc, closeFunc, nil
}
