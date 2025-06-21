package processor

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/kris-hansen/comanda/utils/config"
	"github.com/kris-hansen/comanda/utils/models"
	"gopkg.in/yaml.v3"
)

func createTestServerConfig() *config.ServerConfig {
	return &config.ServerConfig{
		Enabled: true,
	}
}

func TestNormalizeStringSlice(t *testing.T) {
	processor := NewProcessor(&DSLConfig{}, createTestEnvConfig(), createTestServerConfig(), false, "")

	tests := []struct {
		name     string
		input    interface{}
		expected []string
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: []string{},
		},
		{
			name:     "single string",
			input:    "test",
			expected: []string{"test"},
		},
		{
			name:     "string slice",
			input:    []string{"test1", "test2"},
			expected: []string{"test1", "test2"},
		},
		{
			name:     "interface slice",
			input:    []interface{}{"test1", "test2"},
			expected: []string{"test1", "test2"},
		},
		{
			name:     "empty interface slice",
			input:    []interface{}{},
			expected: []string{},
		},
		{
			name:     "mixed type interface slice - only strings extracted",
			input:    []interface{}{"test1", 42, "test2"},
			expected: []string{"test1", "", "test2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.NormalizeStringSlice(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("NormalizeStringSlice() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestNewProcessor(t *testing.T) {
	config := &DSLConfig{}
	envConfig := createTestEnvConfig()
	verbose := true

	processor := NewProcessor(config, envConfig, createTestServerConfig(), verbose, "")

	if processor == nil {
		t.Error("NewProcessor() returned nil")
	}

	if processor.config != config {
		t.Error("NewProcessor() did not set config correctly")
	}

	if processor.envConfig != envConfig {
		t.Error("NewProcessor() did not set envConfig correctly")
	}

	if processor.verbose != verbose {
		t.Error("NewProcessor() did not set verbose correctly")
	}

	if processor.handler == nil {
		t.Error("NewProcessor() did not initialize handler")
	}

	if processor.validator == nil {
		t.Error("NewProcessor() did not initialize validator")
	}

	if processor.providers == nil {
		t.Error("NewProcessor() did not initialize providers map")
	}
}

func TestValidateStepConfig(t *testing.T) {
	processor := NewProcessor(&DSLConfig{}, createTestEnvConfig(), createTestServerConfig(), false, "")

	tests := []struct {
		name          string
		stepName      string
		config        StepConfig
		expectedError string
	}{
		{
			name:     "valid config",
			stepName: "test_step",
			config: StepConfig{
				Input:  "test.txt",
				Model:  "gpt-4o-mini",
				Action: "analyze",
				Output: "STDOUT",
			},
			expectedError: "",
		},
		{
			name:     "missing input tag",
			stepName: "test_step",
			config: StepConfig{
				Model:  "gpt-4o-mini",
				Action: "analyze",
				Output: "STDOUT",
			},
			expectedError: "input tag is required",
		},
		{
			name:     "missing model",
			stepName: "test_step",
			config: StepConfig{
				Input:  "test.txt",
				Action: "analyze",
				Output: "STDOUT",
			},
			expectedError: "model is required",
		},
		{
			name:     "missing action",
			stepName: "test_step",
			config: StepConfig{
				Input:  "test.txt",
				Model:  "gpt-4o-mini",
				Output: "STDOUT",
			},
			expectedError: "action is required",
		},
		{
			name:     "missing output",
			stepName: "test_step",
			config: StepConfig{
				Input:  "test.txt",
				Model:  "gpt-4o-mini",
				Action: "analyze",
			},
			expectedError: "output is required",
		},
		{
			name:     "empty input allowed",
			stepName: "test_step",
			config: StepConfig{
				Input:  "",
				Model:  "gpt-4o-mini",
				Action: "analyze",
				Output: "STDOUT",
			},
			expectedError: "",
		},
		{
			name:     "NA input allowed",
			stepName: "test_step",
			config: StepConfig{
				Input:  "NA",
				Model:  "gpt-4o-mini",
				Action: "analyze",
				Output: "STDOUT",
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := processor.validateStepConfig(tt.stepName, tt.config)
			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("validateStepConfig() returned unexpected error: %v", err)
				}
			} else {
				if err == nil {
					t.Error("validateStepConfig() expected error but got none")
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("validateStepConfig() error = %v, want error containing %v", err, tt.expectedError)
				}
			}
		})
	}
}

func TestProcess(t *testing.T) {
	tests := []struct {
		name        string
		config      DSLConfig
		expectError bool
	}{
		{
			name:        "empty config",
			config:      DSLConfig{},
			expectError: true,
		},
		{
			name: "single step with missing model",
			config: DSLConfig{
				Steps: []Step{
					{
						Name: "step_one",
						Config: StepConfig{
							Action: []string{"test action"},
							Output: []string{"STDOUT"},
						},
					},
				},
			},
			expectError: true,
		},
		{
			name: "valid single step",
			config: DSLConfig{
				Steps: []Step{
					{
						Name: "step_one",
						Config: StepConfig{
							Input:  []string{"NA"},
							Model:  []string{"gpt-4o-mini"},
							Action: []string{"test action"},
							Output: []string{"STDOUT"},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "input file exists as output in later step",
			config: DSLConfig{
				Steps: []Step{
					{
						Name: "step_one",
						Config: StepConfig{
							Input:  []string{"future_output.txt"},
							Model:  []string{"gpt-4o-mini"},
							Action: []string{"test action"},
							Output: []string{"STDOUT"},
						},
					},
					{
						Name: "step_two",
						Config: StepConfig{
							Input:  []string{"NA"},
							Model:  []string{"gpt-4o-mini"},
							Action: []string{"generate"},
							Output: []string{"future_output.txt"},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "non-existent input file with no matching output",
			config: DSLConfig{
				Steps: []Step{
					{
						Name: "step_one",
						Config: StepConfig{
							Input:  []string{"nonexistent.txt"},
							Model:  []string{"gpt-4o-mini"},
							Action: []string{"test action"},
							Output: []string{"STDOUT"},
						},
					},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewProcessor(&tt.config, createTestEnvConfig(), createTestServerConfig(), false, "")
			err := processor.Process()

			if tt.expectError && err == nil {
				t.Error("Process() expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Process() unexpected error: %v", err)
			}
		})
	}
}

func TestDebugf(t *testing.T) {
	tests := []struct {
		name    string
		verbose bool
	}{
		{
			name:    "verbose mode enabled",
			verbose: true,
		},
		{
			name:    "verbose mode disabled",
			verbose: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewProcessor(&DSLConfig{}, createTestEnvConfig(), createTestServerConfig(), tt.verbose, "")
			// Note: This test only verifies that debugf doesn't panic
			// In a real scenario, you might want to capture stdout and verify the output
			processor.debugf("test message %s", "arg")
		})
	}
}

// CustomMockProvider extends MockProvider with custom response handling
type CustomMockProvider struct {
	MockProvider
	responses map[string]string
}

func (m *CustomMockProvider) SendPrompt(model, prompt string) (string, error) {
	// First check if we have a custom response for this prompt
	for key, response := range m.responses {
		if strings.Contains(prompt, key) {
			return response, nil
		}
	}

	// Fall back to the parent implementation
	return "mock response", nil
}

// Override SendPromptWithFile to use our custom responses
func (m *CustomMockProvider) SendPromptWithFile(model, prompt string, file models.FileInput) (string, error) {
	// First check if we have a custom response for this prompt
	for key, response := range m.responses {
		if strings.Contains(prompt, key) {
			return response, nil
		}
	}

	// Fall back to the parent implementation
	return "mock response", nil
}

func TestProcessWithDefer(t *testing.T) {
	// Create a temporary defer.yaml file for the test
	deferYAML := `
determine_poem_type:
  input: STDIN
  model: gpt-4o-mini
  action: "Analyze the input poem"
  output: STDOUT

defer:
  analyze_haiku:
    input: STDIN
    model: gpt-4o-mini
    action: "This is the haiku analysis."
    output: STDOUT
`
	// Load the DSL config from the string
	var dslConfig DSLConfig
	if err := yaml.Unmarshal([]byte(deferYAML), &dslConfig); err != nil {
		t.Fatalf("Failed to unmarshal yaml: %v", err)
	}

	// Create a custom mock provider for this test
	customMockProvider := &CustomMockProvider{
		MockProvider: *NewMockProvider("openai"),
		responses: map[string]string{
			"Analyze the input poem":      `{"step":"analyze_haiku","input":"a test haiku"}`,
			"This is the haiku analysis.": "Haiku analysis complete.",
		},
	}
	customMockProvider.Configure("test-key")

	// Store the original DetectProvider function
	originalDetect := models.DetectProvider

	// Override the DetectProvider function to return our custom mock provider
	models.DetectProvider = func(modelName string) models.Provider {
		return customMockProvider
	}

	// Restore the original function when the test is done
	defer func() { models.DetectProvider = originalDetect }()

	// Create a processor with the default environment config
	processor := NewProcessor(&dslConfig, createTestEnvConfig(), createTestServerConfig(), true, "")
	processor.SetLastOutput("An old silent pond...") // Initial STDIN

	// Process the workflow
	err := processor.Process()
	if err != nil {
		t.Fatalf("Process() failed: %v", err)
	}

	// Check the final output
	expectedOutput := "Haiku analysis complete."
	if processor.LastOutput() != expectedOutput {
		t.Errorf("Expected final output to be '%s', but got '%s'", expectedOutput, processor.LastOutput())
	}
}

func TestUnmarshalYAMLWithDuplicateDeferSteps(t *testing.T) {
	// Define a YAML with duplicate step names in the defer block
	deferYAML := `
determine_poem_type:
  input: STDIN
  model: gpt-4o-mini
  action: "Analyze the input poem"
  output: STDOUT

defer:
  analyze_haiku:
    input: STDIN
    model: gpt-4o-mini
    action: "This is the haiku analysis."
    output: STDOUT
  analyze_haiku:
    input: STDIN
    model: gpt-4o-mini
    action: "This is another haiku analysis."
    output: STDOUT
`
	// Try to unmarshal the YAML
	var dslConfig DSLConfig
	err := yaml.Unmarshal([]byte(deferYAML), &dslConfig)

	// Verify that an error was returned
	if err == nil {
		t.Error("Expected error for duplicate step names in defer block, but got nil")
	} else if !strings.Contains(err.Error(), "already defined") {
		t.Errorf("Expected error message to contain 'already defined', but got: %v", err)
	}
}

func TestProcessWithUncalledDefer(t *testing.T) {
	// Define a YAML where the deferred step should NOT be called
	deferYAML := `
determine_poem_type:
  input: STDIN
  model: gpt-4o-mini
  action: "Analyze the input poem"
  output: STDOUT

defer:
  analyze_sonnet:
    input: STDIN
    model: gpt-4o-mini
    action: "This is the sonnet analysis."
    output: STDOUT
`
	var dslConfig DSLConfig
	if err := yaml.Unmarshal([]byte(deferYAML), &dslConfig); err != nil {
		t.Fatalf("Failed to unmarshal yaml: %v", err)
	}

	// Create a custom mock provider for this test
	customMockProvider := &CustomMockProvider{
		MockProvider: *NewMockProvider("openai"),
		responses: map[string]string{
			"Analyze the input poem":       "Just a regular string output.",
			"This is the sonnet analysis.": "THIS_SHOULD_NOT_BE_RETURNED",
		},
	}
	customMockProvider.Configure("test-key")

	// Store the original DetectProvider function
	originalDetect := models.DetectProvider

	// Override the DetectProvider function to return our custom mock provider
	models.DetectProvider = func(modelName string) models.Provider {
		return customMockProvider
	}

	// Restore the original function when the test is done
	defer func() { models.DetectProvider = originalDetect }()

	// Create a processor with the default environment config
	processor := NewProcessor(&dslConfig, createTestEnvConfig(), createTestServerConfig(), true, "")
	processor.SetLastOutput("An old silent pond...")

	err := processor.Process()
	if err != nil {
		t.Fatalf("Process() failed: %v", err)
	}

	// The final output should be from the first step, as the defer step is never called.
	expectedOutput := "Just a regular string output."
	if processor.LastOutput() != expectedOutput {
		t.Errorf("Expected final output to be '%s', but got '%s'", expectedOutput, processor.LastOutput())
	}
}

func TestIsURL(t *testing.T) {
	processor := NewProcessor(&DSLConfig{}, createTestEnvConfig(), createTestServerConfig(), false, "")

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid http URL",
			input:    "http://example.com",
			expected: true,
		},
		{
			name:     "valid https URL",
			input:    "https://example.com/path?query=value",
			expected: true,
		},
		{
			name:     "invalid URL - no scheme",
			input:    "example.com",
			expected: false,
		},
		{
			name:     "invalid URL - empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "invalid URL - file path",
			input:    "/path/to/file.txt",
			expected: false,
		},
		{
			name:     "invalid URL - relative path",
			input:    "path/to/file.txt",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.isURL(tt.input)
			if result != tt.expected {
				t.Errorf("isURL() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFetchURL(t *testing.T) {
	processor := NewProcessor(&DSLConfig{}, createTestEnvConfig(), createTestServerConfig(), false, "")

	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/text":
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("Hello, World!"))
		case "/html":
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte("<html><body>Hello, World!</body></html>"))
		case "/json":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"message": "Hello, World!"}`))
		case "/error":
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	}))
	defer ts.Close()

	tests := []struct {
		name        string
		url         string
		expectError bool
		contentType string
	}{
		{
			name:        "fetch text content",
			url:         ts.URL + "/text",
			expectError: false,
			contentType: "text/plain",
		},
		{
			name:        "fetch HTML content",
			url:         ts.URL + "/html",
			expectError: false,
			contentType: "text/html",
		},
		{
			name:        "fetch JSON content",
			url:         ts.URL + "/json",
			expectError: false,
			contentType: "application/json",
		},
		{
			name:        "fetch error response",
			url:         ts.URL + "/error",
			expectError: true,
		},
		{
			name:        "invalid URL",
			url:         "http://invalid.url.that.does.not.exist",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpPath, err := processor.fetchURL(tt.url)
			if tt.expectError {
				if err == nil {
					t.Error("fetchURL() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("fetchURL() unexpected error: %v", err)
				}
				if tmpPath == "" {
					t.Error("fetchURL() returned empty path")
				}
				// Clean up temporary file
				if tmpPath != "" {
					if err := processor.processFile(tmpPath); err != nil {
						t.Errorf("Failed to process fetched file: %v", err)
					}
				}
			}
		})
	}
}
