package app

import (
	"context"
	"syscall"

	"github.com/nhtuan0700/GoLoad/internal/handler/grpc"
	"github.com/nhtuan0700/GoLoad/internal/utils"
	"go.uber.org/zap"
)

type StandaloneServer struct {
	grpcServer grpc.Server
	logger     *zap.Logger
}

func NewStandaloneServer(
	grpcServer grpc.Server,
	logger *zap.Logger,
) *StandaloneServer {
	return &StandaloneServer{
		grpcServer: grpcServer,
		logger:     logger,
	}
}

func (s StandaloneServer) Start() error {
	go func() {
		grpcStartErr := s.grpcServer.Start(context.Background())
		s.logger.With(zap.Error(grpcStartErr)).Info("grpc server stopped")
	}()

	utils.BlockUntilSignal(syscall.SIGINT, syscall.SIGTERM)
	return nil
}
