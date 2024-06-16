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
	DownloadType   int16  `db:"download_type"`
	URL            string `db:"url"`
	DownloadStatus int16  `db:"download_status"`
	Metadata       JSON   `db:"metadata"`
}

type DownloadTaskDataAccessor interface {
	CreateDownloadTask(ctx context.Context, downloadTask DownloadTask) (uint64, error)
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

	metaData, err := downloadTask.Metadata.Value()
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to unmarshal metadata")
		return 0, status.Error(codes.Internal, "failed to unmarshal metadata")
	}
	result, err := d.database.
		Insert(TableNameDownloadTask).
		Rows(goqu.Record{
			ColNameDownloadTaskOfAccountID:    downloadTask.OfAccountID,
			ColNameDownloadTaskDownloadType:   downloadTask.DownloadType,
			ColNameDownloadTaskURL:            downloadTask.DownloadStatus,
			ColNameDownloadTaskDownloadStatus: downloadTask.DownloadStatus,
			ColNameDownloadTaskMetadata:       metaData,
		}).
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

func (d *downloadTaskDataAccessor) WithDatabase(database Database) DownloadTaskDataAccessor {
	return &downloadTaskDataAccessor{
		logger:   d.logger,
		database: database,
	}
}
