package logger

// 提供日志初始化功能

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// InitLogger 初始化全局日志记录器
func InitLogger() {
	var cfg zap.Config

	cfg = zap.NewDevelopmentConfig()
	cfg.EncoderConfig.StacktraceKey = ""
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	// 构建日志器并处理可能的错误
	var err error
	logger, err := cfg.Build()
	if err != nil {
		zap.L().Fatal("构建日志器失败", zap.Error(err))
	}

	// 替换全局日志器
	zap.ReplaceGlobals(logger)
}
