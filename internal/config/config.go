package config

import (
	"fmt"
	"regexp"
	"strings"
)

type DBType string

const (
	DBMySQL      DBType = "mysql"
	DBPostgreSQL DBType = "postgresql"
	DBSQLite     DBType = "sqlite"
)

type CacheType string

const (
	CacheRedis  CacheType = "redis"
	CacheValkey CacheType = "valkey"
)

var SelectableFeatures = []string{"api", "mcp", "ui", "db", "cache", "docker", "nomad", "ci", "otel"}

type FeatureSet struct {
	CLI    bool
	API    bool
	MCP    bool
	UI     bool
	DB     bool
	Cache  bool
	Docker bool
	Nomad  bool
	CI     bool
	Otel   bool
}

type ProjectConfig struct {
	AppName      string
	OutputDir    string
	ModulePath   string
	Features     FeatureSet
	DBType       DBType
	UseXDAL      bool
	CacheType    CacheType
	CustomTags   []string
	ResourceName string
}

func (fs *FeatureSet) NeedsSRVResolve() bool {
	return fs.DB || fs.Cache || fs.API
}

func (fs *FeatureSet) HasFeature(name string) bool {
	switch name {
	case "cli":
		return fs.CLI
	case "api":
		return fs.API
	case "mcp":
		return fs.MCP
	case "ui":
		return fs.UI
	case "db":
		return fs.DB
	case "cache":
		return fs.Cache
	case "docker":
		return fs.Docker
	case "nomad":
		return fs.Nomad
	case "ci":
		return fs.CI
	case "otel":
		return fs.Otel
	}
	return false
}

func ResolveFeatures(fs *FeatureSet) {
	fs.CLI = true
	if fs.Nomad {
		fs.Docker = true
	}
	if fs.UI {
		// The UI's built frontend is only reachable through the HTTP server
		// that the api feature provides.
		fs.API = true
	}
}

func EnableFeature(fs *FeatureSet, name string) bool {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "api":
		fs.API = true
	case "mcp":
		fs.MCP = true
	case "ui":
		fs.UI = true
	case "db":
		fs.DB = true
	case "cache":
		fs.Cache = true
	case "docker":
		fs.Docker = true
	case "nomad":
		fs.Nomad = true
	case "ci":
		fs.CI = true
	case "otel":
		fs.Otel = true
	default:
		return false
	}
	return true
}

// NewlyEnabledFeatures returns the SelectableFeatures names that are enabled
// in after but not in before, e.g. to discover features ResolveFeatures
// cascaded on (ui -> api, nomad -> docker) in addition to the one explicitly
// requested.
func NewlyEnabledFeatures(before, after FeatureSet) []string {
	var names []string
	for _, name := range SelectableFeatures {
		if !before.HasFeature(name) && after.HasFeature(name) {
			names = append(names, name)
		}
	}
	return names
}

func ExpandFeatureNames(names []string) []string {
	var expanded []string
	hasAll := false
	for _, name := range names {
		trimmed := strings.TrimSpace(name)
		if strings.EqualFold(trimmed, "all") {
			hasAll = true
			continue
		}
		expanded = append(expanded, trimmed)
	}
	if hasAll {
		return append(append([]string(nil), SelectableFeatures...), expanded...)
	}
	return expanded
}

// appNamePattern restricts app names to characters that are safe in file
// names, Go string literals, Docker image tags and Nomad job names.
var appNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

// modulePathPattern restricts module paths to characters valid in Go import
// paths, so they can be embedded in generated import statements.
var modulePathPattern = regexp.MustCompile(`^[A-Za-z0-9._~/-]+$`)

// ValidateAppName reports whether name is usable as an app name.
func ValidateAppName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("app_name must not be empty")
	}
	if !appNamePattern.MatchString(name) {
		return fmt.Errorf("invalid app_name %q: must start with a letter or digit and contain only letters, digits, '.', '_' or '-'", name)
	}
	return nil
}

// ValidateModulePath reports whether path is usable as a Go module path.
func ValidateModulePath(path string) error {
	if !modulePathPattern.MatchString(path) || strings.HasPrefix(path, "/") || strings.Contains(path, "..") {
		return fmt.Errorf("invalid module_path %q: must be a Go import path such as github.com/org/app", path)
	}
	return nil
}

func (pc *ProjectConfig) Validate() error {
	var errs []string

	if err := ValidateAppName(pc.AppName); err != nil {
		errs = append(errs, err.Error())
	}

	if err := ValidateModulePath(pc.ModulePath); err != nil {
		errs = append(errs, err.Error())
	}

	if strings.TrimSpace(pc.OutputDir) == "" {
		errs = append(errs, "output_dir must not be empty")
	}

	if pc.Features.DB {
		switch pc.DBType {
		case DBMySQL, DBPostgreSQL, DBSQLite:
		default:
			errs = append(errs, fmt.Sprintf("invalid db_type %q: must be mysql, postgresql, or sqlite", pc.DBType))
		}
	} else if pc.UseXDAL {
		errs = append(errs, "use_xdal requires the db feature to be enabled")
	}

	if pc.Features.Cache {
		switch pc.CacheType {
		case CacheRedis, CacheValkey:
		default:
			errs = append(errs, fmt.Sprintf("invalid cache_type %q: must be redis or valkey", pc.CacheType))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errs, "; "))
	}

	return nil
}
