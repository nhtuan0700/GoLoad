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
	accountLogic      logic.Account
	downloadTaskLogic logic.DownloadTask
}

func NewHandler(
	accountLogic logic.Account,
	downloadTaskLogic logic.DownloadTask,
) go_load.GoLoadServiceServer {
	return Handler{
		accountLogic:      accountLogic,
		downloadTaskLogic: downloadTaskLogic,
	}
}

func (h Handler) getAuthTokenMetadata(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	metadataValues := md.Get(AuthTokenMetadataName)
	if len(metadataValues) == 0 {
		return ""
	}

	return metadataValues[0]
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

func (h Handler) CreateDownloadTask(
	ctx context.Context,
	request *go_load.CreateDownloadTaskRequest,
) (*go_load.CreateDownloadTaskResponse, error) {
	output, err := h.downloadTaskLogic.CreateDownloadTask(ctx, logic.CreateDownloadTaskParams{
		Token:        h.getAuthTokenMetadata(ctx),
		URL:          request.Url,
		DownloadType: request.DownloadType,
	})

	if err != nil {
		return nil, err
	}

	return &go_load.CreateDownloadTaskResponse{
		DownloadTask: output.DownloadTask,
	}, nil
}

func (h Handler) GetDownloadTaskList(
	ctx context.Context,
	request *go_load.GetDownloadTaskListRequest,
) (*go_load.GetDownloadTaskListResponse, error) {
	output, err := h.downloadTaskLogic.GetDownloadTaskList(ctx, logic.GetDownloadTaskListParams{
		Token:  h.getAuthTokenMetadata(ctx),
		Limit:  request.Limit,
		Offset: request.Offset,
	})
	if err != nil {
		return nil, err
	}

	return &go_load.GetDownloadTaskListResponse{
		DownloadTaskList: output.DonwloadTaskList,
		TotalCount:       output.TotalCount,
	}, nil
}
