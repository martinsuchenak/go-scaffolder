package config

import (
	"fmt"
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

type FeatureSet struct {
	CLI    bool
	API    bool
	MCP    bool
	UI     bool
	DB     bool
	Cache  bool
	Docker bool
	Nomad  bool
}

type ProjectConfig struct {
	AppName      string
	OutputDir    string
	ModulePath   string
	Features     FeatureSet
	DBType       DBType
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
	}
	return false
}

func ResolveFeatures(fs *FeatureSet) {
	fs.CLI = true
	if fs.Nomad {
		fs.Docker = true
	}
}

func (pc *ProjectConfig) Validate() error {
	var errs []string

	if strings.TrimSpace(pc.AppName) == "" {
		errs = append(errs, "app_name must not be empty")
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
