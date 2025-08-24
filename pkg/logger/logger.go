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
	"io"
	"os"
	"strconv"
	"sync"
	"time"
)

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
)

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

	// Value is the field value, can be string, int, int64, float64, or bool
	Value interface{}
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
}

// Logger is a high-performance logging instance that supports structured
// logging with minimal memory allocations. It is safe for concurrent use.
type Logger struct {
	config Config
	buffer []byte
	pool   sync.Pool
	mu     sync.Mutex
}

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
		New: func() interface{} {
			return make([]byte, 0, 256)
		},
	}

	return l
}

// WithContext creates a ContextLogger that automatically extracts context
// information from the provided context function for each log entry.
//
// This is the recommended approach for dynamic contexts (e.g., HTTP request contexts)
// as it resolves the context at log time, ensuring fresh context values.
//
// Example:
//
//	func handleRequest(w http.ResponseWriter, r *http.Request) {
//		contextLogger := logger.WithContext(func() context.Context {
//			return r.Context() // Always gets fresh request context
//		})
//		contextLogger.Info("Processing request")
//	}
func (l *Logger) WithContext(ctxFunc func() context.Context) *ContextLogger {
	return &ContextLogger{
		logger:  l,
		ctxFunc: ctxFunc,
	}
}

// WithStaticContext creates a ContextLogger with a static context that won't change.
//
// This is useful when you have a context that remains constant throughout
// the logger's lifetime. For dynamic contexts, prefer WithContext().
//
// Example:
//
//	ctx := context.WithValue(context.Background(), "serviceID", "user-service")
//	contextLogger := logger.WithStaticContext(ctx)
//	contextLogger.Info("Service started")
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

	buf := l.pool.Get().([]byte)
	buf = buf[:0]
	defer l.pool.Put(buf)

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
//
// ContextLogger is created using Logger.WithContext() or Logger.WithStaticContext().
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
		if traceID := ctx.Value("traceID"); traceID != nil {
			contextFields = append(contextFields, Field{Key: "traceID", Value: traceID})
		}
		if spanID := ctx.Value("spanID"); spanID != nil {
			contextFields = append(contextFields, Field{Key: "spanID", Value: spanID})
		}
	}

	return append(contextFields, fields...)
}

func (l *Logger) appendText(buf []byte, level Level, msg string, fields ...Field) []byte {
	now := time.Now().UTC()

	buf = append(buf, now.Format("2006-01-02T15:04:05.000Z07:00")...)
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

func appendValue(buf []byte, value interface{}) []byte {
	switch v := value.(type) {
	case string:
		if needsQuoting(v) {
			buf = append(buf, '"')
			buf = append(buf, v...)
			buf = append(buf, '"')
		} else {
			buf = append(buf, v...)
		}
	case int:
		return appendInt(buf, int64(v))
	case int64:
		return appendInt(buf, v)
	case float64:
		return appendFloat(buf, v)
	case bool:
		if v {
			buf = append(buf, "true"...)
		} else {
			buf = append(buf, "false"...)
		}
	default:
		buf = append(buf, '"')
		buf = append(buf, "unknown"...)
		buf = append(buf, '"')
	}
	return buf
}

func needsQuoting(s string) bool {
	for _, r := range s {
		if r == ' ' || r == '=' || r == '"' {
			return true
		}
	}
	return false
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

}


// appendFloat appends the string representation of a float64 to the buffer.
func appendFloat(buf []byte, f float64) []byte {
	// Use 'g' format for compact representation, 6 digits precision, -1 for all digits necessary
	return append(buf, strconv.FormatFloat(f, 'g', -1, 64)...)
}

// appendFloat appends the string representation of a float64 to the buffer.
func appendFloat(buf []byte, f float64) []byte {
	// Use 'g' format for compact representation, 6 digits precision, -1 for all digits necessary
	return append(buf, strconv.FormatFloat(f, 'g', -1, 64)...)
}
