# Config File Reference

The `--config` flag enables non-interactive mode by reading all scaffolding parameters from a YAML file. This is useful for CI pipelines, scripted project generation, and reproducible setups.

## Schema

```yaml
# Required
app_name: my-service
output_dir: ./output

# Optional: Go module path (defaults to app_name if omitted)
module_path: github.com/yourorg/my-service

# Required: list of features to enable
features:
  - api
  - db
  - cache
  - docker

# Required when "db" is in features
db_type: postgresql    # mysql | postgresql | sqlite

# Required when "cache" is in features
cache_type: redis      # redis | valkey
```

## Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `app_name` | string | Yes | Application name. Used in config file names, Nomad job names |
| `output_dir` | string | Yes | Filesystem path for the generated project |
| `module_path` | string | No | Go module path (e.g. `github.com/yourorg/my-service`). Defaults to `app_name` |
| `features` | list of strings | Yes | Features to enable (see below) |
| `db_type` | string | When DB selected | Database engine: `mysql`, `postgresql`, or `sqlite` |
| `cache_type` | string | When Cache selected | Cache engine: `redis` or `valkey` |

## Recognized Feature Names

| Feature | Description |
|---------|-------------|
| `api` | HTTP REST API |
| `mcp` | Model Context Protocol server |
| `ui` | Web frontend (Vite + TailwindCSS + AlpineJS) |
| `db` | Database persistence |
| `cache` | Cache integration (Redis or Valkey) |
| `docker` | Dockerfile generation |
| `nomad` | Nomad job definition (auto-includes Docker) |

Any feature name not in the list above is treated as a **custom tag** and can be used to activate [external templates](extending.md).

## Validation

The same validation rules apply in both interactive and config file modes:

- `app_name` must not be empty or whitespace-only
- `output_dir` must not be empty
- `db_type` must be valid when DB is selected
- `cache_type` must be valid when Cache is selected
- Nomad automatically includes Docker

If validation fails, the scaffolder prints all errors and exits with a non-zero status. No files are generated.

## Examples

### Minimal CLI-only project

```yaml
app_name: my-tool
output_dir: ./my-tool
features: []
```

### Full-featured API service

```yaml
app_name: user-service
module_path: github.com/yourorg/user-service
output_dir: ./user-service
features:
  - api
  - mcp
  - ui
  - db
  - cache
  - docker
  - nomad
db_type: postgresql
cache_type: redis
```

### SQLite with Valkey

```yaml
app_name: local-service
output_dir: ./local-service
features:
  - api
  - db
  - cache
db_type: sqlite
cache_type: valkey
```

### With custom module path and feature tags

```yaml
app_name: my-service
module_path: github.com/yourorg/my-service
output_dir: ./my-service
features:
  - api
  - grpc          # custom tag for external templates
  - monitoring    # another custom tag
db_type: postgresql
```

## Differences from Interactive Mode

| Behavior | Interactive | Config File |
|----------|------------|-------------|
| Missing required fields | Re-prompts | Error + exit |
| Invalid values | Re-prompts | Error + exit |
| Output dir already has files | Asks for confirmation | Proceeds silently |
| Feature dependency resolution | Same | Same |

## Usage

```sh
bin/go-scaffolder --config scaffold.yaml

# Combined with external templates
bin/go-scaffolder --config scaffold.yaml --templates ./my-templates
```
