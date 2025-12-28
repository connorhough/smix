// internal/llm/iostreams.go
package llm

import (
	"bytes"
	"io"
	"os"

	"golang.org/x/term"
)

// IOStreams abstracts standard I/O for testability and dependency injection.
// Inspired by cli/cli's iostreams package.
//
// This pattern allows:
// - Unit tests to inject mock streams instead of using os.Stdin/Stdout/Stderr
// - TTY detection to be mocked for testing interactive vs non-interactive scenarios
// - Commands to be composable without tight coupling to global state
type IOStreams struct {
	In     io.Reader
	Out    io.Writer
	ErrOut io.Writer

	// isTerminalFunc allows lazy evaluation and mocking of TTY detection
	isTerminalFunc func(fd int) bool
	stdinFd        int
}

// NewIOStreams creates IOStreams connected to os.Stdin/Stdout/Stderr.
// Use this in production code for real terminal I/O.
func NewIOStreams() *IOStreams {
	return &IOStreams{
		In:             os.Stdin,
		Out:            os.Stdout,
		ErrOut:         os.Stderr,
		isTerminalFunc: term.IsTerminal,
		stdinFd:        int(os.Stdin.Fd()),
	}
}

// IsInteractive returns true if stdin is a TTY (terminal).
// This is the proper place to check for interactive capability,
// not inside provider implementations.
func (s *IOStreams) IsInteractive() bool {
	if s.isTerminalFunc == nil {
		return false
	}
	return s.isTerminalFunc(s.stdinFd)
}

// TestIOStreams creates IOStreams for testing with in-memory buffers.
// Returns the streams and the input/output buffers for assertions.
// Simulates a TTY by default (isTerminalFunc returns true).
func TestIOStreams() (*IOStreams, *bytes.Buffer, *bytes.Buffer) {
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	return &IOStreams{
		In:             in,
		Out:            out,
		ErrOut:         out,
		isTerminalFunc: func(int) bool { return true }, // Simulate TTY for testing
		stdinFd:        0,
	}, in, out
}

// TestIOStreamsNonInteractive creates IOStreams for testing non-interactive scenarios.
// Similar to TestIOStreams but simulates a non-TTY environment (pipes, CI/CD).
func TestIOStreamsNonInteractive() (*IOStreams, *bytes.Buffer, *bytes.Buffer) {
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	return &IOStreams{
		In:             in,
		Out:            out,
		ErrOut:         out,
		isTerminalFunc: func(int) bool { return false }, // Simulate non-TTY
		stdinFd:        0,
	}, in, out
}
