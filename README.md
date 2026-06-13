# log

[![Quality](https://github.com/toaweme/log/actions/workflows/tests.yml/badge.svg)](https://github.com/toaweme/log/actions/workflows/tests.yml)
[![Go Reference](https://img.shields.io/badge/Docs-pkg.go.dev-blue)](https://pkg.go.dev/github.com/toaweme/log)
[![GitHub Tag](https://img.shields.io/github/v/tag/toaweme/log?label=Tag&color=green)](https://github.com/toaweme/log/releases)
[![License](https://img.shields.io/badge/License-MIT-blue)](/LICENSE)

## Simple slog wrapper

`github.com/toaweme/log` is a thin layer over the standard library's `log/slog`.
Everything is an ordinary `slog.Handler`, so it composes with the stdlib and any
other handler you already use. It has **zero dependencies**. It adds:

- `log.New(...)` assembles outputs, a level, and filters
  without hand-wiring handlers.
- **Filtering** - drop noisy records or shorten fat attribute values, by level,
  message, or attribute match (with `*` prefix wildcards).
- **Fan-out** - send one record to several outputs at once (console + file + ...).
- **Custom levels** - `TRACE` below `DEBUG` and `FATAL` above `ERROR`, rendered
  with their names instead of slog's numeric fallback.

```sh
go get github.com/toaweme/log
```

## Quick start

For app code that just wants to log, use the package-level helpers. They write
text to stdout at `DEBUG` out of the box, no setup:

```go
log.Info("server", "port", 8080)
log.Error("request", "err", err)
log.Trace("entered", "i", i)

log.SetLevel(slog.LevelInfo) // raise the threshold
```

When you're ready to inject a logger instead of reaching for the global, build
one with `log.New`.

## Build a logger: `log.New`

`log.New` takes a handful of options and assembles the handlers for you. With no
options it writes text to stdout at `DEBUG`.

```go
logger := log.New(
    log.WithText(os.Stdout),         // text output
    log.WithLevel(slog.LevelInfo),
)

logger.Info("ready")
logger = logger.With("svc", "api") // every record now carries svc=api
```

`log.Logger` is the interface you pass around. It is itself a `slog.Handler`, so
it drops into anything that expects one.

| Option | What it adds |
| --- | --- |
| `log.WithText(w)` | a text handler writing to `w` |
| `log.WithJSON(w)` | a JSON handler writing to `w` |
| `log.WithOutput(h)` | any `slog.Handler` you already have (memory sink, exporter, ...) |
| `log.WithLevel(l)` | minimum level for the `Text`/`JSON` outputs (default `DEBUG`) |
| `log.WithFilters(f...)` | wraps every output in a `FilteredLogger` |

Pass as many outputs as you like; they fan out automatically.

## Recipes

### Console + rotating file

This package never imports a rotation library, so it stays dependency-free.
`log.WithJSON` takes an `io.Writer`, so pass your own rotating writer (here
[lumberjack](https://github.com/natefinch/lumberjack)) to it:

```go
logger := log.New(
    log.WithText(os.Stdout),
    log.WithJSON(&lumberjack.Logger{
        Filename:   "/var/log/app.log",
        MaxSize:    20, // MB
        MaxBackups: 5,
        Compress:   true,
    }),
)
```

Human-readable text on the console, structured JSON in a rotated file, from one
logger.

### Make it the global, for the package helpers

Build the logger you want once at startup and install it, so `log.Info` and
friends route through it:

```go
func setupLogging(path string) {
    logger := log.New(
        log.WithText(os.Stdout),
        log.WithJSON(&lumberjack.Logger{Filename: path, MaxSize: 20, MaxBackups: 5, Compress: true}),
        log.WithFilters(
            log.Deny().Attr("path", "/healthz*"),
        ),
    )
    log.SetDefault(logger)
}
```

After `SetDefault`, `log.Info(...)` writes to both outputs and obeys the filters.

### Console + an in-memory sink (e.g. a live log view)

Use `log.WithOutput` to add any handler you have, like one that pushes records to
subscribers for a UI:

```go
mem := NewMemoryHandler(subscribers...) // your own slog.Handler

logger := log.New(
    log.WithText(os.Stdout),
    log.WithOutput(mem),
).With("pid", os.Getpid())
```

> Building a raw handler yourself? Pass `log.Options(level)` as its
> `*slog.HandlerOptions` so it renders the custom `TRACE`/`FATAL` level names
> the same way `Text`/`JSON` do.

### Inject the logger into your types

Depend on the `log.Logger` interface, not a global. It keeps types testable and
mockable:

```go
type Server struct {
    log log.Logger
}

func NewServer(l log.Logger) *Server {
    return &Server{log: l.With("component", "server")}
}

func (s *Server) handle() {
    s.log.Debug("handling request")
}
```

Pass `log.New(...)` in production and `log.Default()` (or a buffer-backed
`log.New(log.WithText(&buf))`) in tests.

## Filtering

`log.WithFilters` wraps your outputs in a `FilteredLogger` that runs an ordered
list of filters over each record. Filters are built fluently:

```go
log.New(
    log.WithText(os.Stdout),
    log.WithFilters(
        // drop everything below Info (a level floor)
        log.Deny().Below(slog.LevelInfo),
        // drop health-check noise; * is a prefix match
        log.Deny().Attr("path", "/healthz*"),
        // truncate fat response bodies to 200 chars
        log.Shorten("body").Limit(200).Message("http response"),
    ),
)
```

Builders:

- `log.Deny()` drops matching records.
- `log.Allow()` passes matching records through unchanged.
- `log.Shorten(keys...)` truncates the given attribute values (default limit 100,
  change with `.Limit(n)`).

Match criteria (chain as many as you need; **all** must match):

- `.Message("...")` - exact message match.
- `.Attr(key, val)` - attribute equals `val`; a `val` ending in `*` is a prefix
  match. The record's message is available under the synthetic `"msg"` key.
- `.Below(level)` - matches records *strictly below* `level`. Paired with `Deny`
  it acts as a floor.

Filters can be changed at runtime on a `*FilteredLogger` via `AddFilter` and
`SetFilters`; both are safe to call while logging.

## Fan-out and custom levels directly

The primitives `log.New` builds on are exported for hand-assembly:

```go
// fan one record out to several handlers
multi := log.NewMultiHandler(
    slog.NewTextHandler(os.Stdout, log.Options(slog.LevelDebug)),
    slog.NewJSONHandler(file, log.Options(slog.LevelDebug)),
)

// wrap any handler in filters
filtered := log.NewFilteredLogger(multi, log.Deny().Below(slog.LevelInfo))

logger := log.Wrap(slog.New(filtered)) // adopt an existing *slog.Logger
```

A `MultiHandler` drops a record only when *every* child would discard it, and one
failing output does not stop the others (errors are joined). `log.Wrap` adopts
any `*slog.Logger` as a `log.Logger`; `logger.Slog()` gets the `*slog.Logger`
back.

The custom levels are `log.LevelTrace` (below `DEBUG`) and `log.LevelFatal`
(above `ERROR`). Every `log.Logger` has `Trace`/`Fatal` helpers, and
`WithLevel` returns a logger at a new threshold while keeping the same outputs:

```go
quiet := logger.WithLevel(slog.LevelError) // same outputs, higher threshold
```

## Opinions

This package makes choices the standard library leaves open. They suit its
intended use (desktop apps and servers you own end to end).

- **`Fatal` does not exit.** It logs a `FATAL` record and returns. `slog` itself
  ships no `Fatal`, and `os.Exit` inside a logging call skips deferred cleanup
  and unflushed writers, including the FATAL record itself. If you want to die,
  call `os.Exit(1)` yourself, after the record is flushed or shipped.
- **`Filter.Below` is a floor, not a ceiling.** It matches records *below* the
  given level. See [Filtering](#filtering).
- **There is a global logger.** Created in `init`, writing text to stdout. It is
  there for convenience; prefer injecting `log.Logger` in code you care about and
  treat the global as a quick-start. `log.SetLevel` only moves the built-in
  default; once you `SetDefault` your own logger, set its level when you build it.
