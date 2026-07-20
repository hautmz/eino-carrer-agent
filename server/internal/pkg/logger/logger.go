// Package logger 提供 Eino Career Agent 的日志初始化功能
// 基于 uber-go/zap 实现，支持开发模式（console）和生产模式（json）
package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// globalLogger 是全局日志实例
var globalLogger *zap.Logger

// globalSugarLogger 是全局语法糖日志实例，提供更便捷的日志调用方式
var globalSugarLogger *zap.SugaredLogger

// Init 初始化日志系统
// mode 参数决定日志格式：debug 模式使用 console 编码器，release 模式使用 json 编码器
func Init(mode string) error {
	var encoder zapcore.Encoder
	var core zapcore.Core

	if mode == "release" {
		// 生产模式：JSON 格式
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.TimeKey = "timestamp"
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		// 开发/测试模式：Console 格式，带颜色
		encoderConfig := zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 构建核心日志组件，输出到 stdout
	core = zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		zapcore.DebugLevel,
	)

	if mode == "release" {
		globalLogger = zap.New(core, zap.AddStacktrace(zapcore.ErrorLevel))
	} else {
		globalLogger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	}

	globalSugarLogger = globalLogger.Sugar()

	return nil
}

// L 返回全局 zap.Logger 实例，用于结构化日志
func L() *zap.Logger {
	if globalLogger == nil {
		fmt.Fprintln(os.Stderr, "[WARN] logger not initialized, using nop logger")
		return zap.NewNop()
	}
	return globalLogger
}

// S 返回全局 zap.SugaredLogger 实例，提供更便捷的日志调用
func S() *zap.SugaredLogger {
	if globalSugarLogger == nil {
		fmt.Fprintln(os.Stderr, "[WARN] logger not initialized, using nop logger")
		return zap.NewNop().Sugar()
	}
	return globalSugarLogger
}

// Sync 刷新日志缓冲区，应在程序退出前调用
func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}
