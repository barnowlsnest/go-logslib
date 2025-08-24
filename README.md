# go-logslib

A high-performance, zero-allocation Go logging library designed for production environments with strict performance requirements.

## Features

- 🚀 **Ultra-fast**: 20-600 ns/op depending on configuration
- 🧠 **Memory efficient**: 0-5 allocations per log operation
- 🎯 **Multiple formats**: Text and JSON output
- 📊 **Structured logging**: Type-safe field logging
- 🔧 **Configurable levels**: Debug, Info, Warn, Error, Fatal, Panic
- 🌐 **Context support**: TraceID, SpanID, and custom metadata
- 📦 **Buffering**: Optional buffering for cloud cost optimization
- 🔒 **Thread-safe**: Concurrent logging support
- 📈 **Production-ready**: Designed for high-load backend services

## Quick Start

```go
package main

import (
	"os"
	
	"github.com/barnowlsnest/go-logslib/pkg"
)

func main() {
	// Create a new logger
	logger := pkg.New(pkg.Config{
		Level:  pkg.InfoLevel,
		Format: pkg.JSONFormat,
		Output: os.Stdout,
	})
	
	// Simple logging
	logger.Info("Application started")
	
	// Structured logging with fields
	logger.Info("User logged in",
		pkg.Field{Key: "userID", Value: 12345},
		pkg.Field{Key: "email", Value: "user@example.com"},
	)
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

- Go 1.24 or later

### Performance Requirements

This library maintains strict performance requirements:
- **Execution time**: 20-600 ns/op
- **Memory allocations**: 0-5 per operation
- **Memory usage**: Minimal footprint

## Acknowledgments

- Optimized for high-load production environments
- Inspired by structured logging best practices
- Built with zero-allocation principles
