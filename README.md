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

### Test the example command
```bash
smix example --message "Test" --count 3
```

### Test the gca review command
```bash
smix gca review owner/repo pr_number
```

This command will:
1. Fetch code review feedback from the gemini-code-assist bot on the specified GitHub PR
2. Process each feedback item with an LLM to generate code patches
3. Launch Charm Crush sessions for each generated patch

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

