package logger

import (
	"context"
	"io"
	"testing"
)

var discardWriter = io.Discard

func BenchmarkLogger_SimpleText(b *testing.B) {
	logger := New(Config{
		Level:  InfoLevel,
		Format: TextFormat,
		Output: discardWriter,
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.Info("simple message")
	}
}

func BenchmarkLogger_SimpleJSON(b *testing.B) {
	logger := New(Config{
		Level:  InfoLevel,
		Format: JSONFormat,
		Output: discardWriter,
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.Info("simple message")
	}
}

func BenchmarkLogger_TextWithFields(b *testing.B) {
	logger := New(Config{
		Level:  InfoLevel,
		Format: TextFormat,
		Output: discardWriter,
	})

	fields := []Field{
		{Key: "user_id", Value: 12345},
		{Key: "action", Value: "login"},
		{Key: "success", Value: true},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.Info("user action", fields...)
	}
}

func BenchmarkLogger_JSONWithFields(b *testing.B) {
	logger := New(Config{
		Level:  InfoLevel,
		Format: JSONFormat,
		Output: discardWriter,
	})

	fields := []Field{
		{Key: "user_id", Value: 12345},
		{Key: "action", Value: "login"},
		{Key: "success", Value: true},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.Info("user action", fields...)
	}
}

func BenchmarkLogger_ManyFields(b *testing.B) {
	logger := New(Config{
		Level:  InfoLevel,
		Format: JSONFormat,
		Output: discardWriter,
	})

	fields := []Field{
		{Key: "field1", Value: "value1"},
		{Key: "field2", Value: "value2"},
		{Key: "field3", Value: 42},
		{Key: "field4", Value: 3.14},
		{Key: "field5", Value: true},
		{Key: "field6", Value: "another value"},
		{Key: "field7", Value: 999},
		{Key: "field8", Value: false},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.Info("complex message", fields...)
	}
}

func BenchmarkLogger_WithContext(b *testing.B) {
	logger := New(Config{
		Level:  InfoLevel,
		Format: JSONFormat,
		Output: discardWriter,
	})

	ctx := context.WithValue(context.Background(), contextKey("traceID"), "trace123456")
	ctx = context.WithValue(ctx, contextKey("spanID"), "span789012")
	contextLogger := logger.WithContext(func() context.Context { return ctx })

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		contextLogger.Info("request processed", Field{Key: "status", Value: 200})
	}
}

func BenchmarkLogger_Buffered(b *testing.B) {
	logger := New(Config{
		Level:      InfoLevel,
		Format:     JSONFormat,
		Output:     discardWriter,
		BufferSize: 4096,
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.Info("buffered message", Field{Key: "iteration", Value: i})
	}
}

func BenchmarkLogger_LevelFiltering(b *testing.B) {
	logger := New(Config{
		Level:  WarnLevel,
		Format: JSONFormat,
		Output: discardWriter,
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.Debug("debug message that should be filtered")
		logger.Info("info message that should be filtered")
	}
}

func BenchmarkLogger_ConcurrentAccess(b *testing.B) {
	logger := New(Config{
		Level:  InfoLevel,
		Format: JSONFormat,
		Output: discardWriter,
	})

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("concurrent message", Field{Key: "worker", Value: "test"})
		}
	})
}

func BenchmarkAppendInt(b *testing.B) {
	buf := make([]byte, 0, 64)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf = buf[:0]
		buf = appendInt(buf, int64(i))
	}
}

func BenchmarkAppendJSON(b *testing.B) {
	logger := New(Config{
		Level:  InfoLevel,
		Format: JSONFormat,
		Output: discardWriter,
	})

	buf := make([]byte, 0, 256)
	fields := []Field{
		{Key: "key1", Value: "value1"},
		{Key: "key2", Value: 42},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf = buf[:0]
		buf = logger.appendJSON(buf, InfoLevel, "test message", fields...)
	}
}

func BenchmarkFieldAppend(b *testing.B) {
	buf := make([]byte, 0, 128)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf = buf[:0]
		buf = appendValue(buf, "test string value")
		buf = appendValue(buf, 12345)
		buf = appendValue(buf, true)
		buf = appendValue(buf, 3.14159)
	}
}
