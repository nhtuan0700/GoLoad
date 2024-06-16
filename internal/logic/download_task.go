package logic

import (
	"context"

	"github.com/doug-martin/goqu/v9"
	"github.com/nhtuan0700/GoLoad/internal/dataaccess/database"
	"github.com/nhtuan0700/GoLoad/internal/generated/grpc/go_load"
	"go.uber.org/zap"
)

type CreateDownloadTaskParams struct {
	Token        string
	URL          string
	DownloadType go_load.DownloadType
}

type CreateDownloadTaskOutput struct {
	DownloadTask *go_load.DownloadTask
}

type DownloadTask interface {
	CreateDownloadTask(ctx context.Context, params CreateDownloadTaskParams) (CreateDownloadTaskOutput, error)
}

type downloadTask struct {
	goquDatabase             *goqu.Database
	downloadTaskDataAccessor database.DownloadTaskDataAccessor
	accountDataAccessor      database.AccountDataAccessor
	tokenLogic               Token
	logger                   *zap.Logger
}

func NewDownloadTask(
	goquDatabase *goqu.Database,
	downloadTaskDataAccessor database.DownloadTaskDataAccessor,
	accountDataAccessor database.AccountDataAccessor,
	tokenLogic Token,
	logger *zap.Logger,
) DownloadTask {
	return &downloadTask{
		goquDatabase:             goquDatabase,
		downloadTaskDataAccessor: downloadTaskDataAccessor,
		accountDataAccessor:      accountDataAccessor,
		tokenLogic:               tokenLogic,
		logger:                   logger,
	}
}

func (d *downloadTask) databaseDownloadTaskToProtoDownloadTask(
	downloadTask database.DownloadTask,
	account database.Account,
) *go_load.DownloadTask {
	return &go_load.DownloadTask{
		Id: downloadTask.ID,
		OfAccount: &go_load.Account{
			Id:          account.ID,
			AccountName: account.Name,
		},
		DownloadType:   go_load.DownloadType(downloadTask.DownloadType),
		Url:            downloadTask.URL,
		DownloadStatus: go_load.DownloadStatus(downloadTask.DownloadStatus),
	}
}

func (d *downloadTask) CreateDownloadTask(
	ctx context.Context,
	params CreateDownloadTaskParams,
) (CreateDownloadTaskOutput, error) {
	accountID, _, err := d.tokenLogic.GetAccountIDAndExpireTime(ctx, params.Token)
	if err != nil {
		return CreateDownloadTaskOutput{}, err
	}

	account, err := d.accountDataAccessor.GetAccountByID(ctx, accountID)
	if err != nil {
		return CreateDownloadTaskOutput{}, err
	}

	downloadTask := database.DownloadTask{
		OfAccountID:    account.ID,
		DownloadType:   int16(params.DownloadType),
		URL:            params.URL,
		DownloadStatus: int16(go_load.DownloadStatus_DOWNLOAD_STATUS_PENDING),
		Metadata: database.JSON{
			Data: make(map[string]any),
		},
	}
	txErr := d.goquDatabase.WithTx(func(td *goqu.TxDatabase) error {
		downloadTaskID, createDownloadTaskErr := d.downloadTaskDataAccessor.
			WithDatabase(td).
			CreateDownloadTask(ctx, downloadTask)
		if createDownloadTaskErr != nil {
			return createDownloadTaskErr
		}

		downloadTask.ID = downloadTaskID
		// TODO producer
		return nil
	})

	if txErr != nil {
		return CreateDownloadTaskOutput{}, txErr
	}

	return CreateDownloadTaskOutput{
		DownloadTask: d.databaseDownloadTaskToProtoDownloadTask(downloadTask, account),
	}, nil
}
