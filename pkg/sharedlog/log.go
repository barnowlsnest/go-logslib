// Package sharedlog exposes frequently used logging functions that use singleton logger to simplify getting started.
// Itn this particular package ignores "LOG_FORMAT" and "LOG_USE_UTC" env vars and always uses JSON format and UTC timestamps.
package sharedlog

import (
	"sync"

	"github.com/barnowlsnest/go-logslib/v2/pkg/logger"
)

var (
	once         sync.Once
	sharedLogger *logger.Logger
)

func setSharedLogger() {
	cfg := logger.ConfigFromEnv()
	cfg.Format = logger.JSONFormat
	cfg.UseUTC = true
	sharedLogger = logger.New(cfg)
}

func Logger() *logger.Logger {
	once.Do(setSharedLogger)
	return sharedLogger
}

func F(name string, value any) logger.Field {
	return logger.Field{Key: name, Value: value}
}

func Debug(msg string, fields ...logger.Field) {
	Logger().Debug(msg, fields...)
}

func Error(err error, fields ...logger.Field) {
	Logger().Error(err.Error(), fields...)
}

func Panic(err error, fields ...logger.Field) {
	Logger().Panic(err.Error(), fields...)
}

func Info(msg string, fields ...logger.Field) {
	Logger().Info(msg, fields...)
}
