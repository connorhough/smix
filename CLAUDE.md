# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

smix is a Go CLI toolkit that provides AI-powered development utilities. It uses Cobra for command structure and Viper for configuration management.

## Essential Commands

### Build & Install
```bash
make build          # Build binary to builds/ directory with version injection
make install        # Install binary to GOPATH/bin
make clean          # Remove build artifacts
go mod tidy         # Update dependencies
```

### Testing
```bash
make test                                    # Run all tests
go test -v ./path/to/package -run TestName  # Run single test
```

### Code Quality
```bash
make lint   # Run golangci-lint checks
```

### Cross-compilation
```bash
make build-darwin-arm64   # Build for macOS ARM64
make build-linux-amd64    # Build for Linux x86_64
```

## Architecture

### Package Structure

The codebase follows strict separation of concerns:

- **cmd/**: CLI wiring only (commands, flags, argument parsing)
  - Commands MUST NOT contain business logic
  - Commands return errors to main.go rather than calling os.Exit()
  - Each command file defines Cobra command structure and delegates to internal packages

- **internal/**: All business logic
  - `pr/`: GitHub PR code review processing with gemini-code-assist bot
    - `fetch.go`: Fetches PR review comments and creates prompt files
    - `process.go`: Generates patches via LLM and launches Claude Code sessions
  - `do/`: Natural language to shell command translation
  - `ask/`: Answers short technical questions
  - `llm/`: Provider interface, error types, retry logic, and options
  - `llm/claude/`: Claude provider implementation (wraps Claude Code CLI)
  - `llm/gemini/`: Gemini provider implementation (uses Google AI SDK)
  - `providers/`: Provider factory with caching
  - `config/`: Configuration management wrapper around Viper
  - `version/`: Version info injected at build time

### Configuration System

Uses Viper with XDG-compliant configuration paths:
1. `$XDG_CONFIG_HOME/smix/config.yaml`
2. `~/.config/smix/config.yaml`
3. `~/.smix.yaml`

Config files are automatically created from a template if they don't exist. Environment variables prefixed with `SMIX_` override config file values.

### Global Flags

The root command supports these persistent flags across all subcommands:
- `--config <path>`: Specify custom config file location
- `--debug`: Enable debug output (overrides config `log_level`)
- `--provider <name>`: Override LLM provider (claude, gemini)
- `--model <name>`: Override model name

### Version Injection Pattern

Version information is injected at build time via ldflags (see Makefile lines 12-15):
- `internal/version.Version` = git tag or commit hash
- `internal/version.GitCommit` = short commit SHA
- `internal/version.BuildDate` = build timestamp

This allows tracking exact builds without hardcoding versions.

## Key Commands

### pr

Fetches code review feedback from gemini-code-assist bot on GitHub PRs and launches interactive Claude Code sessions to analyze and implement the suggested changes.

```bash
smix pr review owner/repo pr_number
smix pr review --dir pr_review_pr123  # Process existing feedback directory
```

**Requirements:**
- `GITHUB_TOKEN` env var (optional, increases rate limits)
- `claude` CLI installed (Claude Code)

**Workflow:**
1. Fetches review comments and diff context from GitHub PR
2. Filters for gemini-code-assist bot comments
3. Creates individual prompt files with:
   - Line-numbered code snapshots
   - Git diff hunks showing PR changes
   - Direct links to comment threads
   - Structured decision format (APPLY/REJECT)
4. For each feedback item:
   - Launches Claude Code session with explicit target file and constraints
   - Claude evaluates feedback against codebase patterns and correctness
   - Implements changes with clear reasoning and documentation
   - Provides batch progress (item X of Y)

**Prompt Features:**
- Explicit target file paths for modifications
- Line-numbered code snippets for precise navigation
- PR diff context to understand what changed
- Structured output format (Decision/Reasoning/Changes)
- Clear autonomy constraints (no tests, no commits)
- Project conventions awareness (.editorconfig, CONVENTIONS.md)

### do

Translates natural language task descriptions into shell commands.

```bash
smix do "list all files in the current directory"
smix do --provider gemini "find large files"
```

**Requirements:**
- Configured LLM provider (Claude or Gemini)
- Default: Claude (requires Claude Code CLI)
- Gemini: Set `SMIX_GEMINI_API_KEY` environment variable

Generates safe, POSIX-compliant shell commands using your configured provider.

### ask

Answers short technical questions with concise, accurate responses.

```bash
smix ask "what is the difference between TCP and UDP"
smix ask --provider gemini "how do I list all running processes on Linux"
```

**Requirements:**
- Configured LLM provider (Claude or Gemini)
- Default: Claude (requires Claude Code CLI)
- Gemini: Set `SMIX_GEMINI_API_KEY` environment variable

Great for quick lookups and technical questions where you need a brief, informative answer without searching documentation or web resources.

### config

Manage smix configuration values.

```bash
smix config get <key>     # Get a configuration value
smix config set <key> <value>  # Set a configuration value
```

**Examples:**
```bash
smix config get provider           # Get current provider
smix config set provider gemini    # Set global provider to Gemini
smix config set commands.ask.model haiku  # Set ask command to use Haiku model
```

## LLM Integration

smix supports multiple LLM providers through a unified interface:

### Architecture

- **`internal/llm/`** - Core provider interface, error types, retry logic, and options
- **`internal/llm/claude/`** - Claude provider (wraps Claude Code CLI)
- **`internal/llm/gemini/`** - Gemini provider (uses Google AI SDK)
- **`internal/providers/`** - Provider factory with caching

### Supported Providers

**Claude (via Claude Code CLI):**
- Wraps `claude -p "prompt"` in subprocess
- Models: `haiku`, `sonnet`, `opus`
- Requires: Claude Code CLI installed and authenticated

**Gemini (via Google AI SDK):**
- Uses `google.golang.org/genai` SDK
- Models: `gemini-3-flash-preview`, `gemini-3-pro-preview`
- Requires: `SMIX_GEMINI_API_KEY` environment variable

### Adding New LLM-Powered Features

Use the provider interface for consistent behavior:

```go
import (
    "context"
    "github.com/connorhough/smix/internal/config"
    "github.com/connorhough/smix/internal/llm"
    "github.com/connorhough/smix/internal/providers"
)

// Get provider from factory (API keys retrieved from env vars automatically)
cfg := config.ResolveProviderConfig("commandname")
provider, err := providers.GetProvider(cfg.Provider)

// Generate response
opts := []llm.Option{llm.WithModel(cfg.Model)}
response, err := provider.Generate(ctx, prompt, opts...)
```

Key benefits:
- Automatic retry with exponential backoff
- Typed error handling (auth failures, rate limits, etc.)
- Provider caching for performance
- Configurable per command or globally

## Adding New Commands

Follow the established pattern to maintain separation of concerns:

1. Create business logic in `internal/newcommand/newcommand.go`:
```go
package newcommand

func Run(param string) error {
    // Business logic here
    return nil
}
```

2. Create CLI wiring in `cmd/newcommand.go`:
```go
package cmd

func newNewCommandCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "newcommand",
        Short: "Description",
        RunE: func(cmd *cobra.Command, args []string) error {
            return newcommand.Run(param)
        },
    }
    return cmd
}

func init() {
    rootCmd.AddCommand(newNewCommandCmd())
}
```

3. Commands MUST return errors rather than calling os.Exit() directly (main.go handles exit codes)

## Code Style

- Always check errors explicitly
- Comments are for explaining complex functionality or improving future developer's ability to read and understand code quickly. They should be used intentionally, and sparingly so as to not clutter the file.
- Do not add any emojis or the "genereated with/co-authored by claude" content to commits. It adds too much bloat to the git history


