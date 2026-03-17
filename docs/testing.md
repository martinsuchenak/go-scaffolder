# Testing

The scaffolder uses three levels of testing to ensure correctness.

## Running Tests

```sh
# Unit and property tests only (fast)
task test

# Integration tests only (scaffolds real projects, slower)
task test-integration

# All tests
task test-all

# Verbose output
task test-verbose

# With coverage
go test ./internal/... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

## Property-Based Tests

Property-based tests use [rapid](https://github.com/flyingmutant/rapid) to verify invariants across randomly generated valid configurations. Each test runs 100+ iterations with different inputs.

| Property | What it verifies |
|----------|-----------------|
| 1. Feature resolution invariants | CLI always true; Nomad implies Docker |
| 2. App name validation | Empty/whitespace names are rejected |
| 3. Feature-conditional file inclusion | Output files match exactly the enabled feature guards |
| 4. AppName substitution | AppName appears in go.mod, config file, openapi.yaml, nomad job |
| 5. Source import dependencies | Correct library imports for enabled features |
| 6. No hardcoded versions | go.mod contains no version strings |
| 7. Config TOML sections | Correct sections present/absent based on features |
| 8. Taskfile tasks | Build/test/lint tasks present; frontend task when UI enabled |
| 9. Test file parity | Every .go source file has a corresponding _test.go |
| 10. Dockerfile correctness | Multi-stage build, CGO_ENABLED=0, frontend stage when UI |
| 11. Config wiring in main.go | TOML, dotenv, and env var loading present |
| 12. DB schema UUIDv7 | Primary keys use UUIDv7-compatible types |
| 13. MCP wiring | MCP server started in serve command |
| 14. API endpoints | /health and /metrics registered |
| 15. No partial output | Template errors produce no files |
| 16. SRV resolution usage | DB/Cache init code calls resolve.LookupSRV |
| 17. Config file parity | Config file mode produces same output as interactive |
| 18. Missing required fields | Config file with missing fields produces descriptive error |

## Unit Tests

Unit tests cover specific examples and edge cases:

- **config** -- validation of various valid/invalid configurations, `HasFeature` method
- **configfile** -- YAML loading, missing files, parse errors, state file write/read
- **patcher** -- marker-based patching: insert above/below, replace blocks, missing markers, missing files
- **prompt** -- validation functions for app name, output dir, and resource name
- **writer** -- file writing, directory creation, error handling
- **postgen** -- `go mod tidy` execution and error handling
- **engine/manifest** -- external template loading, overrides, custom feature tags

## Integration Tests

Integration tests scaffold complete projects and verify they work end-to-end. They are tagged with `//go:build integration` and run separately.

Each test:
1. Scaffolds a project with a specific feature combination
2. Runs `go mod tidy`
3. Runs `go build ./...`
4. Runs `go vet ./...`
5. Runs `go test ./...`

Feature combinations tested:

| Test | Features |
|------|----------|
| MinimalCLI | None (CLI only) |
| APIandDB | API + DB(PostgreSQL) + Docker |
| AllFeatures | API + MCP + UI + DB(PostgreSQL) + Cache(Redis) + Docker + Nomad |
| CacheSQLite | DB(SQLite) + Cache(Valkey) |
| MySQLRedis | API + DB(MySQL) + Cache(Redis) |
| AddCLICommand | Scaffold minimal + add cli-command "migrate" |
| AddAPIEndpoint | Scaffold with API + add api-endpoint "user" |
| AddMCPTool | Scaffold with MCP + add mcp-tool "search" |
| EnableAPI | Scaffold minimal + enable API feature |
| EnableMCP | Scaffold minimal + enable MCP feature |
| EnableDB | Scaffold minimal + enable DB(PostgreSQL) feature |
| EnableCache | Scaffold minimal + enable Cache(Redis) feature |
| EnableMultipleFeatures | Scaffold minimal + enable API, MCP, DB sequentially |
| ConfigFileParity | Verifies identical output between two renders |
| ConfigFileScaffolding | End-to-end via the actual binary with --config |

## External Template Tests

Tests for the extension system verify:

- Loading a valid manifest and template files
- Error on missing manifest.yaml
- Error on missing template files referenced in manifest
- Overriding built-in templates by matching output path
- Adding new templates
- Custom feature tag gating (included only when tag is active)

## Test Location

```
internal/
├── config/config_test.go           # Property tests (1, 2) + unit tests
├── configfile/
│   ├── configfile_test.go          # Property tests (17, 18) + unit tests
│   └── statefile_test.go           # State file write/read tests
├── patcher/
│   └── patcher_test.go             # Marker patching, replace blocks, edge cases
├── engine/
│   ├── engine_test.go              # Property tests (3-16)
│   └── manifest_test.go            # External template tests
├── writer/writer_test.go           # Unit tests
├── postgen/postgen_test.go         # Unit tests
└── prompt/prompt_test.go           # Unit tests (incl. resource name validation)
tests/
└── integration_test.go             # Integration tests (build tag: integration)
```
