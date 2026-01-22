package config

import (
	"os"
	"time"
)

// Config 应用配置
type Config struct {
	Server  ServerConfig
	DB      DBConfig
	Redis   RedisConfig
	Cache   CacheConfig
	JWT     JWTConfig
	Session SessionConfig
	CORS    CORSConfig
	Log     LogConfig
}

type LogConfig struct {
	Level string // debug, info, warn, error
	IsDev bool   // 开发模式：彩色控制台输出；生产模式：JSON 格式
}

type ServerConfig struct {
	Port string
}

type DBConfig struct {
	DSN string
}

type SessionConfig struct {
	Secret string
	Name   string
}

type RedisConfig struct {
	Addr     string
	Password string
}

type CacheConfig struct {
	UserExpiration time.Duration // 用户缓存过期时间
}

type JWTConfig struct {
	SecretKey         string
	ExpireTime        time.Duration // Access Token 有效期
	RefreshExpireTime time.Duration // Refresh Token 有效期
}

type CORSConfig struct {
	AllowOrigins []string
	MaxAge       time.Duration
}

// Load 从环境变量加载配置，未设置则使用默认值
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", ":8080"),
		},
		DB: DBConfig{
			DSN: getEnv("DB_DSN", "root:root@tcp(localhost:13316)/webook"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
		},
		Cache: CacheConfig{
			UserExpiration: 15 * time.Minute,
		},
		JWT: JWTConfig{
			SecretKey:         getEnv("JWT_SECRET", "your-jwt-secret-key"),
			ExpireTime:        30 * time.Minute,   // Access Token 30 分钟
			RefreshExpireTime: 7 * 24 * time.Hour, // Refresh Token 7 天
		},
		Session: SessionConfig{
			Secret: getEnv("SESSION_SECRET", "your-secret-key-change-in-production"),
			Name:   getEnv("SESSION_NAME", "mysession"),
		},
		CORS: CORSConfig{
			AllowOrigins: []string{getEnv("CORS_ORIGIN", "https://localhost:3000")},
			MaxAge:       12 * time.Hour,
		},
		Log: LogConfig{
			Level: getEnv("LOG_LEVEL", "info"),
			IsDev: getEnv("APP_ENV", "dev") == "dev",
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
