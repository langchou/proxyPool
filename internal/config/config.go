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
