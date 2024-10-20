package consumers

import (
	"context"
	"encoding/json"

	"github.com/nhtuan0700/GoLoad/internal/dataaccess/mq/consumer"
	"github.com/nhtuan0700/GoLoad/internal/dataaccess/mq/producer"
	"go.uber.org/zap"
)

type Root interface {
	Start(ctx context.Context) error
}

type root struct {
	mqConsumer                 consumer.Consumer
	downloadTaskCreatedHandler DownloadTaskCreated
	logger                     *zap.Logger
}

func NewRoot(
	mqConsumer consumer.Consumer,
	downloadTaskCreatedHandler DownloadTaskCreated,
	logger *zap.Logger,
) Root {
	return &root{
		mqConsumer:                 mqConsumer,
		downloadTaskCreatedHandler: downloadTaskCreatedHandler,
		logger:                     logger,
	}
}

func (r root) Start(ctx context.Context) error {
	r.mqConsumer.RegisterHandler(
		producer.MessageQueueDownloadTaskCreated,
		func(ctx context.Context, _ string, payload []byte) error {
			var event producer.DownloadTaskCreated
			if err := json.Unmarshal(payload, &event); err != nil {
				return err
			}

			return r.downloadTaskCreatedHandler.Handle(ctx, event)
		},
	)

	r.logger.Info("Starting root consumer server")

	return r.mqConsumer.Start(ctx)
}
