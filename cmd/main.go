package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/langchou/proxyPool/internal/api"
	"github.com/langchou/proxyPool/internal/checker"
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

	// 初始化检查器
	checker := checker.NewChecker(store, validator)
	logger.Log.Info("Proxy checker initialized")

	// 启动后台爬虫任务
	go func() {
		logger.Log.Info("Starting crawler goroutine")

		// 创建一个用于取消的context
		ctx := context.Background()

		// 首次执行爬虫任务
		logger.Log.Info("Running initial proxy crawling...")
		if err := crawler.Run(ctx); err != nil {
			logger.Log.Error("Initial crawling failed", zap.Error(err))
		}

		// 创建定时器
		interval := config.GlobalConfig.GetCrawlerInterval()
		logger.Log.Info("Setting up crawler timer", zap.Duration("interval", interval))
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				logger.Log.Info("Starting scheduled proxy crawling...")
				if err := crawler.Run(ctx); err != nil {
					logger.Log.Error("Scheduled crawling failed", zap.Error(err))
				}
			}
		}
	}()

	// 启动定时检查任务
	go func() {
		logger.Log.Info("Starting checker goroutine")

		ctx := context.Background()

		// 等待30秒后开始第一次检查，给爬虫一些时间先获取代理
		logger.Log.Info("Waiting 30 seconds before first check...")
		time.Sleep(30 * time.Second)

		// 执行第一次检查
		logger.Log.Info("Running initial proxy check...")
		if err := checker.Run(ctx); err != nil {
			logger.Log.Error("Initial proxy check failed", zap.Error(err))
		}

		// 创建定时器
		interval := config.GlobalConfig.GetCheckInterval()
		logger.Log.Info("Setting up checker timer", zap.Duration("interval", interval))
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				logger.Log.Info("Starting scheduled proxy check...")
				if err := checker.Run(ctx); err != nil {
					logger.Log.Error("Scheduled proxy check failed", zap.Error(err))
				}
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
