# Getting Started

## Prerequisites

- Go 1.26 or later
- [Task](https://taskfile.dev) (optional, for build commands)

## Installation

Clone the repository and build:

```sh
git clone <repo-url>
cd go-scaffolder
task build
```

The binary is placed in `bin/go-scaffolder`.

Alternatively, build directly with Go:

```sh
go build -o bin/go-scaffolder .
```

## Quick Start

### Interactive mode

Run without arguments to be guided through project setup:

```sh
bin/go-scaffolder
```

You will be prompted for:

1. **App name** -- the name of your microservice (used in config files, Nomad jobs)
2. **Module path** -- Go module path (e.g. `github.com/yourorg/my-service`); press enter to default to the app name
3. **Output directory** -- where the project will be generated
4. **Features** -- select from: API, MCP, UI, DB, Cache, Docker, Nomad
5. **DB type** -- if DB was selected: MySQL, PostgreSQL, or SQLite
6. **Cache type** -- if Cache was selected: Redis or Valkey

The scaffolder then generates all files, runs `go mod tidy`, and reports success.

### Non-interactive mode

Provide a YAML config file to skip all prompts:

```sh
bin/go-scaffolder --config scaffold.yaml
```

See [Config File Reference](config-file.md) for the full YAML schema.

### With custom templates

Extend or override built-in templates:

```sh
bin/go-scaffolder --config scaffold.yaml --templates ./my-templates
```

See [Extending with Custom Templates](extending.md) for details.

## CLI Reference

```
go-scaffolder [flags]                  Scaffold a new project
go-scaffolder add cli-command [flags]  Add a CLI command to existing project
go-scaffolder add api-endpoint [flags] Add an API endpoint to existing project
go-scaffolder add mcp-tool [flags]     Add an MCP tool to existing project
go-scaffolder add feature [flags]      Enable a feature on existing project

Scaffold flags:
    --config <path>      Path to YAML config file (non-interactive mode)
    --templates <path>   Path to external templates directory
    -v, --version        Show version
    -h, --help           Show help

Add flags:
    --name <name>        Name of the new component (prompted if omitted)

Feature flags:
    --db-type <type>     Database type: mysql, postgresql, sqlite
    --cache-type <type>  Cache type: redis, valkey
```

## What Happens During Scaffolding

1. **Input collection** -- interactive prompts or YAML config file
2. **Validation** -- app name, output dir, feature combinations are validated
3. **Feature resolution** -- CLI is always enabled; Nomad auto-includes Docker
4. **Template rendering** -- all applicable templates are rendered to memory
5. **File writing** -- files are written to the output directory (atomic: all or nothing)
6. **State file** -- `.go-scaffolder.yaml` is written to record the project config for future `add` operations
7. **Post-generation** -- `go mod tidy` resolves dependencies to latest versions

If any template fails to render, no files are written. If `go mod tidy` fails, the generated files are preserved and you can run it manually.

## Adding Components to an Existing Project

After scaffolding, you can add new components without re-scaffolding:

```sh
cd my-service

# Add a new CLI command
go-scaffolder add cli-command --name migrate

# Add a new API endpoint (requires API feature)
go-scaffolder add api-endpoint --name user

# Add a new MCP tool (requires MCP feature)
go-scaffolder add mcp-tool --name search

# Enable a feature that wasn't originally selected
go-scaffolder add feature
go-scaffolder add feature --db-type postgresql
```

See [Adding Components](adding-components.md) for full details.
