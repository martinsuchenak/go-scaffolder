package config

import (
	"strings"
	"testing"

	"pgregory.net/rapid"
)

func genFeatureSet() *rapid.Generator[FeatureSet] {
	return rapid.Custom(func(t *rapid.T) FeatureSet {
		return FeatureSet{
			CLI:    rapid.Bool().Draw(t, "CLI"),
			API:    rapid.Bool().Draw(t, "API"),
			MCP:    rapid.Bool().Draw(t, "MCP"),
			UI:     rapid.Bool().Draw(t, "UI"),
			DB:     rapid.Bool().Draw(t, "DB"),
			Cache:  rapid.Bool().Draw(t, "Cache"),
			Docker: rapid.Bool().Draw(t, "Docker"),
			Nomad:  rapid.Bool().Draw(t, "Nomad"),
		}
	})
}

func genValidProjectConfig() *rapid.Generator[ProjectConfig] {
	return rapid.Custom(func(t *rapid.T) ProjectConfig {
		appName := rapid.StringMatching(`[a-z][a-z0-9\-]{2,29}`).Draw(t, "AppName")
		fs := genFeatureSet().Draw(t, "Features")
		ResolveFeatures(&fs)

		var dbType DBType
		if fs.DB {
			dbType = rapid.SampledFrom([]DBType{DBMySQL, DBPostgreSQL, DBSQLite}).Draw(t, "DBType")
		}

		var cacheType CacheType
		if fs.Cache {
			cacheType = rapid.SampledFrom([]CacheType{CacheRedis, CacheValkey}).Draw(t, "CacheType")
		}

		return ProjectConfig{
			AppName:    appName,
			OutputDir:  "/tmp/test-" + appName,
			ModulePath: appName,
			Features:   fs,
			DBType:     dbType,
			CacheType:  cacheType,
		}
	})
}

// Feature: go-scaffolder, Property 1: Feature resolution invariants
func TestProperty1_FeatureResolutionInvariants(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		fs := genFeatureSet().Draw(t, "features")
		ResolveFeatures(&fs)

		if !fs.CLI {
			t.Fatal("CLI must always be true after ResolveFeatures")
		}
		if fs.Nomad && !fs.Docker {
			t.Fatal("if Nomad is true, Docker must also be true after ResolveFeatures")
		}
	})
}

// Feature: go-scaffolder, Property 2: App name validation rejects empty/whitespace input
func TestProperty2_AppNameValidationRejectsEmptyWhitespace(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		ws := rapid.StringMatching(`^[\s]*$`).Draw(t, "whitespace-name")
		pc := ProjectConfig{
			AppName:   ws,
			OutputDir: "/tmp/test",
		}
		if err := pc.Validate(); err == nil {
			t.Fatalf("expected validation error for app_name %q, got nil", ws)
		}
	})
}

func TestNeedsSRVResolve(t *testing.T) {
	tests := []struct {
		name   string
		fs     FeatureSet
		expect bool
	}{
		{"no features", FeatureSet{}, false},
		{"DB only", FeatureSet{DB: true}, true},
		{"Cache only", FeatureSet{Cache: true}, true},
		{"API only", FeatureSet{API: true}, true},
		{"all three", FeatureSet{DB: true, Cache: true, API: true}, true},
		{"unrelated features", FeatureSet{CLI: true, Docker: true, Nomad: true}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fs.NeedsSRVResolve(); got != tt.expect {
				t.Errorf("NeedsSRVResolve() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		pc      ProjectConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid minimal config",
			pc:      ProjectConfig{AppName: "myapp", OutputDir: "/tmp/out"},
			wantErr: false,
		},
		{
			name:    "empty app name",
			pc:      ProjectConfig{AppName: "", OutputDir: "/tmp/out"},
			wantErr: true,
			errMsg:  "app_name must not be empty",
		},
		{
			name:    "whitespace app name",
			pc:      ProjectConfig{AppName: "   ", OutputDir: "/tmp/out"},
			wantErr: true,
			errMsg:  "app_name must not be empty",
		},
		{
			name:    "empty output dir",
			pc:      ProjectConfig{AppName: "myapp", OutputDir: ""},
			wantErr: true,
			errMsg:  "output_dir must not be empty",
		},
		{
			name: "DB enabled without valid type",
			pc: ProjectConfig{
				AppName:  "myapp",
				OutputDir: "/tmp/out",
				Features: FeatureSet{DB: true},
				DBType:   "invalid",
			},
			wantErr: true,
			errMsg:  "invalid db_type",
		},
		{
			name: "Cache enabled without valid type",
			pc: ProjectConfig{
				AppName:   "myapp",
				OutputDir: "/tmp/out",
				Features:  FeatureSet{Cache: true},
				CacheType: "invalid",
			},
			wantErr: true,
			errMsg:  "invalid cache_type",
		},
		{
			name: "valid with DB postgresql",
			pc: ProjectConfig{
				AppName:  "myapp",
				OutputDir: "/tmp/out",
				Features: FeatureSet{DB: true},
				DBType:   DBPostgreSQL,
			},
			wantErr: false,
		},
		{
			name: "valid with Cache redis",
			pc: ProjectConfig{
				AppName:   "myapp",
				OutputDir: "/tmp/out",
				Features:  FeatureSet{Cache: true},
				CacheType: CacheRedis,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.pc.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}
