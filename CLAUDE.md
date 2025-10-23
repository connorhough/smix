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
  - `gca/`: GitHub PR code review processing with gemini-code-assist bot
    - `fetch.go`: Fetches PR review comments and creates prompt files
    - `process.go`: Generates patches via LLM and launches Charm Crush sessions
  - `do/`: Natural language to shell command translation
  - `llm/`: Shared LLM client (Cerebras API via OpenAI SDK)
  - `config/`: Configuration management wrapper around Viper
  - `version/`: Version info injected at build time

### Configuration System

Uses Viper with XDG-compliant configuration paths:
1. `$XDG_CONFIG_HOME/smix/config.yaml`
2. `~/.config/smix/config.yaml`
3. `~/.smix.yaml`

Environment variables prefixed with `SMIX_` override config file values.

### Version Injection Pattern

Version information is injected at build time via ldflags (see Makefile lines 12-15):
- `internal/version.Version` = git tag or commit hash
- `internal/version.GitCommit` = short commit SHA
- `internal/version.BuildDate` = build timestamp

This allows tracking exact builds without hardcoding versions.

## Key Commands

### gca (Gemini Code Assist)

Fetches code review feedback from gemini-code-assist bot on GitHub PRs, generates patches via LLM, and launches interactive Charm Crush sessions for refinement.

```bash
smix gca review owner/repo pr_number
```

**Requirements:**
- `GITHUB_TOKEN` env var (optional, increases rate limits)
- `CEREBRAS_API_KEY` env var (required for patch generation)
- `crush` CLI installed (Charm Crush)

**Workflow:**
1. Fetches review comments from GitHub PR
2. Filters for gemini-code-assist bot comments
3. Creates individual prompt files with code context
4. For each feedback item:
   - Generates git-style diff patch using LLM
   - Copies review prompt to clipboard
   - Launches Crush session for interactive refinement
5. Saves results to `gca_review_pr<N>/results/`

### do

Translates natural language task descriptions into shell commands.

```bash
smix do "list all files in the current directory"
```

**Requirements:**
- `CEREBRAS_API_KEY` env var

Uses Cerebras Qwen-3-Coder-480B model to generate safe, POSIX-compliant shell commands.

## LLM Integration

All LLM calls use the Cerebras API via OpenAI SDK (`internal/llm/client.go`):
- **Base URL**: https://api.cerebras.ai/v1
- **Model**: qwen-3-coder-480b (defined as `CerebrasProModel`)
- **Authentication**: Via `CEREBRAS_API_KEY` environment variable

To add new LLM-powered features, use `llm.NewCerebrasClient()` and follow the pattern in `internal/do/translate.go` or `internal/gca/process.go`.

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

Based on CRUSH.md conventions:

- Use CamelCase for exported names, camelCase for unexported
- Always check errors explicitly
- Use `fmt.Errorf()` with `%w` verb for error wrapping
- Group imports: stdlib, third-party, internal (alphabetically)
- Use `go fmt` for consistent formatting
- Comment all exported functions, variables, and types
- Return errors early rather than deep nesting
