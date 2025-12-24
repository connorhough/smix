package gemini

// Model name constants for Gemini API
// Gemini requires full model names
const (
	ModelFlash     = "gemini-1.5-flash"
	ModelFlash8B   = "gemini-1.5-flash-8b"
	ModelPro       = "gemini-1.5-pro"
	ModelProLatest = "gemini-1.5-pro-latest"
)

// DefaultModel returns the default Gemini model
func DefaultModel() string {
	return ModelFlash
}
