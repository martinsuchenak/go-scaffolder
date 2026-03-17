# MCP Server

The go-scaffolder includes a built-in [Model Context Protocol](https://modelcontextprotocol.io/) server that exposes all scaffolding operations as tools. This allows LLMs and other MCP clients to scaffold projects, add components, and enable features programmatically.

The server uses the [github.com/paularlott/mcp](https://github.com/paularlott/mcp) library and serves on the `/mcp` HTTP endpoint.

## Starting the Server

```sh
# Default (listen on :8080, no auth)
go-scaffolder serve

# Custom listen address
go-scaffolder serve --listen :9090

# With bearer token authorization
go-scaffolder serve --token my-secret-token

# Token via environment variable
MCP_TOKEN=my-secret-token go-scaffolder serve
```

### Flags

| Flag | Env Var | Default | Description |
|------|---------|---------|-------------|
| `--listen` | -- | `:8080` | Address to listen on |
| `--token` | `MCP_TOKEN` | -- | Bearer token for authorization (optional) |

## Authorization

When `--token` is set (or `MCP_TOKEN` env var is present), the server requires all requests to include an `Authorization: Bearer <token>` header. The comparison uses constant-time comparison to prevent timing attacks. When no token is configured, the server accepts all requests.

## Tools

### `scaffold`

Scaffold a new Go microservice project. Returns unified diff of all generated files.

| Parameter | Required | Description |
|-----------|----------|-------------|
| `app_name` | yes | Application name |
| `module_path` | no | Go module path (defaults to `app_name`) |
| `features` | no | Comma-separated features: `api,mcp,ui,db,cache,docker,nomad` |
| `db_type` | no | Database type: `mysql`, `postgresql`, `sqlite` |
| `cache_type` | no | Cache type: `redis`, `valkey` |

Example:

```json
{
  "name": "scaffold",
  "arguments": {
    "app_name": "my-service",
    "module_path": "github.com/myorg/my-service",
    "features": "api,db,docker",
    "db_type": "postgresql"
  }
}
```

### `add_cli_command`

Add a new CLI command to an existing project. Returns unified diff of new files.

| Parameter | Required | Description |
|-----------|----------|-------------|
| `name` | yes | Name of the new command |
| `project_dir` | no | Path to project root (see [Project Context Resolution](#project-context-resolution)) |
| `state_file` | no | Content of `.go-scaffolder.yaml` (see [Project Context Resolution](#project-context-resolution)) |
| `app_name` | no | Fallback app name when no state file is available |
| `module_path` | no | Fallback module path (defaults to `app_name`) |

### `add_api_endpoint`

Add a new API endpoint resource (handler, service, storage, routes + tests). Requires the API feature to be enabled.

| Parameter | Required | Description |
|-----------|----------|-------------|
| `name` | yes | Name of the new endpoint resource |
| `project_dir` | no | Path to project root |
| `state_file` | no | Content of `.go-scaffolder.yaml` |
| `app_name` | no | Fallback app name |
| `module_path` | no | Fallback module path |

### `add_mcp_tool`

Add a new MCP tool to an existing project. Requires the MCP feature to be enabled.

| Parameter | Required | Description |
|-----------|----------|-------------|
| `name` | yes | Name of the new MCP tool |
| `project_dir` | no | Path to project root |
| `state_file` | no | Content of `.go-scaffolder.yaml` |
| `app_name` | no | Fallback app name |
| `module_path` | no | Fallback module path |

### `enable_feature`

Enable a feature on an existing project. Returns unified diff of new files and patches to existing files.

| Parameter | Required | Description |
|-----------|----------|-------------|
| `feature` | yes | Feature to enable: `api`, `mcp`, `ui`, `db`, `cache`, `docker`, `nomad` |
| `db_type` | no | Required when enabling `db` |
| `cache_type` | no | Required when enabling `cache` |
| `project_dir` | no | Path to project root |
| `state_file` | no | Content of `.go-scaffolder.yaml` |
| `app_name` | no | Fallback app name |
| `module_path` | no | Fallback module path |

When `project_dir` is available, the tool computes actual unified diffs for marker-based patches against existing files. Without filesystem access, it outputs the patch snippets with marker references for manual application.

### `project_context`

Generate a rich context summary of a scaffolded project for LLM consumption. The output includes enabled features, available add operations, features that can still be enabled, and (when `project_dir` is available) the project directory tree.

| Parameter | Required | Description |
|-----------|----------|-------------|
| `project_dir` | no | Path to project root |
| `state_file` | no | Content of `.go-scaffolder.yaml` |
| `app_name` | no | Fallback app name |
| `module_path` | no | Fallback module path |

## Project Context Resolution

Tools that operate on existing projects need to know the project configuration (app name, module path, enabled features). The server resolves this through a 3-tier fallback:

1. **`state_file`** -- if provided, the YAML content is parsed directly. This is the best option for fully remote MCP servers that have no filesystem access to the project. Takes precedence over all other options.

2. **`project_dir`** -- if provided, the server reads `.go-scaffolder.yaml` from that directory. Use this when the server runs on the same machine as the project but in a different working directory.

3. **Working directory** -- if neither is provided, the server looks for `.go-scaffolder.yaml` in its own working directory.

4. **`app_name` fallback** -- if no state file is found anywhere, the server builds a minimal config from `app_name` and `module_path` parameters. The project is treated as having only the CLI feature enabled (no API, MCP, etc.), so `enable_feature` can be used to add features and all generated output will be new-file diffs.

If none of these succeed, the tool returns an error: `"no state file found; provide project_dir, state_file, or at minimum app_name"`.

### Example: local server with project_dir

```json
{
  "name": "add_api_endpoint",
  "arguments": {
    "name": "user",
    "project_dir": "/home/dev/my-service"
  }
}
```

### Example: remote server with state_file

```json
{
  "name": "enable_feature",
  "arguments": {
    "feature": "api",
    "state_file": "app_name: my-service\nmodule_path: github.com/myorg/my-service\nfeatures:\n  - mcp\n  - db\ndb_type: postgresql\n"
  }
}
```

### Example: remote server with app_name fallback

```json
{
  "name": "add_cli_command",
  "arguments": {
    "name": "migrate",
    "app_name": "my-service",
    "module_path": "github.com/myorg/my-service"
  }
}
```

## Output Format

All tools return unified diff output:

- **`scaffold`** -- all files as new-file diffs (`--- /dev/null`)
- **`add_*`** -- new component files as new-file diffs
- **`enable_feature`** -- new feature files as new-file diffs, plus modification diffs for patched files (when `project_dir` is available) or patch snippets with marker references (when remote)
- **`project_context`** -- markdown-formatted project summary (not diff)

The diff output is standard unified diff format, consumable by `git apply`, `patch`, or any LLM file editor.

## Deployment Scenarios

### Local development

Run the server from the project directory. Tools automatically find `.go-scaffolder.yaml`:

```sh
cd my-service
go-scaffolder serve
```

### Shared server

Run the server centrally. Clients pass `project_dir` to point at specific projects:

```sh
go-scaffolder serve --listen :9090 --token $MCP_TOKEN
```

### Remote / cloud server

Run the server without filesystem access to any project. Clients pass `state_file` content or use `app_name` fallback:

```sh
go-scaffolder serve --listen :8080 --token $MCP_TOKEN
```

In this mode, `enable_feature` cannot compute actual file diffs for marker-based patches (since it cannot read the project files). Instead, it outputs the code snippets with marker references that the LLM can apply to the project files.
