package gemini

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
