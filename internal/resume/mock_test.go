package resume

import (
	"context"
	"fmt"
)

type MockPlatform struct {
	WindowTitle  string
	Typed        []string
	EnterPressed bool
	FailTitle    bool
	FailType     bool
	FailEnter    bool
}

func (m *MockPlatform) GetActiveWindowTitle(ctx context.Context) (string, error) {
	if m.FailTitle {
		return "", fmt.Errorf("mock error getting title")
	}
	return m.WindowTitle, nil
}

func (m *MockPlatform) TypeString(ctx context.Context, msg string) error {
	if m.FailType {
		return fmt.Errorf("mock error typing")
	}
	m.Typed = append(m.Typed, msg)
	return nil
}

func (m *MockPlatform) PressEnter(ctx context.Context) error {
	if m.FailEnter {
		return fmt.Errorf("mock error pressing enter")
	}
	m.EnterPressed = true
	return nil
}

// Ensure MockPlatform satisfies the Platform interface
var _ Platform = &MockPlatform{}
