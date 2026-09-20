# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

High-performance Go logging library (`github.com/barnowlsnest/go-logslib/v2`) targeting production backend services. Prioritizes minimal memory allocations (0-5 per op) and fast execution (20-600 ns/op) over features. MIT licensed, published on GitHub as `barnowlsnest/go-logslib`. Requires Go 1.27; the only dependency is `stretchr/testify` (tests only).

## Development Commands

```bash
task sanity     # tidy, fmt, lint, build, vet, test + benchmarks — run after every change
task go-test    # go test -cover ./... then go test -bench=. ./...
task go-lint    # golangci-lint run
task go-build   # go build ./...
```

Plain Go tooling when `task` is unavailable:

```bash
go build ./...
go test ./...
go test -bench=. -benchmem ./...
go vet ./...
go fmt ./...

# Single test / benchmark
go test -v -run TestName ./pkg/logger/
go test -bench=BenchmarkName -benchmem ./pkg/logger/
```

CI runs two workflows on PRs and pushes to `main`: `.github/workflows/build.yml` (build + test via Task) and `.github/workflows/golangci-lint.yml` (golangci-lint v2, config in `.golangci.yml`).

## Architecture

Two packages under `pkg/`:

### `pkg/logger` — Core library
- **`logger.go`** — `Logger`, `Config`, `Level`, `Format`, `Field` types; typed field constructors (`StringField`, `IntField`, `Int8/16/64Field`, `Uint/Uint8/16/32/64Field`, `Float32/64Field`, `BoolField`, `DurationField` — note there is no `Int32Field`); text formatting and value serialization (`appendValue`, `appendString`, `appendInt`, `appendUint`, `appendFloat`, `appendDuration`); `ContextLogger` for trace/span propagation. Uses `sync.Pool` for buffer reuse and `sync.Mutex` for buffered writes.
- **`json.go`** — JSON formatting with hand-rolled serialization (no `encoding/json`). Type coverage mirrors the text path exactly; the two share `appendReflectValue` via format-specific string/nil hooks.
- **`env.go`** — `ConfigFromEnv()` reads `LOG_LEVEL` (default `debug`), `LOG_FORMAT` (default `text`), `LOG_BUFFER_SIZE` (default `0`), `LOG_USE_UTC` (default `false`). Unknown values fall back to the default instead of erroring. `Output` is left nil so `New()` defaults it to stdout.

### `pkg/sharedlog` — Singleton convenience wrapper
- **`log.go`** — `sync.Once` singleton over `logger.Logger`. Always JSON + UTC. Reads `LOG_LEVEL` and `LOG_BUFFER_SIZE` from env but overrides `LOG_FORMAT` and `LOG_USE_UTC`. `Error()` and `Panic()` accept `error` (not `string`) as first arg. `Logger()` exposes the underlying `*logger.Logger`.

## Key Design Patterns

- **Zero-allocation logging**: All formatting uses `[]byte` append operations instead of `fmt.Sprintf` or `encoding/json`.
- **Value handling**: `appendValue` / `appendJSONValue` type-switch over `string`, `[]byte`, all signed and unsigned integer widths, `uintptr`, `float32/64`, `bool`, `time.Duration`, `time.Time`, `error` and `nil`. Named types over those kinds (e.g. `type Status string`) fall through to a shared `reflect` path; anything else renders as `"<unsupported>"`. Non-finite floats render as `"NaN"` / `"+Inf"` / `"-Inf"` strings in both formats. Keep the two type switches in sync — text and JSON must agree.
- **Buffer pooling**: `sync.Pool` of `*[]byte` avoids per-call allocations.
- **Optional buffering**: When `Config.BufferSize > 0`, entries accumulate and flush when the next entry would overflow the buffer, or on `Flush()`.
- **Context logging**: `WithContextFunc(func() context.Context)` for dynamic contexts (HTTP handlers), `WithContext(ctx)` for a fixed context. Extracts `traceID` and `spanID` via the unexported `contextKey` type; `FieldTraceID` / `FieldSpanID` are the exported key strings.
- **Deprecations kept for compatibility**: `Logger.Panic()`, `Logger.WithStaticContext()`, `sharedlog.Panic()`, `sharedlog.F()`. Don't use them in new code or examples, and don't remove them without a major version bump.

## Public API Stability

This is a published, tagged module (latest tag `v2.2.0`) consumed by outside users. Treat exported identifiers as API: additive changes are fine, renames and removals need a major version. Update `README.md` whenever exported API, env vars, or benchmark numbers change.

## Testing

Uses `github.com/stretchr/testify`. Benchmark tests in `pkg/logger/benchmark_test.go` are critical — any change must not regress allocation counts or ns/op.