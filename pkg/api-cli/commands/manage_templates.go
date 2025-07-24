package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// TestTemplate represents a test journey template for AI consumption
// This structure maps directly to YAML test definitions
type TestTemplate struct {
	Journey JourneyTemplate `yaml:"journey"`
}

// JourneyTemplate defines the structure of a test journey
type JourneyTemplate struct {
	Name        string               `yaml:"name"`
	ProjectID   interface{}          `yaml:"project_id"` // Can be string or int
	Description string               `yaml:"description,omitempty"`
	Config      *JourneyConfig       `yaml:"config,omitempty"`
	Variables   []Variable           `yaml:"variables,omitempty"`
	Checkpoints []CheckpointTemplate `yaml:"checkpoints"`
}

// JourneyConfig holds journey-level configuration
type JourneyConfig struct {
	RetryFailedSteps  bool `yaml:"retry_failed_steps,omitempty"`
	ScreenshotOnError bool `yaml:"screenshot_on_error,omitempty"`
	ContinueOnError   bool `yaml:"continue_on_error,omitempty"`
	TimeoutDefault    int  `yaml:"timeout_default,omitempty"`
	MaxRetryCount     int  `yaml:"max_retry_count,omitempty"`
}

// Variable represents a test variable
type Variable struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

// CheckpointTemplate represents a checkpoint in a journey
type CheckpointTemplate struct {
	Name        string         `yaml:"name"`
	Position    int            `yaml:"position,omitempty"`
	Description string         `yaml:"description,omitempty"`
	Steps       []StepTemplate `yaml:"steps"`
}

// StepTemplate represents a single test step
type StepTemplate struct {
	Command     string                 `yaml:"command"`
	Args        []string               `yaml:"args,omitempty"`
	Description string                 `yaml:"description,omitempty"`
	Options     map[string]interface{} `yaml:"options,omitempty"`
	Condition   *ConditionTemplate     `yaml:"condition,omitempty"`
	Then        []StepTemplate         `yaml:"then,omitempty"`
	Else        []StepTemplate         `yaml:"else,omitempty"`
	Try         []StepTemplate         `yaml:"try,omitempty"`
	Catch       []StepTemplate         `yaml:"catch,omitempty"`
	Finally     []StepTemplate         `yaml:"finally,omitempty"`
}

// ConditionTemplate represents a conditional check
type ConditionTemplate struct {
	Command  string   `yaml:"command,omitempty"`
	Args     []string `yaml:"args,omitempty"`
	Variable string   `yaml:"variable,omitempty"`
	Exists   bool     `yaml:"exists,omitempty"`
	Timeout  int      `yaml:"timeout,omitempty"`
}

// LoadTestTemplateCmd provides a command to load and validate test templates
var LoadTestTemplateCmd = &cobra.Command{
	Use:   "load-template [file]",
	Short: "Load and validate a test template YAML file",
	Long: `Load a test template from a YAML file and validate its structure.
This command is useful for AI systems to verify template correctness.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		templateFile := args[0]

		// Read the template file
		data, err := os.ReadFile(templateFile)
		if err != nil {
			return fmt.Errorf("failed to read template file: %w", err)
		}

		// Parse the YAML
		var template TestTemplate
		if err := yaml.Unmarshal(data, &template); err != nil {
			return fmt.Errorf("failed to parse YAML: %w", err)
		}

		// Validate the template
		if err := validateTemplate(&template); err != nil {
			return fmt.Errorf("template validation failed: %w", err)
		}

		// Output based on format
		outputFormat, _ := cmd.Flags().GetString("output")
		return outputTemplateInfo(&template, outputFormat)
	},
}

// validateTemplate checks if a template is valid
func validateTemplate(template *TestTemplate) error {
	if template.Journey.Name == "" {
		return fmt.Errorf("journey name is required")
	}

	if template.Journey.ProjectID == nil {
		return fmt.Errorf("project_id is required")
	}

	if len(template.Journey.Checkpoints) == 0 {
		return fmt.Errorf("at least one checkpoint is required")
	}

	for i, checkpoint := range template.Journey.Checkpoints {
		if checkpoint.Name == "" {
			return fmt.Errorf("checkpoint %d: name is required", i+1)
		}

		if len(checkpoint.Steps) == 0 {
			return fmt.Errorf("checkpoint '%s': at least one step is required", checkpoint.Name)
		}

		for j, step := range checkpoint.Steps {
			if step.Command == "" {
				return fmt.Errorf("checkpoint '%s', step %d: command is required", checkpoint.Name, j+1)
			}
		}
	}

	return nil
}

// outputTemplateInfo outputs template information in the requested format
func outputTemplateInfo(template *TestTemplate, format string) error {
	info := map[string]interface{}{
		"valid":            true,
		"journey_name":     template.Journey.Name,
		"project_id":       template.Journey.ProjectID,
		"checkpoint_count": len(template.Journey.Checkpoints),
		"total_steps":      countSteps(template.Journey.Checkpoints),
		"has_config":       template.Journey.Config != nil,
		"variable_count":   len(template.Journey.Variables),
		"checkpoints":      summarizeCheckpoints(template.Journey.Checkpoints),
	}

	bc := NewBaseCommand()
	output, err := bc.FormatOutput(info, format)
	if err != nil {
		return fmt.Errorf("failed to format template info output: %w", err)
	}

	fmt.Println(output)
	return nil
}

// countSteps counts total steps across all checkpoints
func countSteps(checkpoints []CheckpointTemplate) int {
	count := 0
	for _, cp := range checkpoints {
		count += countStepsRecursive(cp.Steps)
	}
	return count
}

// countStepsRecursive counts steps including conditional branches
func countStepsRecursive(steps []StepTemplate) int {
	count := len(steps)
	for _, step := range steps {
		count += countStepsRecursive(step.Then)
		count += countStepsRecursive(step.Else)
		count += countStepsRecursive(step.Try)
		count += countStepsRecursive(step.Catch)
		count += countStepsRecursive(step.Finally)
	}
	return count
}

// summarizeCheckpoints creates a summary of checkpoints
func summarizeCheckpoints(checkpoints []CheckpointTemplate) []map[string]interface{} {
	summary := make([]map[string]interface{}, len(checkpoints))
	for i, cp := range checkpoints {
		summary[i] = map[string]interface{}{
			"name":             cp.Name,
			"position":         cp.Position,
			"step_count":       countStepsRecursive(cp.Steps),
			"has_conditionals": hasConditionals(cp.Steps),
		}
	}
	return summary
}

// hasConditionals checks if steps contain conditional logic
func hasConditionals(steps []StepTemplate) bool {
	for _, step := range steps {
		if step.Condition != nil || len(step.Then) > 0 || len(step.Try) > 0 {
			return true
		}
		if hasConditionals(step.Then) || hasConditionals(step.Else) ||
			hasConditionals(step.Try) || hasConditionals(step.Catch) ||
			hasConditionals(step.Finally) {
			return true
		}
	}
	return false
}

// GenerateCommandsCmd generates CLI commands from a test template
var GenerateCommandsCmd = &cobra.Command{
	Use:   "generate-commands [template-file]",
	Short: "Generate CLI commands from a test template",
	Long: `Generate executable CLI commands from a YAML test template.
This is useful for AI systems to convert templates into actual test executions.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		templateFile := args[0]

		// Read and parse template
		data, err := os.ReadFile(templateFile)
		if err != nil {
			return fmt.Errorf("failed to read template: %w", err)
		}

		var template TestTemplate
		if err := yaml.Unmarshal(data, &template); err != nil {
			return fmt.Errorf("failed to parse YAML: %w", err)
		}

		// Generate commands
		commands := generateCommands(&template)

		// Output format
		outputFormat, _ := cmd.Flags().GetString("output")
		scriptMode, _ := cmd.Flags().GetBool("script")

		if scriptMode || outputFormat == "script" {
			// Output as executable script
			fmt.Println("#!/bin/bash")
			fmt.Println("# Generated from:", templateFile)
			fmt.Println("# Journey:", template.Journey.Name)
			fmt.Println()

			// Add variables
			for _, v := range template.Journey.Variables {
				fmt.Printf("export %s=\"%s\"\n", v.Name, v.Value)
			}
			fmt.Println()

			// Add commands
			for _, cmd := range commands {
				fmt.Println(cmd)
			}
		} else {
			// Output as JSON/YAML/etc
			bc := NewBaseCommand()
			output, _ := bc.FormatOutput(commands, outputFormat)
			fmt.Println(output)
		}

		return nil
	},
}

// generateCommands converts a template into CLI commands
func generateCommands(template *TestTemplate) []string {
	var commands []string

	// Create journey (placeholder - would need actual IDs)
	commands = append(commands, fmt.Sprintf(
		"# Create journey: %s", template.Journey.Name))
	commands = append(commands, fmt.Sprintf(
		"api-cli create-journey %v \"%s\"",
		template.Journey.ProjectID, template.Journey.Name))
	commands = append(commands, "JOURNEY_ID=$LAST_ID  # Capture from response")
	commands = append(commands, "")

	// Create checkpoints and steps
	for i, checkpoint := range template.Journey.Checkpoints {
		position := checkpoint.Position
		if position == 0 {
			position = i + 1
		}

		commands = append(commands, fmt.Sprintf(
			"# Checkpoint %d: %s", position, checkpoint.Name))
		commands = append(commands, fmt.Sprintf(
			"api-cli create-checkpoint $JOURNEY_ID \"%s\" %d",
			checkpoint.Name, position))
		commands = append(commands, "CHECKPOINT_ID=$LAST_ID")
		commands = append(commands, "export VIRTUOSO_SESSION_ID=$CHECKPOINT_ID")
		commands = append(commands, "")

		// Generate step commands
		stepCommands := generateStepCommands(checkpoint.Steps, 1)
		commands = append(commands, stepCommands...)
		commands = append(commands, "")
	}

	return commands
}

// generateStepCommands converts steps to CLI commands
func generateStepCommands(steps []StepTemplate, startPos int) []string {
	var commands []string
	position := startPos

	for _, step := range steps {
		if step.Command == "conditional" {
			// Handle conditional logic
			commands = append(commands, "# Conditional block")
			if step.Condition != nil && step.Condition.Command != "" {
				commands = append(commands, fmt.Sprintf(
					"if api-cli %s %s --output json | jq -e '.success' > /dev/null; then",
					step.Condition.Command, joinArgs(step.Condition.Args)))

				thenCommands := generateStepCommands(step.Then, position)
				for _, cmd := range thenCommands {
					commands = append(commands, "  "+cmd)
				}
				position += len(step.Then)

				if len(step.Else) > 0 {
					commands = append(commands, "else")
					elseCommands := generateStepCommands(step.Else, position)
					for _, cmd := range elseCommands {
						commands = append(commands, "  "+cmd)
					}
					position += len(step.Else)
				}

				commands = append(commands, "fi")
			}
		} else if step.Command == "try-catch" {
			// Handle try-catch blocks
			commands = append(commands, "# Try-catch block")
			commands = append(commands, "set +e  # Allow errors")

			tryCommands := generateStepCommands(step.Try, position)
			commands = append(commands, tryCommands...)
			position += len(step.Try)

			commands = append(commands, "if [ $? -ne 0 ]; then")
			catchCommands := generateStepCommands(step.Catch, position)
			for _, cmd := range catchCommands {
				commands = append(commands, "  "+cmd)
			}
			position += len(step.Catch)
			commands = append(commands, "fi")

			if len(step.Finally) > 0 {
				finallyCommands := generateStepCommands(step.Finally, position)
				commands = append(commands, finallyCommands...)
				position += len(step.Finally)
			}

			commands = append(commands, "set -e  # Re-enable error handling")
		} else {
			// Regular command
			cmd := fmt.Sprintf("api-cli %s", step.Command)

			// Add arguments
			if len(step.Args) > 0 {
				cmd += " " + joinArgs(step.Args)
			}

			// Add options as flags
			for key, value := range step.Options {
				cmd += fmt.Sprintf(" --%s %v", key, value)
			}

			// Add description as comment
			if step.Description != "" {
				commands = append(commands, fmt.Sprintf("# %s", step.Description))
			}

			commands = append(commands, cmd)
			position++
		}
	}

	return commands
}

// joinArgs joins arguments with proper quoting
func joinArgs(args []string) string {
	result := ""
	for i, arg := range args {
		if i > 0 {
			result += " "
		}
		// Quote if contains spaces or special chars
		if containsSpecial(arg) {
			result += fmt.Sprintf("'%s'", arg)
		} else {
			result += arg
		}
	}
	return result
}

// containsSpecial checks if string needs quoting
func containsSpecial(s string) bool {
	for _, c := range s {
		if c == ' ' || c == '\t' || c == '\n' || c == '"' || c == '\'' || c == '$' {
			return true
		}
	}
	return false
}

func init() {
	// Add flags to commands
	GenerateCommandsCmd.Flags().Bool("script", false, "Output as executable bash script")
}

// GetTestTemplatesCmd lists available test templates
var GetTestTemplatesCmd = &cobra.Command{
	Use:   "get-templates [directory]",
	Short: "List available test templates",
	Long:  `List all YAML test templates in the specified directory (default: ./examples)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "./examples"
		if len(args) > 0 {
			dir = args[0]
		}

		// Find all YAML files
		matches, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
		if err != nil {
			return fmt.Errorf("failed to scan directory '%s' for YAML files: %w", dir, err)
		}

		templates := make([]map[string]interface{}, 0)
		for _, file := range matches {
			// Read and parse each template
			data, err := os.ReadFile(file)
			if err != nil {
				continue
			}

			var template TestTemplate
			if err := yaml.Unmarshal(data, &template); err != nil {
				continue
			}

			templates = append(templates, map[string]interface{}{
				"file":        filepath.Base(file),
				"path":        file,
				"name":        template.Journey.Name,
				"description": template.Journey.Description,
				"checkpoints": len(template.Journey.Checkpoints),
				"steps":       countSteps(template.Journey.Checkpoints),
			})
		}

		// Output
		outputFormat, _ := cmd.Flags().GetString("output")
		bc := NewBaseCommand()
		output, _ := bc.FormatOutput(templates, outputFormat)
		fmt.Println(output)

		return nil
	},
}
