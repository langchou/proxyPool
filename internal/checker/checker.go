package checker

import (
	"context"
	"time"

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

// Run 运行代理检查
func (c *Checker) Run(ctx context.Context) error {
	logger.Log.Info("Starting proxy check")

	proxies, err := c.storage.GetAll(ctx)
	if err != nil {
		logger.Log.Error("Failed to get proxies for checking", zap.Error(err))
		return err
	}

	logger.Log.Info("Retrieved proxies for checking", zap.Int("count", len(proxies)))

	for _, proxy := range proxies {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			valid, speed := c.validator.Validate(proxy)
			if !valid {
				// 移除无效代理
				logger.Log.Debug("Removing invalid proxy",
					zap.String("ip", proxy.IP),
					zap.String("port", proxy.Port))
				c.storage.Remove(ctx, proxy.IP+":"+proxy.Port)
			} else {
				// 更新代理信息
				proxy.Speed = speed
				proxy.LastCheck = time.Now()
				// 根据响应时间调整分数
				if speed < 1000 { // 小于1秒
					proxy.Score += 1
				} else {
					proxy.Score -= 1
				}
				if proxy.Score < 0 {
					proxy.Score = 0
				}
				if proxy.Score > 100 {
					proxy.Score = 100
				}
				c.storage.Save(ctx, proxy)
			}
		}
	}

	logger.Log.Info("Completed proxy check")
	return nil
}
