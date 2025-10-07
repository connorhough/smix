# Crush Configuration for smix Codebase

## Build Commands
- `make build` - Build the binary with version information
- `make install` - Install the binary with version information

## Test Commands
- `make test` - Run all tests with verbose output
- To run a single test: `go test -v ./path/to/package -run TestName`

## Lint Commands
- `make lint` - Run golangci-lint (requires golangci-lint to be installed)

## Code Style Guidelines

### Naming Conventions
- Use CamelCase for functions and variables
- Choose descriptive names that clearly indicate purpose
- Exported functions/variables start with uppercase letter

### Error Handling
- Always check errors explicitly
- Use `fmt.Errorf()` with `%w` verb to wrap errors when appropriate
- Return errors early rather than nesting deeply

### Imports
- Group imports in three sections separated by blank lines:
  1. Standard library packages
  2. Third-party packages
  3. Internal packages
- Sort imports alphabetically within each group

### Formatting
- Use `go fmt` for consistent code formatting
- Maintain proper indentation (tabs, not spaces)
- Keep line length reasonable (< 120 characters)

### Types
- Define types close to where they're used when possible
- Use appropriate type names that reflect their purpose
- Prefer explicit types over implicit ones

### Comments
- Comment all exported functions, variables, and types
- Use clear and concise comments
- Follow Go comment conventions (no space between // and comment text)

## Platform Specific Info
- Cross-compilation targets available in Makefile for different OS/architectures