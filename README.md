# log

[![Quality](https://github.com/toaweme/log/actions/workflows/quality.yml/badge.svg)](https://github.com/toaweme/log/actions/workflows/quality.yml)
<a href="https://code.toawe.me/toaweme/log/health">
    <picture>
        <source media="(prefers-color-scheme: dark)" srcset="https://code.toawe.me/toaweme/log/badge-dark.svg">
        <source media="(prefers-color-scheme: light)" srcset="https://code.toawe.me/toaweme/log/badge.svg">
        <img alt="log health" src="https://code.toawe.me/toaweme/log/badge.svg">
    </picture>
</a>
[![Go Reference](https://img.shields.io/badge/Docs-pkg.go.dev-blue)](https://pkg.go.dev/github.com/toaweme/log)
[![GitHub Tag](https://img.shields.io/github/v/tag/toaweme/log?label=Tag&color=green)](https://github.com/toaweme/log/releases)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue)](/LICENSE)

## Simple slog wrapper

`github.com/toaweme/log` is a thin layer over the standard library's `log/slog`.
Outputs are ordinary `slog.Handler`s, so it composes with the stdlib and any
handler you already use. It has **zero dependencies**. It adds:

- `log.New(...)` assembles outputs without hand-wiring handlers.
- **Fan-out** - send one record to several outputs at once (console + file + ...).
- **One process-wide level** - `SetLevel`, `Level` and `New` share it.
- **`log.ParseLevel(s)`** - `LOG_LEVEL` string to `slog.Level`.
- **A `TRACE` level** below `DEBUG`, rendered by name.
- **`log.Discard()`** - a silent `log.Logger` for tests and libraries that should
  produce no output.

**[Documentation](https://toawe.me/docs/log)** | [toawe.me](https://toawe.me)

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

Read the threshold at boot:

```go
lvl, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
if err != nil {
    lvl = slog.LevelInfo
}
log.SetLevel(lvl)
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

`log.Logger` is the interface you pass around. `logger.Slog()` returns the
underlying `*slog.Logger`.

| Option | What it adds |
| --- | --- |
| `log.WithText(w)` | a text handler writing to `w` |
| `log.WithJSON(w)` | a JSON handler writing to `w` |
| `log.WithOutput(h)` | any `slog.Handler` you already have (memory sink, exporter, ...) |
| `log.WithLevel(l)` | sets the process-wide minimum level (default `DEBUG`) |

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
    )
    log.SetDefault(logger)
}
```

After `SetDefault`, `log.Info(...)` writes to both outputs, and `log.SetLevel`
still moves the threshold they share.

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

> Building a raw handler yourself? Pass `log.HandlerOptions(log.Level())` as its
> `*slog.HandlerOptions` so it renders the custom `TRACE` level name the same way
> `Text`/`JSON` do.

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

### A silent logger: `log.Discard()`

When a test or a library just needs a `log.Logger` that produces no output, use
`log.Discard()`. It drops every record.

```go
srv := NewServer(log.Discard()) // logs nothing

func TestThing(t *testing.T) {
    thing := New(log.Discard()) // keep test output clean
    // ...
}
```

It is the idiomatic null logger for this package, the equivalent of wiring up
`slog.New(slog.DiscardHandler)` yourself.

## Fan-out and custom levels directly

The primitives `log.New` builds on are exported for hand-assembly:

```go
// fan one record out to several handlers
multi := log.NewMultiHandler(
    slog.NewTextHandler(os.Stdout, log.HandlerOptions(log.Level())),
    slog.NewJSONHandler(file, log.HandlerOptions(log.Level())),
)

logger := log.Wrap(slog.New(multi)) // adopt an existing *slog.Logger
```

A `MultiHandler` drops a record only when *every* child would discard it, and one
failing output does not stop the others (errors are joined). `log.Wrap` adopts
any `*slog.Logger` as a `log.Logger`; `logger.Slog()` gets the `*slog.Logger`
back.

The custom level is `log.LevelTrace`, below `DEBUG`. Every `log.Logger` has a
`Trace` helper alongside the usual four.

## Opinions

- **There is one level for the whole process.** `log.SetLevel`, `log.Level` and
  `log.WithLevel` all move and read the same value, and every logger `log.New`
  builds reads it too, so `log.Level()` never reports a threshold that is not in
  effect. Set it once at boot from `LOG_LEVEL`. The exception is a handler you
  hand to `log.WithOutput`, which keeps whatever level it was built with.
- **There is a global logger.** Built on the first call to `log.Default()`,
  writing text to stdout, so importing the package opens nothing. It is there for
  convenience; prefer injecting `log.Logger` in code you care about.
- **There is no `Fatal`.** A logging call that exits skips deferred cleanup and
  unflushed writers, and one that does not exit is a trap for anyone reading the
  name. Log at `Error` and call `os.Exit(1)` yourself.

## Contributing

`log` uses an issue-first workflow. Open an issue describing the change and wait for a maintainer to approve the approach (the `approved` label) before you open a pull request. PRs that don't reference an approved issue are flagged by a bot and usually closed, so the issue step saves you wasted work.

Every commit must be signed off for the [Developer Certificate of Origin](https://developercertificate.org/) with `git commit -s`. A CI check enforces this on every commit in a pull request.

Full flow in [CONTRIBUTING.md](CONTRIBUTING.md), ground rules in [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).

## Hosted code and health reports

Reports for this repo are hosted by our <a href="https://code.toawe.me">code viewer</a>, which also serves the badges and cards above.

<p align="center">
  <a href="https://code.toawe.me/toaweme/log/health"><picture><source media="(prefers-color-scheme: dark)" srcset="https://code.toawe.me/toaweme/log/card-dark.svg"><source media="(prefers-color-scheme: light)" srcset="https://code.toawe.me/toaweme/log/card.svg"><img alt="log health" src="https://code.toawe.me/toaweme/log/card.svg" width="48%"></picture></a>
  <a href="https://code.toawe.me/toaweme/log/code"><picture><source media="(prefers-color-scheme: dark)" srcset="https://code.toawe.me/toaweme/log/code-card-dark.svg"><source media="(prefers-color-scheme: light)" srcset="https://code.toawe.me/toaweme/log/code-card.svg"><img alt="log code" src="https://code.toawe.me/toaweme/log/code-card.svg" width="48%"></picture></a>
</p>

---

Made with ❤️ in Lithuania 🇱🇹.

