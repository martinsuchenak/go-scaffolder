# Features

The scaffolder supports the following optional features. CLI is always included.

## CLI (always included)

Every generated project includes:

- `main.go` with `paularlott/cli` command framework
- `cmd/register.go` -- command registry with self-registration support
- `cmd/serve.go` -- the main `serve` sub-command (self-registers via `init()`)
- `cmd/init.go` -- generates a default config file (self-registers via `init()`)
- `cmd/completion.go` -- shell completion scripts (self-registers via `init()`)
- Configuration loading from TOML files, `.env` files, and environment variables
- `build/version.go` with `Version` and `Date` injected via `-ldflags`
- `Taskfile.yml` with `build`/`build-local` (AMD64 + ARM64, CGO_ENABLED=0), `test`, `test:race`, `lint`, `vet`, `tidy`, `check` (vet+lint+test:race), and `release` (tag + push, gated on `check`+`build`)
- `<app-name>-config.toml` with a `[log]` section

New CLI commands can be added post-scaffold with `go-scaffolder add cli-command`.

## API

HTTP REST API using Go's standard `net/http` with Go 1.22+ enhanced routing.

Generated files:

- `cmd/routes/api_routes.go` -- route registry with `GET /health`, `GET /metrics`, and self-registration support
- `cmd/routes/sample_routes.go` -- sample resource routes (self-registers via `init()`)
- `internal/rest/helpers.go` -- JSON response helpers
- `internal/auth/auth.go` -- API key and Bearer token middleware placeholders
- `internal/ctxkeys/ctxkeys.go` -- typed context keys
- `internal/sample/` -- sample resource with Handler, Service, Storage layers
- `openapi.yaml` -- OpenAPI v3.1 skeleton

Config additions: `[server]` section with `host` and `port`.

New API endpoints can be added post-scaffold with `go-scaffolder add api-endpoint`.

## MCP

Model Context Protocol server using `paularlott/mcp`.

Generated files:

- `cmd/mcp/mcp.go` -- MCP server with tool registry and self-registration support
- `cmd/mcp/sample_tool.go` -- sample tool (self-registers via `init()`)

The MCP server starts on a separate port and exposes tools via HTTP.

New MCP tools can be added post-scaffold with `go-scaffolder add mcp-tool`.

## UI

Web frontend using Vite, TailwindCSS, AlpineJS, and TypeScript.

**Automatically enables API** -- the built frontend is only reachable through the API's HTTP server, so selecting UI always pulls in the API feature too (selecting UI alone generates both). This mirrors how Nomad always pulls in Docker.

Generated files:

- `cmd/routes/ui_routes.go` -- serves `web/dist` at `/static/` and renders `web/templates/base.html` at `GET /` (self-registers via `init()`)
- `web/embed.go` -- `go:embed` for static assets and templates
- `web/src/main.ts` -- AlpineJS entry point, with `@alpinejs/focus` for modal focus-trapping and system/dark/light theme handling (`applyTheme`/`cycleTheme`, persisted to `localStorage`)
- `web/src/style.css` -- TailwindCSS import, `dark:` variant wired to a `.dark` class on `<html>`
- `web/templates/base.html` -- base HTML template with a pre-paint theme script (avoids a flash of the wrong theme) and a theme-toggle button
- `web/package.json` -- bun as package manager, `alpinejs` + `@alpinejs/focus`, Tailwind Vite plugin, and `@types/alpinejs`/`@types/alpinejs__focus` dev dependencies
- `web/tsconfig.json` -- strict TypeScript config for the frontend
- `web/vite.config.ts` -- Vite configuration with `@tailwindcss/vite`, `emptyOutDir: true`, and flat `dist/app.js`/`dist/style.css` output (no hashed `assets/` subfolder, matching what `web/embed.go` embeds and `base.html` references)
- `web/dist/.gitkeep` -- placeholder until first frontend build

Config additions:

- `Taskfile.yml`'s `build` and `build-local` tasks depend on a new `frontend-build` task, since the frontend is embedded into the Go binary and must exist before compiling
- `frontend-build` passes the same git-derived `VERSION` used for the Go binary's `-ldflags` into the frontend build as `VITE_APP_VERSION`, so `import.meta.env.VITE_APP_VERSION` in `main.ts` matches `build.Version` at runtime -- `package.json`'s own `"version"` field is left untouched (no file mutation on every build)

When combined with Docker, the Dockerfile includes a frontend build stage.

## DB

Database persistence layer with driver selection.

| DB Type | Driver | go.mod dependency |
|---------|--------|-------------------|
| PostgreSQL | `lib/pq` | `github.com/lib/pq` |
| MySQL | `go-sql-driver/mysql` | `github.com/go-sql-driver/mysql` |
| SQLite | `modernc.org/sqlite` | `modernc.org/sqlite` |

Generated files:

- `internal/db/db.go` -- connection setup with SRV resolution, optionally wrapped with `xdal`
- `internal/db/schema.sql` -- sample schema with UUIDv7 primary keys

Config additions: `[database]` section.

When DB is selected interactively, the scaffolder asks whether the database may need to be swapped later. If yes, it can generate the DB layer on top of [`github.com/martinsuchenak/xdal`](https://github.com/martinsuchenak/xdal), exposing both the raw `*sql.DB` connection and an `xdal.DBInterface`.

## Cache

Cache client integration. Redis and Valkey are mutually exclusive.

| Cache Type | Library |
|------------|---------|
| Redis | `github.com/redis/go-redis/v9` |
| Valkey | `github.com/valkey-io/valkey-go` |

Generated files:

- `internal/redis/redis.go` or `internal/valkey/valkey.go` -- client initialization with SRV resolution

Config additions: `[redis]` or `[valkey]` section.

## Docker

Multi-stage Dockerfile for containerization.

- `CGO_ENABLED=0` static binary
- When UI is also enabled, includes a frontend build stage using bun

## Nomad

HashiCorp Nomad job definition.

- Generates `<app-name>.nomad` with Docker driver configuration
- Includes health check via `/health` endpoint
- **Automatically includes Docker** -- selecting Nomad always enables Docker

## CI

GitHub Actions workflow for continuous integration.

Generated files:

- `.github/workflows/ci.yml` -- Workflow that runs `task lint`, `task test`, and optionally builds the Docker image (if the `docker` feature is enabled).

## Otel

OpenTelemetry initialization for distributed tracing and metrics.

Generated files:

- `internal/otel/otel.go` -- Initializes OpenTelemetry provider with `otlptracegrpc` and `otlpmetricgrpc` exporters.

Configures the `serve` command to automatically initialize OpenTelemetry on startup and gracefully shut it down on exit.

## SRV Resolution

Not a user-selectable feature. Automatically included when any of **DB**, **Cache**, or **API** is enabled.

Uses the external library `github.com/fortix/go-libs/netx/dns` for DNS SRV record resolution with:

- DNS SRV record lookup with fallback to original host
- Result caching with configurable TTL
- RFC 2782 weighted SRV record ordering
- `srv+` URL prefix support for transparent SRV resolution

Generated code initializes `dns.NewDNSResolver` in DB and Cache connection setup.

## Feature Dependencies

| If you select... | Then also included... |
|------------------|-----------------------|
| UI | API |
| Nomad | Docker |
| DB, Cache, or API | SRV Resolution (`internal/resolve/`) |
| (any selection) | CLI (always) |
