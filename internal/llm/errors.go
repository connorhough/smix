package llm

import "fmt"

// ProviderError represents a provider-specific error
type ProviderError struct {
	Provider string
	Msg      string
	Err      error
}

func (e *ProviderError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Msg, e.Err)
	}
	return e.Msg
}

func (e *ProviderError) Unwrap() error {
	return e.Err
}

// ErrProviderNotAvailable indicates the provider is not available (CLI not found, SDK init failed)
func ErrProviderNotAvailable(provider string, err error) error {
	return &ProviderError{
		Provider: provider,
		Msg:      fmt.Sprintf("provider '%s' not available", provider),
		Err:      err,
	}
}

// ErrAuthenticationFailed indicates authentication failure (invalid API key, etc.)
func ErrAuthenticationFailed(provider string, err error) error {
	return &ProviderError{
		Provider: provider,
		Msg:      fmt.Sprintf("authentication failed for provider '%s'", provider),
		Err:      err,
	}
}

// ErrRateLimitExceeded indicates the provider's rate limit was hit
func ErrRateLimitExceeded(provider string) error {
	return &ProviderError{
		Provider: provider,
		Msg:      fmt.Sprintf("rate limit exceeded for provider '%s'", provider),
	}
}

// ErrModelNotFound indicates the specified model doesn't exist for the provider
func ErrModelNotFound(model, provider string) error {
	return &ProviderError{
		Provider: provider,
		Msg:      fmt.Sprintf("model '%s' not found for provider '%s'", model, provider),
	}
}
