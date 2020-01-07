package server

/*
  Interceptors are GRPC middleware.

  # GRPC error codes
  https://godoc.org/google.golang.org/grpc/codes#Code
  # GRPC error codes to http.Errors
  https://github.com/grpc-ecosystem/grpc-gateway/blob/master/runtime/errors.go#L15

*/

import (
	"regexp"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_tags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/dcwangmit01/goapi/config"
	"github.com/dcwangmit01/goapi/jwt"
)

var (
	CommonInterceptors      grpc.ServerOption
	interceptorAuthUriRegex = regexp.MustCompile(`^.*/[Aa]uth$`)
)

func init() {
	logger := log.WithFields(log.Fields{
		"goapi": "server",
	})

	CommonInterceptors = grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
		// order matters
		grpc_logrus.UnaryServerInterceptor(logger),
		AuthInterceptor(),
		// grpc_tags will take http headers and place into tags struct
		//   grpc_tags.Extract(ctx) returns *Tags which is "map[string]interface"
		grpc_tags.UnaryServerInterceptor(),
	))
}

func AuthInterceptor() grpc.UnaryServerInterceptor {

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		// If the request is for the /auth endpoint, then let the
		// request through without checking for auth.
		if interceptorAuthUriRegex.MatchString(info.FullMethod) {
			return handler(ctx, req)
		}
		// Otherwise, continue below and validate the JWT token

		// Check the authorization header
		tokenStr, err := grpc_auth.AuthFromMD(ctx, "Bearer")
		if err != nil {
			return req, grpc.Errorf(codes.Unauthenticated, err.Error())
		}

		// Parse the JWT token string into a token object
		token, claims, err := jwt.ParseJwt(tokenStr)
		if err != nil || token == nil {
			return req, grpc.Errorf(codes.Unauthenticated, err.Error())
		} else if !token.Valid {
			return req, grpc.Errorf(codes.Unauthenticated, "Invalid Token")
		}

		// Find the user
		user, err := config.AppConfig.GetUserById(claims.Id)
		if err != nil {
			return req, grpc.Errorf(codes.Unauthenticated, "Invalid User Id")
		}

		newCtx := config.UserNewContext(ctx, user)
		return handler(newCtx, req)
	}
}
