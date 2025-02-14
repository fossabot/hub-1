package service

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"sync"

	"github.com/plgd-dev/hub/cloud2cloud-gateway/store/mongodb"
	pbGRPC "github.com/plgd-dev/hub/grpc-gateway/pb"
	"github.com/plgd-dev/hub/pkg/fn"
	"github.com/plgd-dev/hub/pkg/log"
	kitNetGrpc "github.com/plgd-dev/hub/pkg/net/grpc"
	grpcClient "github.com/plgd-dev/hub/pkg/net/grpc/client"
	kitNetHttp "github.com/plgd-dev/hub/pkg/net/http"
	"github.com/plgd-dev/hub/pkg/net/listener"
	cmClient "github.com/plgd-dev/hub/pkg/security/certManager/client"
	"github.com/plgd-dev/hub/pkg/security/jwt/validator"
	"github.com/plgd-dev/hub/pkg/sync/task/queue"
	raClient "github.com/plgd-dev/hub/resource-aggregate/client"
	natsClient "github.com/plgd-dev/hub/resource-aggregate/cqrs/eventbus/nats/client"
	"github.com/plgd-dev/hub/resource-aggregate/cqrs/eventbus/nats/subscriber"
	"github.com/plgd-dev/hub/resource-aggregate/cqrs/utils"
)

//Server handle HTTP request
type Server struct {
	server           *http.Server
	listener         *listener.Server
	cancelSubMgrFunc context.CancelFunc
	subMgrDoneWg     *sync.WaitGroup
}

// https://openconnectivity.org/draftspecs/Gaborone/OCF_Cloud_API_for_Cloud_Services.pdf
var authRules = map[string][]kitNetHttp.AuthArgs{
	http.MethodGet: {
		{
			URI: regexp.MustCompile(`[\/]+api[\/]+v1[\/]+devices[\/]*$`),
			Scopes: []*regexp.Regexp{
				regexp.MustCompile(`r:.*`),
			},
		},
		{
			URI: regexp.MustCompile(`[\/]+api[\/]+v1[\/]+devices[\/]*\?` + ContentQuery + `=` + ContentQueryBaseValue + `$`),
			Scopes: []*regexp.Regexp{
				regexp.MustCompile(`r:.*`),
			},
		},
		{
			URI: regexp.MustCompile(`[\/]+api[\/]+v1[\/]+devices[\/]*\?` + ContentQuery + `=` + ContentQueryAllValue + `$`),
			Scopes: []*regexp.Regexp{
				regexp.MustCompile(`r:.*`),
			},
		},
		{
			URI: regexp.MustCompile(`[\/]+api[\/]+v1[\/]+devices[\/]+[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12}[\/]*$`),
			Scopes: []*regexp.Regexp{
				regexp.MustCompile(`r:.*`),
			},
		},
		{
			URI: regexp.MustCompile(`[\/]+api[\/]+v1[\/]+devices[\/]+[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12}[\/]*\?` + ContentQuery + `=` + ContentQueryBaseValue + `$`),
			Scopes: []*regexp.Regexp{
				regexp.MustCompile(`r:.*`),
			},
		},
		{
			URI: regexp.MustCompile(`[\/]+api[\/]+v1[\/]+devices[\/]+[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12}[\/]*\?` + ContentQuery + `=` + ContentQueryAllValue + `$`),
			Scopes: []*regexp.Regexp{
				regexp.MustCompile(`r:.*`),
			},
		},
		{
			URI: regexp.MustCompile(`[\/]+api[\/]+v1[\/]+devices[\/]+[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12}[\/]+.*[\/]*$`),
			Scopes: []*regexp.Regexp{
				regexp.MustCompile(`r:.*`),
			},
		},
	},
	http.MethodPost: {
		{
			URI: regexp.MustCompile(`[\/]+api[\/]+v1[\/]+devices[\/]+subscriptions[\/]*$`),
			Scopes: []*regexp.Regexp{
				regexp.MustCompile(`r:.*`),
			},
		},
		{
			URI: regexp.MustCompile(`[\/]+api[\/]+v1[\/]+devices[\/]+[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12}[\/]+subscriptions[\/]*$`),
			Scopes: []*regexp.Regexp{
				regexp.MustCompile(`r:.*`),
			},
		},
		{
			URI: regexp.MustCompile(`[\/]+api[\/]+v1[\/]+devices[\/]+[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12}[\/]+.*[\/]+subscriptions[\/]*$`),
			Scopes: []*regexp.Regexp{
				regexp.MustCompile(`r:.*`),
			},
		},
		{
			URI: regexp.MustCompile(`[\/]+api[\/]+v1[\/]+devices[\/]+[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12}[\/]+.*[\/]*$`),
			Scopes: []*regexp.Regexp{
				regexp.MustCompile(`w:.*`),
			},
		},
	},
	http.MethodDelete: {
		{
			URI: regexp.MustCompile(`[\/]+api[\/]+v1[\/]+devices[\/]+subscriptions[\/]+[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12}[\/]*$`),
			Scopes: []*regexp.Regexp{
				regexp.MustCompile(`r:.*`),
			},
		},
		{
			URI: regexp.MustCompile(`[\/]+api[\/]+v1[\/]+devices[\/]+[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12}[\/]+subscriptions[\/]+[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12}[\/]*$`),
			Scopes: []*regexp.Regexp{
				regexp.MustCompile(`r:.*`),
			},
		},
		{
			URI: regexp.MustCompile(`[\/]+api[\/]+v1[\/]+devices[\/]+[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12}[\/]+.*[\/]+subscriptions[\/]+[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12}[\/]*$`),
			Scopes: []*regexp.Regexp{
				regexp.MustCompile(`r:.*`),
			},
		},
	},
}

func newGrpcGatewayClient(config GrpcGatewayConfig, logger log.Logger) (pbGRPC.GrpcGatewayClient, func(), error) {
	var fl fn.FuncList
	conn, err := grpcClient.New(config.Connection, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot connect to grpc-gateway: %w", err)
	}
	fl.AddFunc(func() {
		if err := conn.Close(); err != nil && !kitNetGrpc.IsContextCanceled(err) {
			logger.Errorf("error occurs during closing of the connection to grpc-gateway: %w", err)
		}
	})
	client := pbGRPC.NewGrpcGatewayClient(conn.GRPC())
	return client, fl.ToFunction(), nil
}

func newResourceSubscriber(config Config, logger log.Logger) (*subscriber.Subscriber, func(), error) {
	var fl fn.FuncList
	nats, err := natsClient.New(config.Clients.Eventbus.NATS, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create nats client: %w", err)
	}
	fl.AddFunc(nats.Close)

	pool, err := queue.New(config.TaskQueue)
	if err != nil {
		fl.Execute()
		return nil, nil, fmt.Errorf("cannot create job queue %w", err)
	}
	fl.AddFunc(pool.Release)

	resourceSubscriber, err := subscriber.New(nats.GetConn(),
		config.Clients.Eventbus.NATS.PendingLimits,
		logger,
		subscriber.WithGoPool(func(f func()) error { return pool.Submit(f) }),
		subscriber.WithUnmarshaler(utils.Unmarshal))
	if err != nil {
		fl.Execute()
		return nil, nil, fmt.Errorf("cannot create eventbus subscriber: %w", err)
	}
	fl.AddFunc(resourceSubscriber.Close)

	return resourceSubscriber, fl.ToFunction(), nil
}

func newResourceAggregateClient(config ResourceAggregateConfig, subscriber *subscriber.Subscriber, logger log.Logger) (*raClient.Client, func(), error) {
	var fl fn.FuncList
	conn, err := grpcClient.New(config.Connection, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot connect to resource aggregate: %w", err)
	}
	fl.AddFunc(func() {
		if err := conn.Close(); err != nil && !kitNetGrpc.IsContextCanceled(err) {
			logger.Errorf("error occurs during closing of the connection to resource-aggregate: %w", err)
		}
	})
	client := raClient.New(conn.GRPC(), subscriber)
	return client, fl.ToFunction(), nil
}

func newSubscriptionManager(ctx context.Context, cfg Config, gwClient pbGRPC.GrpcGatewayClient, emitEvent emitEventFunc, logger log.Logger) (*SubscriptionManager, func(), error) {
	var fl fn.FuncList
	certManager, err := cmClient.New(cfg.Clients.Storage.MongoDB.TLS, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create cert manager: %w", err)
	}
	fl.AddFunc(certManager.Close)

	store, err := mongodb.NewStore(ctx, cfg.Clients.Storage.MongoDB, certManager.GetTLSConfig())
	if err != nil {
		fl.Execute()
		return nil, nil, fmt.Errorf("cannot create mongodb subscription store: %w", err)
	}
	fl.AddFunc(func() {
		if err := store.Close(ctx); err != nil {
			log.Errorf("failed to close subscription store: %w", err)
		}
	})

	subMgr := NewSubscriptionManager(ctx, store, gwClient, cfg.Clients.Subscription.HTTP.ReconnectInterval, emitEvent)
	if err := subMgr.LoadSubscriptions(); err != nil {
		fl.Execute()
		return nil, nil, fmt.Errorf("cannot load subscriptions: %w", err)
	}

	return subMgr, fl.ToFunction(), nil
}

// New parses configuration and creates new Server with provided store and bus
func New(ctx context.Context, config Config, logger log.Logger) (*Server, error) {
	listener, err := listener.New(config.APIs.HTTP.Connection, logger)
	if err != nil {
		return nil, fmt.Errorf("cannot create http server: %w", err)
	}
	closeListener := func() {
		if err := listener.Close(); err != nil {
			logger.Errorf("cannot create http server: %w", err)
		}
	}

	validator, err := validator.New(ctx, config.APIs.HTTP.Authorization, logger)
	if err != nil {
		closeListener()
		return nil, fmt.Errorf("cannot create validator: %w", err)
	}
	listener.AddCloseFunc(validator.Close)
	auth := kitNetHttp.NewInterceptorWithValidator(validator, authRules)

	gwClient, closeGwClient, err := newGrpcGatewayClient(config.Clients.GrpcGateway, logger)
	if err != nil {
		closeListener()
		return nil, fmt.Errorf("cannot create grpc client: %w", err)
	}
	listener.AddCloseFunc(closeGwClient)

	subscriber, closeSubscriberFn, err := newResourceSubscriber(config, logger)
	if err != nil {
		closeListener()
		return nil, fmt.Errorf("cannot create resource subscriber: %w", err)
	}
	listener.AddCloseFunc(closeSubscriberFn)

	raClient, closeRaClient, err := newResourceAggregateClient(config.Clients.ResourceAggregate, subscriber, logger)
	if err != nil {
		closeListener()
		return nil, fmt.Errorf("cannot create resource-aggregate client: %w", err)
	}
	listener.AddCloseFunc(closeRaClient)

	emitEvent, closeEmitEventFn, err := createEmitEventFunc(config.Clients.Subscription.HTTP.TLS, config.Clients.Subscription.HTTP.EmitEventTimeout, logger)
	if err != nil {
		closeListener()
		return nil, fmt.Errorf("cannot create emit event function: %w", err)
	}
	listener.AddCloseFunc(closeEmitEventFn)

	ctx, cancelSubMgrFunc := context.WithCancel(context.Background())
	subMgr, closeSubMgrFn, err := newSubscriptionManager(ctx, config, gwClient, emitEvent, logger)
	if err != nil {
		cancelSubMgrFunc()
		closeListener()
		return nil, fmt.Errorf("cannot create subscription manager: %w", err)
	}
	listener.AddCloseFunc(closeSubMgrFn)

	var subMgrWg sync.WaitGroup
	subMgrWg.Add(1)
	go func() {
		defer subMgrWg.Done()
		subMgr.Run()
	}()

	requestHandler := NewRequestHandler(gwClient, raClient, subMgr, emitEvent)

	server := Server{
		server:           NewHTTP(requestHandler, auth),
		listener:         listener,
		cancelSubMgrFunc: cancelSubMgrFunc,
		subMgrDoneWg:     &subMgrWg,
	}

	return &server, nil
}

// Serve starts the service's HTTP server and blocks
func (s *Server) Serve() error {
	defer func() {
		s.subMgrDoneWg.Wait()
	}()
	return s.server.Serve(s.listener)
}

// Shutdown ends serving
func (s *Server) Shutdown() error {
	s.cancelSubMgrFunc()
	return s.server.Shutdown(context.Background())
}
