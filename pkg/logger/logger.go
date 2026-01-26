package logger

import (
	"time"
)

// Logger 定义日志接口，便于切换底层实现
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)

	// With 返回带有预设字段的新 Logger
	With(fields ...Field) Logger
}

// Field 日志字段
type Field struct {
	Key   string
	Value any
}

// 字段构造函数

func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

func Error(err error) Field {
	return Field{Key: "error", Value: err}
}

func Any(key string, value any) Field {
	return Field{Key: key, Value: value}
}

func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value}
}
