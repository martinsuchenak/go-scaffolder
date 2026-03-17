package engine

import (
	"bytes"
	"fmt"
	"io/fs"
	"strings"
	"text/template"
	"unicode"

	"github.com/martinsuchenak/go-scaffolder/internal/config"
)

type TemplateEntry struct {
	TemplatePath     string
	OutputPath       string
	RequiredFeatures []string
}

type templateSource struct {
	fs      fs.FS
	entries []TemplateEntry
}

type Engine struct {
	sources []templateSource
	funcMap template.FuncMap
}

func New(embedded fs.FS) *Engine {
	e := &Engine{}
	e.sources = append(e.sources, templateSource{fs: embedded})
	return e
}

func (e *Engine) AddExternalTemplates(externalFS fs.FS, entries []TemplateEntry) {
	e.sources = append(e.sources, templateSource{fs: externalFS, entries: entries})
}

func (e *Engine) buildFuncMap(cfg *config.ProjectConfig) template.FuncMap {
	return template.FuncMap{
		"toLower":  strings.ToLower,
		"toUpper":  strings.ToUpper,
		"toCamel":  toCamel,
		"toPascal": toPascal,
		"toSnake":  toSnake,
		"toKebab":  toKebab,
		"toString": func(v interface{}) string { return fmt.Sprintf("%v", v) },
		"hasFeature": func(name string) bool {
			return e.featureEnabled(cfg, strings.ToLower(name))
		},
		"needsSRV": func() bool {
			return cfg.Features.NeedsSRVResolve()
		},
	}
}

func (e *Engine) TemplateManifest() []TemplateEntry {
	return []TemplateEntry{
		// Base templates (always included)
		{TemplatePath: "base/main.go.tmpl", OutputPath: "main.go", RequiredFeatures: nil},
		{TemplatePath: "base/go.mod.tmpl", OutputPath: "go.mod", RequiredFeatures: nil},
		{TemplatePath: "base/build/version.go.tmpl", OutputPath: "build/version.go", RequiredFeatures: nil},
		{TemplatePath: "base/Taskfile.yml.tmpl", OutputPath: "Taskfile.yml", RequiredFeatures: nil},
		{TemplatePath: "base/config.toml.tmpl", OutputPath: "{{.AppName}}-config.toml", RequiredFeatures: nil},

		// CLI feature (always included)
		{TemplatePath: "cmd/register.go.tmpl", OutputPath: "cmd/register.go", RequiredFeatures: []string{"cli"}},
		{TemplatePath: "cmd/serve.go.tmpl", OutputPath: "cmd/serve.go", RequiredFeatures: []string{"cli"}},
		{TemplatePath: "cmd/init.go.tmpl", OutputPath: "cmd/init.go", RequiredFeatures: []string{"cli"}},
		{TemplatePath: "cmd/completion.go.tmpl", OutputPath: "cmd/completion.go", RequiredFeatures: []string{"cli"}},

		// API feature
		{TemplatePath: "api/cmd/routes/api_routes.go.tmpl", OutputPath: "cmd/routes/api_routes.go", RequiredFeatures: []string{"api"}},
		{TemplatePath: "api/internal/rest/helpers.go.tmpl", OutputPath: "internal/rest/helpers.go", RequiredFeatures: []string{"api"}},
		{TemplatePath: "api/internal/auth/auth.go.tmpl", OutputPath: "internal/auth/auth.go", RequiredFeatures: []string{"api"}},
		{TemplatePath: "api/internal/ctxkeys/ctxkeys.go.tmpl", OutputPath: "internal/ctxkeys/ctxkeys.go", RequiredFeatures: []string{"api"}},
		{TemplatePath: "api/internal/sample/handler.go.tmpl", OutputPath: "internal/sample/handler.go", RequiredFeatures: []string{"api"}},
		{TemplatePath: "api/internal/sample/service.go.tmpl", OutputPath: "internal/sample/service.go", RequiredFeatures: []string{"api"}},
		{TemplatePath: "api/internal/sample/storage.go.tmpl", OutputPath: "internal/sample/storage.go", RequiredFeatures: []string{"api"}},
		{TemplatePath: "api/openapi.yaml.tmpl", OutputPath: "openapi.yaml", RequiredFeatures: []string{"api"}},

		// MCP feature
		{TemplatePath: "mcp/cmd/mcp/mcp.go.tmpl", OutputPath: "cmd/mcp/mcp.go", RequiredFeatures: []string{"mcp"}},

		// UI feature
		{TemplatePath: "ui/web/embed.go.tmpl", OutputPath: "web/embed.go", RequiredFeatures: []string{"ui"}},
		{TemplatePath: "ui/web/templates/base.html.tmpl", OutputPath: "web/templates/base.html", RequiredFeatures: []string{"ui"}},
		{TemplatePath: "ui/web/package.json.tmpl", OutputPath: "web/package.json", RequiredFeatures: []string{"ui"}},
		{TemplatePath: "ui/web/dist/.gitkeep.tmpl", OutputPath: "web/dist/.gitkeep", RequiredFeatures: []string{"ui"}},
		{TemplatePath: "ui/web/src/main.ts.tmpl", OutputPath: "web/src/main.ts", RequiredFeatures: []string{"ui"}},
		{TemplatePath: "ui/web/src/style.css.tmpl", OutputPath: "web/src/style.css", RequiredFeatures: []string{"ui"}},
		{TemplatePath: "ui/web/vite.config.ts.tmpl", OutputPath: "web/vite.config.ts", RequiredFeatures: []string{"ui"}},

		// DB feature
		{TemplatePath: "db/internal/db/db.go.tmpl", OutputPath: "internal/db/db.go", RequiredFeatures: []string{"db"}},
		{TemplatePath: "db/internal/db/schema.sql.tmpl", OutputPath: "internal/db/schema.sql", RequiredFeatures: []string{"db"}},

		// Cache Redis feature
		{TemplatePath: "cache/redis/internal/redis/redis.go.tmpl", OutputPath: "internal/redis/redis.go", RequiredFeatures: []string{"cache_redis"}},

		// Cache Valkey feature
		{TemplatePath: "cache/valkey/internal/valkey/valkey.go.tmpl", OutputPath: "internal/valkey/valkey.go", RequiredFeatures: []string{"cache_valkey"}},

		// SRV resolve (when DB, Cache, or API)
		{TemplatePath: "resolve/internal/resolve/resolve.go.tmpl", OutputPath: "internal/resolve/resolve.go", RequiredFeatures: []string{"srv"}},

		// Docker feature
		{TemplatePath: "docker/Dockerfile.tmpl", OutputPath: "Dockerfile", RequiredFeatures: []string{"docker"}},

		// Nomad feature
		{TemplatePath: "nomad/nomad.tmpl", OutputPath: "{{.AppName}}.nomad", RequiredFeatures: []string{"nomad"}},

		// Test file templates - base
		{TemplatePath: "tests/base/build/version_test.go.tmpl", OutputPath: "build/version_test.go", RequiredFeatures: nil},

		// Test file templates - CLI
		{TemplatePath: "tests/cmd/register_test.go.tmpl", OutputPath: "cmd/register_test.go", RequiredFeatures: []string{"cli"}},
		{TemplatePath: "tests/cmd/serve_test.go.tmpl", OutputPath: "cmd/serve_test.go", RequiredFeatures: []string{"cli"}},
		{TemplatePath: "tests/cmd/init_test.go.tmpl", OutputPath: "cmd/init_test.go", RequiredFeatures: []string{"cli"}},
		{TemplatePath: "tests/cmd/completion_test.go.tmpl", OutputPath: "cmd/completion_test.go", RequiredFeatures: []string{"cli"}},

		// Test file templates - API
		{TemplatePath: "tests/api/cmd/routes/api_routes_test.go.tmpl", OutputPath: "cmd/routes/api_routes_test.go", RequiredFeatures: []string{"api"}},
		{TemplatePath: "tests/api/internal/rest/helpers_test.go.tmpl", OutputPath: "internal/rest/helpers_test.go", RequiredFeatures: []string{"api"}},
		{TemplatePath: "tests/api/internal/auth/auth_test.go.tmpl", OutputPath: "internal/auth/auth_test.go", RequiredFeatures: []string{"api"}},
		{TemplatePath: "tests/api/internal/ctxkeys/ctxkeys_test.go.tmpl", OutputPath: "internal/ctxkeys/ctxkeys_test.go", RequiredFeatures: []string{"api"}},
		{TemplatePath: "tests/api/internal/sample/handler_test.go.tmpl", OutputPath: "internal/sample/handler_test.go", RequiredFeatures: []string{"api"}},
		{TemplatePath: "tests/api/internal/sample/service_test.go.tmpl", OutputPath: "internal/sample/service_test.go", RequiredFeatures: []string{"api"}},
		{TemplatePath: "tests/api/internal/sample/storage_test.go.tmpl", OutputPath: "internal/sample/storage_test.go", RequiredFeatures: []string{"api"}},

		// Test file templates - MCP
		{TemplatePath: "tests/mcp/cmd/mcp/mcp_test.go.tmpl", OutputPath: "cmd/mcp/mcp_test.go", RequiredFeatures: []string{"mcp"}},

		// Test file templates - DB
		{TemplatePath: "tests/db/internal/db/db_test.go.tmpl", OutputPath: "internal/db/db_test.go", RequiredFeatures: []string{"db"}},

		// Test file templates - Cache
		{TemplatePath: "tests/cache/redis/internal/redis/redis_test.go.tmpl", OutputPath: "internal/redis/redis_test.go", RequiredFeatures: []string{"cache_redis"}},
		{TemplatePath: "tests/cache/valkey/internal/valkey/valkey_test.go.tmpl", OutputPath: "internal/valkey/valkey_test.go", RequiredFeatures: []string{"cache_valkey"}},

		// Test file templates - SRV resolve
		{TemplatePath: "tests/resolve/internal/resolve/resolve_test.go.tmpl", OutputPath: "internal/resolve/resolve_test.go", RequiredFeatures: []string{"srv"}},
	}
}

func (e *Engine) featureEnabled(cfg *config.ProjectConfig, feature string) bool {
	switch feature {
	case "cli":
		return cfg.Features.CLI
	case "api":
		return cfg.Features.API
	case "mcp":
		return cfg.Features.MCP
	case "ui":
		return cfg.Features.UI
	case "db":
		return cfg.Features.DB
	case "cache":
		return cfg.Features.Cache
	case "cache_redis":
		return cfg.Features.Cache && cfg.CacheType == config.CacheRedis
	case "cache_valkey":
		return cfg.Features.Cache && cfg.CacheType == config.CacheValkey
	case "docker":
		return cfg.Features.Docker
	case "nomad":
		return cfg.Features.Nomad
	case "srv":
		return cfg.Features.NeedsSRVResolve()
	}
	for _, tag := range cfg.CustomTags {
		if tag == feature {
			return true
		}
	}
	return false
}

func (e *Engine) MergedManifest() []TemplateEntry {
	builtIn := e.TemplateManifest()

	var external []TemplateEntry
	for _, src := range e.sources[1:] {
		external = append(external, src.entries...)
	}

	overridden := make(map[string]bool)
	for _, ext := range external {
		overridden[ext.OutputPath] = true
	}

	var merged []TemplateEntry
	for _, entry := range builtIn {
		if !overridden[entry.OutputPath] {
			merged = append(merged, entry)
		}
	}
	merged = append(merged, external...)
	return merged
}

func (e *Engine) readTemplate(entry TemplateEntry) ([]byte, error) {
	// Try external sources in reverse order (last added wins)
	for i := len(e.sources) - 1; i >= 1; i-- {
		src := e.sources[i]
		for _, ext := range src.entries {
			if ext.TemplatePath == entry.TemplatePath {
				return fs.ReadFile(src.fs, entry.TemplatePath)
			}
		}
	}
	// Fall back to embedded (first source)
	return fs.ReadFile(e.sources[0].fs, entry.TemplatePath)
}

func (e *Engine) RenderAll(cfg *config.ProjectConfig) (map[string][]byte, error) {
	funcMap := e.buildFuncMap(cfg)
	manifest := e.MergedManifest()
	files := make(map[string][]byte)

	for _, entry := range manifest {
		if !e.shouldInclude(cfg, entry) {
			continue
		}

		content, err := e.readTemplate(entry)
		if err != nil {
			return nil, fmt.Errorf("reading template %s: %w", entry.TemplatePath, err)
		}

		tmpl, err := template.New(entry.TemplatePath).Funcs(funcMap).Parse(string(content))
		if err != nil {
			return nil, fmt.Errorf("parsing template %s: %w", entry.TemplatePath, err)
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, cfg); err != nil {
			return nil, fmt.Errorf("rendering template %s: %w", entry.TemplatePath, err)
		}

		outputPath, err := e.resolveOutputPath(entry.OutputPath, cfg)
		if err != nil {
			return nil, fmt.Errorf("resolving output path %s: %w", entry.OutputPath, err)
		}

		files[outputPath] = buf.Bytes()
	}

	return files, nil
}

func (e *Engine) shouldInclude(cfg *config.ProjectConfig, entry TemplateEntry) bool {
	if len(entry.RequiredFeatures) == 0 {
		return true
	}
	for _, f := range entry.RequiredFeatures {
		if !e.featureEnabled(cfg, f) {
			return false
		}
	}
	return true
}

func (e *Engine) resolveOutputPath(path string, cfg *config.ProjectConfig) (string, error) {
	if !strings.Contains(path, "{{") {
		return path, nil
	}
	tmpl, err := template.New("path").Parse(path)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, cfg); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func toCamel(s string) string {
	words := splitWords(s)
	for i := 1; i < len(words); i++ {
		words[i] = capitalize(words[i])
	}
	if len(words) > 0 {
		words[0] = strings.ToLower(words[0])
	}
	return strings.Join(words, "")
}

func toPascal(s string) string {
	words := splitWords(s)
	for i := range words {
		words[i] = capitalize(words[i])
	}
	return strings.Join(words, "")
}

func toSnake(s string) string {
	words := splitWords(s)
	for i := range words {
		words[i] = strings.ToLower(words[i])
	}
	return strings.Join(words, "_")
}

func toKebab(s string) string {
	words := splitWords(s)
	for i := range words {
		words[i] = strings.ToLower(words[i])
	}
	return strings.Join(words, "-")
}

func splitWords(s string) []string {
	var words []string
	var current []rune

	for _, r := range s {
		if r == '-' || r == '_' || r == ' ' {
			if len(current) > 0 {
				words = append(words, string(current))
				current = nil
			}
		} else if unicode.IsUpper(r) && len(current) > 0 {
			words = append(words, string(current))
			current = []rune{r}
		} else {
			current = append(current, r)
		}
	}
	if len(current) > 0 {
		words = append(words, string(current))
	}
	return words
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
