package model

import "time"

// ProxyType 代理类型
type ProxyType string

const (
	ProxyTypeHTTP   ProxyType = "http"
	ProxyTypeHTTPS  ProxyType = "https"
	ProxyTypeSOCKS4 ProxyType = "socks4"
	ProxyTypeSOCKS5 ProxyType = "socks5"
)

type Proxy struct {
	IP        string    `json:"ip"`
	Port      string    `json:"port"`
	Type      ProxyType `json:"type"`      // 代理类型
	Anonymous bool      `json:"anonymous"` // 是否高匿
	Speed     int64     `json:"speed"`     // 响应速度（毫秒）
	Score     int       `json:"score"`     // 可用性评分
	LastCheck time.Time `json:"last_check"`
}

type ProxyList []*Proxy

// IsValid 检查代理类型是否有效
func (t ProxyType) IsValid() bool {
	switch t {
	case ProxyTypeHTTP, ProxyTypeHTTPS, ProxyTypeSOCKS4, ProxyTypeSOCKS5:
		return true
	default:
		return false
	}
}
