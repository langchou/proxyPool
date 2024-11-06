package storage

import (
	"context"
	"encoding/json"
	"math/rand"
	"time"

	"github.com/langchou/proxyPool/internal/logger"
	"github.com/langchou/proxyPool/internal/model"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Storage interface {
	Save(context.Context, *model.Proxy) error
	GetAll(context.Context) ([]*model.Proxy, error)
	GetRandom(context.Context) (*model.Proxy, error)
	Remove(context.Context, string) error
	UpdateScore(context.Context, string, int) error
}

type RedisStorage struct {
	client *redis.Client
}

func NewRedisStorage(addr, password string, db int) *RedisStorage {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisStorage{client: client}
}

// 实现 Storage 接口的方法...

func (s *RedisStorage) Save(ctx context.Context, proxy *model.Proxy) error {
	key := "proxy:" + proxy.IP + ":" + proxy.Port
	logger.Log.Debug("Saving proxy to Redis", zap.String("key", key))

	// 将代理对象序列化为 JSON
	data, err := json.Marshal(proxy)
	if err != nil {
		logger.Log.Error("Failed to marshal proxy", zap.Error(err))
		return err
	}

	// 使用 SET 命令而不是 HSET
	err = s.client.Set(ctx, key, data, 24*time.Hour).Err() // 设置 24 小时过期
	if err != nil {
		logger.Log.Error("Failed to save proxy", zap.String("key", key), zap.Error(err))
	}
	return err
}

func (s *RedisStorage) GetAll(ctx context.Context) ([]*model.Proxy, error) {
	keys, err := s.client.Keys(ctx, "proxy:*").Result()
	if err != nil {
		return nil, err
	}

	proxies := make([]*model.Proxy, 0, len(keys))
	for _, key := range keys {
		// 使用 GET 命令获取 JSON 数据
		data, err := s.client.Get(ctx, key).Result()
		if err != nil {
			if err != redis.Nil {
				logger.Log.Error("Failed to get proxy", zap.String("key", key), zap.Error(err))
			}
			continue
		}

		var proxy model.Proxy
		if err := json.Unmarshal([]byte(data), &proxy); err != nil {
			logger.Log.Error("Failed to unmarshal proxy", zap.String("key", key), zap.Error(err))
			continue
		}

		proxies = append(proxies, &proxy)
	}

	return proxies, nil
}

func (s *RedisStorage) GetRandom(ctx context.Context) (*model.Proxy, error) {
	keys, err := s.client.Keys(ctx, "proxy:*").Result()
	if err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return nil, redis.Nil
	}

	// 随机选择一个代理
	key := keys[rand.Intn(len(keys))]

	// 使用 GET 命令获取 JSON 数据
	data, err := s.client.Get(ctx, key).Result()
	if err != nil {
		logger.Log.Error("Failed to get random proxy", zap.String("key", key), zap.Error(err))
		return nil, err
	}

	var proxy model.Proxy
	if err := json.Unmarshal([]byte(data), &proxy); err != nil {
		logger.Log.Error("Failed to unmarshal proxy", zap.String("key", key), zap.Error(err))
		return nil, err
	}

	return &proxy, nil
}

func (s *RedisStorage) Remove(ctx context.Context, key string) error {
	return s.client.Del(ctx, "proxy:"+key).Err()
}

func (s *RedisStorage) UpdateScore(ctx context.Context, key string, score int) error {
	fullKey := "proxy:" + key

	// 先获取现有数据
	data, err := s.client.Get(ctx, fullKey).Result()
	if err != nil {
		return err
	}

	var proxy model.Proxy
	if err := json.Unmarshal([]byte(data), &proxy); err != nil {
		return err
	}

	// 更新分数
	proxy.Score = score

	// 重新保存
	return s.Save(ctx, &proxy)
}

// 在 RedisStorage 结构体中添加获取客户端的方法
func (s *RedisStorage) GetRedisClient() *redis.Client {
	return s.client
}
