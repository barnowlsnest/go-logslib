// Package logger provides a high-performance, zero-allocation logging library
// designed for production environments with strict performance requirements.
//
// The logger supports multiple output formats (text and JSON), configurable
// log levels, context-aware logging, and optional buffering for cloud cost
// optimization. It is optimized for minimal memory allocations and fast
// execution times (20-600 ns/op).
//
// Example usage:
//
//	logger := logger.New(logger.Config{
//		Level:  logger.InfoLevel,
//		Format: logger.JSONFormat,
//		Output: os.Stdout,
//	})
//
//	logger.Info("Application started", logger.Field{Key: "version", Value: "1.0.0"})
package logger

import (
	"context"
	"errors"
	"io"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

// Level represents the severity level of a log entry.
// Lower values indicate more verbose logging.
type Level int8

const (
	// DebugLevel logs are typically voluminous, and are usually disabled in
	// production.
	DebugLevel Level = iota - 1

	// InfoLevel is the default logging priority.
	InfoLevel

	// WarnLevel logs are more important than Info, but don't need individual
	// human review.
	WarnLevel

	// ErrorLevel logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-level logs.
	ErrorLevel

	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel

	// PanicLevel logs a message, then panics.
	PanicLevel

	DefaultTimeFormat = "2006-01-02T15:04:05.000Z07:00"
)

var ErrUnknownLogLevel = errors.New("unknown log level")

func LogLevelFromString(str string) (Level, error) {
	str = strings.ToLower(str)
	switch str {
	case "debug":
		return DebugLevel, nil
	case "info":
		return InfoLevel, nil
	case "warn":
		return WarnLevel, nil
	case "error":
		return ErrorLevel, nil
	case "fatal":
		return FatalLevel, nil
	case "panic":
		return PanicLevel, nil
	default:
		return -2, ErrUnknownLogLevel
	}
}

// String returns the string representation of the log level.
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	case PanicLevel:
		return "PANIC"
	default:
		return "UNKNOWN"
	}
}

// Format represents the output format for log entries.
type Format int8

const (
	// TextFormat outputs logs in a human-readable text format.
	// Example: "2024-01-20T15:04:05.000Z INFO User logged in userID=12345"
	TextFormat Format = iota

	// JSONFormat outputs logs in structured JSON format.
	// Example: {"timestamp":"2024-01-20T15:04:05.000Z","level":"INFO","message":"User logged in","userID":12345}
	JSONFormat
)

// Field represents a key-value pair that can be attached to a log entry.
// Fields are used for structured logging to provide additional context.
type Field struct {
	// Key is the field name
	Key string

	// Value is the field value. Supported types are string, int, int64, uint,
	// uint64, float32, float64, bool, and time.Duration (rendered as its string
	// form, e.g. "1.5s"). Other types are rendered as "unknown".
	Value any
}

// Config holds the configuration for a Logger instance.
type Config struct {
	// Level sets the minimum log level that will be output.
	// Log entries below this level will be discarded.
	Level Level

	// Format determines the output format (TextFormat or JSONFormat).
	Format Format

	// Output specifies where log entries will be written.
	// If nil, defaults to os.Stdout.
	Output io.Writer

	// BufferSize enables buffering when > 0. Log entries are buffered
	// until the buffer is full or Flush() is called. Useful for reducing
	// I/O operations in cloud environments.
	BufferSize int

	// UseUTC determines whether timestamps are in UTC (true) or local timezone (false).
	// Defaults to false (local timezone).
	UseUTC bool
}

// Logger is a high-performance logging instance that supports structured
// logging with minimal memory allocations. It is safe for concurrent use.
type Logger struct {
	config Config
	buffer []byte
	pool   sync.Pool
	mu     sync.Mutex
}

const (
	FieldTraceID = "traceID"
	FieldSpanID  = "spanID"
)

// New creates a new Logger instance with the given configuration.
//
// If config.Output is nil, it defaults to os.Stdout.
// The logger is safe for concurrent use and optimized for minimal
// memory allocations using object pooling.
//
// Example:
//
//	logger := logger.New(logger.Config{
//		Level:      logger.InfoLevel,
//		Format:     logger.JSONFormat,
//		Output:     os.Stdout,
//		BufferSize: 4096, // Optional buffering
//	})
func New(config Config) *Logger {
	if config.Output == nil {
		config.Output = os.Stdout
	}

	l := &Logger{
		config: config,
		buffer: make([]byte, 0, config.BufferSize),
	}

	l.pool = sync.Pool{
		New: func() any {
			buf := make([]byte, 0, 256)
			return &buf
		},
	}

	return l
}

// StringField creates a Field with a string value.
func StringField(name, value string) Field {
	return Field{Key: name, Value: value}
}

// IntField creates a Field with an int value.
func IntField(name string, value int) Field {
	return Field{Key: name, Value: value}
}

func Int8Field(name string, value int8) Field {
	return Field{Key: name, Value: value}
}

func Int16Field(name string, value int16) Field {
	return Field{Key: name, Value: value}
}

func Int64Field(name string, value int64) Field {
	return Field{Key: name, Value: value}
}

func UintField(name string, value uint) Field {
	return Field{Key: name, Value: value}
}

func Uint8Field(name string, value uint8) Field {
	return Field{Key: name, Value: value}
}

func Uint16Field(name string, value uint16) Field {
	return Field{Key: name, Value: value}
}

func Uint64Field(name string, value uint64) Field {
	return Field{Key: name, Value: value}
}

func Uint32Field(name string, value uint32) Field {
	return Field{Key: name, Value: value}
}

func Float32Field(name string, value float32) Field {
	return Field{Key: name, Value: value}
}

func Float64Field(name string, value float64) Field {
	return Field{Key: name, Value: value}
}

func BoolField(name string, value bool) Field {
	return Field{Key: name, Value: value}
}

func DurationField(name string, value time.Duration) Field {
	return Field{Key: name, Value: value}
}

// WithContextFunc creates a ContextLogger that automatically extracts context
// information from the provided context function for each log entry.
//
// This is the recommended approach for dynamic contexts (e.g., HTTP request contexts)
// as it resolves the context at log time, ensuring fresh context values.
func (l *Logger) WithContextFunc(ctxFunc func() context.Context) *ContextLogger {
	return &ContextLogger{
		logger:  l,
		ctxFunc: ctxFunc,
	}
}

// WithContext creates a ContextLogger that automatically extracts context
func (l *Logger) WithContext(ctx context.Context) *ContextLogger {
	if ctx == nil {
		ctx = context.Background()
	}

	return l.WithContextFunc(func() context.Context { return ctx })
}

// WithStaticContext creates a ContextLogger with a context that won't change.
//
// Deprecated: Use WithContext() instead.
func (l *Logger) WithStaticContext(ctx context.Context) *ContextLogger {
	return &ContextLogger{
		logger:  l,
		ctxFunc: func() context.Context { return ctx },
	}
}

func (l *Logger) log(level Level, msg string, fields ...Field) {
	if level < l.config.Level {
		return
	}

	bufPtr := l.pool.Get().(*[]byte)
	defer l.pool.Put(bufPtr)

	buf := (*bufPtr)[:0]

	switch l.config.Format {
	case JSONFormat:
		buf = l.appendJSON(buf, level, msg, fields...)
	default:
		buf = l.appendText(buf, level, msg, fields...)
	}

	l.write(buf)
}

// Debug logs a message at DebugLevel. Debug logs are typically voluminous
// and are usually disabled in production.
func (l *Logger) Debug(msg string, fields ...Field) {
	l.log(DebugLevel, msg, fields...)
}

// Info logs a message at InfoLevel. This is the default logging priority
// for general application information.
func (l *Logger) Info(msg string, fields ...Field) {
	l.log(InfoLevel, msg, fields...)
}

// Warn logs a message at WarnLevel. Warning logs are more important than Info,
// but don't need individual human review.
func (l *Logger) Warn(msg string, fields ...Field) {
	l.log(WarnLevel, msg, fields...)
}

// Error logs a message at ErrorLevel. Error logs are high-priority.
// If an application is running smoothly, it shouldn't generate any error-level logs.
func (l *Logger) Error(msg string, fields ...Field) {
	l.log(ErrorLevel, msg, fields...)
}

// Fatal logs a message at FatalLevel, then calls os.Exit(1).
// This function does not return.
func (l *Logger) Fatal(msg string, fields ...Field) {
	l.log(FatalLevel, msg, fields...)
	os.Exit(1)
}

// Panic logs a message at PanicLevel, then panics with the message.
// This function does not return.
//
// Deprecated: Use Error() or Fatal() instead. Panic-level logging is generally discouraged in production code.
func (l *Logger) Panic(msg string, fields ...Field) {
	l.log(PanicLevel, msg, fields...)
	panic(msg)
}

func (l *Logger) write(buf []byte) {
	if l.config.BufferSize > 0 {
		l.mu.Lock()
		defer l.mu.Unlock()

		if len(l.buffer)+len(buf) > l.config.BufferSize {
			l.flush()
		}
		l.buffer = append(l.buffer, buf...)
		l.buffer = append(l.buffer, '\n')
	} else {
		_, _ = l.config.Output.Write(buf)
		_, _ = l.config.Output.Write([]byte{'\n'})
	}
}

// Flush forces all buffered log entries to be written to the output.
// This method is only effective when BufferSize > 0 in the Config.
// It is safe to call concurrently with other logger methods.
func (l *Logger) Flush() {
	if l.config.BufferSize > 0 {
		l.mu.Lock()
		defer l.mu.Unlock()
		l.flush()
	}
}

// flush is an internal method that writes all buffered content to the output.
// It must be called with l.mu held.
func (l *Logger) flush() {
	if len(l.buffer) > 0 {
		_, _ = l.config.Output.Write(l.buffer)
		l.buffer = l.buffer[:0]
	}
}

// ContextLogger is a logger that automatically extracts context information
// for each log entry. It provides the same logging methods as Logger but
// includes context fields like traceID and spanID.
type ContextLogger struct {
	logger  *Logger
	ctxFunc func() context.Context
}

// Debug logs a message at DebugLevel, automatically including context fields
// such as traceID and spanID if present in the context.
func (cl *ContextLogger) Debug(msg string, fields ...Field) {
	cl.logger.log(DebugLevel, msg, cl.extractContextFields(fields)...)
}

// Info logs a message at InfoLevel, automatically including context fields
// such as traceID and spanID if present in the context.
func (cl *ContextLogger) Info(msg string, fields ...Field) {
	cl.logger.log(InfoLevel, msg, cl.extractContextFields(fields)...)
}

// Warn logs a message at WarnLevel, automatically including context fields
// such as traceID and spanID if present in the context.
func (cl *ContextLogger) Warn(msg string, fields ...Field) {
	cl.logger.log(WarnLevel, msg, cl.extractContextFields(fields)...)
}

// Error logs a message at ErrorLevel, automatically including context fields
// such as traceID and spanID if present in the context.
func (cl *ContextLogger) Error(msg string, fields ...Field) {
	cl.logger.log(ErrorLevel, msg, cl.extractContextFields(fields)...)
}

// Fatal logs a message at FatalLevel with context fields, then calls os.Exit(1).
// This function does not return.
func (cl *ContextLogger) Fatal(msg string, fields ...Field) {
	cl.logger.log(FatalLevel, msg, cl.extractContextFields(fields)...)
	os.Exit(1)
}

// Panic logs a message at PanicLevel with context fields, then panics with the message.
// This function does not return.
func (cl *ContextLogger) Panic(msg string, fields ...Field) {
	cl.logger.log(PanicLevel, msg, cl.extractContextFields(fields)...)
	panic(msg)
}

func (cl *ContextLogger) extractContextFields(fields []Field) []Field {
	contextFields := make([]Field, 0, 4)

	if cl.ctxFunc != nil {
		ctx := cl.ctxFunc()
		if traceID := ctx.Value(contextKey(FieldTraceID)); traceID != nil {
			contextFields = append(contextFields, Field{Key: FieldTraceID, Value: traceID})
		}
		if spanID := ctx.Value(contextKey(FieldSpanID)); spanID != nil {
			contextFields = append(contextFields, Field{Key: FieldSpanID, Value: spanID})
		}
	}

	return append(contextFields, fields...)
}

func (l *Logger) appendText(buf []byte, level Level, msg string, fields ...Field) []byte {
	now := time.Now()
	if l.config.UseUTC {
		now = now.UTC()
	}

	buf = append(buf, now.Format(DefaultTimeFormat)...)
	buf = append(buf, ' ')
	buf = append(buf, level.String()...)
	buf = append(buf, ' ')
	buf = append(buf, msg...)

	for _, field := range fields {
		buf = append(buf, ' ')
		buf = append(buf, field.Key...)
		buf = append(buf, '=')
		buf = appendValue(buf, field.Value)
	}

	return buf
}

//nolint:gocyclo
func appendValue(buf []byte, value any) []byte {
	switch v := value.(type) {
	case nil:
		return append(buf, "<nil>"...)
	case string:
		return appendString(buf, v)
	case []byte:
		return appendString(buf, string(v))
	case int:
		return appendInt(buf, int64(v))
	case int64:
		return appendInt(buf, v)
	case int32:
		return appendInt(buf, int64(v))
	case int16:
		return appendInt(buf, int64(v))
	case int8:
		return appendInt(buf, int64(v))
	case uint:
		return appendUint(buf, uint64(v))
	case uint64:
		return appendUint(buf, v)
	case uint32:
		return appendUint(buf, uint64(v))
	case uint16:
		return appendUint(buf, uint64(v))
	case uint8:
		return appendUint(buf, uint64(v))
	case uintptr:
		return appendUint(buf, uint64(v))
	case bool:
		if v {
			return append(buf, "true"...)
		}
		return append(buf, "false"...)
	case float64:
		return appendFloat(buf, v, 64)
	case float32:
		return appendFloat(buf, float64(v), 32)
	case time.Duration:
		return appendDuration(buf, v)
	case time.Time:
		return v.AppendFormat(buf, time.RFC3339Nano)
	case error:
		return appendString(buf, v.Error())
	default:
		return appendReflect(buf, value)
	}
}

func appendReflect(buf []byte, value any) []byte {
	return appendReflectValue(buf, value, appendString, "<nil>")
}

// appendReflectValue handles named types whose underlying kind is representable
// as a scalar (e.g. `type Status string`), which a concrete type switch cannot
// match directly. The format-specific parts are injected: appendStr controls how
// string kinds are rendered (quoted vs. quote-if-needed) and nilLit is emitted
// for an invalid/nil value.
func appendReflectValue(buf []byte, value any, appendStr func([]byte, string) []byte, nilLit string) []byte {
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.String:
		return appendStr(buf, rv.String())
	case reflect.Bool:
		if rv.Bool() {
			return append(buf, "true"...)
		}
		return append(buf, "false"...)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.AppendInt(buf, rv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.AppendUint(buf, rv.Uint(), 10)
	case reflect.Float32:
		return appendFloat(buf, rv.Float(), 32)
	case reflect.Float64:
		return appendFloat(buf, rv.Float(), 64)
	case reflect.Invalid:
		return append(buf, nilLit...)
	default:
		return append(buf, `"<unsupported>"`...)
	}
}

const hex = "0123456789abcdef"

func appendString(buf []byte, s string) []byte {
	if !needsQuoting(s) {
		return append(buf, s...)
	}
	buf = append(buf, '"')
	start := 0
	for i := 0; i < len(s); {
		if c := s[i]; c < utf8.RuneSelf {
			if safeChar(c) {
				i++
				continue
			}
			buf = append(buf, s[start:i]...)
			switch c {
			case '"':
				buf = append(buf, '\\', '"')
			case '\\':
				buf = append(buf, '\\', '\\')
			case '\n':
				buf = append(buf, '\\', 'n')
			case '\r':
				buf = append(buf, '\\', 'r')
			case '\t':
				buf = append(buf, '\\', 't')
			default:
				buf = append(buf, '\\', 'u', '0', '0', hex[c>>4], hex[c&0xf])
			}
			i++
			start = i
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			buf = append(buf, s[start:i]...)
			buf = append(buf, `\ufffd`...)
			i += size
			start = i
			continue
		}
		i += size
	}
	buf = append(buf, s[start:]...)

	return append(buf, '"')
}

func safeChar(c byte) bool {
	return c >= 0x20 && c != '"' && c != '\\' && c != 0x7f
}

func needsQuoting(s string) bool {
	if s == "" {
		return true
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c <= ' ' || c == '"' || c == '\\' || c == '=' || c >= utf8.RuneSelf {
			return true
		}
	}

	return false
}

func appendFloat(buf []byte, f float64, bits int) []byte {
	switch {
	case math.IsNaN(f):
		return append(buf, `"NaN"`...)
	case math.IsInf(f, 1):
		return append(buf, `"+Inf"`...)
	case math.IsInf(f, -1):
		return append(buf, `"-Inf"`...)
	}

	return strconv.AppendFloat(buf, f, 'g', -1, bits)
}

func appendDuration(buf []byte, d time.Duration) []byte {
	var a [32]byte

	return append(buf, a[:copy(a[:], d.String())]...)
}

func appendInt(buf []byte, i int64) []byte {
	if i == 0 {
		return append(buf, '0')
	}

	if i < 0 {
		buf = append(buf, '-')
		i = -i
	}

	var tmp [20]byte
	idx := 20
	for i > 0 {
		idx--
		tmp[idx] = byte('0' + i%10)
		i /= 10
	}

	return append(buf, tmp[idx:]...)
}

// appendUint appends the base-10 representation of an unsigned integer to the buffer.
func appendUint(buf []byte, u uint64) []byte {
	if u == 0 {
		return append(buf, '0')
	}

	var tmp [20]byte
	idx := 20
	for u > 0 {
		idx--
		tmp[idx] = byte('0' + u%10)
		u /= 10
	}

	return append(buf, tmp[idx:]...)
}
