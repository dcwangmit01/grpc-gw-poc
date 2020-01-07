package server

import (
	"net/http"
	"regexp"
	"time"

	"github.com/justinas/alice"
	log "github.com/sirupsen/logrus"

	"github.com/dcwangmit01/goapi/jwt"
)

/*
  We may not use http.Handler middleware functions for auth or logging.  We may
    use grpc interceptors (middleware) instead.  GRCP interceptors are one
    level down where we can work with context objects.  Leaving this code here
    for now.

  # Instructions
  srv := &http.Server{p
    Addr:    fmt.Sprintf("%s:%d", config.Host, config.Port),
    Handler: triageHandlerFunc(CommonMiddleware.Then(grpcServer), CommonMiddleware.Then(mux)),

  # HTTP status codes:
  https://golang.org/pkg/net/http/#pkg-constants
*/

var (
	CommonMiddleware        = alice.New(loggingHandler, authHandler)
	middlewareAuthUriRegex  = regexp.MustCompile(`^.*/[Aa]uth$`)
	middlewareAuthBodyRegex = regexp.MustCompile(`^Bearer (?P<jwt>\S+)$`)
)

const (
	authHeader = "Authorization"
)

func loggingHandler(next http.Handler) http.Handler {
	// Taken from: https://github.com/raowl/goapi/blob/master/handlers/middleware.go
	fn := func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		next.ServeHTTP(w, r)
		t2 := time.Now()

		log.Printf("REST [%s] %q %v %v\n", r.Method, r.URL.String(), t2.Sub(t1), r.Header)
	}
	return http.HandlerFunc(fn)
}

func authHandler(next http.Handler) http.Handler {

	// The grpc-gateway will specifically pass the http "Authorization"
	// header (along with "X-Forwarded-For" and "X-Forwarded-Host") to the
	// grpc server as metadata (see: grpc-gateway/runtime/context
	// AnnotateContext).  Thus, we are able to enforce JWT token auth here,
	// which works for both grpc and grpc-gw.
	fn := func(w http.ResponseWriter, r *http.Request) {

		// If the request is for the /auth endpoint, then let the
		// request through without checking for auth.
		if middlewareAuthUriRegex.MatchString(r.URL.String()) {
			next.ServeHTTP(w, r)
			return
		}

		// For all other requests, verify the JWT token and return 401
		// if invalid.
		authString, ok := r.Header[authHeader] // authString is []string
		if !ok {
			http.Error(w, "Authorization Header not Present", http.StatusUnauthorized)
			return
		}

		matches := middlewareAuthBodyRegex.FindStringSubmatch(authString[0])
		if len(matches) != 2 {
			http.Error(w, "Authorization Header Invalid: Bearer Token Not Found", http.StatusUnauthorized)
			return
		}

		token, _, err := jwt.ParseJwt(matches[1])
		if err != nil || token == nil || !token.Valid {
			http.Error(w, "Authorization Token Invalid: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Token successfully validated
		next.ServeHTTP(w, r)
		return
	}
	return http.HandlerFunc(fn)
}
