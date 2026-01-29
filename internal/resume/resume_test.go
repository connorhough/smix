package resume

import (
	"context"
	"testing"
	"time"
)

type MockClock struct {
	CurrentTime time.Time
	SleepCalls  []time.Duration
}

func (m *MockClock) Now() time.Time {
	return m.CurrentTime
}

func (m *MockClock) Sleep(ctx context.Context, d time.Duration) error {
	m.SleepCalls = append(m.SleepCalls, d)
	return nil
}

func TestRun_Success(t *testing.T) {
	// Setup Mock Platform
	mockPlat := &MockPlatform{
		WindowTitle: "Terminal",
	}
	CurrentPlatform = mockPlat
	defer func() { CurrentPlatform = nil }()

	// Setup Mock Clock
	start := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	mockClock := &MockClock{CurrentTime: start}
	originalClock := SystemClock
	SystemClock = mockClock
	defer func() { SystemClock = originalClock }()

	// Target: 5 seconds later
	ctx := context.Background()
	targetTime := start.Add(5 * time.Second)
	msg := "continue"

	err := Run(ctx, targetTime, msg, "")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Verification
	if len(mockPlat.Typed) != 1 {
		t.Errorf("Expected 1 typed message, got %d", len(mockPlat.Typed))
	} else if mockPlat.Typed[0] != msg {
		t.Errorf("Expected message %q, got %q", msg, mockPlat.Typed[0])
	}

	if !mockPlat.EnterPressed {
		t.Error("Expected Enter to be pressed")
	}

	// Check Sleep calls
	if len(mockClock.SleepCalls) != 1 {
		t.Errorf("Expected 1 sleep call, got %d", len(mockClock.SleepCalls))
	} else {
		if mockClock.SleepCalls[0] != 5*time.Second {
			t.Errorf("Expected sleep 5s, got %v", mockClock.SleepCalls[0])
		}
	}
}

func TestRun_WaitLong(t *testing.T) {
	// Setup Mock Platform
	mockPlat := &MockPlatform{
		WindowTitle: "Terminal",
	}
	CurrentPlatform = mockPlat
	defer func() { CurrentPlatform = nil }()

	// Setup Mock Clock
	start := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	mockClock := &MockClock{CurrentTime: start}
	originalClock := SystemClock
	SystemClock = mockClock
	defer func() { SystemClock = originalClock }()

	// Target: 20 seconds later (logic splits if > 10s)
	targetTime := start.Add(20 * time.Second)
	msg := "continue"

	ctx := context.Background()
	err := Run(ctx, targetTime, msg, "")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Verification
	// Should have 2 sleep calls: Wait-10s (10s) and 10s.
	if len(mockClock.SleepCalls) != 2 {
		t.Errorf("Expected 2 sleep calls, got %d", len(mockClock.SleepCalls))
	} else {
		if mockClock.SleepCalls[0] != 10*time.Second {
			t.Errorf("Expected first sleep 10s (wait-10), got %v", mockClock.SleepCalls[0])
		}
		if mockClock.SleepCalls[1] != 10*time.Second {
			t.Errorf("Expected second sleep 10s, got %v", mockClock.SleepCalls[1])
		}
	}
}

func TestRun_Tomorrow(t *testing.T) {
	// Setup Mock Platform
	mockPlat := &MockPlatform{
		WindowTitle: "Terminal",
	}
	CurrentPlatform = mockPlat
	defer func() { CurrentPlatform = nil }()

	// Setup Mock Clock
	// Current: 12:00
	start := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	mockClock := &MockClock{CurrentTime: start}
	originalClock := SystemClock
	SystemClock = mockClock
	defer func() { SystemClock = originalClock }()

	// Target: 11:00 (Past) -> Should be treated as tomorrow 11:00
	targetTime := time.Date(2023, 1, 1, 11, 0, 0, 0, time.UTC)
	msg := "continue"

	ctx := context.Background()
	err := Run(ctx, targetTime, msg, "")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Verify duration.
	// 12:00 today to 11:00 tomorrow = 23 hours.
	expectedDuration := 23 * time.Hour

	// Since 23h > 10s, it splits.
	// Sleep 1: 23h - 10s
	// Sleep 2: 10s

	if len(mockClock.SleepCalls) != 2 {
		t.Errorf("Expected 2 sleep calls, got %d", len(mockClock.SleepCalls))
	} else {
		if mockClock.SleepCalls[0] != expectedDuration-10*time.Second {
			t.Errorf("Expected first sleep %v, got %v", expectedDuration-10*time.Second, mockClock.SleepCalls[0])
		}
	}
}

func TestRun_WindowFilter(t *testing.T) {
	mockPlat := &MockPlatform{
		WindowTitle: "Correct Window",
	}
	CurrentPlatform = mockPlat
	defer func() { CurrentPlatform = nil }()

	mockClock := &MockClock{CurrentTime: time.Now()}
	originalClock := SystemClock
	SystemClock = mockClock
	defer func() { SystemClock = originalClock }()

	// Case 1: Match
	err := Run(context.Background(), mockClock.CurrentTime.Add(1*time.Second), "msg", "Correct")
	if err != nil {
		t.Errorf("Expected success with matching filter, got %v", err)
	}

	// Case 2: Mismatch
	err = Run(context.Background(), mockClock.CurrentTime.Add(1*time.Second), "msg", "Wrong")
	if err == nil {
		t.Error("Expected error with mismatching filter, got nil")
	} else if err.Error() != "active window 'Correct Window' does not match filter 'Wrong'" {
		t.Errorf("Unexpected error message: %v", err)
	}
}
