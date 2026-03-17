package prompt

import (
	"testing"
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
