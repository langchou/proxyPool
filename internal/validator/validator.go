package validator

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"github.com/langchou/proxyPool/internal/logger"
	"github.com/langchou/proxyPool/internal/model"
	"time"

	"go.uber.org/zap"
	"golang.org/x/net/proxy"
)

type Validator struct {
	timeout time.Duration
}

// IPInfo 响应结构
type IPInfo struct {
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	Org      string `json:"org"`
}

func NewValidator(timeout time.Duration) *Validator {
	return &Validator{timeout: timeout}
}

func (v *Validator) Validate(p *model.Proxy) (bool, int64) {
	logger.Log.Debug("Validating proxy",
		zap.String("ip", p.IP),
		zap.String("port", p.Port),
		zap.String("type", string(p.Type)))

	var client *http.Client
	var err error

	switch p.Type {
	case model.ProxyTypeHTTP, model.ProxyTypeHTTPS:
		client, err = v.createHTTPClient(p)
	case model.ProxyTypeSOCKS4, model.ProxyTypeSOCKS5:
		client, err = v.createSocksClient(p)
	default:
		logger.Log.Warn("Invalid proxy type", zap.String("type", string(p.Type)))
		return false, 0
	}

	if err != nil {
		logger.Log.Error("Failed to create HTTP client", zap.Error(err))
		return false, 0
	}

	start := time.Now()

	resp, err := client.Get("http://ipinfo.io/json")
	if err != nil {
		logger.Log.Debug("Proxy validation failed",
			zap.String("ip", p.IP),
			zap.Error(err))
		return false, 0
	}
	defer resp.Body.Close()

	speed := time.Since(start).Milliseconds()
	logger.Log.Debug("Proxy response time",
		zap.String("ip", p.IP),
		zap.Int64("speed_ms", speed))

	if resp.StatusCode != http.StatusOK {
		logger.Log.Debug("Proxy returned non-200 status",
			zap.String("ip", p.IP),
			zap.Int("status", resp.StatusCode))
		return false, speed
	}

	// 读取并解析响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, speed
	}

	var ipInfo IPInfo
	if err := json.Unmarshal(body, &ipInfo); err != nil {
		return false, speed
	}

	// 验证返回的IP是否与代理IP匹配
	// 注意：某些代理可能会返回不同的IP（级联代理），所以这里只验证是否成功获取到了IP信息
	return ipInfo.IP != "", speed
}

func (v *Validator) createHTTPClient(p *model.Proxy) (*http.Client, error) {
	proxyURL := fmt.Sprintf("%s://%s:%s", p.Type, p.IP, p.Port)
	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(parsedURL),
		},
		Timeout: v.timeout,
	}, nil
}

func (v *Validator) createSocksClient(p *model.Proxy) (*http.Client, error) {
	dialer, err := proxy.SOCKS5("tcp", p.IP+":"+p.Port, nil, proxy.Direct)
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Transport: &http.Transport{
			Dial: dialer.Dial,
		},
		Timeout: v.timeout,
	}, nil
}
