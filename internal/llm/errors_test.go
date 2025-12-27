package llm

import (
	"errors"
	"testing"
)

func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantMsg string
	}{
		{
			name:    "provider not available",
			err:     ErrProviderNotAvailable("claude", errors.New("not found")),
			wantMsg: "provider 'claude' not available: not found",
		},
		{
			name:    "authentication failed",
			err:     ErrAuthenticationFailed("gemini", errors.New("invalid key")),
			wantMsg: "authentication failed for provider 'gemini': invalid key",
		},
		{
			name:    "rate limit exceeded",
			err:     ErrRateLimitExceeded("gemini", nil),
			wantMsg: "rate limit exceeded for provider 'gemini'",
		},
		{
			name:    "model not found",
			err:     ErrModelNotFound("invalid-model", "gemini", nil),
			wantMsg: "model 'invalid-model' not found for provider 'gemini'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.wantMsg {
				t.Errorf("got %q, want %q", tt.err.Error(), tt.wantMsg)
			}
		})
	}
}

func TestErrorUnwrapping(t *testing.T) {
	underlying := errors.New("network error")
	err := ErrProviderNotAvailable("test", underlying)

	if !errors.Is(err, underlying) {
		t.Error("error should unwrap to underlying error")
	}
}
