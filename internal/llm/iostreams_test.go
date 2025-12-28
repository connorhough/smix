// internal/llm/iostreams_test.go
package llm

import (
	"bytes"
	"testing"
)

func TestIOStreams_IsInteractive_WithTerminal(t *testing.T) {
	streams := &IOStreams{
		In:             &bytes.Buffer{},
		Out:            &bytes.Buffer{},
		ErrOut:         &bytes.Buffer{},
		isTerminalFunc: func(int) bool { return true },
		stdinFd:        0,
	}

	if !streams.IsInteractive() {
		t.Error("expected IsInteractive to return true when isTerminalFunc returns true")
	}
}

func TestIOStreams_IsInteractive_WithoutTerminal(t *testing.T) {
	streams := &IOStreams{
		In:             &bytes.Buffer{},
		Out:            &bytes.Buffer{},
		ErrOut:         &bytes.Buffer{},
		isTerminalFunc: func(int) bool { return false },
		stdinFd:        0,
	}

	if streams.IsInteractive() {
		t.Error("expected IsInteractive to return false when isTerminalFunc returns false")
	}
}

func TestIOStreams_IsInteractive_NilFunc(t *testing.T) {
	streams := &IOStreams{
		In:             &bytes.Buffer{},
		Out:            &bytes.Buffer{},
		ErrOut:         &bytes.Buffer{},
		isTerminalFunc: nil,
		stdinFd:        0,
	}

	if streams.IsInteractive() {
		t.Error("expected IsInteractive to return false when isTerminalFunc is nil")
	}
}

func TestTestIOStreams_CreatesBuffers(t *testing.T) {
	streams, in, out := TestIOStreams()

	if streams.In != in {
		t.Error("expected In to be the input buffer")
	}

	if streams.Out != out {
		t.Error("expected Out to be the output buffer")
	}

	if streams.ErrOut != out {
		t.Error("expected ErrOut to be the output buffer")
	}

	// Should simulate TTY by default for testing
	if !streams.IsInteractive() {
		t.Error("expected TestIOStreams to simulate interactive terminal")
	}
}

func TestNewIOStreams_UsesOsStreams(t *testing.T) {
	// Can't test actual os.Stdin/Stdout in unit tests, but verify constructor works
	streams := NewIOStreams()

	if streams == nil {
		t.Fatal("expected NewIOStreams to return non-nil")
	}

	if streams.In == nil {
		t.Error("expected In to be set")
	}

	if streams.Out == nil {
		t.Error("expected Out to be set")
	}

	if streams.ErrOut == nil {
		t.Error("expected ErrOut to be set")
	}

	if streams.isTerminalFunc == nil {
		t.Error("expected isTerminalFunc to be set")
	}
}
