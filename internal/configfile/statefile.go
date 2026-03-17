package configfile

import (
	"fmt"
	"os"
	"strings"

	"github.com/martinsuchenak/go-scaffolder/internal/config"

	"gopkg.in/yaml.v3"
)

const StateFileName = ".go-scaffolder.yaml"

func WriteStateFile(path string, cfg *config.ProjectConfig) error {
	sc := ScaffoldConfig{
		AppName:    cfg.AppName,
		ModulePath: cfg.ModulePath,
		DBType:     string(cfg.DBType),
		CacheType:  string(cfg.CacheType),
	}

	if cfg.Features.API {
		sc.Features = append(sc.Features, "api")
	}
	if cfg.Features.MCP {
		sc.Features = append(sc.Features, "mcp")
	}
	if cfg.Features.UI {
		sc.Features = append(sc.Features, "ui")
	}
	if cfg.Features.DB {
		sc.Features = append(sc.Features, "db")
	}
	if cfg.Features.Cache {
		sc.Features = append(sc.Features, "cache")
	}
	if cfg.Features.Docker {
		sc.Features = append(sc.Features, "docker")
	}
	if cfg.Features.Nomad {
		sc.Features = append(sc.Features, "nomad")
	}
	sc.Features = append(sc.Features, cfg.CustomTags...)

	data, err := yaml.Marshal(sc)
	if err != nil {
		return fmt.Errorf("marshaling state file: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func LoadStateFile(path string) (*config.ProjectConfig, error) {
	sc, err := Load(path)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(sc.OutputDir) == "" {
		sc.OutputDir = "."
	}

	return sc.ToProjectConfig()
}

func ParseStateContent(content string) (*config.ProjectConfig, error) {
	var sc ScaffoldConfig
	if err := yaml.Unmarshal([]byte(content), &sc); err != nil {
		return nil, fmt.Errorf("parsing state content: %w", err)
	}

	if strings.TrimSpace(sc.OutputDir) == "" {
		sc.OutputDir = "."
	}

	return sc.ToProjectConfig()
}
