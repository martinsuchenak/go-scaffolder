# Features

The scaffolder supports the following optional features. CLI is always included.

## CLI (always included)

Every generated project includes:

- `main.go` with `paularlott/cli` command framework
- `cmd/serve.go` -- the main `serve` sub-command
- `cmd/init.go` -- generates a default config file
- `cmd/completion.go` -- shell completion scripts (bash, zsh, fish)
- Configuration loading from TOML files, `.env` files, and environment variables
- `build/version.go` with `Version` and `Date` injected via `-ldflags`
- `Taskfile.yml` with build, test, and lint tasks (AMD64 + ARM64, CGO_ENABLED=0)
- `<app-name>-config.toml` with a `[log]` section

## API

HTTP REST API using Go's standard `net/http` with Go 1.22+ enhanced routing.

Generated files:

- `cmd/routes/api_routes.go` -- route registration with `GET /health` and `GET /metrics`
- `internal/rest/helpers.go` -- JSON response helpers
- `internal/auth/auth.go` -- API key and Bearer token middleware placeholders
- `internal/ctxkeys/ctxkeys.go` -- typed context keys
- `internal/sample/` -- sample resource with Handler, Service, Storage layers
- `openapi.yaml` -- OpenAPI v3.1 skeleton

Config additions: `[server]` section with `host` and `port`.

## MCP

Model Context Protocol server using `paularlott/mcp`.

Generated files:

- `cmd/mcp/mcp.go` -- MCP server with a sample tool, wired into the serve command

The MCP server starts on a separate port and exposes tools via HTTP.

## UI

Web frontend using Vite, TailwindCSS, AlpineJS, and TypeScript.

Generated files:

- `web/embed.go` -- `go:embed` for static assets and templates
- `web/src/main.ts` -- AlpineJS entry point
- `web/src/style.css` -- TailwindCSS import
- `web/templates/base.html` -- base HTML template
- `web/package.json` -- with bun as package manager
- `web/vite.config.ts` -- Vite configuration
- `web/dist/.gitkeep` -- placeholder until first frontend build

Config additions: `frontend-build` task added to `Taskfile.yml`.

When combined with Docker, the Dockerfile includes a frontend build stage.

## DB

Database persistence layer with driver selection.

| DB Type | Driver | go.mod dependency |
|---------|--------|-------------------|
| PostgreSQL | `lib/pq` | `github.com/lib/pq` |
| MySQL | `go-sql-driver/mysql` | `github.com/go-sql-driver/mysql` |
| SQLite | `modernc.org/sqlite` | `modernc.org/sqlite` |

Generated files:

- `internal/db/db.go` -- connection setup with SRV resolution
- `internal/db/schema.sql` -- sample schema with UUIDv7 primary keys

Config additions: `[database]` section.

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

## SRV Resolution

Not a user-selectable feature. Automatically included when any of **DB**, **Cache**, or **API** is enabled.

Generates `internal/resolve/` with:

- DNS SRV record lookup with fallback to original host
- Result caching with configurable TTL
- RFC 2782 weighted SRV record ordering
- `LookupSRV`, `LookupAllSRV`, `LookupIP`, `ResolveSRVAddress`, `ResolveSRVHttp` functions
- `srv+` URL prefix support for transparent SRV resolution

## Feature Dependencies

| If you select... | Then also included... |
|------------------|-----------------------|
| Nomad | Docker |
| DB, Cache, or API | SRV Resolution (`internal/resolve/`) |
| (any selection) | CLI (always) |
