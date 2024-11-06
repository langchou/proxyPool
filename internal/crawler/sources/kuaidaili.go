package sources

import (
	"fmt"
	"net/http"
	"github.com/langchou/proxyPool/internal/logger"
	"github.com/langchou/proxyPool/internal/model"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"go.uber.org/zap"
)

type KuaidailiSource struct {
	BaseSource
}

func NewKuaidailiSource() *KuaidailiSource {
	return &KuaidailiSource{
		BaseSource: BaseSource{name: "kuaidaili"},
	}
}

func (s *KuaidailiSource) Fetch() ([]*model.Proxy, error) {
	logger.Log.Info("Starting to fetch proxies from kuaidaili")
	proxies := make([]*model.Proxy, 0)

	urls := map[string]string{
		"http":  "https://www.kuaidaili.com/free/intr/",
		"https": "https://www.kuaidaili.com/free/inha/",
		"socks": "https://www.kuaidaili.com/ops/proxylist/",
	}

	for proxyType, baseURL := range urls {
		logger.Log.Debug("Fetching proxy type", zap.String("type", proxyType))
		for i := 1; i <= 3; i++ {
			url := fmt.Sprintf("%s%d/", baseURL, i)
			newProxies, err := s.fetchPage(url, proxyType)
			if err != nil {
				logger.Log.Error("Failed to fetch page",
					zap.String("url", url),
					zap.Error(err))
				continue
			}
			logger.Log.Debug("Fetched proxies from page",
				zap.String("url", url),
				zap.Int("count", len(newProxies)))
			proxies = append(proxies, newProxies...)
			time.Sleep(1 * time.Second)
		}
	}

	logger.Log.Info("Finished fetching proxies",
		zap.String("source", s.Name()),
		zap.Int("total", len(proxies)))
	return proxies, nil
}

func (s *KuaidailiSource) fetchPage(url, pageType string) ([]*model.Proxy, error) {
	proxies := make([]*model.Proxy, 0)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	doc.Find("table tbody tr").Each(func(i int, selection *goquery.Selection) {
		ip := strings.TrimSpace(selection.Find("td[data-title='IP']").Text())
		port := strings.TrimSpace(selection.Find("td[data-title='PORT']").Text())
		typeStr := strings.ToLower(strings.TrimSpace(selection.Find("td[data-title='类型']").Text()))
		anonymous := strings.Contains(strings.ToLower(selection.Find("td[data-title='匿名度']").Text()), "高匿")

		if ip != "" && port != "" {
			proxyType := s.parseProxyType(typeStr)
			if proxyType.IsValid() {
				proxies = append(proxies, &model.Proxy{
					IP:        ip,
					Port:      port,
					Type:      proxyType,
					Anonymous: anonymous,
					LastCheck: time.Now(),
				})
			}
		}
	})

	return proxies, nil
}

func (s *KuaidailiSource) parseProxyType(typeStr string) model.ProxyType {
	switch strings.ToLower(typeStr) {
	case "http":
		return model.ProxyTypeHTTP
	case "https":
		return model.ProxyTypeHTTPS
	case "socks4":
		return model.ProxyTypeSOCKS4
	case "socks5":
		return model.ProxyTypeSOCKS5
	default:
		return ""
	}
}
