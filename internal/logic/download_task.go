package logic

import (
	"context"
	"errors"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/nhtuan0700/GoLoad/internal/dataaccess/database"
	"github.com/nhtuan0700/GoLoad/internal/dataaccess/file"
	"github.com/nhtuan0700/GoLoad/internal/dataaccess/mq/producer"
	"github.com/nhtuan0700/GoLoad/internal/generated/grpc/go_load"
	"github.com/nhtuan0700/GoLoad/internal/utils"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

const (
	downloadTaskMetadataFieldNameFileName = "file-name"
)

type CreateDownloadTaskParams struct {
	Token        string
	URL          string
	DownloadType go_load.DownloadType
}

type CreateDownloadTaskOutput struct {
	DownloadTask *go_load.DownloadTask
}

type GetDownloadTaskListParams struct {
	Token  string
	Limit  uint64
	Offset uint64
}

type GetDownloadTaskListOutput struct {
	DonwloadTaskList []*go_load.DownloadTask
	TotalCount       uint64
}

type DownloadTask interface {
	CreateDownloadTask(context.Context, CreateDownloadTaskParams) (CreateDownloadTaskOutput, error)
	ExecuteDownloadTask(context.Context, uint64) error
	GetDownloadTaskList(context.Context, GetDownloadTaskListParams) (GetDownloadTaskListOutput, error)
}

type downloadTask struct {
	goquDatabase                *goqu.Database
	downloadTaskDataAccessor    database.DownloadTaskDataAccessor
	accountDataAccessor         database.AccountDataAccessor
	downloadTaskCreatedProducer producer.DownloadTaskCreatedProducer
	fileClient                  file.Client
	tokenLogic                  Token
	logger                      *zap.Logger
}

func NewDownloadTask(
	goquDatabase *goqu.Database,
	downloadTaskDataAccessor database.DownloadTaskDataAccessor,
	accountDataAccessor database.AccountDataAccessor,
	downloadTaskCreatedProducer producer.DownloadTaskCreatedProducer,
	fileClient file.Client,
	tokenLogic Token,
	logger *zap.Logger,
) DownloadTask {
	return &downloadTask{
		goquDatabase:                goquDatabase,
		downloadTaskDataAccessor:    downloadTaskDataAccessor,
		accountDataAccessor:         accountDataAccessor,
		downloadTaskCreatedProducer: downloadTaskCreatedProducer,
		fileClient:                  fileClient,
		tokenLogic:                  tokenLogic,
		logger:                      logger,
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
		DownloadType:   int32(params.DownloadType),
		URL:            params.URL,
		DownloadStatus: int32(go_load.DownloadStatus_DOWNLOAD_STATUS_PENDING),
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

		producerErr := d.downloadTaskCreatedProducer.Produce(ctx, producer.DownloadTaskCreated{
			ID: downloadTaskID,
		})
		if producerErr != nil {
			return producerErr
		}

		return nil
	})

	if txErr != nil {
		return CreateDownloadTaskOutput{}, txErr
	}

	return CreateDownloadTaskOutput{
		DownloadTask: d.databaseDownloadTaskToProtoDownloadTask(downloadTask, account),
	}, nil
}

func (d downloadTask) updateDownloadStatusFromPendingToDownloading(ctx context.Context, id uint64) (bool, database.DownloadTask, error) {
	var (
		logger       = utils.LoggerWithContext(ctx, d.logger).With(zap.Uint64("id", id))
		updated      = false
		downloadTask database.DownloadTask
		err          error
	)

	txErr := d.goquDatabase.WithTx(func(td *goqu.TxDatabase) error {
		downloadTask, err = d.downloadTaskDataAccessor.WithDatabase(td).GetDownloadTaskWithXLock(ctx, id)
		if err != nil {
			if errors.Is(err, database.ErrAccountNotFound) {
				logger.Warn("download task not found, will skip download")
				return nil
			}
			return err
		}

		if downloadTask.DownloadStatus != int32(go_load.DownloadStatus_DOWNLOAD_STATUS_PENDING) {
			logger.Warn("download is not pending status, will not execute")
			updated = false
			return nil
		}

		downloadTask.DownloadStatus = int32(go_load.DownloadStatus_DOWNLOAD_STATUS_DOWNLOADING)
		err = d.downloadTaskDataAccessor.WithDatabase(td).UpdateDownloadTask(ctx, downloadTask)
		if err != nil {
			return err
		}

		updated = true
		return nil
	})

	if txErr != nil {
		return false, database.DownloadTask{}, err
	}

	return updated, downloadTask, nil
}

func (d downloadTask) updateDownloadStatusFromDownloadingToFailed(ctx context.Context, downloadTask database.DownloadTask) error {
	logger := utils.LoggerWithContext(ctx, d.logger)

	downloadTask.DownloadStatus = int32(go_load.DownloadStatus_DOWNLOAD_STATUS_FAILED)
	updateDownloadErr := d.downloadTaskDataAccessor.UpdateDownloadTask(ctx, downloadTask)
	if updateDownloadErr != nil {
		logger.With(zap.Error(updateDownloadErr)).Error("failed to update download task to failed")
		return updateDownloadErr
	}

	return nil
}

func (d downloadTask) ExecuteDownloadTask(ctx context.Context, id uint64) error {
	logger := utils.LoggerWithContext(ctx, d.logger).With(zap.Uint64("id", id))
	updated, downloadTask, err := d.updateDownloadStatusFromPendingToDownloading(ctx, id)
	if err != nil {
		return err
	}
	if !updated {
		return nil
	}

	var downloader Downloader
	switch downloadTask.DownloadType {
	case int32(go_load.DownloadType_DOWNLOAD_TYPE_HTTP):
		downloader = NewDownloader(downloadTask.URL, d.logger)

	default:
		logger.With(zap.Any("download_type", downloadTask.DownloadType)).Error("unsupported download type")
		err := d.updateDownloadStatusFromDownloadingToFailed(ctx, downloadTask)
		if err != nil {
			return err
		}
		return nil
	}

	fileName := fmt.Sprintf("download_file_%d", id)
	fileWriterCloser, err := d.fileClient.Writer(ctx, fileName)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get file writer")
		if err := d.updateDownloadStatusFromDownloadingToFailed(ctx, downloadTask); err != nil {
			return err
		}
		return err
	}
	defer fileWriterCloser.Close()

	metadata, err := downloader.Download(ctx, fileWriterCloser)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to download task")
		if err := d.updateDownloadStatusFromDownloadingToFailed(ctx, downloadTask); err != nil {
			return err
		}
		return err
	}

	metadata[downloadTaskMetadataFieldNameFileName] = fileName
	downloadTask.DownloadStatus = int32(go_load.DownloadStatus_DOWNLOAD_STATUS_SUCCESS)
	downloadTask.Metadata = database.JSON{
		Data: metadata,
	}

	err = d.downloadTaskDataAccessor.UpdateDownloadTask(ctx, downloadTask)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to update download task status to success")
		return err
	}

	logger.Info("Download task executed successfully")
	return nil
}

func (d downloadTask) GetDownloadTaskList(ctx context.Context, params GetDownloadTaskListParams) (GetDownloadTaskListOutput, error) {
	accountID, _, err := d.tokenLogic.GetAccountIDAndExpireTime(ctx, params.Token)
	if err != nil {
		return GetDownloadTaskListOutput{}, err
	}

	account, err := d.accountDataAccessor.GetAccountByID(ctx, accountID)
	if err != nil {
		return GetDownloadTaskListOutput{}, err
	}

	downloadTaskList, count, err := d.downloadTaskDataAccessor.GetDownloadTaskListByAccount(ctx, accountID, params.Limit, params.Offset)
	if err != nil {
		return GetDownloadTaskListOutput{}, err
	}

	return GetDownloadTaskListOutput{
		DonwloadTaskList: lo.Map(downloadTaskList, func(item database.DownloadTask, _ int) *go_load.DownloadTask {
			return d.databaseDownloadTaskToProtoDownloadTask(item, account)
		}),
		TotalCount: count,
	}, nil
}
