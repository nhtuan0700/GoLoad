package producer

import (
	"context"

	"github.com/IBM/sarama"
	"github.com/nhtuan0700/GoLoad/internal/configs"
	"github.com/nhtuan0700/GoLoad/internal/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Client interface {
	Produce(ctx context.Context, queueName string, payload []byte) error
}

type client struct {
	samaraSyncProducer sarama.SyncProducer
	logger             *zap.Logger
}

func newSaramaConfig(mqConfig configs.MQ) *sarama.Config {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.MaxVersion
	saramaConfig.Producer.Retry.Max = 1
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.ClientID = mqConfig.ClientID
	saramaConfig.Metadata.Full = true

	return saramaConfig
}

func NewClient(
	mqConfig configs.MQ,
	logger *zap.Logger,
) (Client, error) {
	saramaProducer, err := sarama.NewSyncProducer(mqConfig.Addresses, newSaramaConfig(mqConfig))
	if err != nil {
		return nil, err
	}

	return &client{
		samaraSyncProducer: saramaProducer,
		logger:             logger,
	}, nil
}

func (c client) Produce(ctx context.Context, queueName string, payload []byte) error {
	logger := utils.LoggerWithContext(ctx, c.logger).
		With(zap.String("queue_name", queueName)).
		With(zap.ByteString("payload", payload))

	_, _, err := c.samaraSyncProducer.SendMessage(&sarama.ProducerMessage{
		Topic: queueName,
		Value: sarama.ByteEncoder(payload),
	})
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to produce message")
		return status.Error(codes.Internal, "failed to produce message")
	}

	return nil
}
