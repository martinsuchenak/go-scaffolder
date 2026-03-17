package patcher

import (
	"fmt"

	"github.com/martinsuchenak/go-scaffolder/internal/config"
)

func FeaturePatches(feature string, cfg *config.ProjectConfig) []Patch {
	switch feature {
	case "api":
		return apiPatches(cfg)
	case "mcp":
		return mcpPatches(cfg)
	case "ui":
		return uiPatches(cfg)
	case "db":
		return dbPatches(cfg)
	case "cache":
		return cachePatches(cfg)
	case "docker":
		return nil
	case "nomad":
		return nil
	}
	return nil
}

func apiPatches(cfg *config.ProjectConfig) []Patch {
	return []Patch{
		{
			File:        "cmd/serve.go",
			Marker:      "// go-scaffolder:serve-imports",
			InsertAbove: true,
			Content: fmt.Sprintf(`	"fmt"
	"net/http"

	"%s/cmd/routes"
`, cfg.ModulePath),
			Description: "Add API imports",
		},
		{
			File:   "cmd/serve.go",
			Marker: "// go-scaffolder:serve-flags",
			Content: `		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:         "server-host",
				DefaultValue: "0.0.0.0",
				Usage:        "Server listen host",
				ConfigPath:   []string{"server.host"},
				EnvVars:      []string{"SERVER_HOST"},
			},
			&cli.IntFlag{
				Name:         "server-port",
				DefaultValue: 8080,
				Usage:        "Server listen port",
				ConfigPath:   []string{"server.port"},
				EnvVars:      []string{"SERVER_PORT"},
			},
		},`,
			Description: "Add API server flags",
		},
		{
			File:        "cmd/serve.go",
			Description: "Replace default select{} with HTTP server startup",
			Replace: &ReplaceBlock{
				StartMarker: "// go-scaffolder:serve-start",
				EndMarker:   "select {}",
				Content: `// go-scaffolder:serve-start
			return http.ListenAndServe(addr, mux)`,
			},
		},
		{
			File:   "cmd/serve.go",
			Marker: "// go-scaffolder:serve-init",
			Content: `
			mux := http.NewServeMux()
			routes.RegisterRoutes(mux)

			addr := fmt.Sprintf("%s:%d", cmd.GetString("server-host"), cmd.GetInt("server-port"))
			log.Info("starting HTTP server", "addr", addr)`,
			Description: "Add HTTP server setup",
		},
		{
			File:        fmt.Sprintf("%s-config.toml", cfg.AppName),
			Marker:      "# go-scaffolder:config-sections",
			InsertAbove: true,
			Content: `
[server]
host = "0.0.0.0"
port = 8080`,
			Description: "Add [server] config section",
		},
	}
}

func mcpPatches(cfg *config.ProjectConfig) []Patch {
	return []Patch{
		{
			File:        "cmd/serve.go",
			Marker:      "// go-scaffolder:serve-imports",
			InsertAbove: true,
			Content:     fmt.Sprintf("\n\tmcpserver \"%s/cmd/mcp\"", cfg.ModulePath),
			Description: "Add MCP import",
		},
		{
			File:   "cmd/serve.go",
			Marker: "// go-scaffolder:serve-init",
			Content: `
			mcpserver.StartMCPServer(log)`,
			Description: "Add MCP server startup",
		},
	}
}

func uiPatches(cfg *config.ProjectConfig) []Patch {
	return []Patch{
		{
			File:   "Taskfile.yml",
			Marker: "# go-scaffolder:taskfile-tasks",
			Content: `
  frontend-build:
    desc: Build frontend assets
    dir: web
    cmds:
      - bun install
      - bun run build`,
			Description: "Add frontend-build task",
		},
	}
}

func dbPatches(cfg *config.ProjectConfig) []Patch {
	var dbConfig string
	switch cfg.DBType {
	case config.DBPostgreSQL:
		dbConfig = fmt.Sprintf(`
[database]
host = "localhost"
port = 5432
user = "postgres"
password = "postgres"
name = "%s"
sslmode = "disable"`, cfg.AppName)
	case config.DBMySQL:
		dbConfig = fmt.Sprintf(`
[database]
host = "localhost"
port = 3306
user = "root"
password = "root"
name = "%s"`, cfg.AppName)
	case config.DBSQLite:
		dbConfig = fmt.Sprintf(`
[database]
path = "%s.db"`, cfg.AppName)
	}

	return []Patch{
		{
			File:        fmt.Sprintf("%s-config.toml", cfg.AppName),
			Marker:      "# go-scaffolder:config-sections",
			InsertAbove: true,
			Content:     dbConfig,
			Description: "Add [database] config section",
		},
	}
}

func cachePatches(cfg *config.ProjectConfig) []Patch {
	var cacheConfig string
	switch cfg.CacheType {
	case config.CacheRedis:
		cacheConfig = `
[redis]
host = "localhost"
port = 6379
password = ""
db = 0`
	case config.CacheValkey:
		cacheConfig = `
[valkey]
host = "localhost"
port = 6379
password = ""`
	}

	return []Patch{
		{
			File:        fmt.Sprintf("%s-config.toml", cfg.AppName),
			Marker:      "# go-scaffolder:config-sections",
			InsertAbove: true,
			Content:     cacheConfig,
			Description: fmt.Sprintf("Add [%s] config section", cfg.CacheType),
		},
	}
}
