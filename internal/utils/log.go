package utils

import (
	"context"

	"github.com/nhtuan0700/GoLoad/internal/configs"
	"go.uber.org/zap"
)

func getZapLoggerLevel(level string) zap.AtomicLevel {
	switch level {
	case "debug":
		return zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		return zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		return zap.NewAtomicLevelAt(zap.ErrorLevel)
	case "panic":
		return zap.NewAtomicLevelAt(zap.PanicLevel)
	}
	return zap.NewAtomicLevelAt(zap.InfoLevel)
}

func InitializeLogger(logConfig configs.Log) (*zap.Logger, func(), error) {
	zapLoggerConfig := zap.NewProductionConfig()
	zapLoggerConfig.Level = getZapLoggerLevel(logConfig.Level)

	logger, err := zapLoggerConfig.Build()
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		_ = logger.Sync()
	}

	return logger, cleanup, nil
}

func LoggerWithContext(_ context.Context, logger *zap.Logger) *zap.Logger {
	// TODO: Add request id to context

	return logger
}
