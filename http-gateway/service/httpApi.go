package service

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/plgd-dev/hub/grpc-gateway/client"
	"github.com/plgd-dev/hub/grpc-gateway/pb"
	"github.com/plgd-dev/hub/http-gateway/uri"
	"github.com/plgd-dev/hub/pkg/log"
	kitHttp "github.com/plgd-dev/hub/pkg/net/http"

	//	"github.com/tmc/grpc-websocket-proxy/wsproxy"
	"github.com/plgd-dev/hub/http-gateway/grpc-websocket-proxy/wsproxy"

	router "github.com/gorilla/mux"
)

//RequestHandler for handling incoming request
type RequestHandler struct {
	client *client.Client
	config *Config
	mux    *runtime.ServeMux
}

//NewRequestHandler factory for new RequestHandler
func NewRequestHandler(config *Config, client *client.Client) *RequestHandler {
	return &RequestHandler{
		client: client,
		config: config,
		mux: runtime.NewServeMux(
			runtime.WithErrorHandler(errorHandler),
			runtime.WithMarshalerOption(uri.ApplicationProtoJsonContentType, newJsonpbMarshaler()),
			runtime.WithMarshalerOption(runtime.MIMEWildcard, newJsonMarshaler()),
		),
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := httputil.DumpRequest(r, false)
		if err != nil {
			log.Infof("Request: %v %v", r.Method, r.RequestURI)
		} else {
			log.Infof("Request: %v", string(data))
		}

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

func makeQueryCaseInsensitive(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, err := url.ParseRequestURI(r.RequestURI)
		if err != nil {
			log.Errorf("cannot make query case insensitive: %v", err)
			next.ServeHTTP(w, r)
			return
		}
		queries := u.Query()
		newQueries := make(url.Values)
		for key, val := range queries {
			newKey, ok := uri.QueryCaseInsensitive[strings.ToLower(key)]
			if ok {
				newQueries[newKey] = val
			} else {
				newQueries[key] = val
			}
		}
		r.URL.RawQuery = newQueries.Encode()
		r.RequestURI = u.String()
		next.ServeHTTP(w, r)
	})
}

func trailSlashSuffix(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, err := url.ParseRequestURI(r.RequestURI)
		if err != nil {
			log.Errorf("cannot trail slash suffix: %v", err)
			next.ServeHTTP(w, r)
			return
		}
		l := len(u.Path)
		u.Path = strings.TrimRight(u.Path, "/")
		if l > 0 && len(u.Path) == 0 {
			u.Path = "/"
		}
		r.RequestURI = u.String()
		r.URL.Path = u.Path
		next.ServeHTTP(w, r)
	})
}

func splitURIPath(requestURI, prefix string) []string {
	v := kitHttp.CanonicalHref(requestURI)
	p := strings.TrimPrefix(v, prefix) // remove core prefix
	if p == v {
		return nil
	}
	p = strings.TrimLeft(p, "/")
	return strings.Split(p, "/")
}

func resourcePendingCommandsMatcher(r *http.Request, rm *router.RouteMatch) bool {
	paths := splitURIPath(r.RequestURI, uri.Devices)
	if len(paths) > 3 && paths[1] == uri.ResourcesPathKey && strings.Contains(paths[len(paths)-1], uri.PendingCommandsPathKey) {
		if rm.Vars == nil {
			rm.Vars = make(map[string]string)
		}
		rm.Vars[uri.DeviceIDKey] = paths[0]
		rm.Vars[uri.ResourceHrefKey] = strings.Split("/"+strings.Join(paths[2:len(paths)-1], "/"), "?")[0]
		return true
	}
	if len(paths) > 4 && paths[1] == uri.ResourcesPathKey && strings.Contains(paths[len(paths)-2], uri.PendingCommandsPathKey) {
		if rm.Vars == nil {
			rm.Vars = make(map[string]string)
		}
		rm.Vars[uri.DeviceIDKey] = paths[0]
		rm.Vars[uri.ResourceHrefKey] = "/" + strings.Join(paths[2:len(paths)-2], "/")
		rm.Vars[uri.CorrelationIDKey] = strings.Split(paths[len(paths)-1], "?")[0]
		return true
	}
	return false
}

func resourceMatcher(r *http.Request, rm *router.RouteMatch) bool {
	paths := splitURIPath(r.RequestURI, uri.Devices)
	if len(paths) > 2 &&
		paths[1] == uri.ResourcesPathKey &&
		!strings.HasPrefix(paths[len(paths)-1], uri.EventsPathKey) {
		if rm.Vars == nil {
			rm.Vars = make(map[string]string)
		}
		rm.Vars[uri.DeviceIDKey] = paths[0]
		rm.Vars[uri.ResourceHrefKey] = strings.Split("/"+strings.Join(paths[2:], "/"), "?")[0]
		return true
	}
	return false
}

func resourceLinksMatcher(r *http.Request, rm *router.RouteMatch) bool {
	paths := splitURIPath(r.RequestURI, uri.Devices)
	if len(paths) > 2 && paths[1] == uri.ResourceLinksPathKey {
		if rm.Vars == nil {
			rm.Vars = make(map[string]string)
		}
		rm.Vars[uri.DeviceIDKey] = paths[0]
		rm.Vars[uri.ResourceHrefKey] = strings.Split("/"+strings.Join(paths[2:], "/"), "?")[0]
		return true
	}
	return false
}

func resourceEventsMatcher(r *http.Request, rm *router.RouteMatch) bool {
	paths := splitURIPath(r.RequestURI, uri.Devices)
	// /api/v1/devices/{deviceId}/resources/{resourceHref}/events
	// /api/v1/devices/{deviceId}/resources/{resourceHref}/events?timestampFilter={timestamp}
	if len(paths) > 3 &&
		paths[1] == uri.ResourcesPathKey &&
		strings.HasPrefix(paths[len(paths)-1], uri.EventsPathKey) {
		if rm.Vars == nil {
			rm.Vars = make(map[string]string)
		}
		rm.Vars[uri.DeviceIDKey] = paths[0]
		rm.Vars[uri.ResourceHrefKey] = "/" + strings.Join(paths[2:len(paths)-1], "/")
		return true
	}
	return false
}

// NewHTTP returns HTTP server
func NewHTTP(requestHandler *RequestHandler, authInterceptor kitHttp.Interceptor) (*http.Server, error) {
	r0 := router.NewRouter()
	r0.Use(loggingMiddleware)
	r0.Use(kitHttp.CreateAuthMiddleware(authInterceptor, func(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
		writeError(w, fmt.Errorf("cannot access to %v: %w", r.RequestURI, err))
	}))
	r0.Use(makeQueryCaseInsensitive)
	r0.Use(trailSlashSuffix)
	r := router.NewRouter()
	r0.PathPrefix("/").Handler(r)

	// Aliases
	r.HandleFunc(uri.AliasDevice, requestHandler.getDevice).Methods(http.MethodGet)
	r.HandleFunc(uri.AliasDevice, requestHandler.deleteDevice).Methods(http.MethodDelete)
	r.HandleFunc(uri.AliasDeviceResourceLinks, requestHandler.getDeviceResourceLinks).Methods(http.MethodGet)
	r.HandleFunc(uri.AliasDeviceResources, requestHandler.getDeviceResources).Methods(http.MethodGet)
	r.HandleFunc(uri.AliasDevicePendingCommands, requestHandler.getDevicePendingCommands).Methods(http.MethodGet)
	r.HandleFunc(uri.AliasDevicePendingMetadataUpdates, requestHandler.getPendingMetadataUpdates).Methods(http.MethodGet)
	r.HandleFunc(uri.AliasDevicePendingMetadataUpdate, requestHandler.cancelPendingMetadataUpdate).Methods(http.MethodDelete)
	r.HandleFunc(uri.AliasDeviceEvents, requestHandler.getEvents).Methods(http.MethodGet)
	r.HandleFunc(uri.HubConfiguration, requestHandler.getHubConfiguration).Methods(http.MethodGet)

	r.PathPrefix(uri.Devices).Methods(http.MethodPost).MatcherFunc(resourceLinksMatcher).HandlerFunc(requestHandler.createResource)
	r.PathPrefix(uri.Devices).Methods(http.MethodGet).MatcherFunc(resourcePendingCommandsMatcher).HandlerFunc(requestHandler.getResourcePendingCommands)
	r.PathPrefix(uri.Devices).Methods(http.MethodDelete).MatcherFunc(resourcePendingCommandsMatcher).HandlerFunc(requestHandler.CancelPendingCommands)
	r.PathPrefix(uri.Devices).Methods(http.MethodGet).MatcherFunc(resourceMatcher).HandlerFunc(requestHandler.getResource)
	r.PathPrefix(uri.Devices).Methods(http.MethodPut).MatcherFunc(resourceMatcher).HandlerFunc(requestHandler.updateResource)
	r.PathPrefix(uri.Devices).Methods(http.MethodGet).MatcherFunc(resourceEventsMatcher).HandlerFunc(requestHandler.getEvents)

	// register grpc-proxy handler
	if err := pb.RegisterGrpcGatewayHandlerClient(context.Background(), requestHandler.mux, requestHandler.client.GrpcGatewayClient()); err != nil {
		return nil, fmt.Errorf("failed to register grpc-gateway handler: %w", err)
	}

	// ws grpc-proxy
	ws := wsproxy.WebsocketProxy(requestHandler.mux,
		wsproxy.WithMaxRespBodyBufferSize(requestHandler.config.APIs.HTTP.WebSocket.StreamBodyLimit),
		wsproxy.WithPingControl(requestHandler.config.APIs.HTTP.WebSocket.PingFrequency),
		wsproxy.WithRequestMutator(func(incoming, outgoing *http.Request) *http.Request {
			outgoing.Method = http.MethodPost
			accept := getAccept(incoming)
			if accept != "" {
				outgoing.Header.Set(uri.AcceptHeaderKey, accept)
			}
			return outgoing
		}))
	r.PathPrefix(uri.APIWS + "/").HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		ws.ServeHTTP(rw, r)
	})

	// api grpc-proxy
	r.Handle(uri.HubConfiguration, requestHandler.mux)
	r.PathPrefix(uri.API + "/").HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		requestHandler.mux.ServeHTTP(rw, r)
	})

	// serve www directory
	if requestHandler.config.UI.Enabled {
		r.HandleFunc(uri.WebConfiguration, requestHandler.getWebConfiguration).Methods(http.MethodGet)
		fs := http.FileServer(http.Dir(requestHandler.config.UI.Directory))
		r.PathPrefix("/").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := httptest.NewRecorder()
			fs.ServeHTTP(c, r)
			if c.Code == http.StatusNotFound {
				c = httptest.NewRecorder()
				r.URL.Path = "/"
				fs.ServeHTTP(c, r)
			}
			for k, v := range c.Header() {
				w.Header().Set(k, strings.Join(v, ""))
			}
			w.WriteHeader(c.Code)
			if _, err := c.Body.WriteTo(w); err != nil {
				log.Errorf("failed to write response body: %w", err)
			}
		}))
	}

	return &http.Server{Handler: r0}, nil
}
