package gemini

// APIKeyEnvVar is the environment variable used for the Gemini API key
const APIKeyEnvVar = "SMIX_GEMINI_API_KEY"

// Model name constants for Gemini API
// Gemini requires full model names
const (
	ModelFlash = "gemini-3-flash-preview"
	ModelPro   = "gemini-3-pro-preview"
)

// DefaultModel returns the default Gemini model
func DefaultModel() string {
	return ModelFlash
}
