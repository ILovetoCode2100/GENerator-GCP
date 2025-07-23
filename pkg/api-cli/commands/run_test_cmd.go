package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// TestDefinition represents a simplified test format
type TestDefinition struct {
	Name        string                   `yaml:"name" json:"name"`
	Description string                   `yaml:"description,omitempty" json:"description,omitempty"`
	Project     interface{}              `yaml:"project,omitempty" json:"project,omitempty"` // Can be ID or name
	Variables   []TestVariable           `yaml:"variables,omitempty" json:"variables,omitempty"`
	Steps       []map[string]interface{} `yaml:"steps" json:"steps"`
	Config      TestConfig               `yaml:"config,omitempty" json:"config,omitempty"`
}

// TestVariable represents a test variable
type TestVariable struct {
	Name  string `yaml:"name" json:"name"`
	Value string `yaml:"value" json:"value"`
}

// TestConfig represents test configuration
type TestConfig struct {
	ContinueOnError bool `yaml:"continue_on_error,omitempty" json:"continue_on_error,omitempty"`
	Timeout         int  `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}

// TestResult represents the result of running a test
type TestResult struct {
	Success      bool             `json:"success"`
	ProjectID    string           `json:"project_id"`
	GoalID       string           `json:"goal_id"`
	JourneyID    string           `json:"journey_id"`
	CheckpointID string           `json:"checkpoint_id"`
	Steps        []TestStepResult `json:"steps"`
	Links        TestLinks        `json:"links"`
	Error        string           `json:"error,omitempty"`
}

// TestStepResult represents the result of a single test step
type TestStepResult struct {
	Position    int    `json:"position"`
	Command     string `json:"command"`
	Success     bool   `json:"success"`
	StepID      string `json:"step_id,omitempty"`
	Error       string `json:"error,omitempty"`
	Description string `json:"description,omitempty"`
}

// TestLinks provides URLs to view the test in Virtuoso UI
type TestLinks struct {
	Project    string `json:"project,omitempty"`
	Goal       string `json:"goal,omitempty"`
	Journey    string `json:"journey,omitempty"`
	Checkpoint string `json:"checkpoint,omitempty"`
}

func newRunTestCmd() *cobra.Command {
	var (
		dryRun       bool
		execute      bool
		outputFormat string
		projectName  string
	)

	cmd := &cobra.Command{
		Use:   "run-test [file]",
		Short: "Run a test from a simplified YAML/JSON definition",
		Long: `Run a test from a simplified YAML or JSON definition file.

This command provides a single interface for creating and optionally executing tests.
It automatically handles project, goal, journey, and checkpoint creation.

The test definition focuses on the actual test steps, minimizing boilerplate.

Example test definition (YAML):
  name: "Login Test"
  steps:
    - navigate: "https://example.com"
    - click: "#login"
    - write:
        selector: "#email"
        text: "test@example.com"
    - assert: "Welcome"

Example test definition (JSON):
  {
    "name": "Login Test",
    "steps": [
      {"navigate": "https://example.com"},
      {"click": "#login"},
      {"write": {"selector": "#email", "text": "test@example.com"}},
      {"assert": "Welcome"}
    ]
  }

The command accepts input from:
  - A file: api-cli run-test test.yaml
  - Stdin: cat test.yaml | api-cli run-test -
  - Stdin: echo '{"name":"Test","steps":[...]}' | api-cli run-test -`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create context
			ctx, cancel := CommandContext()
			defer cancel()

			// Initialize client
			apiClient := client.NewClient(cfg)

			// Read test definition
			var input []byte
			var err error

			if len(args) == 0 || args[0] == "-" {
				// Read from stdin
				input, err = io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("failed to read from stdin: %w", err)
				}
			} else {
				// Read from file
				input, err = os.ReadFile(args[0])
				if err != nil {
					return fmt.Errorf("failed to read file: %w", err)
				}
			}

			// Parse test definition
			var testDef TestDefinition

			// Try YAML first
			if err := yaml.Unmarshal(input, &testDef); err != nil {
				// Try JSON
				if err := json.Unmarshal(input, &testDef); err != nil {
					return fmt.Errorf("failed to parse input as YAML or JSON: %w", err)
				}
			}

			// Validate test definition
			if len(testDef.Steps) == 0 {
				return fmt.Errorf("test definition must contain at least one step")
			}

			// Generate name if not provided
			if testDef.Name == "" {
				testDef.Name = fmt.Sprintf("Test Run %s", time.Now().Format("2006-01-02 15:04:05"))
			}

			// Create test result
			result := TestResult{
				Success: true,
				Steps:   []TestStepResult{},
			}

			// Dry run mode - just show what would be created
			if dryRun {
				return outputDryRun(testDef, outputFormat)
			}

			// Create or find project
			projectID, err := resolveProject(ctx, apiClient, testDef.Project, projectName)
			if err != nil {
				result.Success = false
				result.Error = fmt.Sprintf("Failed to resolve project: %v", err)
				return outputResult(result, outputFormat)
			}
			result.ProjectID = projectID

			// Create goal
			goalName := fmt.Sprintf("%s - Goal", testDef.Name)
			// Convert projectID string to int
			projectIDInt, err := strconv.Atoi(projectID)
			if err != nil {
				result.Success = false
				result.Error = fmt.Sprintf("Invalid project ID: %v", err)
				return outputResult(result, outputFormat)
			}

			goal, err := callWithContext(ctx, func() (*client.Goal, error) {
				// CreateGoal needs projectID, name, and URL
				return apiClient.CreateGoal(projectIDInt, goalName, "")
			})
			if err != nil {
				result.Success = false
				result.Error = fmt.Sprintf("Failed to create goal: %v", err)
				return outputResult(result, outputFormat)
			}
			result.GoalID = fmt.Sprintf("%d", goal.ID)

			// Create journey
			journeyName := testDef.Name
			// Parse snapshot ID from goal
			snapshotIDInt, err := strconv.Atoi(goal.SnapshotID)
			if err != nil {
				result.Success = false
				result.Error = fmt.Sprintf("Invalid snapshot ID: %v", err)
				return outputResult(result, outputFormat)
			}

			journey, err := callWithContext(ctx, func() (*client.Journey, error) {
				return apiClient.CreateJourney(goal.ID, snapshotIDInt, journeyName)
			})
			if err != nil {
				result.Success = false
				result.Error = fmt.Sprintf("Failed to create journey: %v", err)
				return outputResult(result, outputFormat)
			}
			result.JourneyID = fmt.Sprintf("%d", journey.ID)

			// Create checkpoint
			checkpointName := "Test Steps"
			checkpoint, err := callWithContext(ctx, func() (*client.Checkpoint, error) {
				return apiClient.CreateCheckpoint(goal.ID, snapshotIDInt, checkpointName)
			})
			if err != nil {
				result.Success = false
				result.Error = fmt.Sprintf("Failed to create checkpoint: %v", err)
				return outputResult(result, outputFormat)
			}
			result.CheckpointID = fmt.Sprintf("%d", checkpoint.ID)

			// Add links
			result.Links = TestLinks{
				Project:    fmt.Sprintf("https://app.virtuoso.qa/#/project/%s", projectID),
				Goal:       fmt.Sprintf("https://app.virtuoso.qa/#/goal/%d", goal.ID),
				Journey:    fmt.Sprintf("https://app.virtuoso.qa/#/journey/%d", journey.ID),
				Checkpoint: fmt.Sprintf("https://app.virtuoso.qa/#/checkpoint/%d", checkpoint.ID),
			}

			// Create steps
			for i, stepDef := range testDef.Steps {
				position := i + 1
				stepResult := TestStepResult{
					Position: position,
					Success:  true,
				}

				// Convert simplified step format to CLI command
				command, args, err := parseStepDefinition(stepDef)
				if err != nil {
					stepResult.Success = false
					stepResult.Error = err.Error()
					stepResult.Command = fmt.Sprintf("Invalid step: %v", stepDef)
				} else {
					stepResult.Command = fmt.Sprintf("%s %s", command, strings.Join(args, " "))

					// Execute the step command
					stepID, err := executeStep(apiClient, checkpoint.ID, position, command, args)
					if err != nil {
						stepResult.Success = false
						stepResult.Error = err.Error()
						if !testDef.Config.ContinueOnError {
							result.Success = false
							result.Steps = append(result.Steps, stepResult)
							result.Error = fmt.Sprintf("Step %d failed: %v", position, err)
							return outputResult(result, outputFormat)
						}
					} else {
						stepResult.StepID = stepID
					}
				}

				result.Steps = append(result.Steps, stepResult)
			}

			// Execute the test if requested
			if execute {
				// TODO: Implement test execution
				fmt.Fprintln(os.Stderr, "Note: Test execution not yet implemented. Test created successfully.")
			}

			return outputResult(result, outputFormat)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be created without actually creating it")
	cmd.Flags().BoolVar(&execute, "execute", false, "Execute the test after creation")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "human", "Output format: human, json, yaml")
	cmd.Flags().StringVar(&projectName, "project-name", "", "Create new project with this name (overrides project field)")

	return cmd
}

// resolveProject finds or creates a project based on the input
func resolveProject(ctx context.Context, apiClient *client.Client, projectRef interface{}, projectName string) (string, error) {
	// If project name flag is provided, create new project
	if projectName != "" {
		project, err := callWithContext(ctx, func() (*client.Project, error) {
			return apiClient.CreateProject(projectName, "")
		})
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", project.ID), nil
	}

	// If no project reference, create a new one
	if projectRef == nil {
		projectName := fmt.Sprintf("Test Project %s", time.Now().Format("2006-01-02"))
		project, err := callWithContext(ctx, func() (*client.Project, error) {
			return apiClient.CreateProject(projectName, "")
		})
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", project.ID), nil
	}

	// Check if it's a number (ID) or string (name)
	switch v := projectRef.(type) {
	case int:
		return fmt.Sprintf("%d", v), nil
	case float64:
		return fmt.Sprintf("%d", int(v)), nil
	case string:
		// Try to parse as ID first
		if _, err := fmt.Sscanf(v, "%d", new(int)); err == nil {
			return v, nil
		}
		// Otherwise, create project with this name
		project, err := callWithContext(ctx, func() (*client.Project, error) {
			return apiClient.CreateProject(v, "")
		})
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", project.ID), nil
	default:
		return "", fmt.Errorf("invalid project reference type: %T", v)
	}
}

// parseStepDefinition converts simplified step format to CLI command and args
func parseStepDefinition(stepDef map[string]interface{}) (string, []string, error) {
	// Handle simple string commands
	for cmd, value := range stepDef {
		switch cmd {
		case "navigate":
			if url, ok := value.(string); ok {
				return "step-navigate", []string{"to", url}, nil
			}
		case "click":
			if selector, ok := value.(string); ok {
				return "step-interact", []string{"click", selector}, nil
			}
		case "write":
			if writeMap, ok := value.(map[string]interface{}); ok {
				selector := writeMap["selector"].(string)
				text := writeMap["text"].(string)
				return "step-interact", []string{"write", selector, text}, nil
			}
		case "assert":
			if text, ok := value.(string); ok {
				return "step-assert", []string{"exists", text}, nil
			}
		case "wait":
			switch v := value.(type) {
			case string:
				return "step-wait", []string{"element", v}, nil
			case int:
				return "step-wait", []string{"time", fmt.Sprintf("%d", v)}, nil
			case float64:
				return "step-wait", []string{"time", fmt.Sprintf("%d", int(v))}, nil
			}
		case "hover":
			if selector, ok := value.(string); ok {
				return "step-interact", []string{"hover", selector}, nil
			}
		case "key":
			if key, ok := value.(string); ok {
				return "step-interact", []string{"key", key}, nil
			}
		case "select":
			if selectMap, ok := value.(map[string]interface{}); ok {
				selector := selectMap["selector"].(string)
				option := selectMap["option"].(string)
				return "step-interact", []string{"select", "option", selector, option}, nil
			}
		case "store":
			if storeMap, ok := value.(map[string]interface{}); ok {
				selector := storeMap["selector"].(string)
				variable := storeMap["as"].(string)
				storeType := "element-text"
				if t, ok := storeMap["type"].(string); ok {
					storeType = t
				}
				return "step-data", []string{"store", storeType, selector, variable}, nil
			}
		case "comment":
			if text, ok := value.(string); ok {
				return "step-misc", []string{"comment", text}, nil
			}
		case "execute":
			if script, ok := value.(string); ok {
				return "step-misc", []string{"execute", script}, nil
			}
		}
	}

	return "", nil, fmt.Errorf("unknown step format: %v", stepDef)
}

// executeStep runs a single step command
func executeStep(apiClient *client.Client, checkpointID int, position int, command string, args []string) (string, error) {
	// Build full args including checkpoint ID and position
	fullArgs := append([]string{fmt.Sprintf("%d", checkpointID)}, args...)
	fullArgs = append(fullArgs, fmt.Sprintf("%d", position))

	// Get the command function based on command name
	switch command {
	case "step-navigate":
		// Handle navigation commands
		if len(args) > 0 && args[0] == "to" && len(args) > 1 {
			stepID, err := apiClient.CreateNavigationStep(checkpointID, args[1], position)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("%d", stepID), nil
		}
	case "step-interact":
		// Handle interaction commands
		if len(args) > 0 {
			switch args[0] {
			case "click":
				if len(args) > 1 {
					stepID, err := apiClient.CreateClickStep(checkpointID, args[1], position)
					if err != nil {
						return "", err
					}
					return fmt.Sprintf("%d", stepID), nil
				}
			case "write":
				if len(args) > 2 {
					stepID, err := apiClient.CreateWriteStep(checkpointID, args[1], args[2], position)
					if err != nil {
						return "", err
					}
					return fmt.Sprintf("%d", stepID), nil
				}
			case "hover":
				if len(args) > 1 {
					stepID, err := apiClient.CreateHoverStep(checkpointID, args[1], position)
					if err != nil {
						return "", err
					}
					return fmt.Sprintf("%d", stepID), nil
				}
			}
		}
	case "step-assert":
		// Handle assertion commands
		if len(args) > 0 && args[0] == "exists" && len(args) > 1 {
			stepID, err := apiClient.CreateAssertExistsStep(checkpointID, args[1], position)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("%d", stepID), nil
		}
	case "step-wait":
		// Handle wait commands
		if len(args) > 0 {
			switch args[0] {
			case "element":
				if len(args) > 1 {
					stepID, err := apiClient.CreateWaitElementStep(checkpointID, args[1], position)
					if err != nil {
						return "", err
					}
					return fmt.Sprintf("%d", stepID), nil
				}
			case "time":
				if len(args) > 1 {
					// Convert string to int for time in milliseconds
					timeMs, err := strconv.Atoi(args[1])
					if err != nil {
						return "", fmt.Errorf("invalid time value: %s", args[1])
					}
					// API expects seconds, so convert from milliseconds
					seconds := timeMs / 1000
					stepID, err := apiClient.CreateWaitTimeStep(checkpointID, seconds, position)
					if err != nil {
						return "", err
					}
					return fmt.Sprintf("%d", stepID), nil
				}
			}
		}
	}

	// For other commands, return a generic implementation
	// In a real implementation, this would call the appropriate command handler
	return "", fmt.Errorf("step command not yet implemented: %s", command)
}

// outputDryRun shows what would be created
func outputDryRun(testDef TestDefinition, format string) error {
	switch format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(testDef)
	case "yaml":
		enc := yaml.NewEncoder(os.Stdout)
		return enc.Encode(testDef)
	default:
		fmt.Printf("Test Definition (Dry Run)\n")
		fmt.Printf("========================\n\n")
		fmt.Printf("Name: %s\n", testDef.Name)
		if testDef.Description != "" {
			fmt.Printf("Description: %s\n", testDef.Description)
		}
		if testDef.Project != nil {
			fmt.Printf("Project: %v\n", testDef.Project)
		} else {
			fmt.Printf("Project: <will create new>\n")
		}
		fmt.Printf("\nSteps (%d):\n", len(testDef.Steps))
		for i, step := range testDef.Steps {
			cmd, args, err := parseStepDefinition(step)
			if err != nil {
				fmt.Printf("  %d. ERROR: %v\n", i+1, err)
			} else {
				fmt.Printf("  %d. %s %s\n", i+1, cmd, strings.Join(args, " "))
			}
		}
		fmt.Printf("\n")
	}
	return nil
}

// outputResult outputs the test result in the requested format
func outputResult(result TestResult, format string) error {
	switch format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	case "yaml":
		enc := yaml.NewEncoder(os.Stdout)
		return enc.Encode(result)
	default:
		if result.Success {
			fmt.Printf("✓ Test created successfully!\n\n")
			fmt.Printf("Infrastructure:\n")
			fmt.Printf("  Project:    %s\n", result.ProjectID)
			fmt.Printf("  Goal:       %s\n", result.GoalID)
			fmt.Printf("  Journey:    %s\n", result.JourneyID)
			fmt.Printf("  Checkpoint: %s\n", result.CheckpointID)
			fmt.Printf("\nSteps created: %d\n", len(result.Steps))

			failedSteps := 0
			for _, step := range result.Steps {
				if !step.Success {
					failedSteps++
					fmt.Printf("  ✗ Step %d: %s\n", step.Position, step.Error)
				}
			}

			if failedSteps > 0 {
				fmt.Printf("\n⚠️  %d steps failed\n", failedSteps)
			} else {
				fmt.Printf("  All steps created successfully\n")
			}

			fmt.Printf("\nView in Virtuoso:\n")
			fmt.Printf("  Checkpoint: %s\n", result.Links.Checkpoint)
		} else {
			fmt.Printf("✗ Test creation failed\n")
			fmt.Printf("Error: %s\n", result.Error)
			if len(result.Steps) > 0 {
				fmt.Printf("\nSteps attempted: %d\n", len(result.Steps))
				for _, step := range result.Steps {
					if step.Success {
						fmt.Printf("  ✓ Step %d: %s\n", step.Position, step.Command)
					} else {
						fmt.Printf("  ✗ Step %d: %s - %s\n", step.Position, step.Command, step.Error)
					}
				}
			}
		}
	}
	return nil
}
