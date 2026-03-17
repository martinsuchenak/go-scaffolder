# Extending with Custom Templates

The scaffolder can be extended with external templates that **add new files** or **override built-in templates**. This is done via the `--templates` flag pointing to a directory containing a `manifest.yaml` and template files.

## Directory Structure

```
my-templates/
├── manifest.yaml              # Required: declares template entries
├── monitoring/
│   └── monitoring.go.tmpl     # A new template
├── Taskfile.yml.tmpl          # Overrides the built-in Taskfile
└── custom/
    └── worker.go.tmpl         # Another new template
```

## manifest.yaml

The manifest declares each template's source file, output path, and optional feature guards:

```yaml
templates:
  # Add a new file -- always included (no features guard)
  - template: monitoring/monitoring.go.tmpl
    output: internal/monitoring/monitoring.go

  # Add a file guarded by a custom feature tag
  - template: custom/worker.go.tmpl
    output: internal/worker/worker.go
    features:
      - worker

  # Override a built-in template (matched by output path)
  - template: Taskfile.yml.tmpl
    output: Taskfile.yml
```

### Fields

| Field | Required | Description |
|-------|----------|-------------|
| `template` | Yes | Path to the template file, relative to the templates directory |
| `output` | Yes | Output path in the generated project (supports `{{.AppName}}` substitution) |
| `features` | No | List of feature guards. Template is included only when all are enabled |

## Overriding Built-in Templates

To override a built-in template, create an external template entry with the **same output path**. The external version completely replaces the built-in one.

For example, to replace the generated `Taskfile.yml`:

```yaml
# manifest.yaml
templates:
  - template: Taskfile.yml.tmpl
    output: Taskfile.yml
```

```
# Taskfile.yml.tmpl -- your custom version
version: '3'

tasks:
  build:
    desc: Custom build for {{.AppName}}
    cmds:
      - go build -o {{.AppName}} .
```

The override is matched by the `output` path. All other built-in templates remain unchanged.

## Custom Feature Tags

External templates can use any string as a feature guard. Feature names that don't match a built-in feature (api, mcp, ui, db, cache, docker, nomad) are treated as **custom tags**.

Custom tags are activated by listing them in the config file's `features` array:

```yaml
# scaffold.yaml
app_name: my-service
module_path: github.com/yourorg/my-service
output_dir: ./output
features:
  - api
  - worker           # custom tag -- activates templates guarded by "worker"
  - monitoring        # another custom tag
```

```yaml
# manifest.yaml
templates:
  - template: worker.go.tmpl
    output: internal/worker/worker.go
    features:
      - worker        # only included when "worker" is in the features list

  - template: monitoring.go.tmpl
    output: internal/monitoring/monitoring.go
    features:
      - monitoring    # only included when "monitoring" is in the features list
```

Custom tags are also available inside templates via the `hasFeature` function:

```go
{{- if hasFeature "worker"}}
// Worker-specific code
{{- end}}
```

## Running with External Templates

```sh
# Interactive mode with custom templates
bin/go-scaffolder --templates ./my-templates

# Config file mode with custom templates
bin/go-scaffolder --config scaffold.yaml --templates ./my-templates
```

## Template Data and Functions

External templates have access to the same data and functions as built-in templates. See [How Templates Work](templates.md) for the full reference.

## Multiple Template Directories

Currently only one `--templates` directory is supported per invocation. To combine multiple extension sets, merge them into a single directory with a single `manifest.yaml`.

## Validation

When loading external templates, the scaffolder validates:

- `manifest.yaml` exists and parses as valid YAML
- Every `template` path in the manifest points to an existing file
- Every entry has both `template` and `output` fields

If validation fails, the scaffolder exits with an error before generating any files.

## Example: Adding a gRPC Feature

```
grpc-templates/
├── manifest.yaml
├── proto/
│   └── service.proto.tmpl
├── internal/
│   └── grpc/
│       └── server.go.tmpl
└── cmd/
    └── grpc_routes.go.tmpl
```

```yaml
# manifest.yaml
templates:
  - template: proto/service.proto.tmpl
    output: proto/{{.AppName}}.proto
    features:
      - grpc

  - template: internal/grpc/server.go.tmpl
    output: internal/grpc/server.go
    features:
      - grpc

  - template: cmd/grpc_routes.go.tmpl
    output: cmd/grpc_routes.go
    features:
      - grpc
```

```yaml
# scaffold.yaml
app_name: my-service
module_path: github.com/yourorg/my-service
output_dir: ./output
features:
  - api
  - grpc
```

```sh
bin/go-scaffolder --config scaffold.yaml --templates ./grpc-templates
```
