package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/paularlott/cli"

	"github.com/martinsuchenak/go-scaffolder/internal/config"
	"github.com/martinsuchenak/go-scaffolder/internal/configfile"
	"github.com/martinsuchenak/go-scaffolder/internal/engine"
	"github.com/martinsuchenak/go-scaffolder/internal/postgen"
	"github.com/martinsuchenak/go-scaffolder/internal/prompt"
	"github.com/martinsuchenak/go-scaffolder/internal/writer"
	"github.com/martinsuchenak/go-scaffolder/templates"
)

var configPath string
var templatesDir string

func main() {
	app := &cli.Command{
		Name:    "go-scaffolder",
		Usage:   "Scaffold a Go microservice",
		Version: "0.1.0",
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
		},
		Run: runScaffold,
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
