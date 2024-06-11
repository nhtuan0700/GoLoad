package grpc

import (
	"context"

	"github.com/nhtuan0700/GoLoad/internal/generated/grpc/go_load"
	"github.com/nhtuan0700/GoLoad/internal/logic"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	//nolint:gosec // This is just to specify the metadata name
	AuthTokenMetadataName = "GOLOAD_AUTH"
)

type Handler struct {
	go_load.UnimplementedGoLoadServiceServer
	accountLogic logic.Account
}

func NewHandler(
	accountLogic logic.Account,
) go_load.GoLoadServiceServer {
	return Handler{
		accountLogic: accountLogic,
	}
}

func (h Handler) CreateAccount(
	ctx context.Context,
	request *go_load.CreateAccountRequest,
) (*go_load.CreateAccountResponse, error) {
	output, err := h.accountLogic.CreateAccount(ctx, logic.CreateAccountParams{
		AccountName:     request.AccountName,
		AccountPassword: request.Password,
	})
	if err != nil {
		return nil, err
	}

	return &go_load.CreateAccountResponse{
		AccountId: output.ID,
	}, nil
}

func (h Handler) CreateSession(
	ctx context.Context,
	request *go_load.CreateSessionRequest,
) (*go_load.CreateSessionResponse, error) {
	output, err := h.accountLogic.CreateSession(ctx, logic.CreateSessionParams{
		AccountName:     request.AccountName,
		AccountPassword: request.Password,
	})
	if err != nil {
		return nil, err
	}

	err = grpc.SetHeader(ctx, metadata.Pairs(AuthTokenMetadataName, output.Token))
	if err != nil {
		return nil, err
	}

	return &go_load.CreateSessionResponse{
		Account: output.Account,
	}, nil
}
