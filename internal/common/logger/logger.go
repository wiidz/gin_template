package logger

import (
	"github.com/wiidz/goutil/helpers/loggerHelper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var L *zap.Logger

func Init(env string) {
	level := zapcore.InfoLevel
	if env == "dev" || env == "development" {
		level = zapcore.DebugLevel
	} else if env == "prod" || env == "production" {
		level = zapcore.InfoLevel
	}

	helper, err := loggerHelper.NewLoggerHelper(&loggerHelper.Config{
		Filename:        "", // default console only; wire to file later if needed
		Level:           level,
		Json:            false,
		SyncToConsole:   true,
		EncodeTime:      loggerHelper.MyTimeEncoder,
		IsFullPath:      false,
		AddCaller:       false,
		ShowFileAndLine: false,
		MaxSize:         10,
		MaxBackups:      3,
		MaxAge:          28,
		Compress:        true,
	})
	if err != nil {
		panic(err)
	}
	L = helper.Normal
}

func Sync() {
	if L != nil {
		_ = L.Sync()
	}
}

func With(fields ...zap.Field) *zap.Logger {
	if L == nil {
		return zap.NewNop()
	}
	return L.With(fields...)
}
