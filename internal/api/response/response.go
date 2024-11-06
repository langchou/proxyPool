package response

import (
	"net/http"
	"github.com/langchou/proxyPool/internal/model"

	"github.com/gin-gonic/gin"
)

// 响应码定义
const (
	CodeSuccess       = 200
	CodeNotFound      = 404
	CodeInternalError = 500
)

// Response 通用响应结构
type Response struct {
	Code    int         `json:"code"`           // 响应码
	Message string      `json:"message"`        // 响应信息
	Data    interface{} `json:"data,omitempty"` // 数据，可选
}

// ProxyData 代理数据结构
type ProxyData struct {
	IP        string `json:"ip"`        // IP地址
	Port      string `json:"port"`      // 端口
	Type      string `json:"type"`      // 代理类型
	Anonymous bool   `json:"anonymous"` // 是否高匿
	Speed     int64  `json:"speed_ms"`  // 响应速度（毫秒）
	Score     int    `json:"score"`     // 可用性评分
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

// NotFound 未找到响应
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Response{
		Code:    CodeNotFound,
		Message: message,
	})
}

// Error 错误响应
func Error(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, Response{
		Code:    CodeInternalError,
		Message: message,
	})
}

// ConvertProxy 转换代理模型为响应数据
func ConvertProxy(proxy *model.Proxy) ProxyData {
	return ProxyData{
		IP:        proxy.IP,
		Port:      proxy.Port,
		Type:      string(proxy.Type),
		Anonymous: proxy.Anonymous,
		Speed:     proxy.Speed,
		Score:     proxy.Score,
	}
}

// ConvertProxies 转换代理列表
func ConvertProxies(proxies []*model.Proxy) []ProxyData {
	result := make([]ProxyData, len(proxies))
	for i, proxy := range proxies {
		result[i] = ConvertProxy(proxy)
	}
	return result
}
