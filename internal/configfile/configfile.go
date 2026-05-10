package configfile

import (
	"fmt"
	"os"
	"strings"

	"github.com/martinsuchenak/go-scaffolder/internal/config"

	"gopkg.in/yaml.v3"
)

type ScaffoldConfig struct {
	AppName    string   `yaml:"app_name"`
	OutputDir  string   `yaml:"output_dir"`
	ModulePath string   `yaml:"module_path"`
	Features   []string `yaml:"features"`
	DBType     string   `yaml:"db_type"`
	UseXDAL    bool     `yaml:"use_xdal"`
	CacheType  string   `yaml:"cache_type"`
}

func Load(path string) (*ScaffoldConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var sc ScaffoldConfig
	if err := yaml.Unmarshal(data, &sc); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	return &sc, nil
}

func (sc *ScaffoldConfig) ToProjectConfig() (*config.ProjectConfig, error) {
	var missing []string
	if strings.TrimSpace(sc.AppName) == "" {
		missing = append(missing, "app_name")
	}
	if strings.TrimSpace(sc.OutputDir) == "" {
		missing = append(missing, "output_dir")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required fields: %s", strings.Join(missing, ", "))
	}

	fs := config.FeatureSet{}
	var customTags []string
	knownFeatures := make(map[string]bool, len(config.SelectableFeatures))
	for _, feature := range config.SelectableFeatures {
		knownFeatures[feature] = true
	}
	for _, f := range config.ExpandFeatureNames(sc.Features) {
		lower := strings.ToLower(f)
		if config.EnableFeature(&fs, lower) {
			continue
		}
		if !knownFeatures[lower] {
			customTags = append(customTags, lower)
		}
	}

	config.ResolveFeatures(&fs)

	modulePath := sc.ModulePath
	if strings.TrimSpace(modulePath) == "" {
		modulePath = sc.AppName
	}

	pc := &config.ProjectConfig{
		AppName:    sc.AppName,
		OutputDir:  sc.OutputDir,
		ModulePath: modulePath,
		Features:   fs,
		DBType:     config.DBType(sc.DBType),
		UseXDAL:    sc.UseXDAL,
		CacheType:  config.CacheType(sc.CacheType),
		CustomTags: customTags,
	}

	if err := pc.Validate(); err != nil {
		return nil, err
	}

	return pc, nil
}
