package consumer

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/IBM/sarama"
	"github.com/nhtuan0700/GoLoad/internal/configs"
	"github.com/nhtuan0700/GoLoad/internal/utils"
	"go.uber.org/zap"
)

type HandlerFunc func(ctx context.Context, queueName string, payload []byte) error

type consumerHandler struct {
	handlerFunc      HandlerFunc
	exitSignalChanel chan os.Signal
	logger           *zap.Logger
}

func newConsumerHandler(
	handlerFunc HandlerFunc,
	exitSignalChanel chan os.Signal,
	logger *zap.Logger,
) *consumerHandler {
	return &consumerHandler{
		handlerFunc:      handlerFunc,
		exitSignalChanel: exitSignalChanel,
		logger:           logger,
	}
}

func (h consumerHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h consumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h consumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	h.logger.Info("Consumer is ready...")
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				session.Commit()
				return nil
			}
			fmt.Println(message)
			if err := h.handlerFunc(session.Context(), message.Topic, message.Value); err != nil {
				logger := utils.LoggerWithContext(session.Context(), h.logger)
				// Note:
				// - Not return error to make sure no blocking when handler failed
				// - consider to handle another way to handle the failed task such as cronjob
				logger.With(zap.Error(err)).Error("Consumer handler failed")
			}

		case <-h.exitSignalChanel:
			h.logger.Info("All messages committed")
			session.Commit()
			return nil
		}
	}
}

type Consumer interface {
	RegisterHandler(queueName string, handlerFunc HandlerFunc)
	Start(ctx context.Context) error
}

type consumer struct {
	saramaConsumer            sarama.ConsumerGroup
	logger                    *zap.Logger
	queueNameToHandlerFuncMap map[string]HandlerFunc
}

func newSaramaConfig(mqConfig configs.MQ) *sarama.Config {
	saramaConfig := sarama.NewConfig()
	saramaConfig.ClientID = mqConfig.ClientID
	saramaConfig.Metadata.Full = true
	return saramaConfig
}

func NewConsumer(
	mqConfig configs.MQ,
	logger *zap.Logger,
) (Consumer, error) {
	saramaConsumer, err := sarama.NewConsumerGroup(mqConfig.Addresses, mqConfig.ClientID, newSaramaConfig(mqConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to create sarama consumer: %w", err)
	}

	return &consumer{
		saramaConsumer:            saramaConsumer,
		logger:                    logger,
		queueNameToHandlerFuncMap: make(map[string]HandlerFunc),
	}, nil
}

func (c *consumer) RegisterHandler(queueName string, handlerFunc HandlerFunc) {
	// Each queueName(Topic) will be consumed by one handler (consumer) in consumer group [clientID]
	c.queueNameToHandlerFuncMap[queueName] = handlerFunc
}

func (c consumer) Start(ctx context.Context) error {
	logger := utils.LoggerWithContext(ctx, c.logger)

	exitSignalChanel := make(chan os.Signal, 1)
	signal.Notify(exitSignalChanel, syscall.SIGINT, syscall.SIGTERM)

	for queueName, handlerFunc := range c.queueNameToHandlerFuncMap {
		go func(queueName string, handlerFunc HandlerFunc) {
			err := c.saramaConsumer.Consume(ctx, []string{queueName}, newConsumerHandler(handlerFunc, exitSignalChanel, logger))
			if err != nil {
				logger.With(
					zap.String("queue_name", queueName),
					zap.Error(err),
				).Error("failed to consume message from queue")
			}
		}(queueName, handlerFunc)
	}

	<-exitSignalChanel
	return nil
}
