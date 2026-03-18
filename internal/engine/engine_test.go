package engine

import (
	"strings"
	"testing"

	"github.com/martinsuchenak/go-scaffolder/internal/config"
	"github.com/martinsuchenak/go-scaffolder/templates"

	"pgregory.net/rapid"
)

func genValidProjectConfig() *rapid.Generator[config.ProjectConfig] {
	return rapid.Custom(func(t *rapid.T) config.ProjectConfig {
		appName := rapid.StringMatching(`[a-z][a-z0-9\-]{2,15}`).Draw(t, "AppName")
		fs := config.FeatureSet{
			CLI:    rapid.Bool().Draw(t, "CLI"),
			API:    rapid.Bool().Draw(t, "API"),
			MCP:    rapid.Bool().Draw(t, "MCP"),
			UI:     rapid.Bool().Draw(t, "UI"),
			DB:     rapid.Bool().Draw(t, "DB"),
			Cache:  rapid.Bool().Draw(t, "Cache"),
			Docker: rapid.Bool().Draw(t, "Docker"),
			Nomad:  rapid.Bool().Draw(t, "Nomad"),
		}
		config.ResolveFeatures(&fs)

		var dbType config.DBType
		if fs.DB {
			dbType = rapid.SampledFrom([]config.DBType{config.DBMySQL, config.DBPostgreSQL, config.DBSQLite}).Draw(t, "DBType")
		}
		var cacheType config.CacheType
		if fs.Cache {
			cacheType = rapid.SampledFrom([]config.CacheType{config.CacheRedis, config.CacheValkey}).Draw(t, "CacheType")
		}

		return config.ProjectConfig{
			AppName:    appName,
			OutputDir:  "/tmp/test-" + appName,
			ModulePath: appName,
			Features:   fs,
			DBType:     dbType,
			CacheType:  cacheType,
		}
	})
}

func newEngine() *Engine {
	return New(templates.FS)
}

// Feature: go-scaffolder, Property 3: Feature-conditional file inclusion
func TestProperty3_FeatureConditionalFileInclusion(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		cfg := genValidProjectConfig().Draw(t, "config")
		eng := newEngine()

		files, err := eng.RenderAll(&cfg)
		if err != nil {
			t.Fatalf("RenderAll error: %v", err)
		}

		manifest := eng.TemplateManifest()
		for _, entry := range manifest {
			shouldInclude := eng.shouldInclude(&cfg, entry)
			outputPath, _ := eng.resolveOutputPath(entry.OutputPath, &cfg, eng.buildFuncMap(&cfg))
			_, exists := files[outputPath]

			if shouldInclude && !exists {
				t.Fatalf("expected file %q to be included for features %v", outputPath, entry.RequiredFeatures)
			}
			if !shouldInclude && exists {
				t.Fatalf("expected file %q to NOT be included", outputPath)
			}
		}
	})
}

// Feature: go-scaffolder, Property 4: App_Name substitution in rendered output
func TestProperty4_AppNameSubstitution(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		cfg := genValidProjectConfig().Draw(t, "config")
		eng := newEngine()

		files, err := eng.RenderAll(&cfg)
		if err != nil {
			t.Fatalf("RenderAll error: %v", err)
		}

		// go.mod module path
		if gomod, ok := files["go.mod"]; ok {
			if !strings.Contains(string(gomod), "module "+cfg.AppName) {
				t.Fatalf("go.mod should contain module path with AppName %q", cfg.AppName)
			}
		}

		// config file name
		configKey := cfg.AppName + "-config.toml"
		if _, ok := files[configKey]; !ok {
			t.Fatalf("expected config file %q", configKey)
		}

		// openapi.yaml title
		if cfg.Features.API {
			if openapi, ok := files["openapi.yaml"]; ok {
				if !strings.Contains(string(openapi), cfg.AppName) {
					t.Fatalf("openapi.yaml should contain AppName %q", cfg.AppName)
				}
			}
		}

		// Nomad job name
		if cfg.Features.Nomad {
			nomadKey := cfg.AppName + ".nomad"
			if nomad, ok := files[nomadKey]; ok {
				if !strings.Contains(string(nomad), `job "`+cfg.AppName+`"`) {
					t.Fatalf("nomad file should contain job name with AppName %q", cfg.AppName)
				}
			}
		}
	})
}

// Feature: go-scaffolder, Property 5: Generated source files import correct dependencies for enabled features
func TestProperty5_SourceImportDependencies(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		cfg := genValidProjectConfig().Draw(t, "config")
		eng := newEngine()

		files, err := eng.RenderAll(&cfg)
		if err != nil {
			t.Fatalf("RenderAll error: %v", err)
		}

		allContent := collectAllGoFiles(files)

		if !strings.Contains(allContent, "paularlott/cli") {
			t.Fatal("source files must import paularlott/cli")
		}
		if !strings.Contains(allContent, "paularlott/logger") {
			t.Fatal("source files must import paularlott/logger")
		}
		if cfg.Features.MCP {
			if !strings.Contains(allContent, "paularlott/mcp") {
				t.Fatal("source should import paularlott/mcp when MCP enabled")
			}
		}
		if cfg.Features.DB {
			switch cfg.DBType {
			case config.DBPostgreSQL:
				if !strings.Contains(allContent, "lib/pq") {
					t.Fatal("source should import lib/pq for postgresql")
				}
			case config.DBMySQL:
				if !strings.Contains(allContent, "go-sql-driver/mysql") {
					t.Fatal("source should import go-sql-driver/mysql for mysql")
				}
			case config.DBSQLite:
				if !strings.Contains(allContent, "modernc.org/sqlite") {
					t.Fatal("source should import modernc.org/sqlite for sqlite")
				}
			}
		}
		if cfg.Features.Cache {
			switch cfg.CacheType {
			case config.CacheRedis:
				if !strings.Contains(allContent, "redis/go-redis") {
					t.Fatal("source should import redis library for redis cache")
				}
			case config.CacheValkey:
				if !strings.Contains(allContent, "valkey-io/valkey-go") {
					t.Fatal("source should import valkey library for valkey cache")
				}
			}
		}
	})
}

func collectAllGoFiles(files map[string][]byte) string {
	var sb strings.Builder
	for path, content := range files {
		if strings.HasSuffix(path, ".go") {
			sb.Write(content)
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// Feature: go-scaffolder, Property 6: go.mod contains no hardcoded version strings
func TestProperty6_GoModNoHardcodedVersions(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		cfg := genValidProjectConfig().Draw(t, "config")
		eng := newEngine()

		files, err := eng.RenderAll(&cfg)
		if err != nil {
			t.Fatalf("RenderAll error: %v", err)
		}

		gomod := string(files["go.mod"])
		for _, line := range strings.Split(gomod, "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "module") || strings.HasPrefix(line, "go ") {
				continue
			}
			if strings.Contains(line, " v") && !strings.HasPrefix(line, "//") {
				t.Fatalf("go.mod should not contain hardcoded version string: %q", line)
			}
		}
	})
}

// Feature: go-scaffolder, Property 7: Config TOML contains correct sections for enabled features
func TestProperty7_ConfigTOMLSections(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		cfg := genValidProjectConfig().Draw(t, "config")
		eng := newEngine()

		files, err := eng.RenderAll(&cfg)
		if err != nil {
			t.Fatalf("RenderAll error: %v", err)
		}

		configKey := cfg.AppName + "-config.toml"
		toml := string(files[configKey])

		if !strings.Contains(toml, "[log]") {
			t.Fatal("config TOML must always contain [log]")
		}

		if cfg.Features.API {
			if !strings.Contains(toml, "[server]") {
				t.Fatal("config TOML should contain [server] when API enabled")
			}
		} else {
			if strings.Contains(toml, "[server]") {
				t.Fatal("config TOML should NOT contain [server] when API disabled")
			}
		}

		if cfg.Features.DB {
			if !strings.Contains(toml, "[database]") {
				t.Fatal("config TOML should contain [database] when DB enabled")
			}
		} else {
			if strings.Contains(toml, "[database]") {
				t.Fatal("config TOML should NOT contain [database] when DB disabled")
			}
		}

		if cfg.Features.Cache && cfg.CacheType == config.CacheRedis {
			if !strings.Contains(toml, "[redis]") {
				t.Fatal("config TOML should contain [redis] when Cache=Redis")
			}
		}
		if cfg.Features.Cache && cfg.CacheType == config.CacheValkey {
			if !strings.Contains(toml, "[valkey]") {
				t.Fatal("config TOML should contain [valkey] when Cache=Valkey")
			}
		}
		if !cfg.Features.Cache || cfg.CacheType != config.CacheRedis {
			if strings.Contains(toml, "[redis]") {
				t.Fatal("config TOML should NOT contain [redis] when Cache!=Redis")
			}
		}
		if !cfg.Features.Cache || cfg.CacheType != config.CacheValkey {
			if strings.Contains(toml, "[valkey]") {
				t.Fatal("config TOML should NOT contain [valkey] when Cache!=Valkey")
			}
		}
	})
}

// Feature: go-scaffolder, Property 8: Taskfile.yml contains correct tasks for enabled features
func TestProperty8_TaskfileTasks(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		cfg := genValidProjectConfig().Draw(t, "config")
		eng := newEngine()

		files, err := eng.RenderAll(&cfg)
		if err != nil {
			t.Fatalf("RenderAll error: %v", err)
		}

		taskfile := string(files["Taskfile.yml"])

		if !strings.Contains(taskfile, "build:") {
			t.Fatal("Taskfile should contain build task")
		}
		if !strings.Contains(taskfile, "test:") {
			t.Fatal("Taskfile should contain test task")
		}
		if !strings.Contains(taskfile, "lint:") {
			t.Fatal("Taskfile should contain lint task")
		}
		if !strings.Contains(taskfile, "GOARCH=amd64") {
			t.Fatal("Taskfile should contain AMD64 target")
		}
		if !strings.Contains(taskfile, "GOARCH=arm64") {
			t.Fatal("Taskfile should contain ARM64 target")
		}
		if !strings.Contains(taskfile, "CGO_ENABLED=0") {
			t.Fatal("Taskfile should contain CGO_ENABLED=0")
		}
		if !strings.Contains(taskfile, "ldflags") {
			t.Fatal("Taskfile should contain ldflags")
		}

		if cfg.Features.UI {
			if !strings.Contains(taskfile, "frontend-build:") {
				t.Fatal("Taskfile should contain frontend-build task when UI enabled")
			}
		} else {
			if strings.Contains(taskfile, "frontend-build:") {
				t.Fatal("Taskfile should NOT contain frontend-build task when UI disabled")
			}
		}
	})
}

// Feature: go-scaffolder, Property 9: Test file parity with source files
func TestProperty9_TestFileParity(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		cfg := genValidProjectConfig().Draw(t, "config")
		eng := newEngine()

		files, err := eng.RenderAll(&cfg)
		if err != nil {
			t.Fatalf("RenderAll error: %v", err)
		}

		for path, content := range files {
			if !strings.HasSuffix(path, ".go") {
				continue
			}
			if strings.HasSuffix(path, "_test.go") {
				continue
			}
			if path == "main.go" {
				continue
			}
			// For embed.go, we don't require a test file
			if strings.HasSuffix(path, "embed.go") {
				continue
			}

			testPath := strings.TrimSuffix(path, ".go") + "_test.go"
			testContent, ok := files[testPath]
			if !ok {
				t.Fatalf("missing test file %q for source file %q", testPath, path)
			}
			_ = content
			if !strings.Contains(string(testContent), "func Test") {
				t.Fatalf("test file %q should contain at least one Test function", testPath)
			}
		}
	})
}

// Feature: go-scaffolder, Property 10: Dockerfile correctness based on feature combination
func TestProperty10_DockerfileCorrectness(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		cfg := genValidProjectConfig().Draw(t, "config")
		if !cfg.Features.Docker {
			cfg.Features.Docker = true
		}
		eng := newEngine()

		files, err := eng.RenderAll(&cfg)
		if err != nil {
			t.Fatalf("RenderAll error: %v", err)
		}

		dockerfile, ok := files["Dockerfile"]
		if !ok {
			t.Fatal("Dockerfile should be present when Docker enabled")
		}

		df := string(dockerfile)
		if !strings.Contains(df, "CGO_ENABLED=0") {
			t.Fatal("Dockerfile should contain CGO_ENABLED=0")
		}
		if !strings.Contains(df, "FROM") {
			t.Fatal("Dockerfile should be multi-stage (contain FROM)")
		}

		if cfg.Features.UI {
			if !strings.Contains(df, "frontend") {
				t.Fatal("Dockerfile should contain frontend build stage when UI enabled")
			}
		}
	})
}

// Feature: go-scaffolder, Property 11: Configuration wiring in main.go
func TestProperty11_ConfigWiringMainGo(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		cfg := genValidProjectConfig().Draw(t, "config")
		eng := newEngine()

		files, err := eng.RenderAll(&cfg)
		if err != nil {
			t.Fatalf("RenderAll error: %v", err)
		}

		mainGo := string(files["main.go"])
		if !strings.Contains(mainGo, "ConfigFile") {
			t.Fatal("main.go should contain TOML config file setup")
		}
		if !strings.Contains(mainGo, "env.Load") {
			t.Fatal("main.go should contain dotenv loading")
		}
		if !strings.Contains(mainGo, "EnvVars") {
			t.Fatal("main.go should contain environment variable support")
		}
	})
}

// Feature: go-scaffolder, Property 12: DB schema uses UUIDv7 primary keys
func TestProperty12_DBSchemaUUIDv7(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		cfg := genValidProjectConfig().Draw(t, "config")
		if !cfg.Features.DB {
			cfg.Features.DB = true
			cfg.DBType = rapid.SampledFrom([]config.DBType{config.DBMySQL, config.DBPostgreSQL, config.DBSQLite}).Draw(t, "DBType")
			config.ResolveFeatures(&cfg.Features)
		}
		eng := newEngine()

		files, err := eng.RenderAll(&cfg)
		if err != nil {
			t.Fatalf("RenderAll error: %v", err)
		}

		schema, ok := files["internal/db/schema.sql"]
		if !ok {
			t.Fatal("schema.sql should be present when DB enabled")
		}

		s := string(schema)
		switch cfg.DBType {
		case config.DBPostgreSQL:
			if !strings.Contains(s, "UUID") {
				t.Fatal("postgresql schema should use UUID type for primary keys")
			}
		case config.DBMySQL:
			if !strings.Contains(s, "BINARY(16)") {
				t.Fatal("mysql schema should use BINARY(16) for UUIDv7 primary keys")
			}
		case config.DBSQLite:
			if !strings.Contains(s, "TEXT PRIMARY KEY") {
				t.Fatal("sqlite schema should use TEXT for UUIDv7 primary keys")
			}
		}
	})
}

// Feature: go-scaffolder, Property 13: MCP wiring in serve command
func TestProperty13_MCPWiringServeCmd(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		cfg := genValidProjectConfig().Draw(t, "config")
		if !cfg.Features.MCP {
			cfg.Features.MCP = true
			config.ResolveFeatures(&cfg.Features)
		}
		eng := newEngine()

		files, err := eng.RenderAll(&cfg)
		if err != nil {
			t.Fatalf("RenderAll error: %v", err)
		}

		serveGo, ok := files["cmd/serve.go"]
		if !ok {
			t.Fatal("cmd/serve.go should be present")
		}

		if !strings.Contains(string(serveGo), "StartMCPServer") {
			t.Fatal("serve.go should contain MCP server startup when MCP enabled")
		}
	})
}

// Feature: go-scaffolder, Property 14: API health and metrics endpoints registered
func TestProperty14_APIHealthMetricsEndpoints(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		cfg := genValidProjectConfig().Draw(t, "config")
		if !cfg.Features.API {
			cfg.Features.API = true
			config.ResolveFeatures(&cfg.Features)
		}
		eng := newEngine()

		files, err := eng.RenderAll(&cfg)
		if err != nil {
			t.Fatalf("RenderAll error: %v", err)
		}

		routes, ok := files["cmd/routes/api_routes.go"]
		if !ok {
			t.Fatal("api_routes.go should be present when API enabled")
		}

		r := string(routes)
		if !strings.Contains(r, "/health") {
			t.Fatal("routes should contain GET /health registration")
		}
		if !strings.Contains(r, "/metrics") {
			t.Fatal("routes should contain GET /metrics registration")
		}
	})
}

// Feature: go-scaffolder, Property 15: No partial output on template error
func TestProperty15_NoPartialOutputOnError(t *testing.T) {
	eng := New(templates.FS)

	cfg := &config.ProjectConfig{
		AppName:    "test-app",
		OutputDir:  "/tmp/test",
		ModulePath: "test-app",
		Features: config.FeatureSet{
			CLI: true,
			DB:  true,
		},
		DBType: "invalid_db_type",
	}

	files, err := eng.RenderAll(cfg)
	// If there's an error, files should be nil
	if err != nil {
		if files != nil {
			t.Fatal("on error, RenderAll should not return partial files")
		}
	}
}

// Feature: go-scaffolder, Property 16: SRV resolution usage in feature initialization code
func TestProperty16_SRVResolutionUsage(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		cfg := genValidProjectConfig().Draw(t, "config")
		eng := newEngine()

		files, err := eng.RenderAll(&cfg)
		if err != nil {
			t.Fatalf("RenderAll error: %v", err)
		}

		if cfg.Features.DB {
			dbGo, ok := files["internal/db/db.go"]
			if !ok {
				t.Fatal("internal/db/db.go should be present when DB enabled")
			}
			// Check for DNS resolver initialization with logger
			if !strings.Contains(string(dbGo), "dns.NewDNSResolver") {
				t.Fatal("DB init code should initialize dns.NewDNSResolver with logger")
			}
		}

		if cfg.Features.Cache && cfg.CacheType == config.CacheRedis {
			redisGo, ok := files["internal/redis/redis.go"]
			if !ok {
				t.Fatal("internal/redis/redis.go should be present when Cache=Redis")
			}
			if !strings.Contains(string(redisGo), "dns.NewDNSResolver") {
				t.Fatal("Redis init code should initialize dns.NewDNSResolver with logger")
			}
		}

		if cfg.Features.Cache && cfg.CacheType == config.CacheValkey {
			valkeyGo, ok := files["internal/valkey/valkey.go"]
			if !ok {
				t.Fatal("internal/valkey/valkey.go should be present when Cache=Valkey")
			}
			if !strings.Contains(string(valkeyGo), "dns.NewDNSResolver") {
				t.Fatal("Valkey init code should initialize dns.NewDNSResolver with logger")
			}
		}
	})
}
