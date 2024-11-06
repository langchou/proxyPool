package sources

import "github.com/langchou/proxyPool/internal/model"

type Source interface {
	Name() string
	Fetch() ([]*model.Proxy, error)
}

type BaseSource struct {
	name string
}

func (s *BaseSource) Name() string {
	return s.name
}
