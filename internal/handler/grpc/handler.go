package grpc

import (
	"context"
	"errors"
	"io"

	"github.com/nhtuan0700/GoLoad/internal/configs"
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
	accountLogic                                 logic.Account
	downloadTaskLogic                            logic.DownloadTask
	getDownloadTaskFileResponseBufferSizeInBytes uint64
}

func NewHandler(
	accountLogic logic.Account,
	downloadTaskLogic logic.DownloadTask,
	grpcConfig configs.GRPC,
) (go_load.GoLoadServiceServer, error) {
	getDownloadTaskFileResponseBufferSizeInBytes, err := grpcConfig.GetDownloadTaskFile.GetResponseBufferSizeInBytes()
	if err != nil {
		return nil, err
	}

	return Handler{
		accountLogic:      accountLogic,
		downloadTaskLogic: downloadTaskLogic,
		getDownloadTaskFileResponseBufferSizeInBytes: getDownloadTaskFileResponseBufferSizeInBytes,
	}, nil
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
		AccountName:     request.GetAccountName(),
		AccountPassword: request.GetPassword(),
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
		AccountName:     request.GetAccountName(),
		AccountPassword: request.GetPassword(),
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
		URL:          request.GetUrl(),
		DownloadType: request.GetDownloadType(),
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

func (h Handler) GetDownloadTaskFile(
	request *go_load.GetDownloadTaskFileRequest,
	server go_load.GoLoadService_GetDownloadTaskFileServer,
) error {
	outputReader, err := h.downloadTaskLogic.GetDownloadTaskFile(server.Context(), logic.GetDownloadTaskFileParams{
		Token: h.getAuthTokenMetadata(server.Context()),
		ID:    request.GetDownloadTaskId(),
	})
	if err != nil {
		return err
	}
	defer outputReader.Close()

	for {
		dataBuffer := make([]byte, h.getDownloadTaskFileResponseBufferSizeInBytes)
		readByteCount, readErr := outputReader.Read(dataBuffer)
		if readByteCount > 0 {
			sendErr := server.Send(&go_load.GetDownloadTaskFileResponse{
				Data: dataBuffer[:readByteCount],
			})
			if sendErr != nil {
				return sendErr
			}

			continue
		}

		if readErr != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return readErr
		}
	}
	return nil
}

func (h Handler) UpdateDownloadTask(
	ctx context.Context,
	request *go_load.UpdateDownloadTaskRequest,
) (*go_load.UpdateDownloadTaskResponse, error) {
	output, err := h.downloadTaskLogic.UpdateDownloadTask(ctx, logic.UpdateDownloadTaskParams{
		Token: h.getAuthTokenMetadata(ctx),
		ID:    request.GetDownloadTaskId(),
		URL:   request.GetUrl(),
	})
	if err != nil {
		return nil, err
	}

	return &go_load.UpdateDownloadTaskResponse{
		DownloadTask: output.DownloadTask,
	}, nil
}

func (h Handler) DeleteDownloadTask(
	ctx context.Context,
	request *go_load.DeleteDownloadTaskRequest,
) (*go_load.DeleteDownloadTaskResponse, error) {
	err := h.downloadTaskLogic.DeleteDownloadTask(ctx, logic.DeleteDownloadTaskParams{
		Token: h.getAuthTokenMetadata(ctx),
		ID: request.GetDownloadTaskId(),
	})
	if err != nil {
		return nil, err
	}

	return &go_load.DeleteDownloadTaskResponse{}, nil
}
