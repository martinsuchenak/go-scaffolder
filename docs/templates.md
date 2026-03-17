# How Templates Work

## Overview

The scaffolder uses Go's `text/template` engine to render all generated files. Templates are embedded in the binary via `go:embed` and organized by feature.

## Template Directory Structure

```
templates/
├── base/                       # Always included
│   ├── main.go.tmpl
│   ├── go.mod.tmpl
│   ├── build/version.go.tmpl
│   ├── Taskfile.yml.tmpl
│   └── config.toml.tmpl
├── cmd/                        # CLI feature (always included)
│   ├── register.go.tmpl        #   Command registry
│   ├── serve.go.tmpl           #   Self-registers via init()
│   ├── init.go.tmpl            #   Self-registers via init()
│   └── completion.go.tmpl      #   Self-registers via init()
├── api/                        # API feature
│   ├── cmd/routes/
│   │   ├── api_routes.go.tmpl  #   Route registry + health/metrics
│   │   └── sample_routes.go.tmpl # Self-registers via init()
│   ├── internal/rest/helpers.go.tmpl
│   ├── internal/auth/auth.go.tmpl
│   ├── internal/ctxkeys/ctxkeys.go.tmpl
│   ├── internal/sample/handler.go.tmpl
│   ├── internal/sample/service.go.tmpl
│   ├── internal/sample/storage.go.tmpl
│   └── openapi.yaml.tmpl
├── mcp/                        # MCP feature
│   └── cmd/mcp/
│       ├── mcp.go.tmpl         #   Server + tool registry
│       └── sample_tool.go.tmpl #   Self-registers via init()
├── ui/                         # UI feature
│   └── web/...
├── db/                         # DB feature
│   └── internal/db/...
├── cache/                      # Cache feature
│   ├── redis/internal/redis/redis.go.tmpl
│   └── valkey/internal/valkey/valkey.go.tmpl
├── resolve/                    # SRV resolution (auto-included)
│   └── internal/resolve/resolve.go.tmpl
├── docker/                     # Docker feature
│   └── Dockerfile.tmpl
├── nomad/                      # Nomad feature
│   └── nomad.tmpl
├── add/                        # Templates for "add" operations
│   ├── cli_command/            #   New CLI command
│   ├── api_endpoint/           #   New API endpoint (handler/service/storage/routes)
│   └── mcp_tool/               #   New MCP tool
└── tests/                      # Test file templates (mirrors source structure)
    ├── base/...
    ├── cmd/...
    ├── api/...
    └── ...
```

## Template Manifest

Each template has a manifest entry defining:

| Field | Description |
|-------|-------------|
| `TemplatePath` | Path within the embedded filesystem |
| `OutputPath` | Relative output path in the generated project |
| `RequiredFeatures` | Feature flags that must all be enabled for this template to be included |

Example entries:

```
TemplatePath: "api/internal/rest/helpers.go.tmpl"
OutputPath:   "internal/rest/helpers.go"
RequiredFeatures: ["api"]
```

```
TemplatePath: "cache/redis/internal/redis/redis.go.tmpl"
OutputPath:   "internal/redis/redis.go"
RequiredFeatures: ["cache_redis"]
```

Templates with no `RequiredFeatures` are always included (base templates).

## Template Data

Every template receives a `ProjectConfig` struct as its data context:

| Field | Type | Description |
|-------|------|-------------|
| `.AppName` | `string` | Application name |
| `.OutputDir` | `string` | Output directory |
| `.ModulePath` | `string` | Go module path (e.g. `github.com/yourorg/my-service`, defaults to AppName) |
| `.Features.CLI` | `bool` | Always `true` |
| `.Features.API` | `bool` | API feature enabled |
| `.Features.MCP` | `bool` | MCP feature enabled |
| `.Features.UI` | `bool` | UI feature enabled |
| `.Features.DB` | `bool` | DB feature enabled |
| `.Features.Cache` | `bool` | Cache feature enabled |
| `.Features.Docker` | `bool` | Docker feature enabled |
| `.Features.Nomad` | `bool` | Nomad feature enabled |
| `.DBType` | `string` | `"mysql"`, `"postgresql"`, or `"sqlite"` |
| `.CacheType` | `string` | `"redis"` or `"valkey"` |
| `.CustomTags` | `[]string` | Custom feature tags from config |
| `.ResourceName` | `string` | Name of the resource being added (only in `add` templates) |

## Template Functions

The following functions are available in all templates:

| Function | Description | Example |
|----------|-------------|---------|
| `toLower` | Lowercase | `{{toLower .AppName}}` |
| `toUpper` | Uppercase | `{{toUpper .AppName}}` |
| `toCamel` | camelCase | `{{toCamel .AppName}}` |
| `toPascal` | PascalCase | `{{toPascal .AppName}}` |
| `toSnake` | snake_case | `{{toSnake .AppName}}` |
| `toKebab` | kebab-case | `{{toKebab .AppName}}` |
| `toString` | String conversion | `{{toString .DBType}}` |
| `hasFeature` | Check feature/tag | `{{if hasFeature "api"}}...{{end}}` |
| `needsSRV` | Check SRV needed | `{{if needsSRV}}...{{end}}` |

## Dynamic Output Paths

Output paths can contain Go template expressions:

```
OutputPath: "{{.AppName}}-config.toml"
OutputPath: "{{.AppName}}.nomad"
```

These are resolved using the same `ProjectConfig` data.

## Feature Guards

Feature guards control which templates are included in the output. A template is included only when **all** of its required features are enabled.

Built-in feature guard values:

| Guard | Enabled when |
|-------|-------------|
| `cli` | Always (CLI is always on) |
| `api` | API feature selected |
| `mcp` | MCP feature selected |
| `ui` | UI feature selected |
| `db` | DB feature selected |
| `cache` | Cache feature selected |
| `cache_redis` | Cache=Redis |
| `cache_valkey` | Cache=Valkey |
| `docker` | Docker feature selected |
| `nomad` | Nomad feature selected |
| `srv` | DB, Cache, or API enabled |

External templates can use any string as a feature guard. Unknown guards are matched against `CustomTags` in the project config.

## Conditional Content Within Templates

Use standard Go template conditionals to vary content based on features:

```go
{{- if .Features.API}}
import "net/http"
{{- end}}
```

```go
{{- if eq (toString .DBType) "postgresql"}}
    _ "github.com/lib/pq"
{{- end}}
```

## Two-Pass Rendering

The scaffolder uses a two-pass approach:

1. **Render** -- all applicable templates are rendered to in-memory buffers
2. **Write** -- only if all templates rendered successfully, files are written to disk

If any template fails to parse or execute, no files are written. This prevents partial/broken project output.

## Marker Comments

Generated files contain marker comments that enable the `add feature` command to patch shared files when enabling new features. These markers serve as insertion points:

| Marker | File | Purpose |
|--------|------|---------|
| `// go-scaffolder:serve-imports` | `cmd/serve.go` | Insert new import statements |
| `// go-scaffolder:serve-flags` | `cmd/serve.go` | Insert new CLI flags |
| `// go-scaffolder:serve-init` | `cmd/serve.go` | Insert service initialization code |
| `// go-scaffolder:serve-start` | `cmd/serve.go` | Mark the server start section (replaceable) |
| `# go-scaffolder:config-sections` | `<app>-config.toml` | Insert new TOML config sections |
| `# go-scaffolder:taskfile-tasks` | `Taskfile.yml` | Insert new task definitions |

If a marker is removed by the user, the patcher will not modify the file and will instead print the code to add manually.

## Escaping Template Syntax

Since the generated Go files may themselves contain Go template syntax (e.g., HTML templates), use raw string delimiters to escape:

```
{{` + "`" + `{{.Title}}` + "`" + `}}
```

Or restructure Go struct literals to avoid double braces `{{` at the start of a line, as the template engine would interpret them.
