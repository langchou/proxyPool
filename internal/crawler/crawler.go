package crawler

import (
	"context"
	"github.com/langchou/proxyPool/internal/crawler/sources"
	"github.com/langchou/proxyPool/internal/storage"
	"github.com/langchou/proxyPool/internal/validator"
	"sync"
)

type Manager struct {
	sources   []sources.Source
	storage   storage.Storage
	validator *validator.Validator
}

func NewManager(storage storage.Storage, validator *validator.Validator) *Manager {
	return &Manager{
		sources: []sources.Source{
			// sources.NewKuaidailiSource(),
			sources.NewOpenProxyListSource(),
			// 添加更多代理源
		},
		storage:   storage,
		validator: validator,
	}
}

func (m *Manager) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	var errs []error
	var mu sync.Mutex

	for _, source := range m.sources {
		wg.Add(1)
		go func(s sources.Source) {
			defer wg.Done()

			proxies, err := s.Fetch()
			if err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				return
			}

			// 验证和存储代理
			for _, proxy := range proxies {
				valid, speed := m.validator.Validate(proxy)
				if valid {
					proxy.Speed = speed
					proxy.Score = 100 // 初始分数
					if err := m.storage.Save(ctx, proxy); err != nil {
						mu.Lock()
						errs = append(errs, err)
						mu.Unlock()
					}
				}
			}
		}(source)
	}

	wg.Wait()

	// 如果有错误，返回第一个错误
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}
