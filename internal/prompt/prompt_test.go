package prompt

import (
	"testing"

	"github.com/martinsuchenak/go-scaffolder/internal/config"
)

func TestValidateAppName(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"myapp", false},
		{"", true},
		{"   ", true},
		{"valid-name", false},
	}
	for _, tt := range tests {
		if err := ValidateAppName(tt.input); (err != nil) != tt.wantErr {
			t.Errorf("ValidateAppName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
		}
	}
}

func TestValidateOutputDir(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"/tmp/out", false},
		{"", true},
		{"   ", true},
	}
	for _, tt := range tests {
		if err := ValidateOutputDir(tt.input); (err != nil) != tt.wantErr {
			t.Errorf("ValidateOutputDir(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
		}
	}
}

func TestValidateResourceName(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"my-resource", false},
		{"my_resource", false},
		{"MyResource", false},
		{"resource123", false},
		{"", true},
		{"   ", true},
		{"my resource", true},
		{"my.resource", true},
		{"my/resource", true},
	}
	for _, tt := range tests {
		if err := ValidateResourceName(tt.input); (err != nil) != tt.wantErr {
			t.Errorf("ValidateResourceName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
		}
	}
}

type stubPrompter struct {
	askStringValues      []string
	askMultiSelectValues []string
	askSelectValues      []string
	confirmValues        []bool
}

func (p *stubPrompter) AskString(label string, validate func(string) error) (string, error) {
	value := p.askStringValues[0]
	p.askStringValues = p.askStringValues[1:]
	return value, nil
}

func (p *stubPrompter) AskMultiSelect(label string, options []string) ([]string, error) {
	values := append([]string(nil), p.askMultiSelectValues...)
	p.askMultiSelectValues = nil
	return values, nil
}

func (p *stubPrompter) AskSelect(label string, options []string) (string, error) {
	value := p.askSelectValues[0]
	p.askSelectValues = p.askSelectValues[1:]
	return value, nil
}

func (p *stubPrompter) Confirm(label string) (bool, error) {
	value := p.confirmValues[0]
	p.confirmValues = p.confirmValues[1:]
	return value, nil
}

func TestCollectConfig_AllFeaturesAndXDAL(t *testing.T) {
	p := &stubPrompter{
		askStringValues:      []string{"myapp", "github.com/example/myapp", "./out"},
		askMultiSelectValues: []string{"All"},
		askSelectValues:      []string{"postgresql", "redis"},
		confirmValues:        []bool{true, true},
	}

	cfg, err := CollectConfig(p)
	if err != nil {
		t.Fatalf("CollectConfig error: %v", err)
	}

	if !cfg.Features.API || !cfg.Features.MCP || !cfg.Features.UI || !cfg.Features.DB || !cfg.Features.Cache || !cfg.Features.Docker || !cfg.Features.Nomad {
		t.Fatal("all selectable features should be enabled when All is selected")
	}
	if cfg.DBType != config.DBPostgreSQL {
		t.Fatalf("DBType = %q, want %q", cfg.DBType, config.DBPostgreSQL)
	}
	if !cfg.UseXDAL {
		t.Fatal("UseXDAL should be enabled when both confirmations are accepted")
	}
	if cfg.CacheType != config.CacheRedis {
		t.Fatalf("CacheType = %q, want %q", cfg.CacheType, config.CacheRedis)
	}
}

func TestCollectDBOptions_NoSwap(t *testing.T) {
	p := &stubPrompter{
		askSelectValues: []string{"sqlite"},
		confirmValues:   []bool{false},
	}

	dbType, useXDAL, err := CollectDBOptions(p)
	if err != nil {
		t.Fatalf("CollectDBOptions error: %v", err)
	}
	if dbType != config.DBSQLite {
		t.Fatalf("DBType = %q, want %q", dbType, config.DBSQLite)
	}
	if useXDAL {
		t.Fatal("UseXDAL should be false when there is no DB swap plan")
	}
}
