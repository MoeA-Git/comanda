package processor

// StepConfig represents the configuration for a single step
type StepConfig struct {
	Input      interface{} `yaml:"input"`       // Can be string or map[string]interface{}
	Model      interface{} `yaml:"model"`       // Can be string or []string
	Action     interface{} `yaml:"action"`      // Can be string or []string
	Output     interface{} `yaml:"output"`      // Can be string or []string
	NextAction interface{} `yaml:"next-action"` // Can be string or []string
}

// Step represents a named step in the DSL
type Step struct {
	Name   string
	Config StepConfig
}

// DSLConfig represents the structure of the DSL configuration
type DSLConfig struct {
	Steps []Step
}

// NormalizeOptions represents options for string slice normalization
type NormalizeOptions struct {
	AllowEmpty bool // Whether to allow empty strings in the result
}
