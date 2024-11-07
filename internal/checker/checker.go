package checker

import (
	"context"
	"fmt"

	"github.com/langchou/proxyPool/internal/logger"
	"github.com/langchou/proxyPool/internal/storage"
	"github.com/langchou/proxyPool/internal/validator"
	"go.uber.org/zap"
)

type Checker struct {
	storage   storage.Storage
	validator *validator.Validator
}

func NewChecker(storage storage.Storage, validator *validator.Validator) *Checker {
	return &Checker{
		storage:   storage,
		validator: validator,
	}
}

func (c *Checker) Run(ctx context.Context) error {
	logger.Log.Info("Starting to check existing proxies")

	// 从存储中获取所有代理
	proxies, err := c.storage.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to get proxies from storage: %w", err)
	}

	logger.Log.Info("Retrieved proxies for checking", zap.Int("count", len(proxies)))

	for _, proxy := range proxies {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			logger.Log.Debug("Checking proxy",
				zap.String("ip", proxy.IP),
				zap.String("port", proxy.Port))

			// 验证代理
			valid, speed := c.validator.Validate(proxy)
			if valid {
				proxy.Speed = speed
				// 验证成功，更新代理信息
				if err := c.storage.Save(ctx, proxy); err != nil {
					logger.Log.Error("Failed to update proxy",
						zap.String("ip", proxy.IP),
						zap.String("port", proxy.Port),
						zap.Error(err))
				}
				logger.Log.Info("Proxy check passed",
					zap.String("ip", proxy.IP),
					zap.String("port", proxy.Port),
					zap.Int64("speed", speed))
			} else {
				// 验证失败，从存储中删除
				key := proxy.IP + ":" + proxy.Port
				if err := c.storage.Remove(ctx, key); err != nil {
					logger.Log.Error("Failed to remove invalid proxy",
						zap.String("ip", proxy.IP),
						zap.String("port", proxy.Port),
						zap.Error(err))
				}
				logger.Log.Info("Removed invalid proxy",
					zap.String("ip", proxy.IP),
					zap.String("port", proxy.Port))
			}
		}
	}

	logger.Log.Info("Finished checking all proxies")
	return nil
}
