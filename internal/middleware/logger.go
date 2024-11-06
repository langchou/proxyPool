package middleware

import (
	"github.com/langchou/proxyPool/internal/logger"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		logger.Log.Info("Incoming request",
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("method", c.Request.Method))

		// 处理请求
		c.Next()

		// 请求处理完成后记录
		latency := time.Since(start)
		logger.Log.Info("Request completed",
			zap.String("path", path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency))
	}
}
