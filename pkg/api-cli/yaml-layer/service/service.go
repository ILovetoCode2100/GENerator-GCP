package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/marklovelady/api-cli-generator/pkg/api-cli/yaml-layer/compiler"
	"github.com/marklovelady/api-cli-generator/pkg/api-cli/yaml-layer/core"
	"github.com/marklovelady/api-cli-generator/pkg/api-cli/yaml-layer/validation"
	"gopkg.in/yaml.v3"
)

// Service handles the complete YAML test pipeline
type Service struct {
	config    *Config
	validator *validation.Validator
	compiler  *compiler.Compiler
	apiClient *client.Client
	metrics   *Metrics
}

// Config holds service configuration
type Config struct {
	API        APIConfig
	Validation ValidationConfig
	Execution  ExecutionConfig
}

// APIConfig holds API configuration
type APIConfig struct {
	BaseURL   string
	AuthToken string
	OrgID     string
}

// ValidationConfig holds validation settings
type ValidationConfig struct {
	Strict        bool
	MaxErrors     int
	SpellCheck    bool
	BestPractices bool
}

// ExecutionConfig holds execution settings
type ExecutionConfig struct {
	ScreenshotOnFailure bool
	AutoWait            bool
	Parallel            int
	Timeout             time.Duration
}

// Metrics tracks performance
type Metrics struct {
	ParseTime    time.Duration
	ValidateTime time.Duration
	CompileTime  time.Duration
	ExecuteTime  time.Duration
	StepTimes    map[string]time.Duration
}

// NewService creates a new service instance
func NewService(config *Config) *Service {
	return &Service{
		config:    config,
		validator: validation.NewValidator(),
		compiler:  compiler.NewCompiler(config.API.BaseURL),
		metrics:   &Metrics{StepTimes: make(map[string]time.Duration)},
	}
}

// ProcessResult holds the result of processing a YAML test
type ProcessResult struct {
	Success     bool                   `json:"success"`
	TestName    string                 `json:"test_name"`
	Errors      []core.ValidationError `json:"errors,omitempty"`
	Warnings    []core.ValidationError `json:"warnings,omitempty"`
	Commands    []core.CompiledStep    `json:"commands,omitempty"`
	ExecutionID string                 `json:"execution_id,omitempty"`
	Duration    time.Duration          `json:"duration"`
	Metrics     *Metrics               `json:"metrics,omitempty"`
}

// ProcessYAML handles the complete pipeline
func (s *Service) ProcessYAML(ctx context.Context, reader io.Reader) (*ProcessResult, error) {
	startTime := time.Now()
	result := &ProcessResult{
		Success: false,
		Metrics: s.metrics,
	}

	// Read YAML content
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return result, fmt.Errorf("failed to read YAML: %w", err)
	}

	// Parse YAML
	parseStart := time.Now()
	var test core.YAMLTest
	if err := yaml.Unmarshal(content, &test); err != nil {
		result.Errors = append(result.Errors, core.ValidationError{
			Line:    1,
			Message: fmt.Sprintf("YAML parse error: %v", err),
			Fix:     "Check YAML syntax and formatting",
		})
		return result, nil
	}
	s.metrics.ParseTime = time.Since(parseStart)
	result.TestName = test.Test

	// Validate
	validateStart := time.Now()
	valid, errors := s.validator.Validate(content)
	s.metrics.ValidateTime = time.Since(validateStart)

	if !valid {
		result.Errors = errors
		if s.config.Validation.Strict {
			return result, nil
		}
	}

	// Get warnings
	result.Warnings = s.validator.GetWarnings()

	// Compile
	compileStart := time.Now()
	compiled, err := s.compiler.Compile(&test)
	if err != nil {
		result.Errors = append(result.Errors, core.ValidationError{
			Message: fmt.Sprintf("Compilation error: %v", err),
		})
		return result, nil
	}
	s.metrics.CompileTime = time.Since(compileStart)
	result.Commands = compiled.Steps

	// Execute if API client is configured
	if s.apiClient != nil {
		execStart := time.Now()
		execID, err := s.execute(ctx, &test, compiled)
		s.metrics.ExecuteTime = time.Since(execStart)

		if err != nil {
			result.Errors = append(result.Errors, core.ValidationError{
				Message: fmt.Sprintf("Execution error: %v", err),
			})
			return result, nil
		}
		result.ExecutionID = execID
	}

	result.Success = len(result.Errors) == 0
	result.Duration = time.Since(startTime)
	return result, nil
}

// ValidateFile validates a YAML file
func (s *Service) ValidateFile(filePath string) (*ProcessResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Create a context for validation only
	ctx := context.Background()

	// Temporarily disable execution
	oldClient := s.apiClient
	s.apiClient = nil
	defer func() { s.apiClient = oldClient }()

	return s.ProcessYAML(ctx, file)
}

// CompileFile compiles a YAML file without executing
func (s *Service) CompileFile(filePath string) (*core.CompileResult, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var test core.YAMLTest
	if err := yaml.Unmarshal(content, &test); err != nil {
		return nil, fmt.Errorf("YAML parse error: %w", err)
	}

	return s.compiler.Compile(&test)
}

// RunFile runs a YAML test file
func (s *Service) RunFile(ctx context.Context, filePath string) (*ProcessResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return s.ProcessYAML(ctx, file)
}

// RunDirectory runs all YAML tests in a directory
func (s *Service) RunDirectory(ctx context.Context, dirPath string, pattern string) ([]*ProcessResult, error) {
	if pattern == "" {
		pattern = "*.yaml"
	}

	matches, err := filepath.Glob(filepath.Join(dirPath, pattern))
	if err != nil {
		return nil, err
	}

	// Also check for .yml files
	ymlMatches, _ := filepath.Glob(filepath.Join(dirPath, strings.Replace(pattern, ".yaml", ".yml", -1)))
	matches = append(matches, ymlMatches...)

	results := make([]*ProcessResult, 0, len(matches))

	for _, match := range matches {
		log.Printf("Processing: %s", match)
		result, err := s.RunFile(ctx, match)
		if err != nil {
			log.Printf("Error processing %s: %v", match, err)
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

// execute runs the compiled commands
func (s *Service) execute(ctx context.Context, test *core.YAMLTest, compiled *core.CompileResult) (string, error) {
	// Create test infrastructure if needed
	projectID, _, journeyID, checkpointID, err := s.ensureTestInfrastructure(ctx, test.Test)
	if err != nil {
		return "", fmt.Errorf("failed to create test infrastructure: %w", err)
	}

	// Execute each command
	for i, step := range compiled.Steps {
		stepStart := time.Now()

		// Add checkpoint ID to args if needed
		args := s.addCheckpointID(step.Args, checkpointID)

		// Execute based on command type
		err := s.executeCommand(ctx, step.Command, args, step.Options)
		if err != nil {
			if s.config.Execution.ScreenshotOnFailure {
				s.captureScreenshot(ctx, checkpointID, fmt.Sprintf("failure_step_%d", i))
			}
			return "", fmt.Errorf("step %d failed: %w", i+1, err)
		}

		// Track timing
		s.metrics.StepTimes[fmt.Sprintf("step_%d", i)] = time.Since(stepStart)

		// Apply auto-wait if configured
		if s.config.Execution.AutoWait && i < len(compiled.Steps)-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	// Start execution
	executionID, err := s.startExecution(ctx, projectID, journeyID)
	if err != nil {
		return "", fmt.Errorf("failed to start execution: %w", err)
	}

	return executionID, nil
}

// ensureTestInfrastructure creates or retrieves test infrastructure
func (s *Service) ensureTestInfrastructure(ctx context.Context, testName string) (string, string, string, string, error) {
	// For now, return dummy IDs
	// In real implementation, would create project/goal/journey/checkpoint
	projectID := "proj_" + generateID()
	goalID := "goal_" + generateID()
	journeyID := "journey_" + generateID()
	checkpointID := "cp_" + generateID()

	return projectID, goalID, journeyID, checkpointID, nil
}

// executeCommand executes a single CLI command
func (s *Service) executeCommand(ctx context.Context, command string, args []string, options map[string]interface{}) error {
	// Extract checkpoint ID from args (should be first arg for most commands)
	checkpointID := 0
	if len(args) > 0 {
		if id, ok := parseCheckpointID(args[0]); ok {
			checkpointID = id
			args = args[1:] // Remove checkpoint ID from args
		}
	}

	// Map command to API client method
	switch command {
	case "step-navigate":
		if len(args) >= 2 && args[0] == "to" {
			_, err := s.apiClient.CreateStepNavigateWithContext(ctx, checkpointID, args[1], false, 0)
			return err
		} else if len(args) >= 2 && args[0] == "scroll" {
			return s.executeScroll(ctx, checkpointID, args[1:])
		}

	case "step-interact":
		if len(args) >= 2 {
			switch args[0] {
			case "click":
				_, err := s.apiClient.CreateStepClickWithContext(ctx, checkpointID, args[1], 0)
				return err
			case "write":
				if len(args) >= 3 {
					_, err := s.apiClient.CreateStepWriteWithContext(ctx, checkpointID, args[1], args[2], 0)
					return err
				}
			case "hover":
				_, err := s.apiClient.CreateStepHoverWithContext(ctx, checkpointID, args[1], 0)
				return err
			case "key":
				_, err := s.apiClient.CreateStepKeyGlobalWithContext(ctx, checkpointID, args[1], 0)
				return err
			case "select":
				if len(args) >= 4 && args[1] == "option" {
					_, err := s.apiClient.CreateStepPick(checkpointID, args[2], args[3], 0)
					return err
				}
			case "mouse":
				return s.executeMouse(ctx, checkpointID, args[1:])
			}
		}

	case "step-assert":
		if len(args) >= 2 {
			switch args[0] {
			case "exists":
				_, err := s.apiClient.CreateAssertExistsStepWithContext(ctx, checkpointID, args[1], 0)
				return err
			case "not-exists":
				_, err := s.apiClient.CreateAssertNotExistsStepWithContext(ctx, checkpointID, args[1], 0)
				return err
			case "equals":
				if len(args) >= 3 {
					_, err := s.apiClient.CreateAssertEqualsStep(checkpointID, args[1], args[2], 0)
					return err
				}
			case "not-equals":
				if len(args) >= 3 {
					_, err := s.apiClient.CreateAssertNotEqualsStep(checkpointID, args[1], args[2], 0)
					return err
				}
			}
		}

	case "step-wait":
		if len(args) >= 2 {
			switch args[0] {
			case "element":
				_, err := s.apiClient.CreateStepWaitForElementWithContext(ctx, checkpointID, args[1], 0)
				return err
			case "time":
				if ms, err := parseTimeMillis(args[1]); err == nil {
					_, err := s.apiClient.CreateStepWaitTime(checkpointID, ms, 0)
					return err
				}
			}
		}

	case "step-data":
		if len(args) >= 4 && args[0] == "store" {
			_, err := s.apiClient.CreateStepStoreElementTextWithContext(ctx, checkpointID, args[2], args[3], 0)
			return err
		}

	case "step-misc":
		if len(args) >= 2 {
			switch args[0] {
			case "comment":
				_, err := s.apiClient.CreateStepComment(checkpointID, args[1], 0)
				return err
			case "execute":
				_, err := s.apiClient.CreateStepExecuteJs(checkpointID, args[1], 0)
				return err
			}
		}

	case "step-dialog":
		if len(args) >= 1 {
			return s.executeDialog(ctx, checkpointID, args)
		}

	case "step-window":
		if len(args) >= 1 {
			return s.executeWindow(ctx, checkpointID, args)
		}

	case "step-file":
		if len(args) >= 3 && args[0] == "upload" {
			_, err := s.apiClient.CreateStepFileUploadByURLWithContext(ctx, checkpointID, args[2], args[1], 0)
			return err
		}
	}

	return fmt.Errorf("unknown command: %s %v", command, args)
}

// executeScroll handles scroll commands
func (s *Service) executeScroll(ctx context.Context, checkpointID int, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("scroll requires arguments")
	}

	switch args[0] {
	case "top":
		_, err := s.apiClient.CreateStepScrollToTopWithContext(ctx, checkpointID, 0)
		return err
	case "bottom":
		_, err := s.apiClient.CreateStepScrollBottomWithContext(ctx, checkpointID, 0)
		return err
	case "element":
		if len(args) >= 2 {
			_, err := s.apiClient.CreateStepScrollElementWithContext(ctx, checkpointID, args[1], 0)
			return err
		}
	case "position":
		if len(args) >= 3 {
			x, y, err := parseCoordinates(args[1] + "," + args[2])
			if err != nil {
				return err
			}
			_, err = s.apiClient.CreateStepScrollToPositionWithContext(ctx, checkpointID, x, y, 0)
			return err
		}
	}

	return fmt.Errorf("invalid scroll command: %v", args)
}

// executeMouse handles mouse commands
func (s *Service) executeMouse(ctx context.Context, checkpointID int, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("mouse requires action")
	}

	switch args[0] {
	case "down":
		// Mouse down at current position
		_, err := s.apiClient.CreateStepMouseDown(checkpointID, "", 0)
		return err
	case "up":
		// Mouse up at current position
		_, err := s.apiClient.CreateStepMouseUp(checkpointID, "", 0)
		return err
	case "move-to":
		if len(args) >= 2 {
			_, err := s.apiClient.CreateStepMouseMove(checkpointID, args[1], 0)
			return err
		}
	case "move-by":
		if len(args) >= 2 {
			x, y, err := parseCoordinates(args[1])
			if err != nil {
				return err
			}
			_, err = s.apiClient.CreateStepMouseMoveBy(checkpointID, x, y, 0)
			return err
		}
	}

	return fmt.Errorf("invalid mouse command: %v", args)
}

// executeDialog handles dialog commands
func (s *Service) executeDialog(ctx context.Context, checkpointID int, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("dialog requires action")
	}

	switch args[0] {
	case "dismiss-alert":
		_, err := s.apiClient.CreateStepDismissAlert(checkpointID, 0)
		return err
	case "dismiss-confirm":
		accept := false
		for _, arg := range args[1:] {
			if arg == "--accept" {
				accept = true
				break
			}
		}
		if accept {
			_, err := s.apiClient.CreateStepConfirmAcceptWithContext(ctx, checkpointID, 0)
			return err
		} else {
			_, err := s.apiClient.CreateStepConfirmDismissWithContext(ctx, checkpointID, 0)
			return err
		}
	case "dismiss-prompt":
		accept := false
		for _, arg := range args[1:] {
			if arg == "--accept" {
				accept = true
				break
			}
		}
		// For prompt, accept means fill with empty text, reject means dismiss
		if accept {
			_, err := s.apiClient.CreateStepPromptDismissWithTextWithContext(ctx, checkpointID, "", 0)
			return err
		} else {
			_, err := s.apiClient.CreateStepPromptDismissWithContext(ctx, checkpointID, 0)
			return err
		}
	case "dismiss-prompt-with-text":
		if len(args) >= 2 {
			_, err := s.apiClient.CreateStepPromptDismissWithTextWithContext(ctx, checkpointID, args[1], 0)
			return err
		}
	}

	return fmt.Errorf("invalid dialog command: %v", args)
}

// executeWindow handles window commands
func (s *Service) executeWindow(ctx context.Context, checkpointID int, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("window requires action")
	}

	checkpointStr := fmt.Sprintf("%d", checkpointID)

	switch args[0] {
	case "maximize":
		_, err := s.apiClient.CreateStepWindowMaximize(checkpointID, 0)
		return err
	case "resize":
		if len(args) >= 2 {
			// Parse WIDTHxHEIGHT format
			parts := strings.Split(args[1], "x")
			if len(parts) != 2 {
				return fmt.Errorf("invalid resize format: %s (expected WIDTHxHEIGHT)", args[1])
			}
			width, err := strconv.Atoi(parts[0])
			if err != nil {
				return fmt.Errorf("invalid width: %s", parts[0])
			}
			height, err := strconv.Atoi(parts[1])
			if err != nil {
				return fmt.Errorf("invalid height: %s", parts[1])
			}
			_, err = s.apiClient.CreateStepResizeWindowWithContext(ctx, checkpointStr, width, height, 0)
			return err
		}
	case "switch":
		if len(args) >= 3 && args[1] == "tab" {
			action := "tab"
			value := args[2]
			_, err := s.apiClient.CreateStepWindowSwitchWithContext(ctx, checkpointStr, action, value, 0)
			return err
		} else if len(args) >= 3 && args[1] == "iframe" {
			_, err := s.apiClient.CreateSwitchIFrameStep(checkpointID, args[2], 0)
			return err
		} else if len(args) >= 2 && args[1] == "parent-frame" {
			_, err := s.apiClient.CreateStepSwitchParentFrame(checkpointID, 0)
			return err
		}
	}

	return fmt.Errorf("invalid window command: %v", args)
}

// addCheckpointID adds checkpoint ID to command args if needed
func (s *Service) addCheckpointID(args []string, checkpointID string) []string {
	// Check if checkpoint ID is already present
	for _, arg := range args {
		if strings.HasPrefix(arg, "cp_") {
			return args
		}
	}

	// Add checkpoint ID after command and subcommand
	if len(args) >= 2 {
		newArgs := make([]string, 0, len(args)+1)
		newArgs = append(newArgs, args[0], args[1], checkpointID)
		newArgs = append(newArgs, args[2:]...)
		return newArgs
	}

	return args
}

// captureScreenshot captures a screenshot for debugging
func (s *Service) captureScreenshot(ctx context.Context, checkpointID, name string) {
	// In real implementation, would capture screenshot via API
	log.Printf("Would capture screenshot: %s for checkpoint %s", name, checkpointID)
}

// startExecution starts test execution
func (s *Service) startExecution(ctx context.Context, projectID, journeyID string) (string, error) {
	// In real implementation, would start execution via API
	return "exec_" + generateID(), nil
}

// generateID generates a simple ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// parseCheckpointID extracts checkpoint ID from string
func parseCheckpointID(s string) (int, bool) {
	// Remove cp_ prefix if present
	s = strings.TrimPrefix(s, "cp_")
	
	var id int
	_, err := fmt.Sscanf(s, "%d", &id)
	return id, err == nil
}

// parseTimeMillis parses time in milliseconds from string
func parseTimeMillis(s string) (int, error) {
	var ms int
	_, err := fmt.Sscanf(s, "%d", &ms)
	if err != nil {
		return 0, fmt.Errorf("invalid time format: %s", s)
	}
	return ms, nil
}

// parseCoordinates parses x,y coordinates from string
func parseCoordinates(s string) (int, int, error) {
	parts := strings.Split(s, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid coordinates format: %s (expected x,y)", s)
	}
	
	var x, y int
	_, err := fmt.Sscanf(parts[0], "%d", &x)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid x coordinate: %s", parts[0])
	}
	
	_, err = fmt.Sscanf(parts[1], "%d", &y)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid y coordinate: %s", parts[1])
	}
	
	return x, y, nil
}

// SetAPIClient sets the API client for execution
func (s *Service) SetAPIClient(client *client.Client) {
	s.apiClient = client
}

// GenerateReport generates an execution report
func (s *Service) GenerateReport(results []*ProcessResult, format string) error {
	switch format {
	case "json":
		return s.generateJSONReport(results)
	case "html":
		return s.generateHTMLReport(results)
	default:
		return s.generateTextReport(results)
	}
}

// generateJSONReport generates a JSON report
func (s *Service) generateJSONReport(results []*ProcessResult) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile("test-results.json", data, 0644)
}

// generateHTMLReport generates an HTML report
func (s *Service) generateHTMLReport(results []*ProcessResult) error {
	// Simple HTML template
	html := `<!DOCTYPE html>
<html>
<head>
	<title>Virtuoso Test Results</title>
	<style>
		body { font-family: Arial, sans-serif; margin: 20px; }
		.success { color: green; }
		.failure { color: red; }
		.warning { color: orange; }
		table { border-collapse: collapse; width: 100%; }
		th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
		th { background-color: #f2f2f2; }
	</style>
</head>
<body>
	<h1>Virtuoso Test Results</h1>
	<table>
		<tr>
			<th>Test Name</th>
			<th>Status</th>
			<th>Duration</th>
			<th>Errors</th>
			<th>Warnings</th>
		</tr>`

	for _, result := range results {
		status := "success"
		if !result.Success {
			status = "failure"
		}

		html += fmt.Sprintf(`
		<tr>
			<td>%s</td>
			<td class="%s">%s</td>
			<td>%v</td>
			<td>%d</td>
			<td>%d</td>
		</tr>`,
			result.TestName,
			status,
			status,
			result.Duration,
			len(result.Errors),
			len(result.Warnings))
	}

	html += `
	</table>
</body>
</html>`

	return ioutil.WriteFile("test-results.html", []byte(html), 0644)
}

// generateTextReport generates a text report
func (s *Service) generateTextReport(results []*ProcessResult) error {
	var report strings.Builder

	report.WriteString("Virtuoso Test Results\n")
	report.WriteString("====================\n\n")

	passed := 0
	failed := 0

	for _, result := range results {
		if result.Success {
			passed++
			report.WriteString(fmt.Sprintf("✓ %s (%v)\n", result.TestName, result.Duration))
		} else {
			failed++
			report.WriteString(fmt.Sprintf("✗ %s (%v)\n", result.TestName, result.Duration))
			for _, err := range result.Errors {
				report.WriteString(fmt.Sprintf("  ERROR: %s\n", err.Message))
			}
		}

		for _, warn := range result.Warnings {
			report.WriteString(fmt.Sprintf("  WARNING: %s\n", warn.Message))
		}

		report.WriteString("\n")
	}

	report.WriteString(fmt.Sprintf("\nSummary: %d passed, %d failed\n", passed, failed))

	return ioutil.WriteFile("test-results.txt", []byte(report.String()), 0644)
}
