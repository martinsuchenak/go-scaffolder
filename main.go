package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/paularlott/cli"

	"github.com/martinsuchenak/go-scaffolder/internal/config"
	"github.com/martinsuchenak/go-scaffolder/internal/configfile"
	"github.com/martinsuchenak/go-scaffolder/internal/diff"
	"github.com/martinsuchenak/go-scaffolder/internal/engine"
	"github.com/martinsuchenak/go-scaffolder/internal/patcher"
	"github.com/martinsuchenak/go-scaffolder/internal/postgen"
	"github.com/martinsuchenak/go-scaffolder/internal/prompt"
	"github.com/martinsuchenak/go-scaffolder/internal/writer"
	"github.com/martinsuchenak/go-scaffolder/templates"
)

var configPath string
var templatesDir string
var resourceName string
var dbTypeFlag string
var cacheTypeFlag string
var patchMode bool

func main() {
	app := &cli.Command{
		Name:    "go-scaffolder",
		Usage:   "Scaffold a Go microservice",
		Version: "0.2.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config",
				Usage:    "Path to YAML config file for non-interactive mode",
				AssignTo: &configPath,
			},
			&cli.StringFlag{
				Name:     "templates",
				Usage:    "Path to external templates directory (must contain manifest.yaml)",
				AssignTo: &templatesDir,
			},
			&cli.BoolFlag{
				Name:     "patch",
				Usage:    "Output unified diff to stdout instead of writing files",
				AssignTo: &patchMode,
			},
		},
		Commands: []*cli.Command{addCommand(), cli.GenerateCompletionCommand()},
		Run:      runScaffold,
	}

	if err := app.Execute(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runScaffold(ctx context.Context, cmd *cli.Command) error {
	var cfg *config.ProjectConfig
	var err error

	if configPath != "" {
		cfg, err = loadFromConfigFile(configPath)
	} else {
		cfg, err = loadInteractive()
	}
	if err != nil {
		return err
	}

	config.ResolveFeatures(&cfg.Features)

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	eng := engine.New(templates.FS)

	if templatesDir != "" {
		extFS, extEntries, extErr := engine.LoadExternalTemplates(templatesDir)
		if extErr != nil {
			return fmt.Errorf("loading external templates: %w", extErr)
		}
		eng.AddExternalTemplates(extFS, extEntries)
		fmt.Printf("Loaded %d external templates from %s\n", len(extEntries), templatesDir)
	}

	files, err := eng.RenderAll(cfg)
	if err != nil {
		return fmt.Errorf("template rendering error: %w", err)
	}

	if patchMode {
		for path, content := range files {
			fmt.Print(diff.NewFileDiff(path, string(content)))
		}
		return nil
	}

	outputDir, err := filepath.Abs(cfg.OutputDir)
	if err != nil {
		return fmt.Errorf("resolving output directory: %w", err)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	if configPath == "" {
		entries, _ := os.ReadDir(outputDir)
		if len(entries) > 0 {
			p := prompt.NewTerminalPrompter()
			confirmed, err := p.Confirm(fmt.Sprintf("Output directory %s already contains files. Continue?", outputDir))
			if err != nil || !confirmed {
				fmt.Println("Aborted.")
				return nil
			}
		}
	}

	if err := writer.WriteAll(outputDir, files); err != nil {
		return fmt.Errorf("writing files: %w", err)
	}

	stateFilePath := filepath.Join(outputDir, configfile.StateFileName)
	if err := configfile.WriteStateFile(stateFilePath, cfg); err != nil {
		return fmt.Errorf("writing state file: %w", err)
	}

	fmt.Printf("Generated %d files in %s\n", len(files), outputDir)

	fmt.Println("Running go mod tidy...")
	output, err := postgen.RunGoModTidy(outputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: go mod tidy failed: %v\n%s\n", err, output)
		fmt.Println("You can run 'go mod tidy' manually in the output directory.")
	} else {
		fmt.Println("Dependencies resolved successfully.")
	}

	fmt.Printf("\nProject %s scaffolded successfully!\n", cfg.AppName)
	return nil
}

func addCommand() *cli.Command {
	return &cli.Command{
		Name:  "add",
		Usage: "Add a component or enable a feature in an existing scaffolded project",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:     "patch",
				Usage:    "Output unified diff to stdout instead of writing files",
				AssignTo: &patchMode,
				Global:   true,
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "cli-command",
				Usage: "Add a new CLI command",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Usage:    "Name of the new command",
						AssignTo: &resourceName,
					},
				},
				Run: func(ctx context.Context, cmd *cli.Command) error {
					return runAdd(ctx, cmd, "cli-command", "")
				},
			},
			{
				Name:  "api-endpoint",
				Usage: "Add a new API endpoint resource",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Usage:    "Name of the new endpoint resource",
						AssignTo: &resourceName,
					},
				},
				Run: func(ctx context.Context, cmd *cli.Command) error {
					return runAdd(ctx, cmd, "api-endpoint", "api")
				},
			},
			{
				Name:  "mcp-tool",
				Usage: "Add a new MCP tool",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Usage:    "Name of the new MCP tool",
						AssignTo: &resourceName,
					},
				},
				Run: func(ctx context.Context, cmd *cli.Command) error {
					return runAdd(ctx, cmd, "mcp-tool", "mcp")
				},
			},
			enableFeatureCommand(),
		},
	}
}

func enableFeatureCommand() *cli.Command {
	return &cli.Command{
		Name:  "feature",
		Usage: "Enable a feature (api, mcp, ui, db, cache, docker, nomad)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "db-type",
				Usage:    "Database type (mysql, postgresql, sqlite) -- required when enabling db",
				AssignTo: &dbTypeFlag,
			},
			&cli.StringFlag{
				Name:     "cache-type",
				Usage:    "Cache type (redis, valkey) -- required when enabling cache",
				AssignTo: &cacheTypeFlag,
			},
		},
		Run: runEnableFeature,
	}
}

var enableableFeatures = []string{"api", "mcp", "ui", "db", "cache", "docker", "nomad"}

func runEnableFeature(ctx context.Context, cmd *cli.Command) error {
	cfg, err := configfile.LoadStateFile(configfile.StateFileName)
	if err != nil {
		return fmt.Errorf("not a go-scaffolder project (missing %s): %w", configfile.StateFileName, err)
	}

	p := prompt.NewTerminalPrompter()

	var available []string
	for _, f := range enableableFeatures {
		if !cfg.Features.HasFeature(f) {
			available = append(available, f)
		}
	}
	if len(available) == 0 {
		fmt.Println("All features are already enabled.")
		return nil
	}

	feature, err := p.AskSelect("Select feature to enable", available)
	if err != nil {
		return err
	}

	if feature == "db" {
		if dbTypeFlag != "" {
			cfg.DBType = config.DBType(dbTypeFlag)
		} else {
			dbStr, askErr := p.AskSelect("Select database type", prompt.DBTypeOptions)
			if askErr != nil {
				return askErr
			}
			cfg.DBType = config.DBType(dbStr)
		}
	}

	if feature == "cache" {
		if cacheTypeFlag != "" {
			cfg.CacheType = config.CacheType(cacheTypeFlag)
		} else {
			cacheStr, askErr := p.AskSelect("Select cache type", prompt.CacheTypeOptions)
			if askErr != nil {
				return askErr
			}
			cfg.CacheType = config.CacheType(cacheStr)
		}
	}

	if feature == "nomad" {
		cfg.Features.Nomad = true
		cfg.Features.Docker = true
	}

	switch feature {
	case "api":
		cfg.Features.API = true
	case "mcp":
		cfg.Features.MCP = true
	case "ui":
		cfg.Features.UI = true
	case "db":
		cfg.Features.DB = true
	case "cache":
		cfg.Features.Cache = true
	case "docker":
		cfg.Features.Docker = true
	case "nomad":
		cfg.Features.Nomad = true
	}

	eng := engine.New(templates.FS)

	var allEntries []engine.TemplateEntry

	entries := eng.EnableFeatureManifest(feature)
	if entries != nil {
		allEntries = append(allEntries, entries...)
	}

	if feature == "cache" {
		cacheEntries := eng.EnableFeatureCacheManifest(cfg.CacheType)
		allEntries = append(allEntries, cacheEntries...)
	}

	needsSRV := cfg.Features.NeedsSRVResolve()
	if needsSRV {
		if _, statErr := os.Stat("internal/resolve/resolve.go"); os.IsNotExist(statErr) {
			allEntries = append(allEntries, eng.EnableFeatureSRVManifest()...)
		}
	}

	if patchMode {
		if len(allEntries) > 0 {
			files, renderErr := eng.RenderFeatureFiles(cfg, allEntries)
			if renderErr != nil {
				return fmt.Errorf("template rendering error: %w", renderErr)
			}
			for path, content := range files {
				if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
					fmt.Print(diff.NewFileDiff(path, string(content)))
				}
			}
		}

		patches := patcher.FeaturePatches(feature, cfg)
		if len(patches) > 0 {
			computed := patcher.ComputePatches(".", patches)
			for _, cp := range computed {
				if cp.Applied {
					fmt.Print(diff.UnifiedDiff(cp.File, cp.File, cp.OldContent, cp.NewContent))
				} else {
					fmt.Fprintf(os.Stderr, "# Could not compute patch for %s (%s)\n", cp.File, cp.Description)
					fmt.Fprintf(os.Stderr, "# Add manually:\n%s\n", cp.Content)
				}
			}
		}
		return nil
	}

	if len(allEntries) > 0 {
		files, renderErr := eng.RenderFeatureFiles(cfg, allEntries)
		if renderErr != nil {
			return fmt.Errorf("template rendering error: %w", renderErr)
		}

		newFiles := make(map[string][]byte)
		for path, content := range files {
			if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
				newFiles[path] = content
			}
		}

		if len(newFiles) > 0 {
			if writeErr := writer.WriteAll(".", newFiles); writeErr != nil {
				return fmt.Errorf("writing files: %w", writeErr)
			}
			for path := range newFiles {
				fmt.Printf("  created: %s\n", path)
			}
		}
	}

	patches := patcher.FeaturePatches(feature, cfg)
	if len(patches) > 0 {
		results := patcher.ApplyPatches(".", patches)
		patcher.ReportResults(results)
	}

	if err := configfile.WriteStateFile(configfile.StateFileName, cfg); err != nil {
		return fmt.Errorf("updating state file: %w", err)
	}

	fmt.Println("Running go mod tidy...")
	output, tidyErr := postgen.RunGoModTidy(".")
	if tidyErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: go mod tidy failed: %v\n%s\n", tidyErr, output)
	} else {
		fmt.Println("Dependencies resolved successfully.")
	}

	fmt.Printf("\nFeature %q enabled successfully!\n", feature)
	return nil
}

func runAdd(ctx context.Context, cmd *cli.Command, addType string, requiredFeature string) error {
	cfg, err := configfile.LoadStateFile(configfile.StateFileName)
	if err != nil {
		return fmt.Errorf("not a go-scaffolder project (missing %s): %w", configfile.StateFileName, err)
	}

	if requiredFeature != "" && !cfg.Features.HasFeature(requiredFeature) {
		return fmt.Errorf("this project was not scaffolded with the %q feature; enable it first with: go-scaffolder add feature", requiredFeature)
	}

	name := resourceName
	if name == "" {
		p := prompt.NewTerminalPrompter()
		name, err = p.AskString("Resource name", prompt.ValidateResourceName)
		if err != nil {
			return err
		}
	}

	if err := prompt.ValidateResourceName(name); err != nil {
		return err
	}

	cfg.ResourceName = name

	eng := engine.New(templates.FS)
	files, err := eng.RenderAdd(cfg, addType)
	if err != nil {
		return fmt.Errorf("template rendering error: %w", err)
	}

	if patchMode {
		for path, content := range files {
			fmt.Print(diff.NewFileDiff(path, string(content)))
		}
		return nil
	}

	for relPath := range files {
		if _, statErr := os.Stat(relPath); statErr == nil {
			return fmt.Errorf("file already exists: %s (will not overwrite)", relPath)
		}
	}

	if err := writer.WriteAll(".", files); err != nil {
		return fmt.Errorf("writing files: %w", err)
	}

	fmt.Printf("Added %d files for %s %q:\n", len(files), addType, name)
	for path := range files {
		fmt.Printf("  created: %s\n", path)
	}

	fmt.Println("Running go mod tidy...")
	output, tidyErr := postgen.RunGoModTidy(".")
	if tidyErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: go mod tidy failed: %v\n%s\n", tidyErr, output)
	} else {
		fmt.Println("Dependencies resolved successfully.")
	}

	return nil
}

func loadFromConfigFile(path string) (*config.ProjectConfig, error) {
	sc, err := configfile.Load(path)
	if err != nil {
		return nil, err
	}
	return sc.ToProjectConfig()
}

func loadInteractive() (*config.ProjectConfig, error) {
	p := prompt.NewTerminalPrompter()
	return prompt.CollectConfig(p)
}
