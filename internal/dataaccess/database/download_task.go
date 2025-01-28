package database

import (
	"context"

	"github.com/doug-martin/goqu/v9"
	"github.com/nhtuan0700/GoLoad/internal/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	TableNameDownloadTask = goqu.T("download_tasks")

	ErrDownloadTaskNotFound = status.Error(codes.NotFound, "download task not found")
)

const (
	ColNameDownloadTaskID             = "id"
	ColNameDownloadTaskOfAccountID    = "of_account_id"
	ColNameDownloadTaskDownloadType   = "download_type"
	ColNameDownloadTaskURL            = "url"
	ColNameDownloadTaskDownloadStatus = "download_status"
	ColNameDownloadTaskMetadata       = "metadata"
)

type DownloadTask struct {
	ID             uint64 `db:"id" goqu:"skipinsert,skipupdate"`
	OfAccountID    uint64 `db:"of_account_id" goqu:"skipupdate"`
	DownloadType   int32  `db:"download_type"`
	URL            string `db:"url"`
	DownloadStatus int32  `db:"download_status"`
	Metadata       JSON   `db:"metadata"`
}

type DownloadTaskDataAccessor interface {
	CreateDownloadTask(ctx context.Context, downloadTask DownloadTask) (uint64, error)
	UpdateDownloadTask(ctx context.Context, downloadTask DownloadTask) error
	GetDownloadTaskWithXLock(ctx context.Context, id uint64) (DownloadTask, error)
	GetDownloadTaskListByAccount(ctx context.Context, accountID uint64, limit uint64, offset uint64) ([]DownloadTask, uint64, error)
	WithDatabase(database Database) DownloadTaskDataAccessor
}

type downloadTaskDataAccessor struct {
	database Database
	logger   *zap.Logger
}

func NewDownloadTaskDataAccessor(
	database *goqu.Database,
	logger *zap.Logger,
) DownloadTaskDataAccessor {
	return &downloadTaskDataAccessor{
		database: database,
		logger:   logger,
	}
}

func (d *downloadTaskDataAccessor) CreateDownloadTask(ctx context.Context, downloadTask DownloadTask) (uint64, error) {
	logger := utils.LoggerWithContext(ctx, d.logger).With(zap.Any("download_task", downloadTask))

	result, err := d.database.
		Insert(TableNameDownloadTask).
		Rows(downloadTask).
		Executor().
		ExecContext(ctx)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to create download task")
		return 0, status.Error(codes.Internal, "failed to create download task")
	}

	lastInsertedID, err := result.LastInsertId()
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get last inserted id")
		return 0, status.Error(codes.Internal, "failed to get last inserted id")
	}

	return uint64(lastInsertedID), nil
}

func (d downloadTaskDataAccessor) UpdateDownloadTask(ctx context.Context, downloadTask DownloadTask) error {
	logger := utils.LoggerWithContext(ctx, d.logger).With(zap.Any("download_task", downloadTask))

	_, err := d.database.
		Update(TableNameDownloadTask).
		Set(downloadTask).
		Where(goqu.Ex{ColNameDownloadTaskID: downloadTask.ID}).
		Executor().
		ExecContext(ctx)

	if err != nil {
		logger.With(zap.Error(err)).Error("failed to update download task")
		return status.Error(codes.Internal, "failed to update download task")
	}

	return nil
}

func (d downloadTaskDataAccessor) GetDownloadTaskWithXLock(ctx context.Context, id uint64) (DownloadTask, error) {
	logger := utils.LoggerWithContext(ctx, d.logger).With(zap.Uint64("id", id))

	downloadTask := DownloadTask{}
	found, err := d.database.
		Select().
		From(TableNameDownloadTask).
		Where(goqu.Ex{ColNameDownloadTaskID: id}).
		ForUpdate(goqu.Wait).
		ScanStructContext(ctx, &downloadTask)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get download task by id")
		return DownloadTask{}, status.Error(codes.Internal, "failed to get download task by id")
	}

	if !found {
		logger.Warn("download task not found")
		return DownloadTask{}, ErrDownloadTaskNotFound
	}

	return downloadTask, nil
}

func (d *downloadTaskDataAccessor) GetDownloadTaskListByAccount(
	ctx context.Context,
	accountID uint64,
	limit uint64,
	offset uint64,
) ([]DownloadTask, uint64, error) {
	logger := utils.LoggerWithContext(ctx, d.logger).
		With(zap.Uint64("account_id", accountID)).
		With(zap.Uint64("limit", limit)).
		With(zap.Uint64("offset", offset))

	var downloadTaskList []DownloadTask
	if err := d.database.
		Select().
		From(TableNameDownloadTask).
		Where(goqu.Ex{ColNameAccountPasswordOfAccountID: accountID}).
		Limit(uint(limit)).
		Offset(uint(offset)).
		Executor().
		ScanStructsContext(ctx, &downloadTaskList); err != nil {
		logger.With(zap.Error(err)).Error("failed to get download task list by account")
		return nil, 0, status.Error(codes.Internal, "failed to get download task list by account")
	}

	count, err := d.database.
		From(TableNameDownloadTask).
		Where(goqu.Ex{ColNameAccountPasswordOfAccountID: accountID}).
		CountContext(ctx)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get count of download task by account")
		return nil, 0, status.Error(codes.Internal, "failed to get count of download task by account")
	}

	return downloadTaskList, uint64(count), nil
}

func (d *downloadTaskDataAccessor) WithDatabase(database Database) DownloadTaskDataAccessor {
	return &downloadTaskDataAccessor{
		logger:   d.logger,
		database: database,
	}
}
