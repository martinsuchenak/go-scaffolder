package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/martinsuchenak/go-scaffolder/internal/config"
)

type Prompter interface {
	AskString(label string, validate func(string) error) (string, error)
	AskMultiSelect(label string, options []string) ([]string, error)
	AskSelect(label string, options []string) (string, error)
	Confirm(label string) (bool, error)
}

type TerminalPrompter struct {
	reader *bufio.Reader
}

func NewTerminalPrompter() *TerminalPrompter {
	return &TerminalPrompter{reader: bufio.NewReader(os.Stdin)}
}

func (p *TerminalPrompter) AskString(label string, validate func(string) error) (string, error) {
	for {
		fmt.Printf("%s: ", label)
		input, err := p.reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("reading input: %w", err)
		}
		input = strings.TrimSpace(input)
		if validate != nil {
			if err := validate(input); err != nil {
				fmt.Printf("  Error: %s\n", err)
				continue
			}
		}
		return input, nil
	}
}

func (p *TerminalPrompter) AskMultiSelect(label string, options []string) ([]string, error) {
	fmt.Printf("%s (comma-separated numbers):\n", label)
	for i, opt := range options {
		fmt.Printf("  %d) %s\n", i+1, opt)
	}
	fmt.Print("Selection: ")
	input, err := p.reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("reading input: %w", err)
	}

	var selected []string
	parts := strings.Split(strings.TrimSpace(input), ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		idx, err := strconv.Atoi(part)
		if err != nil || idx < 1 || idx > len(options) {
			return nil, fmt.Errorf("invalid selection: %s", part)
		}
		selected = append(selected, options[idx-1])
	}
	return selected, nil
}

func (p *TerminalPrompter) AskSelect(label string, options []string) (string, error) {
	fmt.Printf("%s:\n", label)
	for i, opt := range options {
		fmt.Printf("  %d) %s\n", i+1, opt)
	}
	fmt.Print("Selection: ")
	input, err := p.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("reading input: %w", err)
	}
	idx, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil || idx < 1 || idx > len(options) {
		return "", fmt.Errorf("invalid selection: %s", strings.TrimSpace(input))
	}
	return options[idx-1], nil
}

func (p *TerminalPrompter) Confirm(label string) (bool, error) {
	fmt.Printf("%s [y/N]: ", label)
	input, err := p.reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("reading input: %w", err)
	}
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes", nil
}

func ValidateAppName(s string) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("app name must not be empty")
	}
	return nil
}

func ValidateOutputDir(s string) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("output directory must not be empty")
	}
	return nil
}

func ValidateResourceName(s string) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("resource name must not be empty")
	}
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return fmt.Errorf("resource name must contain only alphanumeric characters, hyphens, and underscores")
		}
	}
	return nil
}

var FeatureOptions = []string{"API", "MCP", "UI", "DB", "Cache", "Docker", "Nomad"}
var DBTypeOptions = []string{"mysql", "postgresql", "sqlite"}
var CacheTypeOptions = []string{"redis", "valkey"}

func CollectConfig(p Prompter) (*config.ProjectConfig, error) {
	appName, err := p.AskString("App Name", ValidateAppName)
	if err != nil {
		return nil, err
	}

	modulePath, err := p.AskString(fmt.Sprintf("Module Path (default: %s)", appName), nil)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(modulePath) == "" {
		modulePath = appName
	}

	outputDir, err := p.AskString("Output Directory", ValidateOutputDir)
	if err != nil {
		return nil, err
	}

	features, err := p.AskMultiSelect("Select features", FeatureOptions)
	if err != nil {
		return nil, err
	}

	fs := config.FeatureSet{}
	for _, f := range features {
		switch strings.ToLower(f) {
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
		}
	}

	config.ResolveFeatures(&fs)

	var dbType config.DBType
	if fs.DB {
		dbStr, err := p.AskSelect("Select database type", DBTypeOptions)
		if err != nil {
			return nil, err
		}
		dbType = config.DBType(dbStr)
	}

	var cacheType config.CacheType
	if fs.Cache {
		cacheStr, err := p.AskSelect("Select cache type", CacheTypeOptions)
		if err != nil {
			return nil, err
		}
		cacheType = config.CacheType(cacheStr)
	}

	pc := &config.ProjectConfig{
		AppName:    appName,
		OutputDir:  outputDir,
		ModulePath: modulePath,
		Features:   fs,
		DBType:     dbType,
		CacheType:  cacheType,
	}

	return pc, nil
}
