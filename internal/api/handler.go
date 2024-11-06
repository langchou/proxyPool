package api

import (
	"github.com/langchou/proxyPool/internal/api/response"
	"github.com/langchou/proxyPool/internal/logger"
	"github.com/langchou/proxyPool/internal/model"
	"github.com/langchou/proxyPool/internal/storage"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Handler struct {
	storage storage.Storage
}

func NewHandler(storage storage.Storage) *Handler {
	return &Handler{storage: storage}
}

// GetProxy 获取代理
// @param type: 代理类型，可选值：http,https,socks4,socks5，多个类型用逗号分隔
// @param count: 返回数量，默认1
// @param anonymous: 是否只返回高匿代理，可选值：true/false
func (h *Handler) GetProxy(c *gin.Context) {
	logger.Log.Info("Received request for proxy")

	// 解析请求参数
	proxyTypes := parseProxyTypes(c.Query("type"))
	count := parseCount(c.Query("count"), 1)
	anonymous := c.Query("anonymous") == "true"

	// 获取所有代理
	proxies, err := h.storage.GetAll(c.Request.Context())
	if err != nil {
		if err == redis.Nil {
			logger.Log.Warn("No proxies available")
			response.Success(c, []response.ProxyData{})
			return
		}
		logger.Log.Error("Failed to get proxies", zap.Error(err))
		response.Error(c, "Failed to get proxies")
		return
	}

	// 过滤代理
	filtered := filterProxies(proxies, proxyTypes, anonymous)
	if len(filtered) == 0 {
		response.Success(c, []response.ProxyData{})
		return
	}

	// 限制返回数量
	if count > len(filtered) {
		count = len(filtered)
	}

	result := filtered[:count]
	logger.Log.Info("Successfully returned proxies",
		zap.Int("requested", count),
		zap.Int("returned", len(result)))

	response.Success(c, response.ConvertProxies(result))
}

func (h *Handler) GetAllProxies(c *gin.Context) {
	logger.Log.Info("Received request for all proxies")

	// 解析请求参数
	proxyTypes := parseProxyTypes(c.Query("type"))
	anonymous := c.Query("anonymous") == "true"

	proxies, err := h.storage.GetAll(c.Request.Context())
	if err != nil {
		logger.Log.Error("Failed to get all proxies", zap.Error(err))
		response.Error(c, "Failed to get proxies")
		return
	}

	// 过滤代理
	filtered := filterProxies(proxies, proxyTypes, anonymous)

	logger.Log.Info("Successfully returned all proxies",
		zap.Int("total", len(filtered)))

	response.Success(c, response.ConvertProxies(filtered))
}

// 解析代理类型
func parseProxyTypes(typeStr string) []model.ProxyType {
	if typeStr == "" {
		return nil
	}

	types := strings.Split(typeStr, ",")
	result := make([]model.ProxyType, 0, len(types))

	for _, t := range types {
		proxyType := model.ProxyType(strings.ToLower(strings.TrimSpace(t)))
		if proxyType.IsValid() {
			result = append(result, proxyType)
		}
	}

	return result
}

// 解析数量
func parseCount(countStr string, defaultValue int) int {
	if countStr == "" {
		return defaultValue
	}

	count, err := strconv.Atoi(countStr)
	if err != nil || count < 1 {
		return defaultValue
	}
	return count
}

// 过滤代理
func filterProxies(proxies []*model.Proxy, types []model.ProxyType, anonymous bool) []*model.Proxy {
	if len(proxies) == 0 {
		return proxies
	}

	filtered := make([]*model.Proxy, 0, len(proxies))
	for _, proxy := range proxies {
		// 类型过滤
		if len(types) > 0 {
			typeMatched := false
			for _, t := range types {
				if proxy.Type == t {
					typeMatched = true
					break
				}
			}
			if !typeMatched {
				continue
			}
		}

		// 匿名性过滤
		if anonymous && !proxy.Anonymous {
			continue
		}

		filtered = append(filtered, proxy)
	}

	return filtered
}
