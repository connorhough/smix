package claude

// Model name constants for Claude CLI
// Claude CLI accepts short model names
const (
	ModelHaiku  = "haiku"
	ModelSonnet = "sonnet"
	ModelOpus   = "opus"
)

// DefaultModel returns the default Claude model
func DefaultModel() string {
	return ModelHaiku
}
