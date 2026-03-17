# Requirements Document

## Introduction

The Go Scaffolder is an interactive tool that generates fully functional Go microservice applications following the Fortix architecture conventions. The tool prompts the user for an application name, output directory, and desired feature set (CLI, API, MCP, UI, DB, Cache, Docker, Nomad), then produces a complete, buildable project skeleton with all required boilerplate, configuration, build tooling, and directory structure.

## Glossary

- **Scaffolder**: The interactive tool that generates new Go microservice projects
- **Scaffolded_App**: The Go microservice project produced by the Scaffolder
- **Feature_Set**: The combination of optional capabilities selected by the user: CLI, API, MCP, UI, DB, Cache, Docker, Nomad
- **CLI_Feature**: Command-line interface capability using `github.com/paularlott/cli` for command parsing and TOML config loading
- **API_Feature**: HTTP REST API capability using Go standard library routing (`net/http` with Go 1.22+ enhanced routing)
- **MCP_Feature**: Model Context Protocol capability using `github.com/paularlott/mcp` for AI tool exposure
- **UI_Feature**: Web frontend capability using Vite, TailwindCSS, AlpineJS, and TypeScript, with assets embedded via `go:embed`
- **DB_Feature**: Database persistence capability providing schema, initialization logic, and driver dependencies for a user-selected database engine
- **DB_Type**: The specific database engine selected by the user when the DB_Feature is enabled (e.g., MySQL, PostgreSQL, SQLite)
- **Cache_Feature**: Cache client integration capability providing connection setup and helper utilities for caching or messaging, where the user selects exactly one Cache_Type (Redis or Valkey); Redis and Valkey are mutually exclusive options
- **Cache_Type**: The specific cache engine selected by the user when the Cache_Feature is enabled: either Redis or Valkey
- **Docker_Feature**: Dockerfile generation capability producing a multi-stage build for the Scaffolded_App
- **Nomad_Feature**: HashiCorp Nomad job definition generation capability producing a deployment-ready Nomad job file
- **Fortix_Conventions**: The set of architectural patterns, directory layout, dependency choices, and coding standards defined for Fortix microservices
- **App_Name**: The user-provided name for the scaffolded microservice, used in module paths, config file names, and Nomad job definitions
- **Output_Directory**: The filesystem path where the Scaffolder places the generated project
- **SRV_Resolution**: The process of looking up DNS SRV records to discover the actual host address and port for a service, commonly used with service discovery systems like Consul
- **Config_File_Mode**: A non-interactive mode where the Scaffolder reads all configuration from a file instead of prompting the user, enabling automated/scripted scaffolding

## Requirements

### Requirement 1: Interactive User Prompting

**User Story:** As a developer, I want the Scaffolder to interactively ask me for project configuration, so that I can customize the generated application to my needs.

#### Acceptance Criteria

1. WHEN the Scaffolder is launched, THE Scaffolder SHALL prompt the user for the App_Name
2. WHEN the App_Name is collected, THE Scaffolder SHALL prompt the user for the Output_Directory
3. WHEN the Output_Directory is collected, THE Scaffolder SHALL prompt the user to select one or more features from the Feature_Set (CLI, API, MCP, UI, DB, Cache, Docker, Nomad)
4. WHEN the user selects the DB_Feature, THE Scaffolder SHALL prompt the user to select a DB_Type from the supported database engines (MySQL, PostgreSQL, SQLite)
5. WHEN the user selects the Cache_Feature, THE Scaffolder SHALL prompt the user to select a Cache_Type from the supported cache engines (Redis, Valkey)
6. WHEN the user selects the Nomad_Feature without selecting the Docker_Feature, THE Scaffolder SHALL automatically include the Docker_Feature and inform the user that Docker is required for Nomad deployments
7. IF the user provides an empty App_Name, THEN THE Scaffolder SHALL display an error message and re-prompt for the App_Name
8. IF the user provides an Output_Directory that does not exist, THEN THE Scaffolder SHALL create the Output_Directory before generating files
9. IF the user provides an Output_Directory that already contains files, THEN THE Scaffolder SHALL display a warning and ask for confirmation before proceeding

### Requirement 2: Core Project Structure Generation

**User Story:** As a developer, I want the Scaffolder to generate the base Fortix directory layout and files, so that every scaffolded app starts with the correct structure.

#### Acceptance Criteria

1. THE Scaffolder SHALL generate a `main.go` file that wires CLI commands using `github.com/paularlott/cli` with a root command, `--config`, `--log-level`, and `--log-format` global flags
2. THE Scaffolder SHALL generate a `go.mod` file with the module path derived from the App_Name and Go version set to 1.26 or later
3. THE Scaffolder SHALL generate a `build/version.go` file exposing `build.Version` and `build.Date` variables populated via `-ldflags`
4. THE Scaffolder SHALL generate a `Taskfile.yml` with build, test, and lint tasks
5. THE Scaffolder SHALL generate a `<App_Name>-config.toml` file with sections for `[log]` and any additional sections required by the selected Feature_Set
6. THE Scaffolder SHALL include `github.com/paularlott/logger` as a dependency in every generated project
7. THE Scaffolder SHALL include `github.com/paularlott/cli` as a dependency in every generated project

### Requirement 3: CLI Feature Scaffolding

**User Story:** As a developer, I want the Scaffolder to generate CLI sub-commands when I select the CLI feature, so that my app has a working command-line interface from the start.

#### Acceptance Criteria

1. WHEN the CLI_Feature is selected, THE Scaffolder SHALL generate a `cmd/serve.go` file containing the `serve` sub-command implementation
2. WHEN the CLI_Feature is selected, THE Scaffolder SHALL generate an `init` sub-command that initializes a default configuration file
3. WHEN the CLI_Feature is selected, THE Scaffolder SHALL generate a `completion` sub-command that outputs shell completion scripts
4. THE Scaffolder SHALL always generate the CLI_Feature scaffolding regardless of user selection, as CLI is a core Fortix requirement
5. THE Scaffolded_App SHALL display help text listing all available commands and their flags WHEN run with no arguments or with the `--help` flag, using the built-in help output provided by `github.com/paularlott/cli`
6. THE Scaffolded_App SHALL support loading configuration from TOML config files, dotenv files, and environment variables using the built-in configuration support provided by `github.com/paularlott/cli`
7. THE Scaffolder SHALL generate a `main.go` that wires up configuration loading from all three sources (TOML config file, dotenv file, and environment variables) via `github.com/paularlott/cli`

### Requirement 4: API Feature Scaffolding

**User Story:** As a developer, I want the Scaffolder to generate HTTP API boilerplate when I select the API feature, so that I have a working REST API skeleton.

#### Acceptance Criteria

1. WHEN the API_Feature is selected, THE Scaffolder SHALL generate route registration files in `cmd/` following the `*_routes.go` naming convention
2. WHEN the API_Feature is selected, THE Scaffolder SHALL generate a health check endpoint registered at `GET /health`
3. WHEN the API_Feature is selected, THE Scaffolder SHALL generate a metrics endpoint registered at `GET /metrics`
4. WHEN the API_Feature is selected, THE Scaffolder SHALL generate an `internal/rest/` package with shared HTTP helper functions
5. WHEN the API_Feature is selected, THE Scaffolder SHALL generate an `internal/auth/` package with placeholder OAuth 2.0 + PKCE and API key authentication logic
6. WHEN the API_Feature is selected, THE Scaffolder SHALL generate an `internal/ctxkeys/` package with typed context key definitions
7. WHEN the API_Feature is selected, THE Scaffolder SHALL generate an `openapi.yaml` file with an OpenAPI v3.1 skeleton referencing the App_Name
8. WHEN the API_Feature is selected, THE Scaffolder SHALL generate a sample resource package under `internal/` with `handler.go`, `service.go`, and `storage.go` files demonstrating the Handler → Service → Storage layering pattern
9. WHEN the API_Feature is selected, THE Scaffolder SHALL add a `[server]` section to the generated `<App_Name>-config.toml`

### Requirement 5: MCP Feature Scaffolding

**User Story:** As a developer, I want the Scaffolder to generate MCP integration boilerplate when I select the MCP feature, so that my app can expose tools to AI agents.

#### Acceptance Criteria

1. WHEN the MCP_Feature is selected, THE Scaffolder SHALL include `github.com/paularlott/mcp` as a dependency in the generated `go.mod`
2. WHEN the MCP_Feature is selected, THE Scaffolder SHALL generate an MCP server registration file that exposes a sample tool definition
3. WHEN the MCP_Feature is selected, THE Scaffolder SHALL wire the MCP server into the `serve` sub-command startup sequence

### Requirement 6: UI Feature Scaffolding

**User Story:** As a developer, I want the Scaffolder to generate a web frontend skeleton when I select the UI feature, so that my app includes an embedded web interface.

#### Acceptance Criteria

1. WHEN the UI_Feature is selected, THE Scaffolder SHALL generate a `web/embed.go` file that uses `go:embed` to embed static assets
2. WHEN the UI_Feature is selected, THE Scaffolder SHALL generate a `web/src/` directory with a minimal Vite + TailwindCSS + AlpineJS + TypeScript project
3. WHEN the UI_Feature is selected, THE Scaffolder SHALL generate a `web/templates/` directory with a base HTML template
4. WHEN the UI_Feature is selected, THE Scaffolder SHALL include a `package.json` in the `web/` directory with `bun` as the expected package manager
5. WHEN the UI_Feature is selected, THE Scaffolder SHALL add a frontend build step to the generated `Taskfile.yml`

### Requirement 7: DB Feature Scaffolding

**User Story:** As a developer, I want the Scaffolder to generate database boilerplate when I select the DB feature, so that I have a working persistence layer from the start.

#### Acceptance Criteria

1. WHEN the DB_Feature is selected, THE Scaffolder SHALL generate an `internal/db/` package containing database initialization logic for the selected DB_Type
2. WHEN the DB_Feature is selected, THE Scaffolder SHALL generate an embedded `schema.sql` file within the `internal/db/` package
3. WHEN the DB_Feature is selected, THE Scaffolder SHALL use UUIDv7 for all primary key fields in generated schema and model code
4. WHEN the DB_Feature is selected with DB_Type set to PostgreSQL, THE Scaffolder SHALL include the `github.com/lib/pq` driver dependency in the generated `go.mod`
5. WHEN the DB_Feature is selected with DB_Type set to MySQL, THE Scaffolder SHALL include the `github.com/go-sql-driver/mysql` driver dependency in the generated `go.mod`
6. WHEN the DB_Feature is selected with DB_Type set to SQLite, THE Scaffolder SHALL include the `modernc.org/sqlite` driver dependency in the generated `go.mod`
7. WHEN the DB_Feature is selected, THE Scaffolder SHALL add a `[database]` section to the generated `<App_Name>-config.toml`

### Requirement 8: Cache Feature Scaffolding

**User Story:** As a developer, I want the Scaffolder to generate cache client boilerplate when I select the Cache feature, so that I have a working cache integration from the start using my preferred cache engine.

#### Acceptance Criteria

1. WHEN the Cache_Feature is selected with Cache_Type set to Redis, THE Scaffolder SHALL generate an `internal/redis/` package containing Redis client initialization and connection helper functions
2. WHEN the Cache_Feature is selected with Cache_Type set to Redis, THE Scaffolder SHALL include a Redis client library dependency in the generated `go.mod`
3. WHEN the Cache_Feature is selected with Cache_Type set to Redis, THE Scaffolder SHALL add a `[redis]` section to the generated `<App_Name>-config.toml`
4. WHEN the Cache_Feature is selected with Cache_Type set to Valkey, THE Scaffolder SHALL generate an `internal/valkey/` package containing Valkey client initialization and connection helper functions
5. WHEN the Cache_Feature is selected with Cache_Type set to Valkey, THE Scaffolder SHALL include a Valkey client library dependency in the generated `go.mod`
6. WHEN the Cache_Feature is selected with Cache_Type set to Valkey, THE Scaffolder SHALL add a `[valkey]` section to the generated `<App_Name>-config.toml`

### Requirement 9: Docker Feature Scaffolding

**User Story:** As a developer, I want the Scaffolder to generate a Dockerfile when I select the Docker feature, so that I can containerize my application.

#### Acceptance Criteria

1. WHEN the Docker_Feature is selected, THE Scaffolder SHALL generate a `Dockerfile` implementing a multi-stage build producing a `CGO_ENABLED=0` static binary
2. WHEN the Docker_Feature is selected and the UI_Feature is also selected, THE Scaffolder SHALL add a frontend build stage to the generated `Dockerfile`
3. WHEN the Nomad_Feature is selected, THE Scaffolder SHALL automatically include the Docker_Feature even if the user did not explicitly select the Docker_Feature

### Requirement 10: Nomad Feature Scaffolding

**User Story:** As a developer, I want the Scaffolder to generate a Nomad job definition when I select the Nomad feature, so that I can deploy my application to a Nomad cluster.

#### Acceptance Criteria

1. WHEN the Nomad_Feature is selected, THE Scaffolder SHALL generate a `<App_Name>.nomad` file with a Nomad job definition referencing the App_Name
2. WHEN the Nomad_Feature is selected, THE Scaffolder SHALL include the Docker_Feature in the generated project as a dependency of Nomad deployments

### Requirement 11: Generated Code Correctness

**User Story:** As a developer, I want the scaffolded app to compile and run without modification, so that I can immediately start building on top of it.

#### Acceptance Criteria

1. THE Scaffolded_App SHALL compile without errors using `go build ./...`
2. THE Scaffolded_App SHALL pass `go vet ./...` without warnings
3. WHEN the API_Feature is selected, THE Scaffolded_App SHALL start an HTTP server that responds to `GET /health` with a 200 status code
4. THE Scaffolded_App SHALL use the Service-First Architecture pattern where all entry points route through the service layer

### Requirement 12: Template Rendering and File Generation

**User Story:** As a developer, I want the Scaffolder to use Go templates for file generation, so that the generated code is consistent and maintainable.

#### Acceptance Criteria

1. THE Scaffolder SHALL use Go `text/template` to render all generated source files
2. THE Scaffolder SHALL substitute the App_Name into module paths, package names, config file names, and Nomad job names
3. THE Scaffolder SHALL conditionally include or exclude files and code blocks based on the selected Feature_Set
4. IF a template rendering error occurs, THEN THE Scaffolder SHALL display the error with the template name and stop generation without writing partial output

### Requirement 13: Cross-Compilation and Build Support

**User Story:** As a developer, I want the scaffolded build tooling to support cross-compilation, so that I can produce binaries for multiple platforms.

#### Acceptance Criteria

1. THE Scaffolded_App SHALL include build targets for AMD64 and ARM64 architectures in the generated `Taskfile.yml`
2. THE Scaffolded_App SHALL set `CGO_ENABLED=0` in all build configurations to produce static binaries
3. THE Scaffolded_App SHALL inject `build.Version` and `build.Date` via `-ldflags` in all build targets

### Requirement 14: Post-Generation Dependency Resolution

**User Story:** As a developer, I want the Scaffolder to resolve all dependencies after generating files, so that the generated `go.mod` contains the correct latest versions without hardcoded version strings in templates.

#### Acceptance Criteria

1. WHEN all files have been generated, THE Scaffolder SHALL run `go mod tidy` in the Output_Directory to resolve all dependencies to their latest versions
2. THE Scaffolder SHALL NOT hardcode library version strings in generated `go.mod` files; instead the Scaffolder SHALL list dependencies without versions and rely on `go mod tidy` to resolve them
3. IF `go mod tidy` fails, THEN THE Scaffolder SHALL display the error output from `go mod tidy` and inform the user that dependency resolution failed
4. WHEN `go mod tidy` completes successfully, THE Scaffolder SHALL display a confirmation message indicating that dependencies were resolved

### Requirement 15: Test File Generation

**User Story:** As a developer, I want the Scaffolder to generate test file stubs alongside source files, so that I have a testing structure ready from the start.

#### Acceptance Criteria

1. THE Scaffolder SHALL generate a corresponding `_test.go` file for each generated `.go` source file
2. THE Scaffolder SHALL include at least one example test function in each generated test file
3. THE Scaffolder SHALL generate test files that pass when run with `go test ./...`

### Requirement 16: SRV Record Resolution in Scaffolded Apps

**User Story:** As a developer, I want the scaffolded app to automatically resolve DNS SRV records for configured service hosts, so that I can use service discovery (e.g. Consul) without manual address/port management.

#### Acceptance Criteria

1. WHEN the DB_Feature is selected, THE Scaffolded_App SHALL check if the configured database host is a DNS SRV record and resolve it to the actual address and port before connecting
2. WHEN the Cache_Feature is selected, THE Scaffolded_App SHALL check if the configured cache host is a DNS SRV record and resolve it to the actual address and port before connecting
3. WHEN the API_Feature is selected and the Scaffolded_App connects to external services, THE Scaffolded_App SHALL check if configured service URLs contain DNS SRV records and resolve them before connecting
4. THE Scaffolder SHALL generate the SRV resolution logic in a shared `internal/resolve/` package within the Scaffolded_App, reusable by all features that perform network connections

### Requirement 17: Non-Interactive Config File Input Mode

**User Story:** As a developer, I want the Scaffolder to accept a config file for fully automated scaffolding, so that I can script project generation in CI pipelines without interactive prompts.

#### Acceptance Criteria

1. THE Scaffolder SHALL accept a `--config` flag pointing to a YAML config file containing all scaffolding parameters (App_Name, Output_Directory, Feature_Set selections, DB_Type, Cache_Type)
2. WHEN a `--config` flag is provided, THE Scaffolder SHALL skip all interactive prompts and use the values from the config file
3. IF the config file is missing required fields (e.g. App_Name or Output_Directory), THEN THE Scaffolder SHALL display an error listing the missing fields and exit with a non-zero status code
4. THE Scaffolder SHALL validate all config file values using the same validation rules as interactive mode (e.g. non-empty App_Name, valid DB_Type when DB is selected, valid Cache_Type when Cache is selected)