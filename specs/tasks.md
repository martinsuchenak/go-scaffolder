# Implementation Plan: Go Scaffolder

## Overview

Build a CLI tool in Go that scaffolds Fortix-convention microservice projects. The implementation follows the pipeline: CLI setup → input collection (interactive or config file) → config building & validation → feature resolution → template rendering → file writing → post-generation (`go mod tidy`). All templates are embedded via `go:embed`. Property-based tests use the `rapid` library.

## Tasks

- [x] 1. Set up project structure and core config package
  - [x] 1.1 Initialize Go module and create directory structure
    - Create `go.mod` for the scaffolder itself
    - Create directories: `internal/config/`, `internal/prompt/`, `internal/configfile/`, `internal/engine/`, `internal/writer/`, `internal/postgen/`, `templates/`, `tests/`
    - _Requirements: 2.1, 2.2_

  - [x] 1.2 Implement `internal/config` types and feature resolution
    - Define `ProjectConfig`, `FeatureSet`, `DBType`, `CacheType` types
    - Implement `ResolveFeatures` (CLI always true, Nomad implies Docker)
    - Implement `NeedsSRVResolve` (true when DB, Cache, or API enabled)
    - Implement `Validate` (mutual exclusivity, required sub-selections, non-empty AppName)
    - _Requirements: 1.6, 1.7, 3.4, 7.3, 7.4, 7.5, 7.6, 8.2, 8.5, 9.3, 10.2_

  - [x] 1.3 Write property tests for feature resolution invariants
    - **Property 1: Feature resolution invariants**
    - Create `FeatureSet` generator producing all valid combinations
    - Assert: after `ResolveFeatures`, CLI is always true; if Nomad then Docker is true
    - **Validates: Requirements 1.6, 3.4, 9.3, 10.2**

  - [x] 1.4 Write property test for app name validation
    - **Property 2: App name validation rejects empty/whitespace input**
    - Create `AppName` generator producing whitespace-only and empty strings
    - Assert: validation returns error for all such inputs
    - **Validates: Requirements 1.7**

- [x] 2. Implement input collection packages
  - [x] 2.1 Implement `internal/prompt` for interactive mode
    - Define `Prompter` interface with `AskString`, `AskMultiSelect`, `AskSelect`, `Confirm`
    - Implement concrete prompter using terminal I/O
    - Implement validation functions for AppName, OutputDir, feature selection, DB_Type, Cache_Type
    - Wire prompts in sequence: AppName → OutputDir → Features → DB_Type (if DB) → Cache_Type (if Cache)
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.7, 1.8, 1.9_

  - [x] 2.2 Implement `internal/configfile` for config file mode
    - Define `ScaffoldConfig` struct with YAML tags
    - Implement `Load(path string)` to read and parse YAML
    - Implement `ToProjectConfig()` converting ScaffoldConfig to ProjectConfig with same validation rules
    - Handle missing required fields (app_name, output_dir) with descriptive errors
    - _Requirements: 17.1, 17.2, 17.3, 17.4_

  - [x] 2.3 Write property test for config file missing required fields
    - **Property 18: Config file missing required fields produces error**
    - Generate YAML configs with missing app_name and/or output_dir
    - Assert: error returned listing missing field names, no output produced
    - **Validates: Requirements 17.3**

  - [x] 2.4 Write property test for config file mode parity
    - **Property 17: Config file mode produces identical output to interactive mode**
    - Generate random valid scaffolding parameters
    - Assert: output from config file mode is identical to interactive mode for same inputs
    - **Validates: Requirements 17.2, 17.4**

- [x] 3. Checkpoint - Core packages
  - Ensure all tests pass, ask the user if questions arise.

- [x] 4. Implement template engine and embedded templates
  - [x] 4.1 Create base templates (always included)
    - `templates/base/main.go.tmpl` — CLI wiring with `paularlott/cli`, root command, `--config`, `--log-level`, `--log-format` flags, config loading from TOML/dotenv/env
    - `templates/base/go.mod.tmpl` — module path from AppName, Go 1.26+, feature-conditional dependencies without version strings
    - `templates/base/build/version.go.tmpl` — `build.Version` and `build.Date` variables
    - `templates/base/Taskfile.yml.tmpl` — build/test/lint tasks, AMD64+ARM64, CGO_ENABLED=0, ldflags
    - `templates/base/{{.AppName}}-config.toml.tmpl` — `[log]` section, conditional sections for API/DB/Cache
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7, 3.6, 3.7, 13.1, 13.2, 13.3, 14.2_

  - [x] 4.2 Create CLI feature templates
    - `templates/cmd/serve.go.tmpl` — serve sub-command
    - `templates/cmd/init.go.tmpl` — init sub-command for default config
    - `templates/cmd/completion.go.tmpl` — shell completion sub-command
    - _Requirements: 3.1, 3.2, 3.3, 3.5_

  - [x] 4.3 Create API feature templates
    - `templates/api/cmd/api_routes.go.tmpl` — route registration with health and metrics endpoints
    - `templates/api/internal/rest/helpers.go.tmpl` — shared HTTP helpers
    - `templates/api/internal/auth/auth.go.tmpl` — OAuth 2.0 + PKCE and API key placeholder
    - `templates/api/internal/ctxkeys/ctxkeys.go.tmpl` — typed context keys
    - `templates/api/internal/sample/handler.go.tmpl`, `service.go.tmpl`, `storage.go.tmpl` — sample resource with Handler→Service→Storage pattern
    - `templates/api/openapi.yaml.tmpl` — OpenAPI v3.1 skeleton with AppName
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6, 4.7, 4.8, 4.9, 11.3, 11.4_

  - [x] 4.4 Create MCP feature templates
    - `templates/mcp/cmd/mcp.go.tmpl` — MCP server registration with sample tool, wired into serve command
    - _Requirements: 5.1, 5.2, 5.3_

  - [x] 4.5 Create UI feature templates
    - `templates/ui/web/embed.go.tmpl` — go:embed for static assets
    - `templates/ui/web/src/` — minimal Vite + TailwindCSS + AlpineJS + TypeScript project
    - `templates/ui/web/templates/base.html.tmpl` — base HTML template
    - `templates/ui/web/package.json.tmpl` — with bun as package manager
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_

  - [x] 4.6 Create DB feature templates
    - `templates/db/internal/db/db.go.tmpl` — DB initialization with driver selection based on DBType
    - `templates/db/internal/db/schema.sql.tmpl` — schema with UUIDv7 primary keys
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.6, 7.7_

  - [x] 4.7 Create Cache feature templates
    - `templates/cache/redis/internal/redis/redis.go.tmpl` — Redis client init
    - `templates/cache/valkey/internal/valkey/valkey.go.tmpl` — Valkey client init
    - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5, 8.6_

  - [x] 4.8 Create SRV resolve templates
    - `templates/resolve/internal/resolve/resolve.go.tmpl` — DNS SRV lookup with fallback
    - Guard: included when `NeedsSRVResolve()` is true (DB, Cache, or API enabled)
    - _Requirements: 16.1, 16.2, 16.3, 16.4_

  - [x] 4.9 Create Docker and Nomad templates
    - `templates/docker/Dockerfile.tmpl` — multi-stage build, CGO_ENABLED=0, conditional UI build stage
    - `templates/nomad/{{.AppName}}.nomad.tmpl` — Nomad job definition with AppName
    - _Requirements: 9.1, 9.2, 10.1, 10.2_

  - [x] 4.10 Create test file templates
    - `templates/tests/` — corresponding `_test.go` templates for each generated `.go` source file
    - Each test file contains at least one `Test` function
    - _Requirements: 15.1, 15.2, 15.3_

- [x] 5. Implement template engine core
  - [x] 5.1 Implement `internal/engine` package
    - Define `Engine` struct with `embed.FS` and `template.FuncMap`
    - Implement FuncMap: `toLower`, `toUpper`, `toCamel`, `toPascal`, `toSnake`, `toKebab`, `hasFeature`, `needsSRV`
    - Define `TemplateEntry` with `TemplatePath`, `OutputPath`, `RequiredFeatures`
    - Implement `TemplateManifest()` returning all template entries with feature guards
    - Implement `RenderAll(cfg)` — filter templates by enabled features, render all to `map[string][]byte`, return error if any template fails (no partial output)
    - _Requirements: 12.1, 12.2, 12.3, 12.4_

  - [x] 5.2 Write property test for feature-conditional file inclusion
    - **Property 3: Feature-conditional file inclusion**
    - Generate random valid `ProjectConfig` values
    - Assert: files produced by `RenderAll` match exactly those whose `RequiredFeatures` are satisfied; `internal/resolve/` present iff `NeedsSRVResolve()` is true
    - **Validates: Requirements 12.3, 4.1, 4.4, 4.5, 4.6, 4.8, 5.2, 6.1, 6.2, 6.3, 6.4, 7.1, 7.2, 8.1, 8.4, 9.1, 10.1, 16.4, 2.1, 2.3, 3.1, 3.2, 3.3**

  - [x] 5.3 Write property test for AppName substitution
    - **Property 4: App_Name substitution in rendered output**
    - Generate random valid configs with non-empty AppName
    - Assert: go.mod module path, config file name, openapi.yaml title, Nomad job name all contain exact AppName
    - **Validates: Requirements 12.2, 2.2, 4.7**

  - [x] 5.4 Write property test for go.mod dependencies
    - **Property 5: go.mod contains correct dependencies for enabled features**
    - Assert: `paularlott/cli` and `paularlott/logger` always present; `paularlott/mcp` when MCP; correct DB driver per DBType; correct cache library per CacheType
    - **Validates: Requirements 2.6, 2.7, 5.1, 7.4, 7.5, 7.6, 8.2, 8.5**

  - [x] 5.5 Write property test for go.mod no hardcoded versions
    - **Property 6: go.mod contains no hardcoded version strings**
    - Assert: rendered go.mod require directives contain no version strings
    - **Validates: Requirements 14.2**

  - [x] 5.6 Write property test for config TOML sections
    - **Property 7: Config TOML contains correct sections for enabled features**
    - Assert: `[log]` always; `[server]` when API; `[database]` when DB; `[redis]` when Cache=Redis; `[valkey]` when Cache=Valkey; no sections for disabled features
    - **Validates: Requirements 2.5, 4.9, 7.7, 8.3, 8.6**

  - [x] 5.7 Write property test for Taskfile.yml tasks
    - **Property 8: Taskfile.yml contains correct tasks for enabled features**
    - Assert: build/test/lint tasks with AMD64+ARM64, CGO_ENABLED=0, ldflags; frontend build task when UI enabled
    - **Validates: Requirements 2.4, 6.5, 13.1, 13.2, 13.3**

  - [x] 5.8 Write property test for test file parity
    - **Property 9: Test file parity with source files**
    - Assert: every `.go` source file (excluding `_test.go` and `main.go`) has a corresponding `_test.go` with at least one `Test` function
    - **Validates: Requirements 15.1, 15.2**

  - [x] 5.9 Write property test for Dockerfile correctness
    - **Property 10: Dockerfile correctness based on feature combination**
    - Assert: multi-stage build with CGO_ENABLED=0; frontend build stage when UI enabled
    - **Validates: Requirements 9.1, 9.2**

  - [x] 5.10 Write property test for main.go configuration wiring
    - **Property 11: Configuration wiring in main.go**
    - Assert: rendered main.go contains TOML config, dotenv, and env variable loading via `paularlott/cli`
    - **Validates: Requirements 3.6, 3.7**

  - [x] 5.11 Write property test for DB schema UUIDv7
    - **Property 12: DB schema uses UUIDv7 primary keys**
    - Assert: when DB enabled, schema.sql uses UUIDv7-compatible column types for primary keys
    - **Validates: Requirements 7.3**

  - [x] 5.12 Write property test for MCP wiring
    - **Property 13: MCP wiring in serve command**
    - Assert: when MCP enabled, cmd/serve.go contains MCP server init and startup
    - **Validates: Requirements 5.3**

  - [x] 5.13 Write property test for API health and metrics endpoints
    - **Property 14: API health and metrics endpoints registered**
    - Assert: when API enabled, route files contain `GET /health` and `GET /metrics` handler registrations
    - **Validates: Requirements 4.2, 4.3**

  - [x] 5.14 Write property test for no partial output on error
    - **Property 15: No partial output on template error**
    - Inject template errors and assert file writer produces no files
    - **Validates: Requirements 12.4**

  - [x] 5.15 Write property test for SRV resolution usage
    - **Property 16: SRV resolution usage in feature initialization code**
    - Assert: when DB/Cache/API enabled, initialization code calls `resolve.LookupSRV`
    - **Validates: Requirements 16.1, 16.2, 16.3**

- [x] 6. Checkpoint - Templates and engine
  - Ensure all tests pass, ask the user if questions arise.

- [x] 7. Implement file writer and post-generation
  - [x] 7.1 Implement `internal/writer` package
    - Implement `WriteAll(outputDir string, files map[string][]byte) error`
    - Create directories as needed, write files atomically
    - Handle I/O errors with descriptive messages
    - _Requirements: 1.8, 12.4_

  - [x] 7.2 Write unit tests for file writer
    - Test directory creation, file writing, error handling
    - _Requirements: 1.8_

  - [x] 7.3 Implement `internal/postgen` package
    - Implement `RunGoModTidy(dir string) (string, error)`
    - Execute `go mod tidy`, capture stdout/stderr, return output and error
    - _Requirements: 14.1, 14.3, 14.4_

  - [x] 7.4 Write unit tests for post-generation
    - Test `go mod tidy` execution and error handling
    - _Requirements: 14.1, 14.3_

- [x] 8. Implement CLI entry point and wire pipeline
  - [x] 8.1 Implement `main.go`
    - Set up root command with `paularlott/cli`
    - Add `--config` flag for config file mode
    - Implement `runScaffold` action wiring the full pipeline:
      1. Check `--config` flag → load config file or run interactive prompts
      2. Build `ProjectConfig` and validate
      3. Resolve features
      4. Render all templates via engine
      5. Write files via writer
      6. Run `go mod tidy` via postgen
      7. Display success/failure messages
    - Handle existing output directory warning/confirmation in interactive mode
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.8, 1.9, 12.4, 14.1, 17.1, 17.2_

- [x] 9. Checkpoint - Full pipeline
  - Ensure all tests pass, ask the user if questions arise.

- [x] 10. Integration tests
  - [x] 10.1 Write integration tests for end-to-end scaffolding
    - Generate projects with major feature combinations (both interactive and config file mode)
    - Run `go build ./...` on generated projects
    - Run `go vet ./...` on generated projects
    - Run `go test ./...` on generated projects
    - Verify `GET /health` returns 200 when API feature enabled
    - Verify config file mode produces identical output to interactive mode
    - Tag tests for separate CI execution
    - _Requirements: 11.1, 11.2, 11.3, 15.3, 17.2_

- [x] 11. Final checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests validate universal correctness properties using the `rapid` library
- Unit tests validate specific examples and edge cases
- Integration tests are slower and should be tagged for separate CI execution
- All templates are embedded via `go:embed` — no external file dependencies
- The design specifies two-pass generation: render all templates to memory first, write to disk only if all succeed
