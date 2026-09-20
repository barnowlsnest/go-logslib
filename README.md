# go-logslib

[![Build and Test](https://github.com/barnowlsnest/go-logslib/actions/workflows/build.yml/badge.svg)](https://github.com/barnowlsnest/go-logslib/actions/workflows/build.yml)
[![Lint](https://github.com/barnowlsnest/go-logslib/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/barnowlsnest/go-logslib/actions/workflows/golangci-lint.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/barnowlsnest/go-logslib/v2.svg)](https://pkg.go.dev/github.com/barnowlsnest/go-logslib/v2)
[![Go Report Card](https://goreportcard.com/badge/github.com/barnowlsnest/go-logslib/v2)](https://goreportcard.com/report/github.com/barnowlsnest/go-logslib/v2)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Simple logging library ready to Go.

A small, dependency-free structured logger for production Go services. Text or
JSON output, typed fields, context propagation and optional buffering — in about
a thousand lines of hand-written `[]byte` appends, with no `fmt.Sprintf` and no
`encoding/json` on the hot path.

## Features

- 🚀 **Relatively Fast**: 2-450 ns/op depending on configuration
- 🧠 **Memory efficient**: 0-3 allocations per log operation
- 🎯 **Flexible**: Text and JSON output formats with UTC or Local time zones
- 📊 **Structured logging**: Type-safe field constructors
- 🔧 **Configurable levels**: Debug, Info, Warn, Error, Fatal, Panic
- 🌐 **Context support**: TraceID and SpanID propagation
- 📦 **Buffering**: Optional buffering for cloud cost optimization
- ⚙️ **Env-driven config**: `ConfigFromEnv()` for 12-factor apps
- 🔒 **Thread-safe**: Concurrent logging support
- 🪶 **No runtime dependencies**: stdlib only (`testify` is used for tests)

## Installation

```bash
go get github.com/barnowlsnest/go-logslib/v2
```

Requires Go 1.27 or later.

## Quick Start

```go
package main

import (
	"os"

	"github.com/barnowlsnest/go-logslib/v2/pkg/logger"
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

	// Structured logging with typed field constructors
	log.Info("User logged in",
		logger.IntField("userID", 12345),
		logger.StringField("email", "user@example.com"),
	)
}
```

Full API documentation lives on
[pkg.go.dev](https://pkg.go.dev/github.com/barnowlsnest/go-logslib/v2).

## Fields

Use the typed constructors — they avoid mistakes at the call site and keep the
value in a type the formatters handle natively.

```go
logger.StringField(name string, value string)
logger.IntField(name string, value int)
logger.Int8Field(name string, value int8)
logger.Int16Field(name string, value int16)
logger.Int64Field(name string, value int64)
logger.UintField(name string, value uint)
logger.Uint8Field(name string, value uint8)
logger.Uint16Field(name string, value uint16)
logger.Uint32Field(name string, value uint32)
logger.Uint64Field(name string, value uint64)
logger.Float32Field(name string, value float32)
logger.Float64Field(name string, value float64)
logger.BoolField(name string, value bool)
logger.DurationField(name string, value time.Duration)
```

A `Field` is a plain struct, so `logger.Field{Key: "k", Value: v}` also works for
types without a constructor.

### Value type support

Text and JSON cover the same set of types:

| Value type                                               | Text format                 | JSON format                 |
|----------------------------------------------------------|-----------------------------|-----------------------------|
| `string`, `[]byte`                                       | ✅ (quoted if needed)        | ✅ (quoted)                  |
| `int`, `int8`, `int16`, `int32`, `int64`                 | ✅                           | ✅                           |
| `uint`, `uint8`, `uint16`, `uint32`, `uint64`, `uintptr` | ✅                           | ✅                           |
| `float32`, `float64`                                     | ✅ (`NaN`/`±Inf` as strings) | ✅ (`NaN`/`±Inf` as strings) |
| `bool`                                                   | ✅                           | ✅                           |
| `time.Duration`                                          | ✅ (`"1.5s"`)                | ✅ (`"1.5s"`)                |
| `time.Time`                                              | ✅ (RFC3339Nano)             | ✅ (RFC3339Nano)             |
| `error`                                                  | ✅ (`.Error()`)              | ✅ (`.Error()`)              |
| `nil`                                                    | `<nil>`                     | `null`                      |
| named types over the above kinds                         | ✅ (via reflection)          | ✅ (via reflection)          |
| anything else                                            | `"<unsupported>"`           | `"<unsupported>"`           |

## Shared Logger

The `sharedlog` package provides a zero-configuration singleton logger that
always outputs JSON in UTC to stdout. It reads `LOG_LEVEL` and `LOG_BUFFER_SIZE`
from the environment but ignores `LOG_FORMAT` and `LOG_USE_UTC`.

```go
package main

import (
	"errors"

	"github.com/barnowlsnest/go-logslib/v2/pkg/logger"
	"github.com/barnowlsnest/go-logslib/v2/pkg/sharedlog"
)

func main() {
	// Simple logging
	sharedlog.Info("Application started")
	sharedlog.Debug("cache miss", logger.StringField("key", "user:42"))

	// Error accepts an error directly
	err := errors.New("connection refused")
	sharedlog.Error(err, logger.StringField("host", "db.local"))

	// Access the underlying *logger.Logger for advanced use
	log := sharedlog.Logger()
	log.Warn("disk usage high", logger.IntField("pct", 92))
	log.Flush()
}
```

## Context Logging

`ContextLogger` extracts `traceID` and `spanID` from a `context.Context` and adds
them to every entry.

```go
// Fixed context
ctxLog := log.WithContext(ctx)
ctxLog.Info("handling request")

// Dynamic context, resolved at log time (HTTP handlers, middleware)
reqLog := log.WithContextFunc(func() context.Context { return r.Context() })
reqLog.Info("handling request")
```

Values must be stored under the package's unexported `contextKey` type, so use
the exported names `logger.FieldTraceID` / `logger.FieldSpanID` as the key
strings when your propagation layer sets them.

## Configuration

### Log Levels

```go
// Available log levels (in order of severity)
logger.DebugLevel  // -1: Detailed information for debugging
logger.InfoLevel   //  0: General information
logger.WarnLevel   //  1: Warning messages
logger.ErrorLevel  //  2: Error conditions
logger.FatalLevel  //  3: Fatal errors (calls os.Exit(1))
logger.PanicLevel  //  4: Panic conditions (calls panic())
```

`logger.LogLevelFromString("warn")` parses a level, returning
`logger.ErrUnknownLogLevel` for anything unrecognized.

### Output Formats

```go
// Text format (human-readable)
log := logger.New(logger.Config{
	Level:  logger.InfoLevel,
	Format: logger.TextFormat,
	Output: os.Stdout,
})
// Output: 2024-01-20T15:04:05.000Z INFO User action userID=12345 action=login

// JSON format (structured)
log := logger.New(logger.Config{
	Level:  logger.InfoLevel,
	Format: logger.JSONFormat,
	Output: os.Stdout,
})
// Output: {"timestamp":"2024-01-20T15:04:05.000Z","level":"INFO","message":"User action","userID":12345,"action":"login"}
```

`TextFormat` is the zero value, so an empty `Config` produces text output to
stdout at `InfoLevel`.

### Environment Configuration

`logger.ConfigFromEnv()` builds a `Config` from the environment. It does not set
`Output`, so `New()` defaults it to stdout.

| Variable          | Values                                          | Default  |
|-------------------|-------------------------------------------------|----------|
| `LOG_LEVEL`       | `debug`, `info`, `warn`, `error`, `fatal`, `panic` | `debug`  |
| `LOG_FORMAT`      | `json`, `text`                                  | `text`   |
| `LOG_BUFFER_SIZE` | integer (bytes); `0` disables buffering         | `0`      |
| `LOG_USE_UTC`     | `true`, `1`                                     | `false`  |

Unrecognized values fall back to the default rather than erroring.

```go
log := logger.New(logger.ConfigFromEnv())
```

### Buffering

Enable buffering for reduced I/O operations and cost optimization in cloud
environments:

```go
log := logger.New(logger.Config{
	Level:      logger.InfoLevel,
	Format:     logger.JSONFormat,
	Output:     os.Stdout,
	BufferSize: 4096, // Buffer up to 4KB before flushing
})

// Manually flush when needed
log.Flush()
```

Entries accumulate until the next one would overflow `BufferSize`, then the
buffer is written out. Call `Flush()` before exit so nothing is lost.

## Performance

Benchmarks on Apple M1 Max, Go 1.27 (`go test -bench=. -benchmem ./pkg/logger/`):

| Operation          | Time (ns/op) | Allocations | Memory (B/op) |
|--------------------|--------------|-------------|---------------|
| Simple text        | 182          | 2           | 33            |
| Simple JSON        | 220          | 2           | 33            |
| Text with fields   | 212          | 2           | 33            |
| JSON with fields   | 276          | 2           | 33            |
| Many fields (8)    | 454          | 2           | 33            |
| With context       | 385          | 3           | 161           |
| Buffered           | 257          | 1           | 40            |
| Concurrent         | 39           | 2           | 33            |
| Level filtering    | 2.8          | 0           | 0             |

Numbers vary with hardware; run the suite yourself before drawing conclusions.

## Versioning and Compatibility

The module follows [semantic versioning](https://semver.org) and is released as
a `/v2` module path. Patch and minor releases keep the exported API
backward-compatible; breaking changes wait for a new major version.

The following are deprecated and kept only for compatibility — avoid them in new
code:

- `logger.Logger.Panic()` — use `Error()` or `Fatal()`
- `logger.Logger.WithStaticContext()` — use `WithContext()`
- `sharedlog.Panic()` — use `sharedlog.Error()`
- `sharedlog.F()` — use the typed field constructors from `logger`

## Contributing

Contributions are welcome — issues, bug reports and pull requests all help.

1. Fork the repository and create a branch off `main`.
2. Make your change, with tests covering it.
3. Run the full check suite and make sure it passes:

   ```bash
   task sanity
   ```

4. Open a pull request describing the change and, for anything touching the hot
   path, include before/after benchmark output.

### Prerequisites

- Go 1.27 or later
- [Task](https://taskfile.dev) and `golangci-lint` for the full check suite

### Commands

```bash
task sanity     # tidy, fmt, lint, build, vet, test + benchmarks
task go-test    # tests with coverage, then benchmarks
task go-lint    # golangci-lint
```

Or with plain Go tooling:

```bash
go test ./...
go test -bench=. -benchmem ./...
```

### Performance Requirements

This library maintains strict performance requirements:
- **Execution time**: 20-600 ns/op
- **Memory allocations**: 0-5 per operation
- **Memory usage**: Minimal footprint

Benchmarks in `pkg/logger/benchmark_test.go` are part of the contract — changes
must not regress allocation counts or ns/op. Pull requests are checked by the
build and lint workflows in `.github/workflows/`.

## License

Released under the [MIT License](LICENSE). © Barn Owls Nest.

## Acknowledgments

- Optimized for high-load production environments
- Inspired by structured logging best practices
- Built with zero-allocation principles