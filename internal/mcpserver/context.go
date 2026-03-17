package mcpserver

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/martinsuchenak/go-scaffolder/internal/config"
)

func GenerateContext(cfg *config.ProjectConfig) string {
	var sb strings.Builder

	sb.WriteString("# Project Context\n\n")
	sb.WriteString(fmt.Sprintf("- **App Name**: %s\n", cfg.AppName))
	sb.WriteString(fmt.Sprintf("- **Module Path**: %s\n", cfg.ModulePath))
	sb.WriteString(fmt.Sprintf("- **Output Directory**: %s\n", cfg.OutputDir))
	sb.WriteString("\n## Enabled Features\n\n")

	features := []struct {
		Name    string
		Enabled bool
	}{
		{"CLI", cfg.Features.CLI},
		{"API", cfg.Features.API},
		{"MCP", cfg.Features.MCP},
		{"UI", cfg.Features.UI},
		{"DB", cfg.Features.DB},
		{"Cache", cfg.Features.Cache},
		{"Docker", cfg.Features.Docker},
		{"Nomad", cfg.Features.Nomad},
	}

	for _, f := range features {
		status := "disabled"
		if f.Enabled {
			status = "enabled"
		}
		sb.WriteString(fmt.Sprintf("- %s: %s\n", f.Name, status))
	}

	if cfg.Features.DB {
		sb.WriteString(fmt.Sprintf("\n**Database Type**: %s\n", cfg.DBType))
	}
	if cfg.Features.Cache {
		sb.WriteString(fmt.Sprintf("\n**Cache Type**: %s\n", cfg.CacheType))
	}

	sb.WriteString("\n## Available Add Operations\n\n")
	sb.WriteString("- `add_cli_command`: Add a new CLI command\n")
	if cfg.Features.API {
		sb.WriteString("- `add_api_endpoint`: Add a new API endpoint resource (handler, service, storage, routes + tests)\n")
	}
	if cfg.Features.MCP {
		sb.WriteString("- `add_mcp_tool`: Add a new MCP tool\n")
	}

	sb.WriteString("\n## Available Features to Enable\n\n")
	enableable := []struct {
		Name    string
		Enabled bool
	}{
		{"api", cfg.Features.API},
		{"mcp", cfg.Features.MCP},
		{"ui", cfg.Features.UI},
		{"db", cfg.Features.DB},
		{"cache", cfg.Features.Cache},
		{"docker", cfg.Features.Docker},
		{"nomad", cfg.Features.Nomad},
	}
	anyAvailable := false
	for _, f := range enableable {
		if !f.Enabled {
			sb.WriteString(fmt.Sprintf("- `%s`\n", f.Name))
			anyAvailable = true
		}
	}
	if !anyAvailable {
		sb.WriteString("All features are already enabled.\n")
	}

	sb.WriteString("\n## Project Structure\n\n")
	sb.WriteString("```\n")
	sb.WriteString(walkProjectTree(".", 0, 3))
	sb.WriteString("```\n")

	return sb.String()
}

func walkProjectTree(dir string, depth, maxDepth int) string {
	if depth > maxDepth {
		return ""
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}

	var sb strings.Builder
	indent := strings.Repeat("  ", depth)

	skipDirs := map[string]bool{
		".git": true, "node_modules": true, "vendor": true,
		".idea": true, ".vscode": true, "dist": true, "tmp": true,
	}

	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") && name != ".go-scaffolder.yaml" {
			continue
		}
		if skipDirs[name] {
			continue
		}

		if entry.IsDir() {
			sb.WriteString(fmt.Sprintf("%s%s/\n", indent, name))
			sb.WriteString(walkProjectTree(filepath.Join(dir, name), depth+1, maxDepth))
		} else {
			sb.WriteString(fmt.Sprintf("%s%s\n", indent, name))
		}
	}

	return sb.String()
}
