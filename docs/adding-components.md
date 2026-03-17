# Adding Components to Existing Projects

After scaffolding a project, you can add new CLI commands, API endpoints, and MCP tools without re-scaffolding. The `add` subcommand generates only the new files and never modifies existing ones.

## Prerequisites

The project must have been scaffolded with go-scaffolder (a `.go-scaffolder.yaml` state file must exist in the project root). Run the `add` command from the project root directory.

## Self-Registration Pattern

Generated projects use a self-registration pattern. Each CLI command, API route group, and MCP tool registers itself via Go's `init()` function, so adding a new component never requires editing existing files -- only new files are dropped in.

### CLI Commands

```go
// cmd/registry.go -- registry (generated once)
var registry []*cli.Command
func Register(c *cli.Command) { registry = append(registry, c) }
func Commands() []*cli.Command { return registry }

// cmd/serve.go -- self-registers via init()
func init() { Register(serveCmd()) }
```

### API Routes

```go
// cmd/routes/api_routes.go -- registry + health/metrics (generated once)
var registrations []func(*http.ServeMux)
func Register(fn func(*http.ServeMux)) { registrations = append(registrations, fn) }
func RegisterRoutes(mux *http.ServeMux) {
    mux.HandleFunc("GET /health", healthHandler)
    mux.HandleFunc("GET /metrics", metricsHandler)
    for _, fn := range registrations { fn(mux) }
}

// cmd/routes/sample_routes.go -- self-registers via init()
func init() { Register(registerSampleRoutes) }
```

### MCP Tools

```go
// cmd/mcp/mcp.go -- registry + server setup (generated once)
var toolRegistrations []func(*mcplib.Server)
func RegisterTool(fn func(*mcplib.Server)) { toolRegistrations = append(toolRegistrations, fn) }

// cmd/mcp/sample_tool.go -- self-registers via init()
func init() { RegisterTool(registerSampleTool) }
```

## Adding a CLI Command

```sh
# Interactive
go-scaffolder add cli-command

# Non-interactive
go-scaffolder add cli-command --name migrate
```

Generated files:

| File | Description |
|------|-------------|
| `cmd/<name>.go` | Command implementation with `init()` self-registration |
| `cmd/<name>_test.go` | Test file |

## Adding an API Endpoint

Requires the project to have been scaffolded with the **API** feature.

```sh
# Interactive
go-scaffolder add api-endpoint

# Non-interactive
go-scaffolder add api-endpoint --name user
```

Generated files:

| File | Description |
|------|-------------|
| `internal/<name>/handler.go` | HTTP handler (List, Get, Create) |
| `internal/<name>/service.go` | Business logic layer |
| `internal/<name>/storage.go` | In-memory storage |
| `cmd/routes/<name>_routes.go` | Route registration with `init()` self-registration |
| `internal/<name>/handler_test.go` | Handler test |
| `internal/<name>/service_test.go` | Service test |
| `internal/<name>/storage_test.go` | Storage test |
| `cmd/routes/<name>_routes_test.go` | Route registration test |

The generated endpoint follows the same **Handler -> Service -> Storage** pattern as the sample resource.

## Adding an MCP Tool

Requires the project to have been scaffolded with the **MCP** feature.

```sh
# Interactive
go-scaffolder add mcp-tool

# Non-interactive
go-scaffolder add mcp-tool --name search
```

Generated files:

| File | Description |
|------|-------------|
| `cmd/mcp/<name>.go` | Tool implementation with `init()` self-registration |
| `cmd/mcp/<name>_test.go` | Test file |

## Resource Naming

The `--name` flag accepts alphanumeric characters, hyphens, and underscores. The name is automatically converted to the appropriate case:

| Convention | Used for | Example (name=user-profile) |
|------------|----------|----------------------------|
| kebab-case | file names, CLI command names, API paths | `user-profile` |
| snake_case | Go package names, directory names | `user_profile` |
| PascalCase | Go type names | `UserProfile` |
| camelCase | Go function names | `userProfile` |

## State File

The `.go-scaffolder.yaml` file is written to the project root during scaffolding. It records the project configuration (app name, module path, enabled features) so that `add` operations can render templates with the correct context.

```yaml
app_name: my-service
module_path: github.com/yourorg/my-service
features:
  - api
  - mcp
  - db
db_type: postgresql
```

This file should be committed to version control.

## Enabling Features

You can enable features on a project that was scaffolded without them:

```sh
# Interactive -- select from available features
go-scaffolder add feature

# With DB type specified upfront
go-scaffolder add feature --db-type postgresql

# With cache type specified upfront
go-scaffolder add feature --cache-type redis
```

This will:

1. **Create new files** for the feature (e.g., `cmd/routes/`, `internal/sample/` for API)
2. **Patch shared files** using marker comments to insert feature-specific code (imports, flags, server startup, config sections)
3. **Update the state file** so future `add` operations know the feature is enabled
4. **Run `go mod tidy`** to resolve new dependencies

### Supported features

| Feature | New files | Shared file patches |
|---------|-----------|---------------------|
| api | Route registry, sample resource, rest helpers, auth, openapi.yaml | `cmd/serve.go` (imports, flags, HTTP server), config TOML (`[server]`) |
| mcp | MCP server, sample tool | `cmd/serve.go` (import, StartMCPServer call) |
| ui | Web frontend (Vite, TailwindCSS, AlpineJS) | `Taskfile.yml` (frontend-build task) |
| db | DB connection, schema | Config TOML (`[database]` section) |
| cache | Redis or Valkey client | Config TOML (`[redis]` or `[valkey]` section) |
| docker | Dockerfile | None |
| nomad | Nomad job definition (also enables Docker) | None |

### Marker comments

Generated files contain marker comments like `// go-scaffolder:serve-imports` that the patcher uses to insert code at the right location. If you've removed or moved these markers, the patcher will **not modify the file** and instead print the code you need to add manually.

Markers in generated files:

| File | Markers |
|------|---------|
| `cmd/serve.go` | `// go-scaffolder:serve-imports`, `// go-scaffolder:serve-flags`, `// go-scaffolder:serve-init`, `// go-scaffolder:serve-start` |
| `<app>-config.toml` | `# go-scaffolder:config-sections` |
| `Taskfile.yml` | `# go-scaffolder:taskfile-tasks` |

### Fallback for removed markers

If a marker is not found, the tool prints the code that needs to be added:

```
The following patches could not be applied automatically.
Please add the following code manually:

--- cmd/serve.go (Add API imports) ---
	"fmt"
	"net/http"
	"my-service/cmd/routes"
```

## Safety

The `add` command will **never overwrite existing files**. If a file already exists at the target path, the command exits with an error. This ensures your existing code is never accidentally replaced.

The `add feature` command patches shared files using markers but will not modify files where markers are missing -- it shows you the code to add manually instead.
