# go-logslib

Simple logging library ready to Go.

## Features

- 🚀 **Relativly Fast**: 20-600 ns/op depending on configuration
- 🧠 **Memory efficient**: 0-5 allocations per log operation
- 🎯 **Flexible**: Text and JSON output formats with UTC or Local time zones
- 📊 **Structured logging**: Type-safe field logging
- 🔧 **Configurable levels**: Debug, Info, Warn, Error, Fatal, Panic
- 🌐 **Context support**: TraceID, SpanID, and custom metadata
- 📦 **Buffering**: Optional buffering for cloud cost optimization
- 🔒 **Thread-safe**: Concurrent logging support

## Quick Start

```go
package main

import (
	"os"
	
	"github.com/barnowlsnest/go-logslib/pkg/logger"
)

func main() {
	// Create a new logger
	log := logger.New(logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.JSONFormat,
		Output: os.Stdout,
	})
	
	// Simple logging
	log.Info("Application started")
	
	// Structured logging with fields
	log.Info("User logged in",
		logger.Field{Key: "userID", Value: 12345},
		logger.Field{Key: "email", Value: "user@example.com"},
	)
}
```

## Shared Logger

The `sharedlog` package provides a zero-configuration singleton logger that always outputs JSON in UTC to stdout. It reads `LOG_LEVEL` and `LOG_BUFFER_SIZE` from the environment but ignores `LOG_FORMAT` and `LOG_USE_UTC` (always JSON + UTC).

```go
package main

import (
	"errors"

	"github.com/barnowlsnest/go-logslib/pkg/sharedlog"
)

func main() {
	// Simple logging
	sharedlog.Info("Application started")
	sharedlog.Debug("cache miss", sharedlog.F("key", "user:42"))

	// Error and Panic accept an error directly
	err := errors.New("connection refused")
	sharedlog.Error(err, sharedlog.F("host", "db.local"))

	// Access the underlying *logger.Logger for advanced use
	log := sharedlog.Logger()
	log.Warn("disk usage high", sharedlog.F("pct", 92))
	log.Flush()
}
```

## Configuration

### Log Levels

```go
// Available log levels (in order of severity)
DebugLevel  // -1: Detailed information for debugging
InfoLevel   //  0: General information (default)
WarnLevel   //  1: Warning messages
ErrorLevel  //  2: Error conditions
FatalLevel  //  3: Fatal errors (calls os.Exit(1))
PanicLevel  //  4: Panic conditions (calls panic())
```

### Output Formats

```go
// Text format (human-readable)
log := logger.New(pkg.Config{
    Level:  pkg.InfoLevel,
    Format: pkg.TextFormat,
    Output: os.Stdout,
})
// Output: 2024-01-20T15:04:05.000Z INFO User action userID=12345 action=login

// JSON format (structured)
log := logger.New(pkg.Config{
    Level:  pkg.InfoLevel,
    Format: pkg.JSONFormat,
    Output: os.Stdout,
})
// Output: {"timestamp":"2024-01-20T15:04:05.000Z","level":"INFO","message":"User action","userID":12345,"action":"login"}
```

### Buffering

Enable buffering for reduced I/O operations and cost optimization in cloud environments:

```go
log := logger.New(pkg.Config{
    Level:      pkg.InfoLevel,
    Format:     pkg.JSONFormat,
    Output:     os.Stdout,
    BufferSize: 4096, // Buffer up to 4KB before flushing
})

// Manually flush when needed
logger.Flush()
```

## Performance

Benchmarks on Apple M1 Max:

| Operation        | Time (ns/op) | Allocations | Memory (B/op) |
|------------------|--------------|-------------|---------------|
| Simple JSON      | 171          | 3           | 57            |
| Text with fields | 262          | 3           | 57            |
| JSON with fields | 241          | 3           | 57            |
| Many fields (8)  | 391          | 3           | 57            |
| With context     | 346          | 4           | 185           |
| Level filtering  | 2.8          | 0           | 0             |

## Development

### Prerequisites

- Go 1.25.7 or later

### Performance Requirements

This library maintains strict performance requirements:
- **Execution time**: 20-600 ns/op
- **Memory allocations**: 0-5 per operation
- **Memory usage**: Minimal footprint

## Acknowledgments

- Optimized for high-load production environments
- Inspired by structured logging best practices
- Built with zero-allocation principles
