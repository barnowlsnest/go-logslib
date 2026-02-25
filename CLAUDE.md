# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

High-performance Go logging library (`github.com/barnowlsnest/go-logslib`) targeting production backend services. Prioritizes minimal memory allocations (0-5 per op) and fast execution (20-600 ns/op) over features.

## Development Commands

```bash
go build ./...          # Build
go test ./...           # Run tests
go test -v ./...        # Verbose tests
go test -bench=. ./...  # Benchmarks (critical for this project)
go test -cover ./...    # Coverage
go fmt ./...            # Format
go vet ./...            # Static analysis
go mod tidy             # Tidy deps

# Run a single test
go test -v -run TestName ./pkg/logger/

# Run a single benchmark
go test -bench=BenchmarkName ./pkg/logger/
```

## Architecture

Two packages under `pkg/`:

### `pkg/logger` — Core library
- **`logger.go`** — `Logger` struct, `Config`, `Level`, `Format`, `Field` types, text formatting, and `ContextLogger` for trace/span propagation. Uses `sync.Pool` for buffer reuse and `sync.Mutex` for buffered writes.
- **`json.go`** — JSON formatting with hand-rolled serialization (no `encoding/json`) for zero-allocation output.
- **`env.go`** — `ConfigFromEnv()` reads `LOG_LEVEL`, `LOG_BUFFER_SIZE`, `LOG_FORMAT`, `LOG_USE_UTC` environment variables.

### `pkg/sharedlog` — Singleton convenience wrapper
- **`log.go`** — `sync.Once` singleton over `logger.Logger`. Always JSON + UTC. Reads `LOG_LEVEL` and `LOG_BUFFER_SIZE` from env but ignores `LOG_FORMAT` and `LOG_USE_UTC`. `Error()` and `Panic()` accept `error` (not `string`) as first arg.

## Key Design Patterns

- **Zero-allocation logging**: All formatting uses `[]byte` append operations instead of `fmt.Sprintf` or `encoding/json`. Field values are type-switched (`string`, `int`, `int64`, `float64`, `bool`).
- **Buffer pooling**: `sync.Pool` of `*[]byte` avoids per-call allocations.
- **Optional buffering**: When `Config.BufferSize > 0`, log entries accumulate in an internal buffer and flush when full or on `Flush()`.
- **Context logging**: `WithContext(func() context.Context)` for dynamic contexts (HTTP handlers), `WithStaticContext(ctx)` for fixed contexts. Extracts `traceID` and `spanID` via `contextKey` typed keys.

## Testing

Uses `github.com/stretchr/testify`. Benchmark tests in `pkg/logger/benchmark_test.go` are critical — any change must not regress allocation counts or ns/op.