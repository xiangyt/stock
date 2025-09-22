package config

import (
	"time"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	App       AppConfig       `mapstructure:"app"`
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Log       LogConfig       `mapstructure:"log"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	Metrics   MetricsConfig   `mapstructure:"metrics"`
	CORS      CORSConfig      `mapstructure:"cors"`
	Notify    NotifyConfig    `mapstructure:"notify"`
}

// AppConfig 应用配置
type AppConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
	Env     string `mapstructure:"env"`
	Port    int    `mapstructure:"port"`
	Debug   bool   `mapstructure:"debug"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	ReadTimeout    time.Duration `mapstructure:"read_timeout"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout"`
	IdleTimeout    time.Duration `mapstructure:"idle_timeout"`
	MaxHeaderBytes int           `mapstructure:"max_header_bytes"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver          string        `mapstructure:"driver"`
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Name            string        `mapstructure:"name"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	File       string `mapstructure:"file"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
	Issuer      string `mapstructure:"issuer"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled bool `mapstructure:"enabled"`
	RPS     int  `mapstructure:"rps"`
	Burst   int  `mapstructure:"burst"`
}

// MetricsConfig 监控配置
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    int    `mapstructure:"port"`
	Path    string `mapstructure:"path"`
}

// CORSConfig CORS配置
type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	ExposedHeaders   []string `mapstructure:"exposed_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"`
}

// NotifyConfig 通知配置
type NotifyConfig struct {
	DingTalk *DingTalkConfig `mapstructure:"dingtalk"`
	WeWork   *WeWorkConfig   `mapstructure:"wework"`
}

// DingTalkConfig 钉钉机器人配置
type DingTalkConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Webhook string `mapstructure:"webhook"`
	Secret  string `mapstructure:"secret"`
}

// WeWorkConfig 企微机器人配置
type WeWorkConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Webhook string `mapstructure:"webhook"`
}

// Load 加载配置
func Load() (*Config, error) {
	viper.SetConfigName("app")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// 设置环境变量前缀
	viper.SetEnvPrefix("STOCK")
	viper.AutomaticEnv()

	// 设置默认值
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func setDefaults() {
	// App defaults
	viper.SetDefault("app.name", "智能选股系统")
	viper.SetDefault("app.version", "1.0.0")
	viper.SetDefault("app.env", "development")
	viper.SetDefault("app.port", 8080)
	viper.SetDefault("app.debug", true)

	// Server defaults
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.idle_timeout", "60s")
	viper.SetDefault("server.max_header_bytes", 1048576)

	// Database defaults
	viper.SetDefault("database.driver", "mysql")
	viper.SetDefault("database.host", "192.168.1.238")
	viper.SetDefault("database.port", 3306)
	viper.SetDefault("database.name", "stock")
	viper.SetDefault("database.user", "root")
	viper.SetDefault("database.password", "123456")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_open_conns", 500)
	viper.SetDefault("database.max_idle_conns", 100)
	viper.SetDefault("database.conn_max_lifetime", "300s")

	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)
	viper.SetDefault("redis.min_idle_conns", 5)

	// Log defaults
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("log.file", "logs/app.log")
	viper.SetDefault("log.max_size", 100)
	viper.SetDefault("log.max_backups", 7)
	viper.SetDefault("log.max_age", 30)
	viper.SetDefault("log.compress", true)

	// JWT defaults
	viper.SetDefault("jwt.secret", "your_jwt_secret_here")
	viper.SetDefault("jwt.expire_hours", 24)
	viper.SetDefault("jwt.issuer", "stock")

	// Rate limit defaults
	viper.SetDefault("rate_limit.enabled", true)
	viper.SetDefault("rate_limit.rps", 100)
	viper.SetDefault("rate_limit.burst", 200)

	// Metrics defaults
	viper.SetDefault("metrics.enabled", true)
	viper.SetDefault("metrics.port", 9090)
	viper.SetDefault("metrics.path", "/metrics")

	// CORS defaults
	viper.SetDefault("cors.allowed_origins", []string{"*"})
	viper.SetDefault("cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	viper.SetDefault("cors.allowed_headers", []string{"*"})
	viper.SetDefault("cors.exposed_headers", []string{})
	viper.SetDefault("cors.allow_credentials", true)
	viper.SetDefault("cors.max_age", 86400)
}
