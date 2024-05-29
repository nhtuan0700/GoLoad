package grpc

import (
	"context"
	"net"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/validator"
	"github.com/nhtuan0700/GoLoad/internal/configs"
	go_load "github.com/nhtuan0700/GoLoad/internal/generated/go_load/v1"
	"github.com/nhtuan0700/GoLoad/internal/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server interface {
	Start(ctx context.Context) error
}

type server struct {
	grpcConfig configs.GRPC
	handler    go_load.GoLoadServiceServer
	logger     *zap.Logger
}

func (s *server) Start(ctx context.Context) error {
	logger := utils.LoggerWithContext(ctx, s.logger)

	listener, err := net.Listen("tcp", s.grpcConfig.Address)
	if err != nil {
		return err
	}
	defer listener.Close()

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			validator.UnaryServerInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			validator.StreamServerInterceptor(),
		),
	)

	go_load.RegisterGoLoadServiceServer(server, s.handler)
	logger.With(zap.String("address", s.grpcConfig.Address)).Info("starting grpc server")

	return server.Serve(listener)
}

func NewServer(
	grpcConfig configs.GRPC,
	handler go_load.GoLoadServiceServer,
	logger *zap.Logger,
) Server {
	return &server{
		grpcConfig: grpcConfig,
		handler:    handler,
		logger:     logger,
	}
}
