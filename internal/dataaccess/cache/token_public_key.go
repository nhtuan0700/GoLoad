package cache

import (
	"context"
	"errors"
	"fmt"

	"github.com/nhtuan0700/GoLoad/internal/utils"
	"go.uber.org/zap"
)

type TokenPublicKeyCache interface {
	Get(ctx context.Context, id uint64) (string, error)
	Set(ctx context.Context, id uint64, data string) error
}

type tokenPublicKeyCache struct {
	client Client
	logger *zap.Logger
}

func NewTokenPublicKeyCache(
	client Client,
	logger *zap.Logger,
) TokenPublicKeyCache {
	return &tokenPublicKeyCache{
		client: client,
		logger: logger,
	}
}

func (c *tokenPublicKeyCache) getTokenPublicKeyCacheKey(id uint64) string {
	return fmt.Sprintf("token_public_key:%d", id)
}

func (c *tokenPublicKeyCache) Get(ctx context.Context, id uint64) (string, error) {
	logger := utils.LoggerWithContext(ctx, c.logger).With(zap.Uint64("id", id))

	cacheKey := c.getTokenPublicKeyCacheKey(id)
	cacheEntry, err := c.client.Get(ctx, cacheKey)
	if err != nil {
		if errors.Is(err, ErrCacheMiss) {
			return "", ErrCacheMiss
		}
		logger.With(zap.Error(err)).Error("failed to get token public key cache")
		return "", err
	}

	publicKey, ok := cacheEntry.(string)
	if !ok {
		logger.Error("cache entry is not type string")
		return "", nil
	}

	return publicKey, nil
}

func (c *tokenPublicKeyCache) Set(ctx context.Context, id uint64, data string) error {
	logger := utils.LoggerWithContext(ctx, c.logger).With(zap.Uint64("id", id))

	cacheKey := c.getTokenPublicKeyCacheKey(id)
	if err := c.client.Set(ctx, cacheKey, data, 0); err != nil {
		logger.With(zap.Error(err)).Error("failed to insert token public key into cache")
		return err
	}

	return nil
}
