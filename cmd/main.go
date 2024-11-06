package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/langchou/proxyPool/internal/api"
	"github.com/langchou/proxyPool/internal/config"
	"github.com/langchou/proxyPool/internal/crawler"
	"github.com/langchou/proxyPool/internal/logger"
	"github.com/langchou/proxyPool/internal/middleware"
	"github.com/langchou/proxyPool/internal/storage"
	"github.com/langchou/proxyPool/internal/validator"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "data/config.toml", "path to config file")
	flag.Parse()

	// 加载配置文件
	if err := config.LoadConfig(*configPath); err != nil {
		panic(err)
	}

	// 初始化日志
	if err := logger.Init(
		config.GlobalConfig.Log.Level,
		config.GlobalConfig.Log.Output,
		config.GlobalConfig.Log.FilePath,
	); err != nil {
		panic(fmt.Sprintf("Initialize logger failed: %v", err))
	}
	defer logger.Log.Sync()

	logger.Log.Info("Starting proxy pool service...")

	// 设置gin模式
	gin.SetMode(config.GlobalConfig.Server.Mode)

	// 初始化存储
	store := storage.NewRedisStorage(
		config.GlobalConfig.GetRedisAddr(),
		config.GlobalConfig.Redis.Password,
		config.GlobalConfig.Redis.DB,
	)
	logger.Log.Info("Redis storage initialized")

	// 初始化验证器
	validator := validator.NewValidator(config.GlobalConfig.GetValidatorTimeout())
	logger.Log.Info("Proxy validator initialized")

	// 初始化爬虫管理器
	crawler := crawler.NewManager(store, validator)
	logger.Log.Info("Crawler manager initialized")

	// 启动后台爬虫任务
	go func() {
		// 首次执行延迟1秒，让API服务器先启动
		time.Sleep(time.Second)

		// 创建一个用于取消的context
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// 首次执行爬虫任务
		logger.Log.Info("Running initial proxy crawling...")
		if err := crawler.Run(ctx); err != nil {
			logger.Log.Error("Initial crawling failed", zap.Error(err))
		}

		// 定时执行爬虫任务
		ticker := time.NewTicker(config.GlobalConfig.GetCrawlerInterval())
		defer ticker.Stop()

		for range ticker.C {
			logger.Log.Info("Starting scheduled proxy crawling...")
			if err := crawler.Run(ctx); err != nil {
				logger.Log.Error("Scheduled crawling failed", zap.Error(err))
			}
		}
	}()

	// 启动API服务
	r := gin.New()
	r.Use(middleware.Logger())
	r.Use(middleware.ErrorHandler())

	// 初始化限流器（如果启用）
	if config.GlobalConfig.Security.RateLimitEnabled {
		rateLimiter := middleware.NewRateLimiter(
			store.GetRedisClient(),
			config.GlobalConfig.Security.RateLimit,
			time.Duration(config.GlobalConfig.Security.RateWindow)*time.Minute,
			time.Duration(config.GlobalConfig.Security.BanDuration)*time.Hour,
		)
		r.Use(rateLimiter.RateLimit())
	}

	r.Use(middleware.BasicAuth())  // 基本认证
	r.Use(middleware.APIKeyAuth()) // API Key 认证
	r.Use(gin.Recovery())

	handler := api.NewHandler(store)
	r.GET("/proxy", handler.GetProxy)
	r.GET("/proxies", handler.GetAllProxies)

	// 添加健康检查接口
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	addr := fmt.Sprintf(":%d", config.GlobalConfig.Server.Port)
	logger.Log.Info("Starting HTTP server", zap.String("addr", addr))
	if err := r.Run(addr); err != nil {
		logger.Log.Fatal("Failed to start HTTP server", zap.Error(err))
	}
}
