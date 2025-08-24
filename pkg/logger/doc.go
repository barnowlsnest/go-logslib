// Package logger provides a high-performance, zero-allocation logging library
// optimized for production environments with strict performance requirements.
//
// # Performance Characteristics
//
// The logger is designed to meet aggressive performance targets:
//   - Execution time: 20-600 ns/op depending on configuration and field count
//   - Memory allocations: 0-5 allocations per log operation
//   - Memory usage: Minimal footprint using object pooling and zero-copy operations
//
// # Key Features
//
//   - Multiple output formats: Human-readable text and structured JSON
//   - Configurable log levels: DEBUG, INFO, WARN, ERROR, FATAL, PANIC
//   - Structured logging: Type-safe field logging with key-value pairs
//   - Context-aware logging: Automatic extraction of traceID, spanID, and custom metadata
//   - Optional buffering: Reduces I/O operations for cloud cost optimization
//   - Thread-safe: Safe for concurrent use across goroutines
//   - Zero external dependencies: Pure Go implementation
//
// # Quick Start
//
// Basic usage with default configuration:
//
//	logger := logger.New(logger.Config{
//		Level:  logger.InfoLevel,
//		Format: logger.JSONFormat,
//		Output: os.Stdout,
//	})
//
//	logger.Info("Application started")
//
// # Structured Logging
//
// Add context with structured fields:
//
//	logger.Info("User authentication",
//		logger.Field{Key: "userID", Value: 12345},
//		logger.Field{Key: "method", Value: "oauth"},
//		logger.Field{Key: "success", Value: true},
//	)
//
// # Context-Aware Logging
//
// For HTTP request tracing and distributed systems:
//
//	// Dynamic context (recommended for HTTP handlers)
//	contextLogger := logger.WithContext(func() context.Context {
//		return r.Context() // Always gets fresh request context
//	})
//
//	// Static context (for long-lived services)
//	ctx := context.WithValue(context.Background(), "serviceID", "user-service")
//	contextLogger := logger.WithStaticContext(ctx)
//
//	contextLogger.Info("Request processed") // Includes traceID, spanID automatically
//
// # Buffering for Cloud Optimization
//
// Reduce I/O costs in cloud environments:
//
//	logger := logger.New(logger.Config{
//		Level:      logger.InfoLevel,
//		Format:     logger.JSONFormat,
//		Output:     os.Stdout,
//		BufferSize: 4096, // Buffer until 4KB, then flush
//	})
//
//	// Manually flush when needed
//	logger.Flush()
//
// # Performance Benchmarks
//
// Measured on Apple M1 Max:
//   - Simple JSON logging: 171 ns/op, 3 allocs, 57 B/op
//   - Text with fields: 262 ns/op, 3 allocs, 57 B/op
//   - JSON with context: 346 ns/op, 4 allocs, 185 B/op
//   - Level filtering (disabled): 2.8 ns/op, 0 allocs, 0 B/op
//
// # Thread Safety
//
// All logger operations are safe for concurrent use. The logger uses:
//   - sync.Pool for buffer management to reduce allocations
//   - sync.Mutex for buffering coordination
//   - Atomic operations where possible for optimal performance
//
// # Supported Field Types
//
// The logger efficiently handles these Go types in structured fields:
//   - string: Properly escaped in JSON, quoted in text if needed
//   - int, int64: Native numeric formatting
//   - float64: Basic precision formatting optimized for performance
//   - bool: Native boolean representation
//   - Other types: Converted to "unknown" string representation
package logger
