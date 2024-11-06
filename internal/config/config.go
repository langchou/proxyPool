package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Validator ValidatorConfig `mapstructure:"validator"`
	Crawler   CrawlerConfig   `mapstructure:"crawler"`
	Log       LogConfig       `mapstructure:"log"`
	Security  SecurityConfig  `mapstructure:"security"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type ValidatorConfig struct {
	Timeout       int    `mapstructure:"timeout"`
	CheckInterval int    `mapstructure:"check_interval"`
	TestURL       string `mapstructure:"test_url"`
}

type CrawlerConfig struct {
	Interval  int `mapstructure:"interval"`
	BatchSize int `mapstructure:"batch_size"`
}

type LogConfig struct {
	Level    string `mapstructure:"level"`
	Output   string `mapstructure:"output"`
	FilePath string `mapstructure:"file_path"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	// 基本认证
	AuthEnabled bool   `mapstructure:"auth_enabled"`
	Username    string `mapstructure:"username"`
	Password    string `mapstructure:"password"`

	// API Key认证
	APIKeyEnabled bool     `mapstructure:"api_key_enabled"`
	APIKeys       []string `mapstructure:"api_keys"`

	// 限流配置
	RateLimit        int  `mapstructure:"rate_limit"`         // 每个时间窗口最大请求数
	RateWindow       int  `mapstructure:"rate_window"`        // 时间窗口（分钟）
	BanDuration      int  `mapstructure:"ban_duration"`       // 封禁时长（小时）
	RateLimitEnabled bool `mapstructure:"rate_limit_enabled"` // 是否启用限流
}

var (
	GlobalConfig Config
)

// LoadConfig 加载配置文件
func LoadConfig(configPath string) error {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("toml")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := viper.Unmarshal(&GlobalConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

// Helper functions for getting config values
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}

func (c *Config) GetValidatorTimeout() time.Duration {
	return time.Duration(c.Validator.Timeout) * time.Second
}

func (c *Config) GetCrawlerInterval() time.Duration {
	return time.Duration(c.Crawler.Interval) * time.Minute
}

func (c *Config) GetCheckInterval() time.Duration {
	return time.Duration(c.Validator.CheckInterval) * time.Minute
}
