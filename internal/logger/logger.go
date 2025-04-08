package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

func Init() {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	var err error
	log, err = config.Build()
	if err != nil {
		panic(err)
	}
}

func Get() *zap.Logger {
	return log
}

func Sync() {
	_ = log.Sync()
}
