package llm

// Option configures provider behavior
type Option func(*GenerateOptions)

// GenerateOptions holds configuration for Generate calls
type GenerateOptions struct {
	Model string
}

// WithModel overrides the model for this generation
func WithModel(model string) Option {
	return func(opts *GenerateOptions) {
		opts.Model = model
	}
}

// buildOptions constructs GenerateOptions from Option functions
func buildOptions(opts []Option) *GenerateOptions {
	options := &GenerateOptions{}
	for _, opt := range opts {
		opt(options)
	}
	return options
}
