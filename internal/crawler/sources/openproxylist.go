package sources

import (
	"bufio"
	"fmt"
	"net/http"
	"github.com/langchou/proxyPool/internal/logger"
	"github.com/langchou/proxyPool/internal/model"
	"strings"
	"time"

	"go.uber.org/zap"
)

type OpenProxyListSource struct {
	BaseSource
}

func NewOpenProxyListSource() *OpenProxyListSource {
	return &OpenProxyListSource{
		BaseSource: BaseSource{name: "openproxylist"},
	}
}

func (s *OpenProxyListSource) Fetch() ([]*model.Proxy, error) {
	logger.Log.Info("Starting to fetch proxies from openproxylist")
	proxies := make([]*model.Proxy, 0)

	// 定义要爬取的URL
	urls := map[string]string{
		"https":  "https://raw.githubusercontent.com/roosterkid/openproxylist/main/HTTPS.txt",
		"socks5": "https://raw.githubusercontent.com/roosterkid/openproxylist/main/SOCKS5.txt",
	}

	for proxyType, url := range urls {
		logger.Log.Debug("Fetching proxy type", zap.String("type", proxyType))
		newProxies, err := s.fetchList(url, proxyType)
		if err != nil {
			logger.Log.Error("Failed to fetch list",
				zap.String("url", url),
				zap.Error(err))
			continue
		}
		proxies = append(proxies, newProxies...)
		// 避免请求过快
		time.Sleep(2 * time.Second)
	}

	logger.Log.Info("Finished fetching proxies",
		zap.String("source", s.Name()),
		zap.Int("total", len(proxies)))
	return proxies, nil
}

func (s *OpenProxyListSource) fetchList(url, proxyType string) ([]*model.Proxy, error) {
	proxies := make([]*model.Proxy, 0)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		// 跳过注释和空行
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" ||
			strings.HasPrefix(line, "Support") || strings.HasPrefix(line, "BTC") ||
			strings.HasPrefix(line, "ETH") || strings.HasPrefix(line, "LTC") ||
			strings.HasPrefix(line, "Doge") || strings.HasPrefix(line, "Format") ||
			strings.HasPrefix(line, "Website") || !strings.Contains(line, "]") {
			continue
		}

		// 解析代理信息
		// 格式: 🇨🇦 67.43.228.250:14395 370ms CA [GloboTech Communications]
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		// 获取IP:PORT部分
		ipPort := strings.Split(parts[1], ":")
		if len(ipPort) != 2 {
			continue
		}

		// 解析响应时间
		speedStr := strings.TrimSuffix(parts[2], "ms")
		var speed int64
		if _, err := fmt.Sscanf(speedStr, "%d", &speed); err != nil {
			speed = 0
		}

		proxy := &model.Proxy{
			IP:        ipPort[0],
			Port:      ipPort[1],
			Type:      s.getProxyType(proxyType),
			Speed:     speed,
			Anonymous: true, // 这些代理通常都是高匿的
			LastCheck: time.Now(),
		}
		proxies = append(proxies, proxy)
	}

	logger.Log.Debug("Fetched proxies from list",
		zap.String("url", url),
		zap.Int("count", len(proxies)))

	return proxies, nil
}

func (s *OpenProxyListSource) getProxyType(typeStr string) model.ProxyType {
	switch strings.ToLower(typeStr) {
	case "https":
		return model.ProxyTypeHTTPS
	case "socks5":
		return model.ProxyTypeSOCKS5
	default:
		return model.ProxyTypeHTTP
	}
}
