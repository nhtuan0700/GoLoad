package http

import (
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/nhtuan0700/GoLoad/internal/configs"
	"github.com/nhtuan0700/GoLoad/internal/generated/grpc/go_load"
	handlerGRPC "github.com/nhtuan0700/GoLoad/internal/handler/grpc"
	"github.com/nhtuan0700/GoLoad/internal/handler/http/servermuxoptions"
	"github.com/nhtuan0700/GoLoad/internal/utils"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	//nolint:gosec // This is just to specify the metadata name
	AuthTokenCookieName = "GOLOAD_AUTH"
)

type Server interface {
	Start(ctx context.Context) error
}

type server struct {
	grpcConfig configs.GRPC
	httpConfig configs.HTTP
	authConfig configs.Auth
	logger     *zap.Logger
}

func NewServer(
	grpcConfig configs.GRPC,
	httpConfig configs.HTTP,
	authConfig configs.Auth,
	logger *zap.Logger,
) Server {
	return &server{
		grpcConfig: grpcConfig,
		httpConfig: httpConfig,
		authConfig: authConfig,
		logger:     logger,
	}
}

func (s server) getGRPCGatewayHandler(ctx context.Context) (http.Handler, error) {
	tokenExpiresIn, err := s.authConfig.Token.GetExpiresInDuration()
	if err != nil {
		return nil, err
	}

	grpcMux := runtime.NewServeMux(
		servermuxoptions.WithAuthCookieToAuthMetadata(AuthTokenCookieName, handlerGRPC.AuthTokenMetadataName),
		servermuxoptions.WithAuthMetadataToAuthCookie(handlerGRPC.AuthTokenMetadataName, AuthTokenCookieName, tokenExpiresIn),
		servermuxoptions.WithRemoveGoAuthMetadata(handlerGRPC.AuthTokenMetadataName),
	)

	err = go_load.RegisterGoLoadServiceHandlerFromEndpoint(
		ctx,
		grpcMux,
		s.grpcConfig.Address,
		[]grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		},
	)

	if err != nil {
		return nil, err
	}

	return grpcMux, nil
}

func (s server) Start(ctx context.Context) error {
	logger := utils.LoggerWithContext(ctx, s.logger)

	grpcGatewayHandler, err := s.getGRPCGatewayHandler(ctx)
	if err != nil {
		return err
	}

	httpServer := http.Server{
		Addr:              s.httpConfig.Address,
		Handler:           grpcGatewayHandler,
		ReadHeaderTimeout: time.Minute,
	}

	logger.With(zap.String("address", s.httpConfig.Address)).Info("Starting http server")
	return httpServer.ListenAndServe()
}
