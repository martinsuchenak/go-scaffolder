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
		if fs.UI && !fs.API {
			t.Fatal("if UI is true, API must also be true after ResolveFeatures, since the built frontend is only served over the api's HTTP server")
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
				AppName:   "myapp",
				OutputDir: "/tmp/out",
				Features:  FeatureSet{DB: true},
				DBType:    "invalid",
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
			name: "xdal without db",
			pc: ProjectConfig{
				AppName:   "myapp",
				OutputDir: "/tmp/out",
				UseXDAL:   true,
			},
			wantErr: true,
			errMsg:  "use_xdal requires the db feature",
		},
		{
			name: "valid with DB postgresql",
			pc: ProjectConfig{
				AppName:   "myapp",
				OutputDir: "/tmp/out",
				Features:  FeatureSet{DB: true},
				DBType:    DBPostgreSQL,
				UseXDAL:   true,
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

func TestEnableFeature(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantOK   bool
		wantFlag func(fs *FeatureSet) bool
	}{
		{"api", "api", true, func(fs *FeatureSet) bool { return fs.API }},
		{"API uppercase", "API", true, func(fs *FeatureSet) bool { return fs.API }},
		{" Api  spaces", " Api ", true, func(fs *FeatureSet) bool { return fs.API }},
		{"mcp", "mcp", true, func(fs *FeatureSet) bool { return fs.MCP }},
		{"db", "db", true, func(fs *FeatureSet) bool { return fs.DB }},
		{"cache", "cache", true, func(fs *FeatureSet) bool { return fs.Cache }},
		{"docker", "docker", true, func(fs *FeatureSet) bool { return fs.Docker }},
		{"nomad", "nomad", true, func(fs *FeatureSet) bool { return fs.Nomad }},
		{"ui", "ui", true, func(fs *FeatureSet) bool { return fs.UI }},
		{"unknown", "unknown", false, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := &FeatureSet{}
			gotOK := EnableFeature(fs, tt.input)
			if gotOK != tt.wantOK {
				t.Errorf("EnableFeature(%q) = %v, want %v", tt.input, gotOK, tt.wantOK)
			}
			if tt.wantFlag != nil && !tt.wantFlag(fs) {
				t.Errorf("feature flag not set for input %q", tt.input)
			}
		})
	}
}

func TestExpandFeatureNames(t *testing.T) {
	tests := []struct {
		name  string
		input []string
	}{
		{"all shortcut", []string{"all"}},
		{"all with extra", []string{"all", "custom-tag"}},
		{"specific features", []string{"api", " db "}},
		{"empty", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExpandFeatureNames(tt.input)
			if len(got) == 0 && len(tt.input) > 0 {
				t.Errorf("ExpandFeatureNames(%v) returned empty", tt.input)
			}
		})
	}
}

func TestHasFeature(t *testing.T) {
	tests := []struct {
		name    string
		fs      FeatureSet
		feature string
		want    bool
	}{
		{"cli true", FeatureSet{CLI: true}, "cli", true},
		{"api true", FeatureSet{API: true}, "api", true},
		{"mcp false", FeatureSet{}, "mcp", false},
		{"ui true", FeatureSet{UI: true}, "ui", true},
		{"db true", FeatureSet{DB: true}, "db", true},
		{"cache true", FeatureSet{Cache: true}, "cache", true},
		{"docker true", FeatureSet{Docker: true}, "docker", true},
		{"nomad true", FeatureSet{Nomad: true}, "nomad", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fs.HasFeature(tt.feature); got != tt.want {
				t.Errorf("HasFeature(%q) = %v, want %v", tt.feature, got, tt.want)
			}
		})
	}
}
