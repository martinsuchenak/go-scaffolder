# Go Scaffolder Documentation

A CLI tool that generates fully functional Go microservice projects.

## Table of Contents

- [Getting Started](getting-started.md) -- installation, quick start, CLI reference, and what happens during scaffolding
- [Features](features.md) -- detailed description of each selectable feature and what it generates
- [Adding Components](adding-components.md) -- add CLI commands, API endpoints, and MCP tools to existing projects
- [MCP Server](mcp-server.md) -- run the scaffolder as an MCP server for LLM integration, tools reference, authorization, project context resolution
- [How Templates Work](templates.md) -- template structure, manifest, data context, functions, feature guards, and rendering pipeline
- [Extending with Custom Templates](extending.md) -- adding new templates, overriding built-in ones, custom feature tags, and examples
- [Config File Reference](config-file.md) -- YAML schema for non-interactive mode, validation rules, and examples
- [Generated Project Structure](generated-project.md) -- directory layout, configuration, build tasks, architecture patterns, and running the app
- [Testing](testing.md) -- property-based tests, unit tests, integration tests, and how to run them

## Architecture Overview

```
                         ┌──────────────┐
                         │   CLI flags   │
                         │  --config     │
                         │  --templates  │
                         └──────┬───────┘
                                │
                    ┌───────────▼───────────┐
                    │   Input Collection     │
                    │  (prompts or YAML)     │
                    └───────────┬───────────┘
                                │
                    ┌───────────▼───────────┐
                    │   Config Validation    │
                    │   Feature Resolution   │
                    └───────────┬───────────┘
                                │
          ┌─────────────────────▼─────────────────────┐
          │            Template Engine                  │
          │                                            │
          │  ┌──────────────┐  ┌───────────────────┐  │
          │  │  Embedded FS  │  │  External FS      │  │
          │  │  (go:embed)   │  │  (--templates)    │  │
          │  └──────┬───────┘  └───────┬───────────┘  │
          │         │    Merged Manifest│              │
          │         └────────┬─────────┘              │
          │                  │                         │
          │         Render to memory                   │
          └──────────────────┬────────────────────────┘
                             │
                   ┌─────────▼─────────┐
                   │   File Writer      │
                   │  (atomic: all or   │
                   │   nothing)         │
                   └─────────┬─────────┘
                             │
                   ┌─────────▼─────────┐
                   │   Post-generation  │
                   │   go mod tidy      │
                   └─────────┬─────────┘
                             │
                   ┌─────────▼─────────┐
                   │   State File       │
                   │  .go-scaffolder    │
                   │  .yaml             │
                   └───────────────────┘

                   ┌───────────────────┐
                   │   Add Component    │
                   │                    │
                   │  Read state file   │
                   │  → Render snippet  │
                   │  → Write new files │
                   │  → go mod tidy     │
                   └───────────────────┘

          ┌────────────────────────────────┐
          │         MCP Server             │
          │                                │
          │  HTTP /mcp endpoint            │
          │  ┌──────────────────────────┐  │
          │  │ scaffold                 │  │
          │  │ add_cli_command          │  │
          │  │ add_api_endpoint         │  │
          │  │ add_mcp_tool             │  │
          │  │ enable_feature           │  │
          │  │ project_context          │  │
          │  └──────────────────────────┘  │
          │                                │
          │  Context resolution:           │
          │  state_file → project_dir      │
          │  → CWD → app_name fallback     │
          │                                │
          │  Output: unified diffs         │
          └────────────────────────────────┘
```
