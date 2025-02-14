package service

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/patrickmn/go-cache"
	"github.com/plgd-dev/hub/pkg/log"
	"github.com/plgd-dev/hub/test/oauth-server/uri"

	router "github.com/gorilla/mux"
)

//RequestHandler for handling incoming request
type RequestHandler struct {
	config             *Config
	authSession        *cache.Cache
	authRestriction    *cache.Cache
	idTokenKey         *rsa.PrivateKey
	idTokenJwkKey      jwk.Key
	accessTokenKey     interface{}
	accessTokenJwkKey  jwk.Key
	refreshRestriction *cache.Cache
}

func createJwkKey(privateKey interface{}) (jwk.Key, error) {
	var alg string
	var publicKey interface{}
	switch v := privateKey.(type) {
	case *rsa.PrivateKey:
		alg = jwa.RS256.String()
		publicKey = &v.PublicKey
	case *ecdsa.PrivateKey:
		alg = jwa.ES256.String()
		publicKey = &v.PublicKey
	}

	jwkKey, err := jwk.New(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create jwk: %w", err)
	}
	data, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal public key: %w", err)
	}

	if err = jwkKey.Set(jwk.KeyIDKey, uuid.NewSHA1(uuid.NameSpaceX500, data).String()); err != nil {
		return nil, fmt.Errorf("failed to set %v: %w", jwk.KeyIDKey, err)
	}
	if err = jwkKey.Set(jwk.AlgorithmKey, alg); err != nil {
		return nil, fmt.Errorf("failed to set %v: %w", jwk.AlgorithmKey, err)
	}
	return jwkKey, nil
}

//NewRequestHandler factory for new RequestHandler
func NewRequestHandler(config *Config, idTokenKey *rsa.PrivateKey, accessTokenKey interface{}) (*RequestHandler, error) {
	idTokenJwkKey, err := createJwkKey(idTokenKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create jwk for idToken: %w", err)
	}
	accessTokenJwkKey, err := createJwkKey(accessTokenKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create jwk for idToken: %w", err)
	}
	return &RequestHandler{
		config:             config,
		authSession:        cache.New(cache.NoExpiration, time.Second*5),
		authRestriction:    cache.New(cache.NoExpiration, time.Second*5),
		idTokenKey:         idTokenKey,
		idTokenJwkKey:      idTokenJwkKey,
		accessTokenJwkKey:  accessTokenJwkKey,
		accessTokenKey:     accessTokenKey,
		refreshRestriction: cache.New(cache.NoExpiration, time.Second*5),
	}, nil
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

// NewHTTP returns HTTP server
func NewHTTP(requestHandler *RequestHandler) *http.Server {
	r := router.NewRouter()
	r.Use(loggingMiddleware)
	r.StrictSlash(true)

	// get JWKs
	r.HandleFunc(uri.JWKs, requestHandler.getJWKs).Methods(http.MethodGet)
	r.HandleFunc(uri.OpenIDConfiguration, requestHandler.getOpenIDConfiguration).Methods(http.MethodGet)

	r.HandleFunc(uri.Authorize, requestHandler.authorize)
	r.HandleFunc(uri.Token, requestHandler.tokenOptions).Methods(http.MethodOptions)
	r.HandleFunc(uri.Token, requestHandler.postToken).Methods(http.MethodPost)
	r.HandleFunc(uri.Token, requestHandler.getToken).Methods(http.MethodGet)
	r.HandleFunc(uri.UserInfo, requestHandler.getUserInfo).Methods(http.MethodGet)
	r.HandleFunc(uri.LogOut, requestHandler.logOut)

	return &http.Server{Handler: r}
}
