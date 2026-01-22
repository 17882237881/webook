package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapLogger 基于 Zap 的 Logger 实现
type ZapLogger struct {
	logger *zap.Logger
}

// NewZapLogger 创建 ZapLogger 实例
// level: debug, info, warn, error
// isDev: 开发模式使用易读格式，生产模式使用 JSON 格式
func NewZapLogger(level string, isDev bool) *ZapLogger {
	// 解析日志级别
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	// 配置编码器
	var encoder zapcore.Encoder
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	if isDev {
		// 开发模式：彩色控制台输出
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		// 生产模式：JSON 格式
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// 创建核心
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		zapLevel,
	)

	// 创建 logger
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return &ZapLogger{logger: zapLogger}
}

// toZapFields 转换 Field 为 zap.Field
func toZapFields(fields []Field) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for _, f := range fields {
		zapFields = append(zapFields, zap.Any(f.Key, f.Value))
	}
	return zapFields
}

func (l *ZapLogger) Debug(msg string, fields ...Field) {
	l.logger.Debug(msg, toZapFields(fields)...)
}

func (l *ZapLogger) Info(msg string, fields ...Field) {
	l.logger.Info(msg, toZapFields(fields)...)
}

func (l *ZapLogger) Warn(msg string, fields ...Field) {
	l.logger.Warn(msg, toZapFields(fields)...)
}

func (l *ZapLogger) Error(msg string, fields ...Field) {
	l.logger.Error(msg, toZapFields(fields)...)
}

func (l *ZapLogger) With(fields ...Field) Logger {
	return &ZapLogger{
		logger: l.logger.With(toZapFields(fields)...),
	}
}

// Sync 刷新缓冲区，应在程序退出前调用
func (l *ZapLogger) Sync() error {
	return l.logger.Sync()
}
