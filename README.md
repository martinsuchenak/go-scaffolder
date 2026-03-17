# Go Scaffolder

A CLI tool that generates fully functional Go microservice projects. Ships as a single binary with all templates embedded via `go:embed`.

## Features

- **Interactive mode** — prompts for app name, output directory, and feature selection
- **Config file mode** — reads all parameters from a YAML file (`--config`) for scripted/CI usage
- **Selectable features** — CLI, API, MCP, UI, DB, Cache, Docker, Nomad
- **Database support** — PostgreSQL, MySQL, SQLite with driver wiring and schema generation
- **Cache support** — Redis or Valkey with client initialization
- **DNS SRV resolution** — auto-generated `internal/resolve/` package when networking features are enabled
- **Test file generation** — every source file gets a corresponding `_test.go`
- **Post-generation** — runs `go mod tidy` to resolve all dependencies

## Requirements

- Go 1.26+
- [Task](https://taskfile.dev) (optional, for build commands)

## Installation

### Homebrew

```sh
brew install martinsuchenak/tap/go-scaffolder
```

### From source

```sh
git clone <repo-url>
cd go-scaffolder
task build
```

The binary is placed in `bin/go-scaffolder`.

## Usage

### Interactive mode

```sh
bin/go-scaffolder
```

You will be prompted for:
1. App name
2. Module path (e.g. `github.com/yourorg/my-service`, defaults to app name)
3. Output directory
4. Features (API, MCP, UI, DB, Cache, Docker, Nomad)
5. DB type (if DB selected)
6. Cache type (if Cache selected)

### Config file mode

```sh
bin/go-scaffolder --config scaffold.yaml
```

Example `scaffold.yaml`:

```yaml
app_name: my-service
module_path: github.com/yourorg/my-service   # optional, defaults to app_name
output_dir: ./output
features:
  - api
  - db
  - cache
  - docker
db_type: postgresql
cache_type: redis
```

### Available features

| Feature | Description |
|---------|-------------|
| API | HTTP REST API with health/metrics endpoints, sample CRUD, auth middleware |
| MCP | Model Context Protocol server with a sample tool |
| UI | Vite + TailwindCSS + AlpineJS + TypeScript frontend with `go:embed` |
| DB | Database layer with PostgreSQL, MySQL, or SQLite support |
| Cache | Redis or Valkey client integration |
| Docker | Multi-stage Dockerfile (includes frontend build stage when UI enabled) |
| Nomad | HashiCorp Nomad job definition (automatically includes Docker) |

CLI is always included regardless of selection.

## Adding components to an existing project

After scaffolding, you can add new components without re-scaffolding. Run these from the project root (where `.go-scaffolder.yaml` is):

```sh
# Add a new CLI command
go-scaffolder add cli-command --name migrate

# Add a new API endpoint (requires API feature)
go-scaffolder add api-endpoint --name user

# Add a new MCP tool (requires MCP feature)
go-scaffolder add mcp-tool --name search
```

Each command generates all necessary files (handler, service, storage, routes, tests) and lists them on output.

## Enabling features on an existing project

Features can be enabled after initial scaffolding:

```sh
# Interactive — select from features not yet enabled
go-scaffolder add feature

# Non-interactive
go-scaffolder add feature --db-type postgresql
go-scaffolder add feature --cache-type redis
```

This creates feature-specific files and patches shared files (`cmd/serve.go`, config TOML, `Taskfile.yml`) using marker comments. If a marker was removed, the tool prints the code to add manually. All created and updated files are listed on output.

## Patch mode (for LLM / MCP integration)

Both scaffolding and all `add` subcommands support `--patch` to output unified diffs to stdout instead of writing files:

```sh
# Full project as patch
go-scaffolder --config scaffold.yaml --patch

# Add components as patch
go-scaffolder add --patch cli-command --name migrate
go-scaffolder add --patch api-endpoint --name user
go-scaffolder add --patch feature --db-type postgresql
```

No files are written, no `go mod tidy` is run. The output is standard unified diff format consumable by `git apply`, `patch`, or any LLM file editor.

## MCP server mode

The scaffolder can run as a remote MCP server, exposing all operations as tools for LLM consumption:

```sh
# Start MCP server on default port
go-scaffolder serve

# Custom listen address
go-scaffolder serve --listen :9090

# With bearer token authorization
go-scaffolder serve --token my-secret-token

# Or via environment variable
MCP_TOKEN=my-secret-token go-scaffolder serve
```

When `--token` is set (or `MCP_TOKEN` env var), all requests must include an `Authorization: Bearer <token>` header. Without it, no auth is required.

### Available MCP tools

| Tool | Description |
|------|-------------|
| `scaffold` | Scaffold a new project (returns unified diff) |
| `add_cli_command` | Add a CLI command to an existing project |
| `add_api_endpoint` | Add an API endpoint resource |
| `add_mcp_tool` | Add an MCP tool |
| `enable_feature` | Enable a feature (api, mcp, ui, db, cache, docker, nomad) |
| `project_context` | Generate project context summary for LLM consumption |

All tools return unified diff output. The server uses the `github.com/paularlott/mcp` library and serves on the `/mcp` endpoint.

### Project context for remote usage

Tools that operate on existing projects (all except `scaffold`) accept two optional parameters for locating the project:

- **`project_dir`** -- absolute path to the project root (where `.go-scaffolder.yaml` lives). The server reads the state file and resolves file paths from this directory.
- **`state_file`** -- the content of `.go-scaffolder.yaml` as a string. Use this for fully remote operation where the server has no filesystem access to the project.

If neither is provided, the server looks for `.go-scaffolder.yaml` in its working directory. When `state_file` is used without `project_dir`, patch operations on existing files (marker-based patching for `enable_feature`) output the code snippets to apply manually since the server cannot read the project files.

## Generated project structure

```
my-service/
├── main.go                  # CLI entry point using paularlott/cli
├── go.mod
├── build/version.go         # Build version/date injected via ldflags
├── Taskfile.yml             # Build, test, lint tasks
├── my-service-config.toml   # TOML config with feature-specific sections
├── cmd/
│   ├── register.go          # Command registration
│   ├── serve.go             # Serve command
│   ├── init.go              # Config file initialization
│   └── completion.go        # Shell completions
├── cmd/routes/              # API routes (when API enabled)
├── cmd/mcp/                 # MCP server (when MCP enabled)
├── internal/
│   ├── rest/                # HTTP helpers (when API enabled)
│   ├── auth/                # Auth middleware (when API enabled)
│   ├── ctxkeys/             # Typed context keys (when API enabled)
│   ├── sample/              # Sample Handler→Service→Storage (when API enabled)
│   ├── db/                  # DB init + schema (when DB enabled)
│   ├── redis/               # Redis client (when Cache=Redis)
│   ├── valkey/              # Valkey client (when Cache=Valkey)
│   └── resolve/             # DNS SRV resolution (when DB, Cache, or API enabled)
├── web/                     # Frontend project (when UI enabled)
├── openapi.yaml             # OpenAPI v3.1 spec (when API enabled)
├── Dockerfile               # Multi-stage build (when Docker enabled)
└── my-service.nomad         # Nomad job (when Nomad enabled)
```

## Development

### Build

```sh
task build
```

### Run tests

```sh
task test              # Unit and property tests
task test-integration  # Integration tests (scaffolds and compiles real projects)
task test-all          # Both
```

### Lint

```sh
task lint
```

### Clean

```sh
task clean
```

## Testing approach

- **Property-based tests** (using [rapid](https://github.com/flyingmutant/rapid)) verify invariants across random valid configurations: feature resolution, file inclusion, dependency correctness, config sections, test parity, and more
- **Unit tests** cover validation, file writing, and post-generation
- **Integration tests** scaffold projects with various feature combinations and verify they compile, pass `go vet`, and pass `go test`

## License

See LICENSE file.
