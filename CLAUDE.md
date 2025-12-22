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
    - `process.go`: Generates patches via LLM and launches Claude Code sessions
  - `do/`: Natural language to shell command translation
  - `ask/`: Answers short technical questions using Claude Code CLI
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

Fetches code review feedback from gemini-code-assist bot on GitHub PRs and launches interactive Claude Code sessions to analyze and implement the suggested changes.

```bash
smix gca review owner/repo pr_number
smix gca review --dir gca_review_pr123  # Process existing feedback directory
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
```

**Requirements:**
- `claude` CLI installed (Claude Code)

Uses Claude Code CLI in non-interactive mode to generate safe, POSIX-compliant shell commands.

### ask

Answers short technical questions with concise, accurate responses.

```bash
smix ask "what is the difference between TCP and UDP"
smix ask "how do I list all running processes on Linux"
```

**Requirements:**
- `claude` CLI installed (Claude Code)

Great for quick lookups and technical questions where you need a brief, informative answer without searching documentation or web resources.

## LLM Integration

All LLM-powered commands use the Claude Code CLI in non-interactive mode:
- **Command**: `claude -p "prompt"`
- **Authentication**: Uses your Claude Code session (requires `claude` CLI installed)
- **Model**: Uses your configured Claude model (Sonnet by default)

To add new LLM-powered features, follow the pattern in `internal/do/translate.go`:
```go
cmd := exec.Command("claude", "-p", yourPrompt)
output, err := cmd.CombinedOutput()
```

For more complex integrations requiring structured output, use `--output-format json` with optional `--json-schema`.

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
- Do not add any emojis or the "genereated with/co-authored by claude" content to commits. It adds too much bloat to the git history

