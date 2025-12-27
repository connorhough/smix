# smix - Go CLI Toolkit

## Build Guide

### Initialize the module
```bash
go mod init github.com/connorhough/smix
```

### Install dependencies
```bash
go mod tidy
```

### Build the application
```bash
make build
```

### Install the application
```bash
make install
```

### Cross-compilation

For macOS (Apple Silicon):
```bash
make build-darwin-arm64
```

For Linux (x86_64):
```bash
make build-linux-amd64
```

### Test the pr review command
```bash
smix pr review owner/repo pr_number
```

This command will:
1. Fetch code review feedback from the gemini-code-assist bot on the specified GitHub PR
2. Process each feedback item with an LLM to generate code patches
3. Launch Claude Code sessions for each generated patch

### Test the do command
```bash
smix do "your natural language task"
```

This command will translate your natural language request into a shell command.

Example:
```bash
smix do "list all files in the current directory"
```

### Test the ask command
```bash
smix ask "your technical question"
```

This command answers short technical questions using your configured LLM provider.

Example:
```bash
smix ask "what is FastAPI"
```

## Configuration

smix supports multiple LLM providers. Configuration is stored in `~/.config/smix/config.yaml` (or `$XDG_CONFIG_HOME/smix/config.yaml`).

On first run, a template configuration file is automatically created with sensible defaults.

### Provider Setup

#### Claude (Default)
- **Requires:** Claude Code CLI installed and authenticated
- **Install:** Visit https://claude.ai/code
- **Models:** `haiku`, `sonnet`, `opus`

#### Gemini
- **Requires:** Google AI Studio API key
- **Setup:** Set `SMIX_GEMINI_API_KEY` environment variable
- **Get API Key:** https://aistudio.google.com/apikey
- **Models:** `gemini-3-flash-preview`, `gemini-3-pro-preview`

### Configuration Examples

**Global default (all commands use Claude):**
```yaml
provider: claude
model: sonnet
```

**Per-command customization (fast/cheap for ask/do, smart for pr):**
```yaml
provider: claude  # global default

commands:
  ask:
    provider: gemini
    model: gemini-3-flash-preview
  do:
    provider: gemini
    model: gemini-3-flash-preview
  pr:
    provider: claude
    model: sonnet
```

**Override with flags:**
```bash
smix ask --provider gemini --model gemini-3-pro-preview "what is FastAPI"
smix do --provider claude --model haiku "list all files"
```

### Configuration Precedence

1. CLI flags (`--provider`, `--model`)
2. Command-specific config (`commands.ask.provider`)
3. Global config (`provider`)

### Debug Mode

Use `--debug` flag to see provider selection and configuration resolution:

```bash
smix ask --debug "what is FastAPI"
```

## Tagging Releases

To create a new version tag for releases:

```bash
git tag -a v0.1.0 -m "Release version 0.1.0"
```

To push tags to remote repository:
```bash
git push origin --tags
```

To create a tag with a semantic version:
```bash
git tag -a v1.0.0 -m "Release version 1.0.0"
```

The versioning system will automatically use the most recent tag when building. If no tags exist, it will fall back to using the commit hash.

## Design Rationale

### Separation of cmd and internal packages
The `cmd` package contains only CLI wiring code (commands, flags), while the `internal` package contains all business logic. This separation:
- Makes business logic easier to test in isolation
- Allows for better code organization and reuse
- Follows the principle of separation of concerns

### Version injection pattern
The version information is injected at build time using ldflags, which allows:
- Tracking of exact builds in production
- No need to hardcode version strings
- Consistent versioning across platforms

### Centralized error handling
Error handling is centralized in main.go by having commands return errors rather than calling os.Exit() directly:
- Ensures consistent error handling across all commands
- Allows main.go to control exit codes
- Prevents commands from terminating the program prematurely

### XDG configuration paths
The configuration search paths follow XDG specifications and cross-platform conventions:
- Respects user's configuration directory preferences
- Provides fallback locations for broader compatibility
- Follows established conventions for CLI tools

### Makefile benefits
The Makefile provides consistent builds across different environments:
- Standardized build process with version injection
- Cross-compilation targets for multiple platforms
- Clear separation of build, install, clean, and test operations

## Extension Guide

To add new commands following the established pattern:

1. Create `internal/newcommand/newcommand.go` with business logic:
   ```go
   package newcommand

   func Run(param string) error {
       // Your business logic here
       return nil
   }
   ```

2. Create `cmd/newcommand.go` with Cobra command definition:
   ```go
   package cmd

   import (
       "github.com/connorhough/smix/internal/newcommand"
       "github.com/spf13/cobra"
   )

   func newNewCommandCmd() *cobra.Command {
       var param string

       newCommandCmd := &cobra.Command{
           Use:   "newcommand",
           Short: "Description of newcommand",
           RunE: func(cmd *cobra.Command, args []string) error {
               return newcommand.Run(param)
           },
       }

       newCommandCmd.Flags().StringVar(&param, "param", "", "Parameter description")
       return newCommandCmd
   }

   func init() {
       rootCmd.AddCommand(newNewCommandCmd())
   }
   ```

