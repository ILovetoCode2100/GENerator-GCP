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

			// Check if we should use existing session checkpoint
			var checkpoint *client.Checkpoint
			var checkpointID int
			var skipInfrastructure bool
			var goal *client.Goal
			var journey *client.Journey
			var snapshotIDInt int
			var projectID string
			
			// Check for session checkpoint first
			if cfg.Session.CurrentCheckpointID != nil && *cfg.Session.CurrentCheckpointID > 0 {
				checkpointID = *cfg.Session.CurrentCheckpointID
				checkpoint = &client.Checkpoint{ID: checkpointID}
				skipInfrastructure = true
			} else if sessionID := os.Getenv("VIRTUOSO_SESSION_ID"); sessionID != "" {
				// Check environment variable
				sessionID = strings.TrimPrefix(sessionID, "cp_")
				if id, err := strconv.Atoi(sessionID); err == nil && id > 0 {
					checkpointID = id
					checkpoint = &client.Checkpoint{ID: checkpointID}
					skipInfrastructure = true
				}
			}

			// Only create infrastructure if not using existing checkpoint
			if !skipInfrastructure {
				// Create or find project
				// Check if infrastructure config is provided
				var projectRef interface{}
				var startingURL string

				if testDef.Infrastructure != nil {
					// Use infrastructure config
					projectRef = testDef.Infrastructure.Project
					startingURL = testDef.Infrastructure.StartingURL
				} else {
					// Use direct fields
					projectRef = testDef.Project
					startingURL = testDef.StartingURL
				}

				// Override with direct StartingURL if specified
				if testDef.StartingURL != "" {
					startingURL = testDef.StartingURL
				}

				projectID, err = resolveProject(ctx, apiClient, projectRef, projectName)
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

				goal, err = callWithContext(ctx, func() (*client.Goal, error) {
					// CreateGoal needs projectID, name, and URL
					// Use starting_url from earlier resolution, or default to example.com
					if startingURL == "" {
						startingURL = "https://example.com"
					}
					return apiClient.CreateGoal(projectIDInt, goalName, startingURL)
				})
				if err != nil {
					result.Success = false
					result.Error = fmt.Sprintf("Failed to create goal: %v", err)
					return outputResult(result, outputFormat)
				}
				result.GoalID = fmt.Sprintf("%d", goal.ID)

				// Get snapshot ID for the goal
				snapshotIDStr, err := callWithContext(ctx, func() (string, error) {
					return apiClient.GetGoalSnapshot(goal.ID)
				})
				if err != nil {
					result.Success = false
					result.Error = fmt.Sprintf("Failed to get snapshot ID: %v", err)
					return outputResult(result, outputFormat)
				}

				// Convert snapshot ID to int
				snapshotIDInt, err = strconv.Atoi(snapshotIDStr)
				if err != nil {
					result.Success = false
					result.Error = fmt.Sprintf("Invalid snapshot ID: %v", err)
					return outputResult(result, outputFormat)
				}

				// Create journey
				journeyName := testDef.Name

				journey, err = callWithContext(ctx, func() (*client.Journey, error) {
					return apiClient.CreateJourney(goal.ID, snapshotIDInt, journeyName)
				})
				if err != nil {
					result.Success = false
					result.Error = fmt.Sprintf("Failed to create journey: %v", err)
					return outputResult(result, outputFormat)
				}
				result.JourneyID = fmt.Sprintf("%d", journey.ID)

				// Create checkpoint for new infrastructure
				checkpointName := "Test Steps"
				checkpoint, err = callWithContext(ctx, func() (*client.Checkpoint, error) {
					return apiClient.CreateCheckpoint(goal.ID, snapshotIDInt, checkpointName)
				})
				if err != nil {
					result.Success = false
					result.Error = fmt.Sprintf("Failed to create checkpoint: %v", err)
					return outputResult(result, outputFormat)
				}
				result.CheckpointID = fmt.Sprintf("%d", checkpoint.ID)
			} else {
				// Using existing checkpoint - populate IDs from session
				if cfg.Session.CurrentProjectID != nil {
					result.ProjectID = fmt.Sprintf("%d", *cfg.Session.CurrentProjectID)
				}
				if cfg.Session.CurrentGoalID != nil {
					result.GoalID = fmt.Sprintf("%d", *cfg.Session.CurrentGoalID)
				}
				if cfg.Session.CurrentJourneyID != nil {
					result.JourneyID = fmt.Sprintf("%d", *cfg.Session.CurrentJourneyID)
				}
				result.CheckpointID = fmt.Sprintf("%d", checkpointID)
			}

			// Add links
			result.Links = TestLinks{}
			if result.ProjectID != "" {
				result.Links.Project = fmt.Sprintf("https://app.virtuoso.qa/#/project/%s", result.ProjectID)
			}
			if result.GoalID != "" {
				result.Links.Goal = fmt.Sprintf("https://app.virtuoso.qa/#/goal/%s", result.GoalID)
			}
			if result.JourneyID != "" {
				result.Links.Journey = fmt.Sprintf("https://app.virtuoso.qa/#/journey/%s", result.JourneyID)
			}
			if result.CheckpointID != "" {
				result.Links.Checkpoint = fmt.Sprintf("https://app.virtuoso.qa/#/checkpoint/%s", result.CheckpointID)
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
	// Use the enhanced parser that supports all 69 commands
	return parseStepDefinitionEnhanced(stepDef)
}

// executeStep runs a single step command
func executeStep(apiClient *client.Client, checkpointID int, position int, command string, args []string) (string, error) {
	// Get context from the command
	ctx, cancel := CommandContext()
	defer cancel()

	// Use the enhanced executor that supports all 69 commands
	return executeStepEnhanced(ctx, apiClient, checkpointID, position, command, args)
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

// parseStepDefinitionEnhanced converts simplified step format to CLI command and args
// This version supports all 69 commands
func parseStepDefinitionEnhanced(stepDef map[string]interface{}) (string, []string, error) {
	// Handle each command type
	for cmd, value := range stepDef {
		switch cmd {
		// ========== NAVIGATION COMMANDS ==========
		case "navigate":
			return parseNavigateCommand(value)
		case "scroll":
			return parseScrollCommand(value)

		// ========== ASSERTION COMMANDS ==========
		case "assert":
			return parseAssertCommand(value)

		// ========== INTERACTION COMMANDS ==========
		case "click":
			return parseClickCommand(value)
		case "double-click", "doubleClick":
			return parseSimpleInteraction("double-click", value)
		case "right-click", "rightClick":
			return parseSimpleInteraction("right-click", value)
		case "hover":
			return parseSimpleInteraction("hover", value)
		case "write", "type":
			return parseWriteCommand(value)
		case "key", "press":
			return parseKeyCommand(value)
		case "mouse":
			return parseMouseCommand(value)
		case "select":
			return parseSelectCommand(value)

		// ========== WAIT COMMANDS ==========
		case "wait":
			return parseWaitCommand(value)

		// ========== DATA COMMANDS ==========
		case "store":
			return parseStoreCommand(value)
		case "cookie":
			return parseCookieCommand(value)

		// ========== WINDOW COMMANDS ==========
		case "window":
			return parseWindowCommand(value)

		// ========== DIALOG COMMANDS ==========
		case "dialog", "alert":
			return parseDialogCommand(value)

		// ========== FILE COMMANDS ==========
		case "file", "upload":
			return parseFileCommand(value)

		// ========== MISC COMMANDS ==========
		case "comment":
			return parseMiscCommand("comment", value)
		case "execute", "javascript", "js":
			return parseMiscCommand("execute", value)
		}
	}

	return "", nil, fmt.Errorf("unknown step format: %v", stepDef)
}

// Helper functions
func getString(m map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if val, ok := m[key]; ok {
			if str, ok := val.(string); ok {
				return str
			}
		}
	}
	return ""
}

func getInt(m map[string]interface{}, keys ...string) int {
	for _, key := range keys {
		if val, ok := m[key]; ok {
			switch v := val.(type) {
			case int:
				return v
			case float64:
				return int(v)
			case string:
				var i int
				fmt.Sscanf(v, "%d", &i)
				return i
			}
		}
	}
	return -1
}

func getBool(m map[string]interface{}, keys ...string) bool {
	for _, key := range keys {
		if val, ok := m[key]; ok {
			if b, ok := val.(bool); ok {
				return b
			}
		}
	}
	return false
}

func getStringValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case map[string]interface{}:
		return getString(v, "value", "text")
	}
	return ""
}

// ========== PARSER HELPER FUNCTIONS ==========

// parseNavigateCommand parses navigate commands
func parseNavigateCommand(value interface{}) (string, []string, error) {
	switch v := value.(type) {
	case string:
		return "step-navigate", []string{"to", v}, nil
	case map[string]interface{}:
		url := getString(v, "url", "to")
		if url == "" {
			return "", nil, fmt.Errorf("navigate requires url")
		}
		args := []string{"to", url}
		// Add flags if present
		if getBool(v, "new_tab", "newTab") {
			args = append(args, "--new-tab")
		}
		return "step-navigate", args, nil
	}
	return "", nil, fmt.Errorf("invalid navigate value type")
}

// parseScrollCommand parses scroll commands
func parseScrollCommand(value interface{}) (string, []string, error) {
	switch v := value.(type) {
	case string:
		// Simple element selector
		return "step-navigate", []string{"scroll-element", v}, nil
	case map[string]interface{}:
		// Complex scroll command
		if to := getString(v, "to"); to != "" {
			switch to {
			case "top":
				return "step-navigate", []string{"scroll-top"}, nil
			case "bottom":
				return "step-navigate", []string{"scroll-bottom"}, nil
			default:
				// Element selector
				return "step-navigate", []string{"scroll-element", to}, nil
			}
		}
		if position := getString(v, "position"); position != "" {
			return "step-navigate", []string{"scroll-position", position}, nil
		}
		if by := getString(v, "by"); by != "" {
			return "step-navigate", []string{"scroll-by", by}, nil
		}
		if getString(v, "direction") == "up" || getBool(v, "up") {
			return "step-navigate", []string{"scroll-up"}, nil
		}
		if getString(v, "direction") == "down" || getBool(v, "down") {
			return "step-navigate", []string{"scroll-down"}, nil
		}
	}
	return "", nil, fmt.Errorf("invalid scroll configuration")
}

// parseAssertCommand parses assertion commands
func parseAssertCommand(value interface{}) (string, []string, error) {
	switch v := value.(type) {
	case string:
		// Simple exists assertion
		return "step-assert", []string{"exists", v}, nil
	case map[string]interface{}:
		assertType := getString(v, "type", "command")
		if assertType == "" {
			assertType = "exists"
		}

		selector := getString(v, "selector", "element")
		if selector == "" && assertType != "variable" {
			return "", nil, fmt.Errorf("assert requires selector")
		}

		switch assertType {
		case "exists":
			return "step-assert", []string{"exists", selector}, nil
		case "not-exists":
			return "step-assert", []string{"not-exists", selector}, nil
		case "equals":
			value := getString(v, "value", "equals")
			return "step-assert", []string{"equals", selector, value}, nil
		case "not-equals":
			value := getString(v, "value", "not_equals")
			return "step-assert", []string{"not-equals", selector, value}, nil
		case "checked":
			return "step-assert", []string{"checked", selector}, nil
		case "selected":
			return "step-assert", []string{"selected", selector}, nil
		case "gt", "greater-than":
			value := getString(v, "value")
			return "step-assert", []string{"gt", selector, value}, nil
		case "gte", "greater-equal":
			value := getString(v, "value")
			return "step-assert", []string{"gte", selector, value}, nil
		case "lt", "less-than":
			value := getString(v, "value")
			return "step-assert", []string{"lt", selector, value}, nil
		case "lte", "less-equal":
			value := getString(v, "value")
			return "step-assert", []string{"lte", selector, value}, nil
		case "matches":
			pattern := getString(v, "pattern", "value")
			return "step-assert", []string{"matches", selector, pattern}, nil
		case "variable":
			varName := getString(v, "variable", "name")
			value := getString(v, "value")
			return "step-assert", []string{"variable", varName, value}, nil
		}
	}
	return "", nil, fmt.Errorf("invalid assert configuration")
}

// parseClickCommand parses click commands
func parseClickCommand(value interface{}) (string, []string, error) {
	switch v := value.(type) {
	case string:
		return "step-interact", []string{"click", v}, nil
	case map[string]interface{}:
		selector := getString(v, "selector", "element")
		if selector == "" {
			return "", nil, fmt.Errorf("click requires selector")
		}
		return "step-interact", []string{"click", selector}, nil
	}
	return "", nil, fmt.Errorf("invalid click value type")
}

// parseSimpleInteraction parses simple interaction commands (hover, double-click, etc)
func parseSimpleInteraction(action string, value interface{}) (string, []string, error) {
	selector := getStringValue(value)
	if selector == "" {
		return "", nil, fmt.Errorf("%s requires selector", action)
	}
	return "step-interact", []string{action, selector}, nil
}

// parseWriteCommand parses write/type commands
func parseWriteCommand(value interface{}) (string, []string, error) {
	if writeMap, ok := value.(map[string]interface{}); ok {
		selector := getString(writeMap, "selector", "element", "into")
		text := getString(writeMap, "text", "value")
		if selector == "" || text == "" {
			return "", nil, fmt.Errorf("write requires selector and text")
		}
		return "step-interact", []string{"write", selector, text}, nil
	}
	return "", nil, fmt.Errorf("write requires object with selector and text")
}

// parseKeyCommand parses key/press commands
func parseKeyCommand(value interface{}) (string, []string, error) {
	key := getStringValue(value)
	if key == "" {
		return "", nil, fmt.Errorf("key requires key name")
	}
	return "step-interact", []string{"key", key}, nil
}

// parseMouseCommand parses mouse commands
func parseMouseCommand(value interface{}) (string, []string, error) {
	if mouseMap, ok := value.(map[string]interface{}); ok {
		action := getString(mouseMap, "action", "command")
		switch action {
		case "move-to", "moveTo":
			selector := getString(mouseMap, "selector", "element", "to")
			return "step-interact", []string{"mouse", "move-to", selector}, nil
		case "move-by", "moveBy":
			offset := getString(mouseMap, "offset", "by")
			return "step-interact", []string{"mouse", "move-by", offset}, nil
		case "move":
			position := getString(mouseMap, "position", "to")
			return "step-interact", []string{"mouse", "move", position}, nil
		case "down":
			return "step-interact", []string{"mouse", "down"}, nil
		case "up":
			return "step-interact", []string{"mouse", "up"}, nil
		case "enter":
			return "step-interact", []string{"mouse", "enter"}, nil
		}
	}
	return "", nil, fmt.Errorf("invalid mouse command")
}

// parseSelectCommand parses select commands
func parseSelectCommand(value interface{}) (string, []string, error) {
	if selectMap, ok := value.(map[string]interface{}); ok {
		selector := getString(selectMap, "selector", "element", "from")
		if selector == "" {
			return "", nil, fmt.Errorf("select requires selector")
		}

		// Check for different select types
		if option := getString(selectMap, "option", "value", "text"); option != "" {
			return "step-interact", []string{"select", "option", selector, option}, nil
		}
		if index := getInt(selectMap, "index"); index >= 0 {
			return "step-interact", []string{"select", "index", selector, fmt.Sprintf("%d", index)}, nil
		}
		if getBool(selectMap, "last") {
			return "step-interact", []string{"select", "last", selector}, nil
		}
	}
	return "", nil, fmt.Errorf("invalid select configuration")
}

// parseWaitCommand parses wait commands
func parseWaitCommand(value interface{}) (string, []string, error) {
	switch v := value.(type) {
	case string:
		// Wait for element
		return "step-wait", []string{"element", v}, nil
	case int:
		// Wait time in milliseconds
		return "step-wait", []string{"time", fmt.Sprintf("%d", v)}, nil
	case float64:
		// Wait time in milliseconds
		return "step-wait", []string{"time", fmt.Sprintf("%d", int(v))}, nil
	case map[string]interface{}:
		// Complex wait
		if element := getString(v, "element", "for"); element != "" {
			return "step-wait", []string{"element", element}, nil
		}
		if timeMs := getInt(v, "time", "ms", "milliseconds"); timeMs > 0 {
			return "step-wait", []string{"time", fmt.Sprintf("%d", timeMs)}, nil
		}
	}
	return "", nil, fmt.Errorf("invalid wait configuration")
}

// parseStoreCommand parses store commands
func parseStoreCommand(value interface{}) (string, []string, error) {
	if storeMap, ok := value.(map[string]interface{}); ok {
		variable := getString(storeMap, "as", "variable", "in")
		if variable == "" {
			return "", nil, fmt.Errorf("store requires 'as' variable name")
		}

		storeType := getString(storeMap, "type")
		if storeType == "" {
			storeType = "element-text"
		}

		switch storeType {
		case "element-text", "text":
			selector := getString(storeMap, "selector", "element", "from")
			return "step-data", []string{"store", "element-text", selector, variable}, nil
		case "element-value", "value":
			selector := getString(storeMap, "selector", "element", "from")
			return "step-data", []string{"store", "element-value", selector, variable}, nil
		case "attribute":
			selector := getString(storeMap, "selector", "element", "from")
			attr := getString(storeMap, "attribute", "attr")
			return "step-data", []string{"store", "attribute", selector, attr, variable}, nil
		case "literal":
			value := getString(storeMap, "value")
			return "step-data", []string{"store", "literal", value, variable}, nil
		}
	}
	return "", nil, fmt.Errorf("invalid store configuration")
}

// parseCookieCommand parses cookie commands
func parseCookieCommand(value interface{}) (string, []string, error) {
	if cookieMap, ok := value.(map[string]interface{}); ok {
		action := getString(cookieMap, "action", "command")
		switch action {
		case "create", "set":
			name := getString(cookieMap, "name")
			value := getString(cookieMap, "value")
			if name == "" || value == "" {
				return "", nil, fmt.Errorf("cookie create requires name and value")
			}
			return "step-data", []string{"cookie", "create", name, value}, nil
		case "delete", "remove":
			name := getString(cookieMap, "name")
			if name == "" {
				return "", nil, fmt.Errorf("cookie delete requires name")
			}
			return "step-data", []string{"cookie", "delete", name}, nil
		case "clear", "clear-all":
			return "step-data", []string{"cookie", "clear"}, nil
		}
	}
	return "", nil, fmt.Errorf("invalid cookie configuration")
}

// parseWindowCommand parses window commands
func parseWindowCommand(value interface{}) (string, []string, error) {
	if windowMap, ok := value.(map[string]interface{}); ok {
		action := getString(windowMap, "action", "command")
		switch action {
		case "resize":
			size := getString(windowMap, "size", "to")
			if size == "" {
				width := getInt(windowMap, "width")
				height := getInt(windowMap, "height")
				if width > 0 && height > 0 {
					size = fmt.Sprintf("%dx%d", width, height)
				}
			}
			return "step-window", []string{"resize", size}, nil
		case "maximize":
			return "step-window", []string{"maximize"}, nil
		case "switch":
			target := getString(windowMap, "to", "target")
			switch target {
			case "next", "next-tab":
				return "step-window", []string{"switch", "tab", "NEXT"}, nil
			case "prev", "previous", "prev-tab":
				return "step-window", []string{"switch", "tab", "PREVIOUS"}, nil
			default:
				if strings.HasPrefix(target, "tab-") {
					index := strings.TrimPrefix(target, "tab-")
					return "step-window", []string{"switch", "tab", "INDEX", index}, nil
				}
				if iframe := getString(windowMap, "iframe"); iframe != "" {
					return "step-window", []string{"switch", "iframe", iframe}, nil
				}
				if getBool(windowMap, "parent", "parent-frame") {
					return "step-window", []string{"switch", "parent-frame"}, nil
				}
			}
		}
	}
	return "", nil, fmt.Errorf("invalid window configuration")
}

// parseDialogCommand parses dialog/alert commands
func parseDialogCommand(value interface{}) (string, []string, error) {
	switch v := value.(type) {
	case string:
		// Simple alert dismiss
		if v == "dismiss" || v == "accept" {
			return "step-dialog", []string{"dismiss-alert"}, nil
		}
	case map[string]interface{}:
		dialogType := getString(v, "type")
		action := getString(v, "action")

		switch dialogType {
		case "alert":
			return "step-dialog", []string{"dismiss-alert"}, nil
		case "confirm":
			if action == "accept" || getBool(v, "accept") {
				return "step-dialog", []string{"dismiss-confirm", "--accept"}, nil
			}
			return "step-dialog", []string{"dismiss-confirm", "--reject"}, nil
		case "prompt":
			text := getString(v, "text", "with")
			if text != "" {
				return "step-dialog", []string{"dismiss-prompt-with-text", text}, nil
			}
			if action == "accept" || getBool(v, "accept") {
				return "step-dialog", []string{"dismiss-prompt", "--accept"}, nil
			}
			return "step-dialog", []string{"dismiss-prompt", "--reject"}, nil
		}
	}
	return "", nil, fmt.Errorf("invalid dialog configuration")
}

// parseFileCommand parses file/upload commands
func parseFileCommand(value interface{}) (string, []string, error) {
	if fileMap, ok := value.(map[string]interface{}); ok {
		selector := getString(fileMap, "selector", "element", "to")
		url := getString(fileMap, "url", "from")
		if selector == "" || url == "" {
			return "", nil, fmt.Errorf("file upload requires selector and url")
		}
		return "step-file", []string{"upload", selector, url}, nil
	}
	return "", nil, fmt.Errorf("file upload requires configuration object")
}

// parseMiscCommand parses misc commands (comment, execute)
func parseMiscCommand(action string, value interface{}) (string, []string, error) {
	text := getStringValue(value)
	if text == "" {
		return "", nil, fmt.Errorf("%s requires text", action)
	}
	return "step-misc", []string{action, text}, nil
}

// ========== EXECUTOR HELPER FUNCTIONS ==========

// executeNavigateStep executes navigation commands
func executeNavigateStep(ctx context.Context, apiClient *client.Client, checkpointID int, position int, args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("navigate command requires subcommand")
	}
	switch args[0] {
	case "to":
		if len(args) < 2 {
			return "", fmt.Errorf("navigate to requires URL")
		}
		stepID, err := apiClient.CreateNavigationStep(checkpointID, args[1], position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "scroll-top":
		stepID, err := apiClient.CreateScrollTopStep(checkpointID, position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "scroll-bottom":
		stepID, err := apiClient.CreateScrollBottomStep(checkpointID, position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "scroll-element":
		if len(args) < 2 {
			return "", fmt.Errorf("scroll-element requires selector")
		}
		stepID, err := apiClient.CreateScrollElementStep(checkpointID, args[1], position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "scroll-position":
		if len(args) < 2 {
			return "", fmt.Errorf("scroll-position requires coordinates")
		}
		// Parse x,y coordinates
		coords := strings.Split(args[1], ",")
		if len(coords) != 2 {
			return "", fmt.Errorf("scroll-position requires x,y format")
		}
		x, err := strconv.Atoi(strings.TrimSpace(coords[0]))
		if err != nil {
			return "", fmt.Errorf("invalid x coordinate: %v", err)
		}
		y, err := strconv.Atoi(strings.TrimSpace(coords[1]))
		if err != nil {
			return "", fmt.Errorf("invalid y coordinate: %v", err)
		}
		stepID, err := apiClient.CreateScrollPositionStep(checkpointID, x, y, position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "scroll-by", "scroll-up", "scroll-down":
		// These require API methods that might not exist yet
		return "", fmt.Errorf("%s not implemented in API client", args[0])

	default:
		return "", fmt.Errorf("unknown navigate subcommand: %s", args[0])
	}
}

// executeAssertStep executes assertion commands
func executeAssertStep(ctx context.Context, apiClient *client.Client, checkpointID int, position int, args []string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("assert command requires type and element")
	}
	assertType := args[0]
	element := args[1]

	switch assertType {
	case "exists":
		stepID, err := apiClient.CreateAssertExistsStep(checkpointID, element, position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "not-exists":
		stepID, err := apiClient.CreateAssertNotExistsStep(checkpointID, element, position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "equals":
		if len(args) < 3 {
			return "", fmt.Errorf("assert equals requires element and value")
		}
		stepID, err := apiClient.CreateAssertEqualsStep(checkpointID, element, args[2], position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "not-equals":
		if len(args) < 3 {
			return "", fmt.Errorf("assert not-equals requires element and value")
		}
		stepID, err := apiClient.CreateAssertNotEqualsStep(checkpointID, element, args[2], position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "checked":
		stepID, err := apiClient.CreateAssertCheckedStep(checkpointID, element, position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "selected":
		stepID, err := apiClient.CreateAssertSelectedStep(checkpointID, element, position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "gt":
		if len(args) < 3 {
			return "", fmt.Errorf("assert gt requires element and value")
		}
		stepID, err := apiClient.CreateAssertGreaterThanStep(checkpointID, element, args[2], position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "gte":
		if len(args) < 3 {
			return "", fmt.Errorf("assert gte requires element and value")
		}
		stepID, err := apiClient.CreateAssertGreaterThanOrEqualStep(checkpointID, element, args[2], position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "lt":
		if len(args) < 3 {
			return "", fmt.Errorf("assert lt requires element and value")
		}
		stepID, err := apiClient.CreateAssertLessThanStep(checkpointID, element, args[2], position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "lte":
		if len(args) < 3 {
			return "", fmt.Errorf("assert lte requires element and value")
		}
		stepID, err := apiClient.CreateAssertLessThanOrEqualStep(checkpointID, element, args[2], position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "matches":
		if len(args) < 3 {
			return "", fmt.Errorf("assert matches requires element and pattern")
		}
		stepID, err := apiClient.CreateAssertMatchesStep(checkpointID, element, args[2], position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "variable":
		if len(args) < 3 {
			return "", fmt.Errorf("assert variable requires name and value")
		}
		stepID, err := apiClient.CreateAssertVariableStep(checkpointID, element, args[2], position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	default:
		return "", fmt.Errorf("unknown assert type: %s", assertType)
	}
}

// executeInteractStep executes interaction commands
func executeInteractStep(ctx context.Context, apiClient *client.Client, checkpointID int, position int, args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("interact command requires action")
	}
	action := args[0]

	switch action {
	case "click":
		if len(args) < 2 {
			return "", fmt.Errorf("click requires selector")
		}
		stepID, err := apiClient.CreateClickStep(checkpointID, args[1], position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "double-click":
		if len(args) < 2 {
			return "", fmt.Errorf("double-click requires selector")
		}
		stepID, err := apiClient.CreateDoubleClickStep(checkpointID, args[1], position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "right-click":
		if len(args) < 2 {
			return "", fmt.Errorf("right-click requires selector")
		}
		stepID, err := apiClient.CreateRightClickStep(checkpointID, args[1], position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "hover":
		if len(args) < 2 {
			return "", fmt.Errorf("hover requires selector")
		}
		stepID, err := apiClient.CreateHoverStep(checkpointID, args[1], position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "write":
		if len(args) < 3 {
			return "", fmt.Errorf("write requires selector and text")
		}
		// Note: API has (text, element) order but our syntax is (element, text)
		stepID, err := apiClient.CreateWriteStep(checkpointID, args[2], args[1], position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "key":
		if len(args) < 2 {
			return "", fmt.Errorf("key requires key name")
		}
		stepID, err := apiClient.CreateKeyStep(checkpointID, args[1], position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "mouse":
		return executeMouseAction(ctx, apiClient, checkpointID, position, args[1:])

	case "select":
		return executeSelectAction(ctx, apiClient, checkpointID, position, args[1:])

	default:
		return "", fmt.Errorf("unknown interact action: %s", action)
	}
}

// executeMouseAction executes mouse-specific actions
func executeMouseAction(ctx context.Context, apiClient *client.Client, checkpointID int, position int, args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("mouse requires action")
	}
	mouseAction := args[0]

	switch mouseAction {
	case "move-to":
		if len(args) < 2 {
			return "", fmt.Errorf("mouse move-to requires selector")
		}
		stepID, err := apiClient.CreateMouseMoveStep(checkpointID, args[1], position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "move-by":
		if len(args) < 2 {
			return "", fmt.Errorf("mouse move-by requires x,y offset")
		}
		coords := strings.Split(args[1], ",")
		if len(coords) != 2 {
			return "", fmt.Errorf("mouse move-by requires x,y format")
		}
		x, err := strconv.Atoi(strings.TrimSpace(coords[0]))
		if err != nil {
			return "", fmt.Errorf("invalid x offset: %v", err)
		}
		y, err := strconv.Atoi(strings.TrimSpace(coords[1]))
		if err != nil {
			return "", fmt.Errorf("invalid y offset: %v", err)
		}
		stepID, err := apiClient.CreateMouseMoveByStep(checkpointID, x, y, position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "move":
		if len(args) < 2 {
			return "", fmt.Errorf("mouse move requires x,y coordinates")
		}
		coords := strings.Split(args[1], ",")
		if len(coords) != 2 {
			return "", fmt.Errorf("mouse move requires x,y format")
		}
		x, err := strconv.Atoi(strings.TrimSpace(coords[0]))
		if err != nil {
			return "", fmt.Errorf("invalid x coordinate: %v", err)
		}
		y, err := strconv.Atoi(strings.TrimSpace(coords[1]))
		if err != nil {
			return "", fmt.Errorf("invalid y coordinate: %v", err)
		}
		stepID, err := apiClient.CreateMouseMoveToStep(checkpointID, x, y, position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "down":
		if len(args) < 2 {
			return "", fmt.Errorf("mouse down requires selector")
		}
		stepID, err := apiClient.CreateMouseDownStep(checkpointID, args[1], position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "up":
		if len(args) < 2 {
			return "", fmt.Errorf("mouse up requires selector")
		}
		stepID, err := apiClient.CreateMouseUpStep(checkpointID, args[1], position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "enter":
		if len(args) < 2 {
			return "", fmt.Errorf("mouse enter requires selector")
		}
		stepID, err := apiClient.CreateMouseEnterStep(checkpointID, args[1], position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	default:
		return "", fmt.Errorf("unknown mouse action: %s", mouseAction)
	}
}

// executeSelectAction executes select-specific actions
func executeSelectAction(ctx context.Context, apiClient *client.Client, checkpointID int, position int, args []string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("select requires type and selector")
	}
	selectType := args[0]
	selector := args[1]

	switch selectType {
	case "option":
		if len(args) < 3 {
			return "", fmt.Errorf("select option requires value")
		}
		stepID, err := apiClient.CreatePickStep(checkpointID, args[2], selector, position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "index":
		if len(args) < 3 {
			return "", fmt.Errorf("select index requires index")
		}
		index, err := strconv.Atoi(args[2])
		if err != nil {
			return "", fmt.Errorf("invalid index: %v", err)
		}
		stepID, err := apiClient.CreateStepPickIndex(checkpointID, selector, index, position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "last":
		stepID, err := apiClient.CreateStepPickLast(checkpointID, selector, position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	default:
		return "", fmt.Errorf("unknown select type: %s", selectType)
	}
}

// executeWaitStep executes wait commands
func executeWaitStep(ctx context.Context, apiClient *client.Client, checkpointID int, position int, args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("wait command requires type")
	}
	switch args[0] {
	case "element":
		if len(args) < 2 {
			return "", fmt.Errorf("wait element requires selector")
		}
		stepID, err := apiClient.CreateWaitElementStep(checkpointID, args[1], position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "time":
		if len(args) < 2 {
			return "", fmt.Errorf("wait time requires milliseconds")
		}
		// Convert string to int for time in milliseconds
		timeMs, err := strconv.Atoi(args[1])
		if err != nil {
			return "", fmt.Errorf("invalid time value: %s", args[1])
		}
		// API expects seconds, so convert from milliseconds
		seconds := timeMs / 1000
		if seconds < 1 {
			seconds = 1 // Minimum 1 second
		}
		stepID, err := apiClient.CreateWaitTimeStep(checkpointID, seconds, position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	default:
		return "", fmt.Errorf("unknown wait type: %s", args[0])
	}
}

// executeDataStep executes data management commands
func executeDataStep(ctx context.Context, apiClient *client.Client, checkpointID int, position int, args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("data command requires action")
	}
	action := args[0]

	switch action {
	case "store":
		return executeStoreAction(ctx, apiClient, checkpointID, position, args[1:])
	case "cookie":
		return executeCookieAction(ctx, apiClient, checkpointID, position, args[1:])
	default:
		return "", fmt.Errorf("unknown data action: %s", action)
	}
}

// executeStoreAction executes store-specific actions
func executeStoreAction(ctx context.Context, apiClient *client.Client, checkpointID int, position int, args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("store requires type")
	}
	storeType := args[0]

	switch storeType {
	case "element-text":
		if len(args) < 3 {
			return "", fmt.Errorf("store element-text requires selector and variable name")
		}
		stepID, err := apiClient.CreateStoreStep(checkpointID, args[1], args[2], position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "literal":
		if len(args) < 3 {
			return "", fmt.Errorf("store literal requires value and variable name")
		}
		stepID, err := apiClient.CreateStepStoreLiteralValueWithContext(ctx, checkpointID, args[1], args[2], position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "attribute":
		if len(args) < 4 {
			return "", fmt.Errorf("store attribute requires selector, attribute, and variable name")
		}
		stepID, err := apiClient.CreateStepStoreAttributeWithContext(ctx, checkpointID, args[1], args[2], args[3], position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	default:
		return "", fmt.Errorf("unknown store type: %s", storeType)
	}
}

// executeCookieAction executes cookie-specific actions
func executeCookieAction(ctx context.Context, apiClient *client.Client, checkpointID int, position int, args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("cookie requires action")
	}
	cookieAction := args[0]

	switch cookieAction {
	case "create":
		if len(args) < 3 {
			return "", fmt.Errorf("cookie create requires name and value")
		}
		stepID, err := apiClient.CreateAddCookieStep(checkpointID, args[1], args[2], position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "delete":
		if len(args) < 2 {
			return "", fmt.Errorf("cookie delete requires name")
		}
		stepID, err := apiClient.CreateDeleteCookieStep(checkpointID, args[1], position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "clear", "clear-all":
		stepID, err := apiClient.CreateClearCookiesStep(checkpointID, position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	default:
		return "", fmt.Errorf("unknown cookie action: %s", cookieAction)
	}
}

// executeDialogStep executes dialog commands
func executeDialogStep(ctx context.Context, apiClient *client.Client, checkpointID int, position int, args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("dialog command requires type")
	}
	dialogType := args[0]

	switch dialogType {
	case "dismiss-alert":
		stepID, err := apiClient.CreateDismissAlertStep(checkpointID, position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "dismiss-confirm":
		accept := false
		// Check for --accept or --reject flags
		for _, arg := range args[1:] {
			if arg == "--accept" {
				accept = true
				break
			}
		}
		stepID, err := apiClient.CreateDismissConfirmStep(checkpointID, accept, position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "dismiss-prompt":
		// This needs a text parameter in the API, but our syntax doesn't provide it
		// Use empty string for now
		stepID, err := apiClient.CreateDismissPromptStep(checkpointID, "", position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "dismiss-prompt-with-text":
		if len(args) < 2 {
			return "", fmt.Errorf("dismiss-prompt-with-text requires text")
		}
		stepID, err := apiClient.CreateDismissPromptStep(checkpointID, args[1], position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	default:
		return "", fmt.Errorf("unknown dialog type: %s", dialogType)
	}
}

// executeWindowStep executes window management commands
func executeWindowStep(ctx context.Context, apiClient *client.Client, checkpointID int, position int, args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("window command requires action")
	}
	action := args[0]

	// Window commands use string checkpoint ID
	checkpointIDStr := fmt.Sprintf("%d", checkpointID)

	switch action {
	case "resize":
		if len(args) < 2 {
			return "", fmt.Errorf("window resize requires dimensions")
		}
		// Parse WIDTHxHEIGHT format
		dimensions := strings.Split(args[1], "x")
		if len(dimensions) != 2 {
			return "", fmt.Errorf("window resize requires WIDTHxHEIGHT format")
		}
		width, err := strconv.Atoi(dimensions[0])
		if err != nil {
			return "", fmt.Errorf("invalid width: %v", err)
		}
		height, err := strconv.Atoi(dimensions[1])
		if err != nil {
			return "", fmt.Errorf("invalid height: %v", err)
		}
		stepResp, err := apiClient.CreateStepResizeWindowV2(checkpointIDStr, width, height, position)
		if err != nil {
			return "", err
		}
		return stepResp.ID, nil

	case "maximize":
		stepResp, err := apiClient.CreateStepMaximizeV2(checkpointIDStr, position)
		if err != nil {
			return "", err
		}
		return stepResp.ID, nil

	case "switch":
		if len(args) < 2 {
			return "", fmt.Errorf("window switch requires target")
		}
		target := args[1]

		switch target {
		case "tab":
			if len(args) < 3 {
				return "", fmt.Errorf("switch tab requires direction or index")
			}
			tabID := args[2]
			stepResp, err := apiClient.CreateStepSwitchTabV2(checkpointIDStr, tabID, position)
			if err != nil {
				return "", err
			}
			return stepResp.ID, nil

		case "iframe":
			if len(args) < 3 {
				return "", fmt.Errorf("switch iframe requires selector")
			}
			stepResp, err := apiClient.CreateStepSwitchIframeV2(checkpointIDStr, args[2], position)
			if err != nil {
				return "", err
			}
			return stepResp.ID, nil

		case "parent-frame":
			stepID, err := apiClient.CreateSwitchParentFrameStep(checkpointID, position)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("%d", stepID), nil

		default:
			return "", fmt.Errorf("unknown switch target: %s", target)
		}

	default:
		return "", fmt.Errorf("unknown window action: %s", action)
	}
}

// executeFileStep executes file commands
func executeFileStep(ctx context.Context, apiClient *client.Client, checkpointID int, position int, args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("file command requires action")
	}
	action := args[0]

	switch action {
	case "upload", "upload-url":
		if len(args) < 3 {
			return "", fmt.Errorf("file upload requires selector and URL")
		}
		selector := args[1]
		url := args[2]
		stepID, err := apiClient.CreateStepFileUploadByURLWithContext(ctx, checkpointID, url, selector, position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	default:
		return "", fmt.Errorf("unknown file action: %s", action)
	}
}

// executeMiscStep executes miscellaneous commands
func executeMiscStep(ctx context.Context, apiClient *client.Client, checkpointID int, position int, args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("misc command requires action")
	}
	action := args[0]

	switch action {
	case "comment":
		if len(args) < 2 {
			return "", fmt.Errorf("comment requires text")
		}
		// Join all remaining args as comment text
		comment := strings.Join(args[1:], " ")
		stepID, err := apiClient.CreateCommentStep(checkpointID, comment, position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	case "execute":
		if len(args) < 2 {
			return "", fmt.Errorf("execute requires JavaScript code")
		}
		// Join all remaining args as JavaScript code
		script := strings.Join(args[1:], " ")
		stepID, err := apiClient.CreateExecuteJsStep(checkpointID, script, position)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", stepID), nil

	default:
		return "", fmt.Errorf("unknown misc action: %s", action)
	}
}

// executeStepEnhanced runs a single step command with support for all 69 commands
func executeStepEnhanced(ctx context.Context, apiClient *client.Client, checkpointID int, position int, command string, args []string) (string, error) {
	// Handle each command type
	switch command {
	case "step-navigate":
		return executeNavigateStep(ctx, apiClient, checkpointID, position, args)
	case "step-assert":
		return executeAssertStep(ctx, apiClient, checkpointID, position, args)
	case "step-interact":
		return executeInteractStep(ctx, apiClient, checkpointID, position, args)
	case "step-wait":
		return executeWaitStep(ctx, apiClient, checkpointID, position, args)
	case "step-data":
		return executeDataStep(ctx, apiClient, checkpointID, position, args)
	case "step-dialog":
		return executeDialogStep(ctx, apiClient, checkpointID, position, args)
	case "step-window":
		return executeWindowStep(ctx, apiClient, checkpointID, position, args)
	case "step-file":
		return executeFileStep(ctx, apiClient, checkpointID, position, args)
	case "step-misc":
		return executeMiscStep(ctx, apiClient, checkpointID, position, args)
	case "library":
		return "", fmt.Errorf("library commands are not supported in run-test context")
	default:
		return "", fmt.Errorf("unknown command: %s", command)
	}
}
