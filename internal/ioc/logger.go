package ioc

import (
	"webook/config"
	"webook/pkg/logger"
)

// NewLogger 创建 Logger 实例
func NewLogger(cfg *config.Config) logger.Logger {
	return logger.NewZapLogger(cfg.Log.Level, cfg.Log.IsDev)
}
