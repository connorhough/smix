# smix - Go CLI Toolkit

## Project Overview

`smix` is a Go-based CLI toolkit designed to enhance terminal workflows with AI capabilities. It follows a standard Go project structure, separating CLI command definitions from core business logic.

**Key Features:**
*   **`pr` (Pull Request):** Fetches code review feedback from `gemini-code-assist` bots on GitHub Pull Requests, processing them into actionable prompts.
*   **`do`:** Translates natural language descriptions into shell commands.
*   **`ask`:** Provides an interface for asking questions directly from the terminal.
*   **Configuration:** Uses `viper` for flexible configuration management, respecting XDG standards.

**Key Technologies:**
*   **Language:** Go (Golang)
*   **CLI Framework:** [Cobra](https://github.com/spf13/cobra)
*   **Configuration:** [Viper](https://github.com/spf13/viper)
*   **GitHub Integration:** [go-github](https://github.com/google/go-github)

## Architecture

The project adheres to the standard Go project layout:

*   **`cmd/`:** Contains the main entry point and CLI command definitions. Each command (e.g., `pr`, `do`, `ask`) has its own file here, responsible for flag parsing and calling into the `internal` package.
*   **`internal/`:** Contains the core business logic, isolated from the CLI interface.
    *   `internal/config/`: Configuration handling.
    *   `internal/gca/`: Logic for fetching and processing GitHub reviews.
    *   `internal/do/`: Logic for the command translation feature.
    *   `internal/version/`: Version information variables.
*   **`main.go`:** The application entry point, which initializes the root command and handles global error reporting.
*   **`Makefile`:** Automates build, test, and release tasks.

## Building and Running

The project uses `make` for build automation.

**Prerequisites:**
*   Go 1.25.1+

**Commands:**

*   **Build:**
    ```bash
    make build
    ```
    Produces the binary in the `builds/` directory (e.g., `builds/smix`).

*   **Install:**
    ```bash
    make install
    ```
    Installs the binary to your `$GOPATH/bin`.

*   **Test:**
    ```bash
    make test
    ```
    Runs all unit tests.

*   **Cross-Compilation:**
    ```bash
    make build-darwin-arm64  # macOS Apple Silicon
    make build-linux-amd64   # Linux x86_64
    ```

*   **Clean:**
    ```bash
    make clean
    ```

## Development Conventions

*   **Code Style:** Follows standard Go formatting (`gofmt`).
*   **Command Structure:**
    *   **Logic Separation:** Business logic **must** reside in `internal/`. `cmd/` is strictly for argument parsing and CLI wiring.
    *   **New Commands:** To add a new command, create a corresponding package in `internal/` for logic and a file in `cmd/` for the Cobra definition.
*   **Error Handling:** Commands should return errors to `main.go` rather than calling `os.Exit()` directly. `main.go` handles the final error output and exit codes.
*   **Configuration:**
    *   Supports config files in `$XDG_CONFIG_HOME/smix/`, `~/.config/smix/`, or `~/.smix.yaml`.
    *   Environment variables are supported with the `SMIX_` prefix (handled by Viper).
*   **Versioning:** Version strings (`Version`, `GitCommit`, `BuildDate`) are injected at build time via `ldflags`. Do not manually edit these in the source for releases.
