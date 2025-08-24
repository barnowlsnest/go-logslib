package logger

import (
	"bytes"
	"context"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogger_Levels(t *testing.T) {
	buf := &bytes.Buffer{}

	logger := New(Config{
		Level:  InfoLevel,
		Format: TextFormat,
		Output: buf,
	})

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	output := buf.String()
	assert.NotContains(t, output, "debug message")
	assert.Contains(t, output, "info message")
	assert.Contains(t, output, "warn message")
	assert.Contains(t, output, "error message")
}

func TestLogger_TextFormat(t *testing.T) {
	buf := &bytes.Buffer{}

	logger := New(Config{
		Level:  InfoLevel,
		Format: TextFormat,
		Output: buf,
	})

	logger.Info("test message", Field{Key: "key1", Value: "value1"}, Field{Key: "key2", Value: 42})

	output := buf.String()
	assert.Contains(t, output, "INFO")
	assert.Contains(t, output, "test message")
	assert.Contains(t, output, "key1=value1")
	assert.Contains(t, output, "key2=42")
}

func TestLogger_JSONFormat(t *testing.T) {
	buf := &bytes.Buffer{}

	logger := New(Config{
		Level:  InfoLevel,
		Format: JSONFormat,
		Output: buf,
	})

	logger.Info("test message", Field{Key: "key1", Value: "value1"}, Field{Key: "key2", Value: 42})

	output := buf.String()
	assert.Contains(t, output, `"level":"INFO"`)
	assert.Contains(t, output, `"message":"test message"`)
	assert.Contains(t, output, `"key1":"value1"`)
	assert.Contains(t, output, `"key2":42`)
}

func TestLogger_WithContext(t *testing.T) {
	buf := &bytes.Buffer{}

	logger := New(Config{
		Level:  InfoLevel,
		Format: JSONFormat,
		Output: buf,
	})

	ctx := context.WithValue(context.Background(), "traceID", "trace123")
	ctx = context.WithValue(ctx, "spanID", "span456")

	contextLogger := logger.WithContext(func() context.Context { return ctx })
	contextLogger.Info("test message", Field{Key: "custom", Value: "field"})

	output := buf.String()
	assert.Contains(t, output, `"traceID":"trace123"`)
	assert.Contains(t, output, `"spanID":"span456"`)
	assert.Contains(t, output, `"custom":"field"`)
}

func TestLogger_Buffering(t *testing.T) {
	buf := &bytes.Buffer{}

	logger := New(Config{
		Level:      InfoLevel,
		Format:     TextFormat,
		Output:     buf,
		BufferSize: 1024,
	})

	logger.Info("message1")
	logger.Info("message2")

	assert.Empty(t, buf.String())

	logger.Flush()

	output := buf.String()
	assert.Contains(t, output, "message1")
	assert.Contains(t, output, "message2")
}

func TestLogger_MemoryAllocations(t *testing.T) {
	buf := &bytes.Buffer{}

	logger := New(Config{
		Level:  InfoLevel,
		Format: TextFormat,
		Output: buf,
	})

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	for i := 0; i < 1000; i++ {
		logger.Info("test message", Field{Key: "iteration", Value: i})
	}

	runtime.GC()
	runtime.ReadMemStats(&m2)

	totalAllocs := m2.TotalAlloc - m1.TotalAlloc
	allocsPerLog := totalAllocs / 1000

	t.Logf("Total allocations: %d bytes", totalAllocs)
	t.Logf("Allocations per log: %d bytes", allocsPerLog)

	require.Less(t, allocsPerLog, uint64(200), "Memory allocation per log should be minimal")
}

func TestJSONEscaping(t *testing.T) {
	buf := &bytes.Buffer{}

	logger := New(Config{
		Level:  InfoLevel,
		Format: JSONFormat,
		Output: buf,
	})

	logger.Info("test message", Field{Key: "special", Value: `test"quote\backslash`})

	output := buf.String()
	assert.Contains(t, output, `test\"quote\\backslash`)
}

func TestFieldTypes(t *testing.T) {
	buf := &bytes.Buffer{}

	logger := New(Config{
		Level:  InfoLevel,
		Format: JSONFormat,
		Output: buf,
	})

	logger.Info("test",
		Field{Key: "string", Value: "test"},
		Field{Key: "int", Value: 42},
		Field{Key: "int64", Value: int64(123)},
		Field{Key: "float", Value: 3.14},
		Field{Key: "bool", Value: true},
	)

	output := buf.String()
	assert.Contains(t, output, `"string":"test"`)
	assert.Contains(t, output, `"int":42`)
	assert.Contains(t, output, `"int64":123`)
	assert.Contains(t, output, `"float":3.14`)
	assert.Contains(t, output, `"bool":true`)
}

func TestLevelString(t *testing.T) {
	assert.Equal(t, "DEBUG", DebugLevel.String())
	assert.Equal(t, "INFO", InfoLevel.String())
	assert.Equal(t, "WARN", WarnLevel.String())
	assert.Equal(t, "ERROR", ErrorLevel.String())
	assert.Equal(t, "FATAL", FatalLevel.String())
	assert.Equal(t, "PANIC", PanicLevel.String())
}

func TestBufferOverflow(t *testing.T) {
	buf := &bytes.Buffer{}

	logger := New(Config{
		Level:      InfoLevel,
		Format:     TextFormat,
		Output:     buf,
		BufferSize: 100,
	})

	longMessage := strings.Repeat("a", 60)
	logger.Info(longMessage)
	logger.Info(longMessage)

	logger.Flush()

	output := buf.String()
	count := strings.Count(output, longMessage)
	assert.Equal(t, 2, count, "Both messages should be written")
}

func TestContextFieldsExtraction(t *testing.T) {
	buf := &bytes.Buffer{}

	logger := New(Config{
		Level:  InfoLevel,
		Format: JSONFormat,
		Output: buf,
	})

	ctx := context.WithValue(context.Background(), "traceID", "trace123")
	contextLogger := logger.WithContext(func() context.Context { return ctx })

	contextLogger.Info("test")

	output := buf.String()
	assert.Contains(t, output, `"traceID":"trace123"`)
	assert.NotContains(t, output, `"spanID"`)
}

func TestLogger_WithStaticContext(t *testing.T) {
	buf := &bytes.Buffer{}

	logger := New(Config{
		Level:  InfoLevel,
		Format: JSONFormat,
		Output: buf,
	})

	ctx := context.WithValue(context.Background(), "traceID", "static123")
	contextLogger := logger.WithStaticContext(ctx)

	contextLogger.Info("test message")

	output := buf.String()
	assert.Contains(t, output, `"traceID":"static123"`)
}

func TestLogger_DynamicContext(t *testing.T) {
	buf := &bytes.Buffer{}

	logger := New(Config{
		Level:  InfoLevel,
		Format: JSONFormat,
		Output: buf,
	})

	traceCounter := 0
	contextLogger := logger.WithContext(func() context.Context {
		traceCounter++
		return context.WithValue(context.Background(), "traceID", "dynamic"+string(rune('0'+traceCounter)))
	})

	contextLogger.Info("first message")
	contextLogger.Info("second message")

	output := buf.String()
	assert.Contains(t, output, `"traceID":"dynamic1"`)
	assert.Contains(t, output, `"traceID":"dynamic2"`)
}

func TestLogger_NilContextFunc(t *testing.T) {
	buf := &bytes.Buffer{}

	logger := New(Config{
		Level:  InfoLevel,
		Format: JSONFormat,
		Output: buf,
	})

	contextLogger := logger.WithContext(nil)
	contextLogger.Info("test message", Field{Key: "custom", Value: "field"})

	output := buf.String()
	assert.NotContains(t, output, `"traceID"`)
	assert.NotContains(t, output, `"spanID"`)
	assert.Contains(t, output, `"custom":"field"`)
}
