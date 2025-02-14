package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	router "github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
	"github.com/plgd-dev/hub/cloud2cloud-connector/events"
	"github.com/plgd-dev/hub/cloud2cloud-connector/store"
	"github.com/plgd-dev/hub/cloud2cloud-connector/uri"
	kitNetHttp "github.com/plgd-dev/hub/pkg/net/http"
	pkgOAuth2 "github.com/plgd-dev/hub/pkg/security/oauth2"
	"github.com/plgd-dev/kit/v2/log"
)

const cloudIDKey = "CloudId"
const accountIDKey = "AccountId"

type provisionCacheData struct {
	linkedAccount store.LinkedAccount
	linkedCloud   store.LinkedCloud
}

// RequestHandler handles incoming requests
type RequestHandler struct {
	ownerClaim     string
	provider       *pkgOAuth2.PlgdProvider
	store          *Store
	provisionCache *cache.Cache
	subManager     *SubscriptionManager
	triggerTask    OnTaskTrigger
}

func logAndWriteErrorResponse(err error, statusCode int, w http.ResponseWriter) {
	log.Errorf("%v", err)
	w.Header().Set(events.ContentTypeKey, "text/plain")
	w.WriteHeader(statusCode)
	if _, err2 := w.Write([]byte(err.Error())); err2 != nil {
		log.Errorf("failed to write error response body: %w", err2)
	}
}

func NewRequestHandler(
	ownerClaim string,
	provider *pkgOAuth2.PlgdProvider,
	subManager *SubscriptionManager,
	store *Store,
	triggerTask OnTaskTrigger,
) *RequestHandler {
	return &RequestHandler{
		ownerClaim:     ownerClaim,
		provider:       provider,
		subManager:     subManager,
		store:          store,
		provisionCache: cache.New(5*time.Minute, 10*time.Minute),
		triggerTask:    triggerTask,
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Infof("%v %v %+v", r.Method, r.RequestURI, r.Header)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// NewHTTP returns HTTP server
func NewHTTP(requestHandler *RequestHandler, authInterceptor kitNetHttp.Interceptor) (*http.Server, error) {
	r := router.NewRouter()
	r.StrictSlash(true)
	r.Use(loggingMiddleware)
	r.Use(kitNetHttp.CreateAuthMiddleware(authInterceptor, func(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
		logAndWriteErrorResponse(fmt.Errorf("cannot process request on %v: %w", r.RequestURI, err), http.StatusUnauthorized, w)
	}))

	// health check
	r.HandleFunc("/", healthCheck).Methods("GET")

	// retrieve all linked clouds
	r.HandleFunc(uri.LinkedClouds, requestHandler.RetrieveLinkedClouds).Methods("GET")
	// add linked cloud
	r.HandleFunc(uri.LinkedClouds, requestHandler.AddLinkedCloud).Methods("POST")
	// delete linked cloud
	r.HandleFunc(uri.LinkedCloud, requestHandler.DeleteLinkedCloud).Methods("DELETE")
	// add linked account
	r.HandleFunc(uri.LinkedAccounts, requestHandler.AddLinkedAccount).Methods("GET")
	// delete linked cloud
	r.HandleFunc(uri.LinkedAccount, requestHandler.DeleteLinkedAccount).Methods("DELETE")
	// notify linked cloud
	r.HandleFunc(uri.Events, requestHandler.ProcessEvent).Methods("POST")

	oauthURL, err := parseOAuthPaths(requestHandler.provider.Config.RedirectURL)
	if err != nil {
		return nil, err
	}
	r.HandleFunc(oauthURL.Path, requestHandler.OAuthCallback).Methods("GET")

	return &http.Server{Handler: r}, nil
}
