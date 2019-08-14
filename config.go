package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewZapConfig create zap config instance
func NewZapConfig(cfg Config) zap.Config {
	var level = zapcore.DebugLevel
	if len(cfg.Level) > 0 {
		if err := level.Set(cfg.Level); err != nil {
			panic(err)
		}
	}
	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(level),
		DisableCaller:     true,
		DisableStacktrace: true,
		Development:       cfg.Development,
		Encoding:          "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}
	return config
}
