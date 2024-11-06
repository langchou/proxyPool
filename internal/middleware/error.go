package middleware

import (
	"net/http"
	"github.com/langchou/proxyPool/internal/api/response"
	"github.com/langchou/proxyPool/internal/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ErrorHandler 处理所有未捕获的路由和错误
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() // 处理请求

		// 如果没有路由匹配，返回 404
		if c.Writer.Status() == http.StatusNotFound {
			logger.Log.Warn("Route not found",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method))

			c.JSON(http.StatusNotFound, response.Response{
				Code:    response.CodeNotFound,
				Message: "Route not found",
			})
			return
		}

		// 处理其他错误状态码
		if c.Writer.Status() >= http.StatusBadRequest {
			logger.Log.Error("Request error",
				zap.String("path", c.Request.URL.Path),
				zap.Int("status", c.Writer.Status()))

			if !c.Writer.Written() {
				c.JSON(c.Writer.Status(), response.Response{
					Code:    c.Writer.Status(),
					Message: http.StatusText(c.Writer.Status()),
				})
			}
		}
	}
}
