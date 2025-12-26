package logger

import (
	"os"
	"strconv"
	"strings"
)

const (
	EnvLogLevel      = "LOG_LEVEL"
	EnvLogBufferSize = "LOG_BUFFER_SIZE"
	EnvLogFormat     = "LOG_FORMAT"
	EnvDebugLevel    = "debug"
	EnvInfoLevel     = "info"
	EnvWarnLevel     = "warn"
	EnvErrorLevel    = "error"
	EnvFatalLevel    = "fatal"
	EnvPanicLevel    = "panic"
	EnvLogFormatJSON = "json"
	EnvLogFormatText = "text"
)

func fromEnvLogLevel() Level {
	var envLevel string
	envLevel = os.Getenv(EnvLogLevel)
	envLevel = strings.ToLower(envLevel)
	switch envLevel {
	case EnvDebugLevel:
		return DebugLevel
	case EnvInfoLevel:
		return InfoLevel
	case EnvWarnLevel:
		return WarnLevel
	case EnvErrorLevel:
		return ErrorLevel
	case EnvFatalLevel:
		return FatalLevel
	case EnvPanicLevel:
		return PanicLevel
	default:
		return DebugLevel
	}
}

func fromEnvBufferSize() int {
	var (
		err           error
		envBufferSize string
		bufSize       int
	)
	envBufferSize = os.Getenv(EnvLogBufferSize)
	bufSize, err = strconv.Atoi(envBufferSize)
	if err != nil {
		bufSize = 0
	}

	return bufSize
}

func fromEnvLogFormat() Format {
	var envFormat string
	envFormat = os.Getenv(EnvLogFormat)
	envFormat = strings.ToLower(envFormat)
	switch envFormat {
	case EnvLogFormatJSON:
		return JSONFormat
	case EnvLogFormatText:
		return TextFormat
	default:
		return TextFormat
	}
}

func ConfigFromEnv() Config {
	return Config{
		Level:      fromEnvLogLevel(),
		Format:     fromEnvLogFormat(),
		BufferSize: fromEnvBufferSize(),
		Output:     os.Stdout,
	}
}
