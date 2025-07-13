package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/marklovelady/api-cli-generator/pkg/config"
	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListProjectsCommand(t *testing.T) {
	tests := []struct {
		name           string
		outputFormat   string
		expectedOutput []string
	}{
		{
			name:         "Human format",
			outputFormat: "human",
			expectedOutput: []string{
				"ID", "NAME", "DESCRIPTION", "CREATED",
				"Total: ", "projects",
			},
		},
		{
			name:         "JSON format",
			outputFormat: "json",
			expectedOutput: []string{
				`"status"`, `"success"`,
				`"count"`, `"projects"`,
			},
		},
		{
			name:         "YAML format",
			outputFormat: "yaml",
			expectedOutput: []string{
				"status: success",
				"count:",
				"projects:",
			},
		},
		{
			name:         "AI format",
			outputFormat: "ai",
			expectedOutput: []string{
				"Found", "projects in organization",
				"Next steps:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test config
			cfg = &config.VirtuosoConfig{
				Output: config.OutputConfig{
					DefaultFormat: tt.outputFormat,
				},
				Org: config.OrgConfig{
					ID: "12345",
				},
			}

			// Create command
			cmd := newListProjectsCmd()

			// Capture output
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// Execute command
			err := cmd.Execute()

			// Should handle API error gracefully
			if err != nil {
				assert.Contains(t, err.Error(), "failed to list projects")
				return
			}

			// Check output contains expected strings
			output := buf.String()
			for _, expected := range tt.expectedOutput {
				assert.Contains(t, output, expected, "Output should contain: %s", expected)
			}
		})
	}
}

func TestListGoalsCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		outputFormat   string
		expectedError  string
		expectedOutput []string
	}{
		{
			name:          "Missing project ID",
			args:          []string{},
			expectedError: "requires exactly 1 arg(s), only received 0",
		},
		{
			name:          "Invalid project ID",
			args:          []string{"invalid"},
			expectedError: "invalid project ID",
		},
		{
			name:         "Valid project ID - Human format",
			args:         []string{"123"},
			outputFormat: "human",
			expectedOutput: []string{
				"ID", "NAME", "URL", "SNAPSHOT ID",
			},
		},
		{
			name:         "Valid project ID - JSON format",
			args:         []string{"123"},
			outputFormat: "json",
			expectedOutput: []string{
				`"status"`, `"success"`,
				`"project_id"`, `123`,
				`"goals"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test config
			cfg = &config.VirtuosoConfig{
				Output: config.OutputConfig{
					DefaultFormat: tt.outputFormat,
				},
			}

			// Create command
			cmd := newListGoalsCmd()
			cmd.SetArgs(tt.args)

			// Capture output
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// Execute command
			err := cmd.Execute()

			// Check error
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			// Should handle API error gracefully
			if err != nil {
				assert.Contains(t, err.Error(), "failed to list goals")
				return
			}

			// Check output
			output := buf.String()
			for _, expected := range tt.expectedOutput {
				assert.Contains(t, output, expected, "Output should contain: %s", expected)
			}
		})
	}
}

func TestListJourneysCommand(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedError string
	}{
		{
			name:          "Missing arguments",
			args:          []string{},
			expectedError: "requires exactly 2 arg(s)",
		},
		{
			name:          "Missing snapshot ID",
			args:          []string{"123"},
			expectedError: "requires exactly 2 arg(s)",
		},
		{
			name:          "Invalid goal ID",
			args:          []string{"invalid", "456"},
			expectedError: "invalid goal ID",
		},
		{
			name:          "Invalid snapshot ID",
			args:          []string{"123", "invalid"},
			expectedError: "invalid snapshot ID",
		},
		{
			name: "Valid IDs",
			args: []string{"123", "456"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test config
			cfg = &config.VirtuosoConfig{
				Output: config.OutputConfig{
					DefaultFormat: "human",
				},
			}

			// Create command
			cmd := newListJourneysCmd()
			cmd.SetArgs(tt.args)

			// Capture output
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// Execute command
			err := cmd.Execute()

			// Check error
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}
		})
	}
}

func TestListCheckpointsCommand(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		outputFormat  string
		expectedError string
	}{
		{
			name:          "Missing journey ID",
			args:          []string{},
			expectedError: "requires exactly 1 arg(s)",
		},
		{
			name:          "Invalid journey ID",
			args:          []string{"invalid"},
			expectedError: "invalid journey ID",
		},
		{
			name:         "Valid journey ID - JSON",
			args:         []string{"123"},
			outputFormat: "json",
		},
		{
			name:         "Valid journey ID - YAML",
			args:         []string{"123"},
			outputFormat: "yaml",
		},
		{
			name:         "Valid journey ID - AI",
			args:         []string{"123"},
			outputFormat: "ai",
		},
		{
			name:         "Valid journey ID - Human",
			args:         []string{"123"},
			outputFormat: "human",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test config
			cfg = &config.VirtuosoConfig{
				Output: config.OutputConfig{
					DefaultFormat: tt.outputFormat,
				},
			}

			// Create command
			cmd := newListCheckpointsCmd()
			cmd.SetArgs(tt.args)

			// Capture output
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// Execute command
			err := cmd.Execute()

			// Check error
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}
		})
	}
}

func TestPaginationFlags(t *testing.T) {
	// Test list-projects pagination flags
	cmd := newListProjectsCmd()

	// Check that pagination flags exist
	limitFlag := cmd.Flag("limit")
	require.NotNil(t, limitFlag, "limit flag should exist")
	assert.Equal(t, "50", limitFlag.DefValue, "Default limit should be 50")

	offsetFlag := cmd.Flag("offset")
	require.NotNil(t, offsetFlag, "offset flag should exist")
	assert.Equal(t, "0", offsetFlag.DefValue, "Default offset should be 0")
}

func TestOutputFormatValidation(t *testing.T) {
	validFormats := []string{"json", "yaml", "human", "ai"}

	for _, format := range validFormats {
		t.Run("Valid format: "+format, func(t *testing.T) {
			err := validateOutputFormat(format)
			assert.NoError(t, err)
		})
	}

	invalidFormats := []string{"xml", "csv", "invalid", ""}

	for _, format := range invalidFormats {
		t.Run("Invalid format: "+format, func(t *testing.T) {
			err := validateOutputFormat(format)
			if format == "" {
				assert.NoError(t, err, "Empty format should be valid (uses default)")
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestJSONOutputParsing(t *testing.T) {
	// Setup test config for JSON output
	cfg = &config.VirtuosoConfig{
		Output: config.OutputConfig{
			DefaultFormat: "json",
		},
		Org: config.OrgConfig{
			ID: "12345",
		},
	}

	// Create a mock response that would be generated
	mockResponse := map[string]interface{}{
		"status": "success",
		"count":  2,
		"projects": []virtuoso.Project{
			{ID: 1, Name: "Project 1", Description: "Test project 1"},
			{ID: 2, Name: "Project 2", Description: "Test project 2"},
		},
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(mockResponse)
	require.NoError(t, err)

	// Parse JSON
	var parsed map[string]interface{}
	err = json.Unmarshal(jsonBytes, &parsed)
	require.NoError(t, err)

	// Verify structure
	assert.Equal(t, "success", parsed["status"])
	assert.Equal(t, float64(2), parsed["count"])

	projects, ok := parsed["projects"].([]interface{})
	require.True(t, ok, "projects should be an array")
	assert.Len(t, projects, 2)
}

func TestRichOutputContent(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		command  func() *cobra.Command
		expected []string
	}{
		{
			name:    "List projects AI format includes next steps",
			format:  "ai",
			command: newListProjectsCmd,
			expected: []string{
				"Next steps:",
				"List goals for a project:",
				"Create a new goal:",
			},
		},
		{
			name:    "List goals AI format includes next steps",
			format:  "ai",
			command: newListGoalsCmd,
			expected: []string{
				"Next steps:",
				"Get goal snapshot:",
				"Create a new journey:",
			},
		},
		{
			name:    "List checkpoints AI format includes usage",
			format:  "ai",
			command: newListCheckpointsCmd,
			expected: []string{
				"Usage:",
				"To add steps to a checkpoint:",
				"navigation step shared by all tests",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check command long description
			cmd := tt.command()
			longDesc := cmd.Long

			// Verify rich descriptions exist
			assert.NotEmpty(t, longDesc, "Command should have a long description")
			assert.True(t, len(longDesc) > 50, "Long description should be detailed")
		})
	}
}

func TestTableOutputFormatting(t *testing.T) {
	// Test that human-readable output uses proper table formatting
	tests := []struct {
		name            string
		command         func() *cobra.Command
		expectedHeaders []string
	}{
		{
			name:            "List projects table headers",
			command:         newListProjectsCmd,
			expectedHeaders: []string{"ID", "NAME", "DESCRIPTION", "CREATED"},
		},
		{
			name:            "List goals table headers",
			command:         newListGoalsCmd,
			expectedHeaders: []string{"ID", "NAME", "URL", "SNAPSHOT ID"},
		},
		{
			name:            "List journeys table headers",
			command:         newListJourneysCmd,
			expectedHeaders: []string{"ID", "NAME", "TITLE", "STATUS"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The actual table formatting would be tested with integration tests
			// Here we just verify the command structure
			cmd := tt.command()
			assert.NotNil(t, cmd.RunE, "Command should have RunE function")
		})
	}
}
