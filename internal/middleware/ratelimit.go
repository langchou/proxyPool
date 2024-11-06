package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/langchou/proxyPool/internal/api/response"
	"github.com/langchou/proxyPool/internal/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RateLimiter struct {
	redisClient *redis.Client
	maxRequests int           // 最大请求次数
	duration    time.Duration // 时间窗口
	banDuration time.Duration // 封禁时长
}

func NewRateLimiter(redisClient *redis.Client, maxRequests int, duration, banDuration time.Duration) *RateLimiter {
	return &RateLimiter{
		redisClient: redisClient,
		maxRequests: maxRequests,
		duration:    duration,
		banDuration: banDuration,
	}
}

func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		// 检查是否被封禁
		banKey := fmt.Sprintf("ban:%s", ip)
		if banned, _ := rl.redisClient.Get(c.Request.Context(), banKey).Bool(); banned {
			logger.Log.Warn("Banned IP attempted access",
				zap.String("ip", ip),
				zap.String("path", c.Request.URL.Path))

			c.JSON(http.StatusForbidden, response.Response{
				Code:    403,
				Message: "Your IP has been banned due to excessive requests",
			})
			c.Abort()
			return
		}

		// 访问计数key
		key := fmt.Sprintf("ratelimit:%s", ip)

		// 使用 Redis 的 MULTI 命令保证原子性
		pipe := rl.redisClient.Pipeline()
		incr := pipe.Incr(c.Request.Context(), key)
		pipe.Expire(c.Request.Context(), key, rl.duration)
		_, err := pipe.Exec(c.Request.Context())

		if err != nil {
			logger.Log.Error("Redis pipeline failed", zap.Error(err))
			c.Next()
			return
		}

		count := incr.Val()

		// 检查是否超过限制
		if count > int64(rl.maxRequests) {
			// 封禁IP
			rl.banIP(c.Request.Context(), ip)

			logger.Log.Warn("IP banned due to rate limit exceeded",
				zap.String("ip", ip),
				zap.Int64("request_count", count))

			c.JSON(http.StatusTooManyRequests, response.Response{
				Code:    429,
				Message: fmt.Sprintf("Rate limit exceeded. Maximum %d requests per %v", rl.maxRequests, rl.duration),
			})
			c.Abort()
			return
		}

		// 设置剩余请求次数的header
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(int64(rl.maxRequests)-count, 10))
		c.Header("X-RateLimit-Limit", strconv.Itoa(rl.maxRequests))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(rl.duration).Unix(), 10))

		c.Next()
	}
}

func (rl *RateLimiter) banIP(ctx context.Context, ip string) {
	banKey := fmt.Sprintf("ban:%s", ip)
	rl.redisClient.Set(ctx, banKey, true, rl.banDuration)
}

// 解封IP的方法（可用于管理API）
func (rl *RateLimiter) UnbanIP(ctx context.Context, ip string) error {
	banKey := fmt.Sprintf("ban:%s", ip)
	return rl.redisClient.Del(ctx, banKey).Err()
}
