package middleware

import (
	"github.com/langchou/proxyPool/internal/api/response"
	"github.com/langchou/proxyPool/internal/config"
	"github.com/langchou/proxyPool/internal/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// BasicAuth 基本认证中间件
func BasicAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.GlobalConfig.Security.AuthEnabled {
			c.Next()
			return
		}

		username, password, ok := c.Request.BasicAuth()
		if !ok || username != config.GlobalConfig.Security.Username ||
			password != config.GlobalConfig.Security.Password {
			logger.Log.Warn("Invalid basic auth attempt",
				zap.String("ip", c.ClientIP()),
				zap.String("username", username))

			c.Header("WWW-Authenticate", "Basic realm=Authorization Required")
			response.Error(c, "Unauthorized")
			c.Abort()
			return
		}
		c.Next()
	}
}

// APIKeyAuth API Key 认证中间件
func APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.GlobalConfig.Security.APIKeyEnabled {
			c.Next()
			return
		}

		apiKey := c.GetHeader("X-API-Key")
		valid := false
		for _, key := range config.GlobalConfig.Security.APIKeys {
			if apiKey == key {
				valid = true
				break
			}
		}

		if !valid {
			logger.Log.Warn("Invalid API key attempt",
				zap.String("ip", c.ClientIP()),
				zap.String("api_key", apiKey))

			response.Error(c, "Invalid API key")
			c.Abort()
			return
		}
		c.Next()
	}
}
