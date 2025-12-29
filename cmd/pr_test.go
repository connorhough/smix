package cmd

import (
	"testing"
)

func TestPRCommandStructure(t *testing.T) {
	root := NewRootCmd()

	// Check if 'pr' command exists
	prCmd, _, err := root.Find([]string{"pr"})
	if err != nil {
		t.Fatalf("could not find 'pr' command: %v", err)
	}
	if prCmd.Use != "pr" {
		t.Errorf("expected command 'pr', got %q", prCmd.Use)
	}

	// Check if 'pr review' subcommand exists
	reviewCmd, _, err := root.Find([]string{"pr", "review"})
	if err != nil {
		t.Fatalf("could not find 'pr review' subcommand: %v", err)
	}
	if !reviewCmd.HasAvailableFlags() {
		t.Error("expected 'pr review' to have available flags")
	}

	flag := reviewCmd.Flags().Lookup("dir")
	if flag == nil {
		t.Error("expected 'pr review' to have 'dir' flag")
	}
}
