package llm

import (
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
)

const (
	CerebrasAPIBaseURL = "https://api.cerebras.ai/v1"
	CerebrasProModel   = "qwen-3-coder-480b"
)

// NewCerebrasClient creates a new authenticated client for the Cerebras API
func NewCerebrasClient() (*openai.Client, error) {
	apiKey := os.Getenv("CEREBRAS_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("CEREBRAS_API_KEY environment variable not set")
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = CerebrasAPIBaseURL
	client := openai.NewClientWithConfig(config)

	return client, nil
}

