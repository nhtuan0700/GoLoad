package grpc

import (
	"context"

	go_load "github.com/nhtuan0700/GoLoad/internal/generated/go_load/v1"
	"github.com/nhtuan0700/GoLoad/internal/logic"
)

const (
	//nolint:gosec // This is just to specify the meta data name
	AuthTokenMetadataName = "GOLOAD_AUTH"
)

type Handler struct {
	go_load.UnimplementedGoLoadServiceServer
	accountLogic logic.Account
}

func NewHandler(
	accountLogic logic.Account,
) (go_load.GoLoadServiceServer, error) {
	return &Handler{
		accountLogic: accountLogic,
	}, nil
}

func (h Handler) CreateAccount(
	ctx context.Context,
	request *go_load.CreateAccountRequest,
) (*go_load.CreateAccountResponse, error) {
	output, err := h.accountLogic.CreateAccount(ctx, logic.CreateAccountParams{
		AccountName: request.AccountName,
		Password: request.Password,
	})

	if err != nil {
		return nil, err
	}

	return &go_load.CreateAccountResponse{
		AccountId: output.ID,
	}, nil
}
