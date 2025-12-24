package gemini

import (
	"context"
	"os"
	"testing"

	"google.golang.org/genai"
)

// Gemini SDK Prototype Test Findings:
//
// 1. Client Creation:
//    - Created via genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
//    - Does NOT have a Close() method (unlike some SDK patterns)
//    - Client is thread-safe and can be reused
//
// 2. Content Generation:
//    - Called via client.Models.GenerateContent(ctx, modelName, contents, config)
//    - Model name format: "gemini-2.0-flash" (simple string, no prefix)
//    - Contents created via genai.Text("prompt") convenience function
//    - Config can be nil for defaults
//
// 3. Error Handling:
//    - Invalid API Key:
//      * Type: genai.APIError
//      * Status: INVALID_ARGUMENT
//      * HTTP Code: 400
//      * Message: "API key not valid. Please pass a valid API key."
//
//    - Invalid Model Name:
//      * Type: genai.APIError
//      * Status: NOT_FOUND
//      * HTTP Code: 404
//      * Message: "models/MODEL_NAME is not found for API version v1beta..."
//
//    - Rate Limiting:
//      * Type: genai.APIError
//      * Status: RESOURCE_EXHAUSTED
//      * HTTP Code: 429
//      * Includes detailed quota violation information
//      * Provides retry delay (e.g., "Please retry in 30s")
//
//    - Context Cancellation:
//      * Type: *fmt.wrapError (wraps context.Canceled)
//      * Error message: "doRequest: error sending request: ... context canceled"
//      * Honors context cancellation properly
//
// 4. Response Structure:
//    - resp.Candidates[0].Content.Parts[0].Text contains the generated text
//    - resp.UsageMetadata provides token counts (prompt, candidates, total)
//    - resp.Candidates[0].FinishReason indicates completion reason
//    - resp.Candidates[0].SafetyRatings contains safety assessment
//    - resp.PromptFeedback contains prompt-level safety information

func TestPrototypeGeminiSDK(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	model := "gemini-2.0-flash"
	prompt := genai.Text("Say 'hello' and nothing else")

	resp, err := client.Models.GenerateContent(ctx, model, prompt, nil)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(resp.Candidates) == 0 {
		t.Fatal("Expected at least one candidate")
	}

	t.Logf("Gemini SDK works! Response: %v", resp.Candidates[0].Content)
}

// TestInvalidAPIKey verifies error handling for invalid API keys
func TestInvalidAPIKey(t *testing.T) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: "invalid-api-key-12345",
	})
	if err != nil {
		t.Logf("Client creation failed with invalid API key: %v", err)
		return
	}

	model := "gemini-2.0-flash"
	prompt := genai.Text("Say hello")

	_, err = client.Models.GenerateContent(ctx, model, prompt, nil)
	if err != nil {
		t.Logf("Invalid API key error format: %v", err)
		t.Logf("Error type: %T", err)
	} else {
		t.Log("WARNING: Invalid API key did not produce error")
	}
}

// TestInvalidModelName verifies error handling for invalid model names
func TestInvalidModelName(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	model := "invalid-model-name-xyz"
	prompt := genai.Text("Say hello")

	_, err = client.Models.GenerateContent(ctx, model, prompt, nil)
	if err != nil {
		t.Logf("Invalid model name error format: %v", err)
		t.Logf("Error type: %T", err)
	} else {
		t.Log("WARNING: Invalid model name did not produce error")
	}
}

// TestResponseStructure examines the structure of successful responses
func TestResponseStructure(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	model := "gemini-2.0-flash"
	prompt := genai.Text("Say 'hello' and nothing else")

	resp, err := client.Models.GenerateContent(ctx, model, prompt, nil)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	t.Logf("Number of candidates: %d", len(resp.Candidates))
	if len(resp.Candidates) > 0 {
		candidate := resp.Candidates[0]
		t.Logf("Candidate finish reason: %v", candidate.FinishReason)
		t.Logf("Candidate safety ratings count: %d", len(candidate.SafetyRatings))

		if candidate.Content != nil {
			t.Logf("Content role: %s", candidate.Content.Role)
			t.Logf("Content parts count: %d", len(candidate.Content.Parts))

			if len(candidate.Content.Parts) > 0 {
				t.Logf("First part type: %T", candidate.Content.Parts[0])
				part := candidate.Content.Parts[0]
				if part.Text != "" {
					t.Logf("Text content: %q", part.Text)
				}
			}
		}
	}

	if resp.PromptFeedback != nil {
		t.Logf("Prompt feedback block reason: %v", resp.PromptFeedback.BlockReason)
		t.Logf("Prompt feedback safety ratings count: %d", len(resp.PromptFeedback.SafetyRatings))
	}

	if resp.UsageMetadata != nil {
		t.Logf("Usage - Prompt tokens: %d", resp.UsageMetadata.PromptTokenCount)
		t.Logf("Usage - Candidates tokens: %d", resp.UsageMetadata.CandidatesTokenCount)
		t.Logf("Usage - Total tokens: %d", resp.UsageMetadata.TotalTokenCount)
	}
}

// TestContextCancellation verifies if the SDK respects context cancellation
func TestContextCancellation(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Create a context that's already cancelled
	cancelledCtx, cancel := context.WithCancel(ctx)
	cancel()

	model := "gemini-2.0-flash"
	prompt := genai.Text("This should not complete")

	_, err = client.Models.GenerateContent(cancelledCtx, model, prompt, nil)
	if err != nil {
		t.Logf("Cancelled context error: %v", err)
		t.Logf("Error type: %T", err)
		if err == context.Canceled {
			t.Log("Context cancellation works correctly!")
		}
	} else {
		t.Log("WARNING: Cancelled context did not produce error")
	}
}
