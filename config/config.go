package config

import (
	"os"
	"time"
)

// Config 应用配置
type Config struct {
	Server  ServerConfig
	DB      DBConfig
	Session SessionConfig
	CORS    CORSConfig
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
		Session: SessionConfig{
			Secret: getEnv("SESSION_SECRET", "your-secret-key-change-in-production"),
			Name:   getEnv("SESSION_NAME", "mysession"),
		},
		CORS: CORSConfig{
			AllowOrigins: []string{getEnv("CORS_ORIGIN", "https://localhost:3000")},
			MaxAge:       12 * time.Hour,
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
