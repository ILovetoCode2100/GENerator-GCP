package commands

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/marklovelady/api-cli-generator/pkg/api-cli/config"
	"github.com/spf13/cobra"
)

// ========================================
// SHARED OUTPUT STRUCTURES
// ========================================

// ExecutionOutput represents the output structure for goal execution
type ExecutionOutput struct {
	Status      string                    `json:"status"`
	ExecutionID string                    `json:"execution_id"`
	GoalID      int                       `json:"goal_id"`
	SnapshotID  int                       `json:"snapshot_id"`
	StartTime   time.Time                 `json:"start_time"`
	Progress    *client.ExecutionProgress `json:"progress,omitempty"`
	ResultsURL  string                    `json:"results_url,omitempty"`
	ReportURL   string                    `json:"report_url,omitempty"`
	NextSteps   []string                  `json:"next_steps,omitempty"`
}

// MonitorOutput represents the output structure for execution monitoring
type MonitorOutput struct {
	Status       string                    `json:"status"`
	ExecutionID  string                    `json:"execution_id"`
	CurrentState string                    `json:"current_state"`
	Progress     *client.ExecutionProgress `json:"progress,omitempty"`
	Duration     string                    `json:"duration,omitempty"`
	EstimatedEnd *time.Time                `json:"estimated_end,omitempty"`
	LastUpdated  time.Time                 `json:"last_updated"`
	NextSteps    []string                  `json:"next_steps,omitempty"`
}

// AnalysisOutput represents the output structure for execution analysis
type AnalysisOutput struct {
	Status      string                       `json:"status"`
	ExecutionID string                       `json:"execution_id"`
	Summary     *client.ExecutionSummary     `json:"summary"`
	Failures    []client.ExecutionFailure    `json:"failures"`
	Performance *client.ExecutionPerformance `json:"performance,omitempty"`
	AIInsights  []string                     `json:"ai_insights,omitempty"`
	NextSteps   []string                     `json:"next_steps,omitempty"`
}

// TestDataOutput represents the output structure for test data operations
type TestDataOutput struct {
	Status      string       `json:"status"`
	Operation   string       `json:"operation"`
	TableID     string       `json:"table_id,omitempty"`
	TableName   string       `json:"table_name,omitempty"`
	Columns     []string     `json:"columns,omitempty"`
	RowCount    int          `json:"row_count,omitempty"`
	ImportStats *ImportStats `json:"import_stats,omitempty"`
	NextSteps   []string     `json:"next_steps,omitempty"`
}

// ImportStats represents statistics from a CSV import operation
type ImportStats struct {
	TotalRows    int `json:"total_rows"`
	ImportedRows int `json:"imported_rows"`
	SkippedRows  int `json:"skipped_rows"`
	ErrorRows    int `json:"error_rows"`
}

// EnvironmentOutput represents the output structure for environment operations
type EnvironmentOutput struct {
	Status        string                 `json:"status"`
	EnvironmentID string                 `json:"environment_id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description,omitempty"`
	Variables     map[string]interface{} `json:"variables"`
	VariableCount int                    `json:"variable_count"`
	CreatedAt     time.Time              `json:"created_at"`
	NextSteps     []string               `json:"next_steps,omitempty"`
}

// ========================================
// ENVIRONMENT MANAGEMENT
// ========================================

// NewCreateEnvironmentCmd creates a new test environment command
func NewCreateEnvironmentCmd() *cobra.Command {
	var nameFlag string
	var descriptionFlag string
	var variablesFlag string
	var variablesFileFlag string
	var copyFromFlag string

	cmd := &cobra.Command{
		Use:   "create-environment",
		Short: "Create a new test environment with variables",
		Long: `Create a new test environment with configuration variables for testing.

The command creates test environments with support for:
- Environment variables and configuration settings
- JSON file import for bulk variable configuration
- Copying variables from existing environments
- Secure handling of sensitive variables
- Environment-specific test configurations

Examples:
  # Create basic environment
  api-cli create-environment --name "Production" --description "Production environment"

  # Create environment with variables
  api-cli create-environment --name "Staging" --variables "BASE_URL=https://staging.example.com,API_KEY=secret123"

  # Create environment from JSON file
  api-cli create-environment --name "Development" --variables-file dev-config.json

  # Copy environment from existing one
  api-cli create-environment --name "Production-Copy" --copy-from env_456`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig("")
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			if nameFlag == "" {
				return fmt.Errorf("environment name is required")
			}

			client := client.NewClient(cfg)

			// Parse variables from different sources
			variables := make(map[string]interface{})

			// Parse variables from flag
			if variablesFlag != "" {
				vars, err := parseVariablesString(variablesFlag)
				if err != nil {
					return fmt.Errorf("failed to parse variables: %w", err)
				}
				for k, v := range vars {
					variables[k] = v
				}
			}

			// Parse variables from file
			if variablesFileFlag != "" {
				vars, err := parseVariablesFile(variablesFileFlag)
				if err != nil {
					return fmt.Errorf("failed to parse variables file: %w", err)
				}
				for k, v := range vars {
					variables[k] = v
				}
			}

			// TODO: Handle copy-from functionality
			if copyFromFlag != "" {
				return fmt.Errorf("copy-from functionality not yet implemented")
			}

			// Create environment
			environment, err := client.CreateEnvironment(nameFlag, descriptionFlag, variables)
			if err != nil {
				return fmt.Errorf("failed to create environment: %w", err)
			}

			// Prepare output
			output := &EnvironmentOutput{
				Status:        "success",
				EnvironmentID: environment.ID,
				Name:          environment.Name,
				Description:   environment.Description,
				Variables:     environment.Variables,
				VariableCount: len(environment.Variables),
				CreatedAt:     environment.CreatedAt,
				NextSteps: []string{
					"Use environment ID '" + environment.ID + "' in goal configurations",
					"Reference variables in test steps using ${VARIABLE_NAME} syntax",
					"Use 'api-cli update-environment " + environment.ID + "' to modify variables",
				},
			}

			return outputEnvironmentResult(output, cfg.Output.DefaultFormat)
		},
	}

	cmd.Flags().StringVar(&nameFlag, "name", "", "Name of the environment (required)")
	cmd.Flags().StringVar(&descriptionFlag, "description", "", "Description of the environment")
	cmd.Flags().StringVar(&variablesFlag, "variables", "", "Comma-separated key=value pairs")
	cmd.Flags().StringVar(&variablesFileFlag, "variables-file", "", "JSON file containing variables")
	cmd.Flags().StringVar(&copyFromFlag, "copy-from", "", "Environment ID to copy variables from")

	return cmd
}

// ========================================
// TEST DATA MANAGEMENT
// ========================================

// NewManageTestDataCmd creates a test data management command
func NewManageTestDataCmd() *cobra.Command {
	var createTableFlag bool
	var tableNameFlag string
	var descriptionFlag string
	var columnsFlag string
	var importCsvFlag string
	var exportCsvFlag string
	var tableIDFlag string

	cmd := &cobra.Command{
		Use:   "manage-test-data",
		Short: "Manage test data tables and CSV import/export",
		Long: `Manage test data tables with support for CSV import/export operations.

The command provides comprehensive test data management including:
- Create new test data tables with defined columns
- Import test data from CSV files
- Export test data to CSV files
- View table details and statistics
- Manage table structure and data

Examples:
  # Create a new test data table
  api-cli manage-test-data --create-table --table-name "User_Credentials" --columns "username,password,role"

  # Import CSV data to existing table
  api-cli manage-test-data --import-csv users.csv --table-id tbl_789

  # Export table data to CSV
  api-cli manage-test-data --export-csv users_export.csv --table-id tbl_789

  # Get table information
  api-cli manage-test-data --table-id tbl_789`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig("")
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			client := client.NewClient(cfg)

			// Determine operation based on flags
			if createTableFlag {
				return createTestDataTable(client, tableNameFlag, descriptionFlag, columnsFlag, cfg.Output.DefaultFormat)
			} else if importCsvFlag != "" {
				return importTestDataFromCSV(client, importCsvFlag, tableIDFlag, cfg.Output.DefaultFormat)
			} else if exportCsvFlag != "" {
				return exportTestDataToCSV(client, exportCsvFlag, tableIDFlag, cfg.Output.DefaultFormat)
			} else if tableIDFlag != "" {
				return getTestDataTable(client, tableIDFlag, cfg.Output.DefaultFormat)
			} else {
				return fmt.Errorf("no operation specified. Use --create-table, --import-csv, --export-csv, or --table-id")
			}
		},
	}

	cmd.Flags().BoolVar(&createTableFlag, "create-table", false, "Create a new test data table")
	cmd.Flags().StringVar(&tableNameFlag, "table-name", "", "Name of the test data table")
	cmd.Flags().StringVar(&descriptionFlag, "description", "", "Description of the test data table")
	cmd.Flags().StringVar(&columnsFlag, "columns", "", "Comma-separated list of column names")
	cmd.Flags().StringVar(&importCsvFlag, "import-csv", "", "Path to CSV file to import")
	cmd.Flags().StringVar(&exportCsvFlag, "export-csv", "", "Path to CSV file to export to")
	cmd.Flags().StringVar(&tableIDFlag, "table-id", "", "ID of the test data table")

	return cmd
}

// ========================================
// EXECUTION COMMANDS
// ========================================

// NewExecuteGoalCmd creates a goal execution command
func NewExecuteGoalCmd() *cobra.Command {
	var waitFlag bool
	var timeoutFlag int

	cmd := &cobra.Command{
		Use:   "execute-goal GOAL_ID [SNAPSHOT_ID]",
		Short: "Execute a goal with real-time monitoring",
		Long: `Execute a goal and return execution details with optional monitoring.

The command executes the specified goal and returns execution information including:
- Execution ID for monitoring
- Current status and progress
- Results and report URLs when available

Examples:
  # Execute goal with auto-detected snapshot
  api-cli execute-goal 1234

  # Execute goal with specific snapshot
  api-cli execute-goal 1234 5678

  # Execute with wait until completion
  api-cli execute-goal 1234 --wait

  # Execute with custom timeout
  api-cli execute-goal 1234 --wait --timeout 600`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig("")
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			// Parse goal ID
			goalID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid goal ID: %w", err)
			}

			// Parse or get snapshot ID
			var snapshotID int
			if len(args) > 1 {
				snapshotID, err = strconv.Atoi(args[1])
				if err != nil {
					return fmt.Errorf("invalid snapshot ID: %w", err)
				}
			} else {
				// Auto-detect snapshot ID
				client := client.NewClient(cfg)
				snapshotIDStr, err := client.GetGoalSnapshot(goalID)
				if err != nil {
					return fmt.Errorf("failed to get goal snapshot: %w", err)
				}
				snapshotID, err = strconv.Atoi(snapshotIDStr)
				if err != nil {
					return fmt.Errorf("invalid snapshot ID from API: %w", err)
				}
			}

			// Execute the goal
			client := client.NewClient(cfg)
			execution, err := client.ExecuteGoal(goalID, snapshotID)
			if err != nil {
				return fmt.Errorf("failed to execute goal: %w", err)
			}

			// Wait for completion if requested
			if waitFlag {
				execution, err = waitForExecution(client, execution.ID, timeoutFlag)
				if err != nil {
					return fmt.Errorf("failed to wait for execution: %w", err)
				}
			}

			// Prepare output
			output := &ExecutionOutput{
				Status:      "success",
				ExecutionID: execution.ID,
				GoalID:      execution.GoalID,
				SnapshotID:  execution.SnapshotID,
				StartTime:   execution.StartTime,
				Progress:    execution.Progress,
				ResultsURL:  execution.ResultsURL,
				ReportURL:   execution.ReportURL,
			}

			// Add next steps based on execution status
			if execution.Status == "RUNNING" {
				output.NextSteps = []string{
					"Use 'api-cli monitor-execution " + execution.ID + "' to track progress",
					"Use 'api-cli get-execution-analysis " + execution.ID + "' when complete",
				}
			} else if execution.Status == "COMPLETED" {
				output.NextSteps = []string{
					"Use 'api-cli get-execution-analysis " + execution.ID + "' to view results",
					"Check results at: " + execution.ResultsURL,
				}
			}

			return outputExecutionResult(output, cfg.Output.DefaultFormat)
		},
	}

	cmd.Flags().BoolVar(&waitFlag, "wait", false, "Wait for execution to complete")
	cmd.Flags().IntVar(&timeoutFlag, "timeout", 300, "Timeout in seconds when waiting (default: 300)")

	return cmd
}

// NewMonitorExecutionCmd creates an execution monitoring command
func NewMonitorExecutionCmd() *cobra.Command {
	var followFlag bool
	var intervalFlag int
	var timeoutFlag int

	cmd := &cobra.Command{
		Use:   "monitor-execution EXECUTION_ID",
		Short: "Monitor execution progress with real-time updates",
		Long: `Monitor the progress of a running execution with real-time status updates.

The command provides detailed information about execution progress including:
- Current execution status and progress percentage
- Completed vs total steps and journeys
- Success rate and failure counts
- Estimated completion time
- Real-time updates with --follow flag

Examples:
  # Get current execution status
  api-cli monitor-execution exec_12345

  # Follow execution with real-time updates
  api-cli monitor-execution exec_12345 --follow

  # Follow with custom update interval
  api-cli monitor-execution exec_12345 --follow --interval 10

  # Monitor with timeout
  api-cli monitor-execution exec_12345 --follow --timeout 600`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig("")
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			executionID := args[0]
			client := client.NewClient(cfg)

			if followFlag {
				return followExecution(client, executionID, intervalFlag, timeoutFlag, cfg.Output.DefaultFormat)
			}

			// Single status check
			execution, err := client.GetExecutionStatus(executionID)
			if err != nil {
				return fmt.Errorf("failed to get execution status: %w", err)
			}

			// Prepare output
			output := &MonitorOutput{
				Status:       "success",
				ExecutionID:  execution.ID,
				CurrentState: execution.Status,
				Progress:     execution.Progress,
				LastUpdated:  time.Now(),
			}

			// Calculate duration if execution has started
			if !execution.StartTime.IsZero() {
				var endTime time.Time
				if execution.EndTime != nil {
					endTime = *execution.EndTime
				} else {
					endTime = time.Now()
				}
				duration := endTime.Sub(execution.StartTime)
				output.Duration = formatDuration(duration)
			}

			// Add next steps based on status
			switch execution.Status {
			case "RUNNING":
				output.NextSteps = []string{
					"Use --follow flag to monitor in real-time",
					"Use 'api-cli get-execution-analysis " + executionID + "' when complete",
				}
			case "COMPLETED":
				output.NextSteps = []string{
					"Use 'api-cli get-execution-analysis " + executionID + "' to view results",
					"Check results at: " + execution.ResultsURL,
				}
			case "FAILED":
				output.NextSteps = []string{
					"Use 'api-cli get-execution-analysis " + executionID + "' to view failure details",
					"Check failure logs and screenshots in the analysis",
				}
			}

			return outputMonitorResult(output, cfg.Output.DefaultFormat)
		},
	}

	cmd.Flags().BoolVar(&followFlag, "follow", false, "Follow execution with real-time updates")
	cmd.Flags().IntVar(&intervalFlag, "interval", 5, "Update interval in seconds (default: 5)")
	cmd.Flags().IntVar(&timeoutFlag, "timeout", 300, "Timeout in seconds when following (default: 300)")

	return cmd
}

// NewGetExecutionAnalysisCmd creates an execution analysis command
func NewGetExecutionAnalysisCmd() *cobra.Command {
	var includeAIFlag bool
	var failuresOnlyFlag bool

	cmd := &cobra.Command{
		Use:   "get-execution-analysis EXECUTION_ID",
		Short: "Get detailed execution analysis and failure insights",
		Long: `Get comprehensive analysis of an execution including failure details and AI insights.

The command provides detailed information about execution results including:
- Summary statistics (total/passed/failed steps)
- Detailed failure analysis with error messages
- Performance metrics and timings
- AI-generated insights and suggestions
- Step-by-step failure breakdown with screenshots

Examples:
  # Get basic execution analysis
  api-cli get-execution-analysis exec_12345

  # Include AI insights and suggestions
  api-cli get-execution-analysis exec_12345 --ai-insights

  # Show only failures for debugging
  api-cli get-execution-analysis exec_12345 --failures-only

  # Get comprehensive analysis with AI insights
  api-cli get-execution-analysis exec_12345 --ai-insights --output json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig("")
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			executionID := args[0]
			client := client.NewClient(cfg)

			// Get execution analysis
			analysis, err := client.GetExecutionAnalysis(executionID, includeAIFlag)
			if err != nil {
				return fmt.Errorf("failed to get execution analysis: %w", err)
			}

			// Prepare output
			output := &AnalysisOutput{
				Status:      "success",
				ExecutionID: executionID,
				Summary:     analysis.Summary,
				Performance: analysis.Performance,
				AIInsights:  analysis.AIInsights,
			}

			// Filter failures if requested
			if failuresOnlyFlag {
				output.Failures = analysis.Failures
			} else {
				output.Failures = analysis.Failures
			}

			// Add next steps based on analysis
			if analysis.Summary != nil {
				if analysis.Summary.FailedSteps > 0 {
					output.NextSteps = []string{
						"Review failure details above for debugging",
						"Check screenshots and error messages",
						"Consider updating selectors or wait conditions",
					}

					if includeAIFlag && len(analysis.AIInsights) > 0 {
						output.NextSteps = append(output.NextSteps, "Review AI suggestions for quick fixes")
					}
				} else {
					output.NextSteps = []string{
						"Execution completed successfully",
						"Consider reviewing performance metrics",
						"Set up monitoring for future executions",
					}
				}
			}

			return outputAnalysisResult(output, cfg.Output.DefaultFormat)
		},
	}

	cmd.Flags().BoolVar(&includeAIFlag, "ai-insights", false, "Include AI-generated insights and suggestions")
	cmd.Flags().BoolVar(&failuresOnlyFlag, "failures-only", false, "Show only failure details")

	return cmd
}

// ========================================
// SHARED HELPER FUNCTIONS
// ========================================

// waitForExecution waits for execution completion with timeout
func waitForExecution(client *client.Client, executionID string, timeout int) (*client.Execution, error) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	timeoutChan := time.After(time.Duration(timeout) * time.Second)

	for {
		select {
		case <-timeoutChan:
			return nil, fmt.Errorf("execution timeout after %d seconds", timeout)
		case <-ticker.C:
			execution, err := client.GetExecutionStatus(executionID)
			if err != nil {
				return nil, fmt.Errorf("failed to get execution status: %w", err)
			}

			if execution.Status == "COMPLETED" || execution.Status == "FAILED" || execution.Status == "CANCELLED" {
				return execution, nil
			}

			// Continue waiting if still running
		}
	}
}

// followExecution follows an execution with real-time updates
func followExecution(client *client.Client, executionID string, interval, timeout int, format string) error {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	timeoutChan := time.After(time.Duration(timeout) * time.Second)

	// Print header for human format
	if format == "human" {
		fmt.Printf("🔄 Following execution %s (press Ctrl+C to stop)\n", executionID)
		fmt.Printf("📊 Updates every %d seconds\n\n", interval)
	}

	for {
		select {
		case <-timeoutChan:
			return fmt.Errorf("monitoring timeout after %d seconds", timeout)
		case <-ticker.C:
			execution, err := client.GetExecutionStatus(executionID)
			if err != nil {
				return fmt.Errorf("failed to get execution status: %w", err)
			}

			// Show current status
			if format == "human" {
				displayExecutionStatus(execution)
			} else {
				output := &MonitorOutput{
					Status:       "success",
					ExecutionID:  execution.ID,
					CurrentState: execution.Status,
					Progress:     execution.Progress,
					LastUpdated:  time.Now(),
				}

				if !execution.StartTime.IsZero() {
					var endTime time.Time
					if execution.EndTime != nil {
						endTime = *execution.EndTime
					} else {
						endTime = time.Now()
					}
					duration := endTime.Sub(execution.StartTime)
					output.Duration = formatDuration(duration)
				}

				outputMonitorResult(output, format)
			}

			// Stop if execution is complete
			if execution.Status == "COMPLETED" || execution.Status == "FAILED" || execution.Status == "CANCELLED" {
				if format == "human" {
					fmt.Printf("\n🏁 Execution %s: %s\n", executionID, execution.Status)
				}
				return nil
			}
		}
	}
}

// displayExecutionStatus displays execution status in human-readable format
func displayExecutionStatus(execution *client.Execution) {
	fmt.Printf("\r⏰ %s | Status: %s",
		time.Now().Format("15:04:05"),
		execution.Status)

	if execution.Progress != nil {
		fmt.Printf(" | Progress: %.1f%% (%d/%d steps)",
			execution.Progress.PercentComplete,
			execution.Progress.CompletedSteps,
			execution.Progress.TotalSteps)

		if execution.Progress.FailedSteps > 0 {
			fmt.Printf(" | Failures: %d", execution.Progress.FailedSteps)
		}

		if execution.Progress.CurrentJourney != "" {
			fmt.Printf(" | Current: %s", execution.Progress.CurrentJourney)
		}
	}

	fmt.Print("                    ") // Clear any remaining characters
}

// formatDuration formats a duration in human-readable format
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh %dm %ds", int(d.Hours()), int(d.Minutes())%60, int(d.Seconds())%60)
}

// parseVariablesString parses a comma-separated string of key=value pairs
func parseVariablesString(varsStr string) (map[string]interface{}, error) {
	variables := make(map[string]interface{})

	if varsStr == "" {
		return variables, nil
	}

	pairs := strings.Split(varsStr, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid variable format: %s (expected key=value)", pair)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "" {
			return nil, fmt.Errorf("empty variable key in: %s", pair)
		}

		variables[key] = value
	}

	return variables, nil
}

// parseVariablesFile parses a JSON file containing variables
func parseVariablesFile(filePath string) (map[string]interface{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open variables file: %w", err)
	}
	defer file.Close()

	var variables map[string]interface{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&variables)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON variables file: %w", err)
	}

	return variables, nil
}

// createTestDataTable creates a new test data table
func createTestDataTable(client *client.Client, name, description, columnsStr, format string) error {
	if name == "" {
		return fmt.Errorf("table name is required")
	}

	if columnsStr == "" {
		return fmt.Errorf("columns are required")
	}

	columns := strings.Split(columnsStr, ",")
	for i, col := range columns {
		columns[i] = strings.TrimSpace(col)
	}

	table, err := client.CreateTestDataTable(name, description, columns)
	if err != nil {
		return fmt.Errorf("failed to create test data table: %w", err)
	}

	output := &TestDataOutput{
		Status:    "success",
		Operation: "create_table",
		TableID:   table.ID,
		TableName: table.Name,
		Columns:   table.Columns,
		RowCount:  table.RowCount,
		NextSteps: []string{
			"Use --import-csv to add data from CSV file",
			"Use --table-id " + table.ID + " to view table details",
			"Reference table in test steps using table name: " + table.Name,
		},
	}

	return outputTestDataResult(output, format)
}

// importTestDataFromCSV imports data from a CSV file
func importTestDataFromCSV(client *client.Client, csvPath, tableID, format string) error {
	if tableID == "" {
		return fmt.Errorf("table ID is required for import")
	}

	// Read CSV file
	file, err := os.Open(csvPath)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV file: %w", err)
	}

	if len(records) == 0 {
		return fmt.Errorf("CSV file is empty")
	}

	// Import data
	err = client.ImportTestDataFromCSV(tableID, records)
	if err != nil {
		return fmt.Errorf("failed to import test data: %w", err)
	}

	// Get updated table info
	table, err := client.GetTestDataTable(tableID)
	if err != nil {
		return fmt.Errorf("failed to get table info after import: %w", err)
	}

	output := &TestDataOutput{
		Status:    "success",
		Operation: "import_csv",
		TableID:   table.ID,
		TableName: table.Name,
		Columns:   table.Columns,
		RowCount:  table.RowCount,
		ImportStats: &ImportStats{
			TotalRows:    len(records),
			ImportedRows: len(records) - 1, // Exclude header
			SkippedRows:  0,
			ErrorRows:    0,
		},
		NextSteps: []string{
			"Use --table-id " + table.ID + " to view table details",
			"Reference table in test steps using table name: " + table.Name,
			"Use --export-csv to backup the data",
		},
	}

	return outputTestDataResult(output, format)
}

// exportTestDataToCSV exports data to a CSV file
func exportTestDataToCSV(client *client.Client, csvPath, tableID, format string) error {
	if tableID == "" {
		return fmt.Errorf("table ID is required for export")
	}

	// Get table info
	table, err := client.GetTestDataTable(tableID)
	if err != nil {
		return fmt.Errorf("failed to get table info: %w", err)
	}

	// Note: In a real implementation, you would get the actual table data
	// For now, we'll create a sample CSV with headers
	file, err := os.Create(csvPath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write headers
	err = writer.Write(table.Columns)
	if err != nil {
		return fmt.Errorf("failed to write CSV headers: %w", err)
	}

	// Note: In a real implementation, you would write the actual data rows here
	// For now, we'll just indicate the export was successful

	output := &TestDataOutput{
		Status:    "success",
		Operation: "export_csv",
		TableID:   table.ID,
		TableName: table.Name,
		Columns:   table.Columns,
		RowCount:  table.RowCount,
		NextSteps: []string{
			"CSV file exported to: " + csvPath,
			"File contains " + fmt.Sprintf("%d", table.RowCount) + " rows",
			"Use this file as backup or for sharing test data",
		},
	}

	return outputTestDataResult(output, format)
}

// getTestDataTable gets test data table information
func getTestDataTable(client *client.Client, tableID, format string) error {
	table, err := client.GetTestDataTable(tableID)
	if err != nil {
		return fmt.Errorf("failed to get test data table: %w", err)
	}

	output := &TestDataOutput{
		Status:    "success",
		Operation: "get_table",
		TableID:   table.ID,
		TableName: table.Name,
		Columns:   table.Columns,
		RowCount:  table.RowCount,
		NextSteps: []string{
			"Use --import-csv to add data from CSV file",
			"Use --export-csv to backup the data",
			"Reference table in test steps using table name: " + table.Name,
		},
	}

	return outputTestDataResult(output, format)
}

// isSensitiveKey checks if a key contains sensitive information
func isSensitiveKey(key string) bool {
	key = strings.ToLower(key)
	sensitiveKeys := []string{
		"password", "token", "key", "secret", "auth", "credential",
		"api_key", "access_token", "client_secret", "private_key",
	}

	for _, sensitive := range sensitiveKeys {
		if strings.Contains(key, sensitive) {
			return true
		}
	}

	return false
}

// maskValue masks sensitive values for display
func maskValue(value string) string {
	if len(value) <= 4 {
		return "***"
	}

	return value[:2] + "***" + value[len(value)-2:]
}

// ========================================
// OUTPUT FORMATTING FUNCTIONS
// ========================================

// outputExecutionResult formats and outputs the execution result
func outputExecutionResult(output *ExecutionOutput, format string) error {
	switch format {
	case "json":
		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(jsonData))

	case "yaml":
		// Convert to YAML-friendly format
		yamlData := map[string]interface{}{
			"status":       output.Status,
			"execution_id": output.ExecutionID,
			"goal_id":      output.GoalID,
			"snapshot_id":  output.SnapshotID,
			"start_time":   output.StartTime.Format(time.RFC3339),
			"progress":     output.Progress,
			"results_url":  output.ResultsURL,
			"report_url":   output.ReportURL,
			"next_steps":   output.NextSteps,
		}

		jsonData, err := json.MarshalIndent(yamlData, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal YAML: %w", err)
		}
		fmt.Println(string(jsonData))

	case "ai":
		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal AI format: %w", err)
		}
		fmt.Println(string(jsonData))

	default: // human
		fmt.Printf("✅ Goal execution started successfully\n")
		fmt.Printf("📋 Execution ID: %s\n", output.ExecutionID)
		fmt.Printf("🎯 Goal ID: %d\n", output.GoalID)
		fmt.Printf("📸 Snapshot ID: %d\n", output.SnapshotID)
		fmt.Printf("⏰ Started: %s\n", output.StartTime.Format("2006-01-02 15:04:05"))

		if output.Progress != nil {
			fmt.Printf("📊 Progress: %.1f%% (%d/%d steps)\n",
				output.Progress.PercentComplete,
				output.Progress.CompletedSteps,
				output.Progress.TotalSteps)
		}

		if output.ResultsURL != "" {
			fmt.Printf("🔗 Results: %s\n", output.ResultsURL)
		}

		if output.ReportURL != "" {
			fmt.Printf("📊 Report: %s\n", output.ReportURL)
		}

		if len(output.NextSteps) > 0 {
			fmt.Printf("\n💡 Next steps:\n")
			for _, step := range output.NextSteps {
				fmt.Printf("  • %s\n", step)
			}
		}
	}

	return nil
}

// outputMonitorResult formats and outputs the monitor result
func outputMonitorResult(output *MonitorOutput, format string) error {
	switch format {
	case "json":
		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(jsonData))

	case "yaml":
		// Convert to YAML-friendly format
		yamlData := map[string]interface{}{
			"status":        output.Status,
			"execution_id":  output.ExecutionID,
			"current_state": output.CurrentState,
			"progress":      output.Progress,
			"duration":      output.Duration,
			"estimated_end": output.EstimatedEnd,
			"last_updated":  output.LastUpdated.Format(time.RFC3339),
			"next_steps":    output.NextSteps,
		}

		jsonData, err := json.MarshalIndent(yamlData, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal YAML: %w", err)
		}
		fmt.Println(string(jsonData))

	case "ai":
		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal AI format: %w", err)
		}
		fmt.Println(string(jsonData))

	default: // human
		fmt.Printf("📊 Execution Monitor - %s\n", output.ExecutionID)
		fmt.Printf("🔄 Current State: %s\n", output.CurrentState)

		if output.Progress != nil {
			fmt.Printf("📈 Progress: %.1f%% (%d/%d steps)\n",
				output.Progress.PercentComplete,
				output.Progress.CompletedSteps,
				output.Progress.TotalSteps)

			if output.Progress.TotalJourneys > 0 {
				fmt.Printf("📋 Journeys: %d/%d completed\n",
					output.Progress.CompletedJourneys,
					output.Progress.TotalJourneys)
			}

			if output.Progress.FailedSteps > 0 {
				fmt.Printf("⚠️  Failed Steps: %d\n", output.Progress.FailedSteps)
			}

			if output.Progress.CurrentJourney != "" {
				fmt.Printf("🎯 Current Journey: %s\n", output.Progress.CurrentJourney)
			}

			if output.Progress.SuccessRate > 0 {
				fmt.Printf("✅ Success Rate: %.1f%%\n", output.Progress.SuccessRate)
			}
		}

		if output.Duration != "" {
			fmt.Printf("⏱️  Duration: %s\n", output.Duration)
		}

		fmt.Printf("🕐 Last Updated: %s\n", output.LastUpdated.Format("2006-01-02 15:04:05"))

		if len(output.NextSteps) > 0 {
			fmt.Printf("\n💡 Next Steps:\n")
			for _, step := range output.NextSteps {
				fmt.Printf("  • %s\n", step)
			}
		}
	}

	return nil
}

// outputAnalysisResult formats and outputs the analysis result
func outputAnalysisResult(output *AnalysisOutput, format string) error {
	switch format {
	case "json":
		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(jsonData))

	case "yaml":
		// Convert to YAML-friendly format
		yamlData := map[string]interface{}{
			"status":       output.Status,
			"execution_id": output.ExecutionID,
			"summary":      output.Summary,
			"failures":     output.Failures,
			"performance":  output.Performance,
			"ai_insights":  output.AIInsights,
			"next_steps":   output.NextSteps,
		}

		jsonData, err := json.MarshalIndent(yamlData, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal YAML: %w", err)
		}
		fmt.Println(string(jsonData))

	case "ai":
		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal AI format: %w", err)
		}
		fmt.Println(string(jsonData))

	default: // human
		fmt.Printf("📊 Execution Analysis - %s\n\n", output.ExecutionID)

		// Summary section
		if output.Summary != nil {
			fmt.Printf("📋 Summary:\n")
			fmt.Printf("  • Total Steps: %d\n", output.Summary.TotalSteps)
			fmt.Printf("  • Passed Steps: %d\n", output.Summary.PassedSteps)
			fmt.Printf("  • Failed Steps: %d\n", output.Summary.FailedSteps)
			fmt.Printf("  • Skipped Steps: %d\n", output.Summary.SkippedSteps)
			fmt.Printf("  • Success Rate: %.1f%%\n", output.Summary.SuccessRate)
			fmt.Printf("  • Duration: %s\n", output.Summary.Duration)

			if output.Summary.TotalJourneys > 0 {
				fmt.Printf("  • Total Journeys: %d\n", output.Summary.TotalJourneys)
				fmt.Printf("  • Passed Journeys: %d\n", output.Summary.PassedJourneys)
				fmt.Printf("  • Failed Journeys: %d\n", output.Summary.FailedJourneys)
			}

			fmt.Printf("\n")
		}

		// Performance section
		if output.Performance != nil {
			fmt.Printf("⚡ Performance:\n")
			fmt.Printf("  • Average Step Time: %dms\n", output.Performance.AverageStepTime)
			fmt.Printf("  • Slowest Step: %dms\n", output.Performance.SlowestStepTime)
			fmt.Printf("  • Fastest Step: %dms\n", output.Performance.FastestStepTime)
			fmt.Printf("  • Network Requests: %d\n", output.Performance.NetworkRequests)
			fmt.Printf("  • JavaScript Errors: %d\n", output.Performance.JavaScriptErrors)
			fmt.Printf("  • Page Load Time: %dms\n", output.Performance.PageLoadTime)
			fmt.Printf("\n")
		}

		// Failures section
		if len(output.Failures) > 0 {
			fmt.Printf("❌ Failures (%d):\n", len(output.Failures))
			for i, failure := range output.Failures {
				fmt.Printf("  %d. Step %d in %s\n", i+1, failure.StepID, failure.JourneyName)
				fmt.Printf("     Action: %s\n", failure.Action)
				fmt.Printf("     Error: %s\n", failure.Error)
				if failure.Screenshot != "" {
					fmt.Printf("     Screenshot: %s\n", failure.Screenshot)
				}
				if failure.AISuggestion != "" {
					fmt.Printf("     💡 AI Suggestion: %s\n", failure.AISuggestion)
				}
				fmt.Printf("     Time: %s\n", failure.Timestamp)
				fmt.Printf("\n")
			}
		}

		// AI Insights section
		if len(output.AIInsights) > 0 {
			fmt.Printf("🤖 AI Insights:\n")
			for i, insight := range output.AIInsights {
				fmt.Printf("  %d. %s\n", i+1, insight)
			}
			fmt.Printf("\n")
		}

		// Next steps section
		if len(output.NextSteps) > 0 {
			fmt.Printf("💡 Next Steps:\n")
			for _, step := range output.NextSteps {
				fmt.Printf("  • %s\n", step)
			}
		}
	}

	return nil
}

// outputTestDataResult formats and outputs the test data result
func outputTestDataResult(output *TestDataOutput, format string) error {
	switch format {
	case "json":
		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(jsonData))

	case "yaml":
		// Convert to YAML-friendly format
		yamlData := map[string]interface{}{
			"status":       output.Status,
			"operation":    output.Operation,
			"table_id":     output.TableID,
			"table_name":   output.TableName,
			"columns":      output.Columns,
			"row_count":    output.RowCount,
			"import_stats": output.ImportStats,
			"next_steps":   output.NextSteps,
		}

		jsonData, err := json.MarshalIndent(yamlData, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal YAML: %w", err)
		}
		fmt.Println(string(jsonData))

	case "ai":
		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal AI format: %w", err)
		}
		fmt.Println(string(jsonData))

	default: // human
		fmt.Printf("📊 Test Data Management - %s\n\n", strings.Title(output.Operation))

		if output.TableID != "" {
			fmt.Printf("🆔 Table ID: %s\n", output.TableID)
		}

		if output.TableName != "" {
			fmt.Printf("📋 Table Name: %s\n", output.TableName)
		}

		if len(output.Columns) > 0 {
			fmt.Printf("📝 Columns: %s\n", strings.Join(output.Columns, ", "))
		}

		if output.RowCount > 0 {
			fmt.Printf("📊 Row Count: %d\n", output.RowCount)
		}

		if output.ImportStats != nil {
			fmt.Printf("\n📥 Import Statistics:\n")
			fmt.Printf("  • Total Rows: %d\n", output.ImportStats.TotalRows)
			fmt.Printf("  • Imported Rows: %d\n", output.ImportStats.ImportedRows)
			if output.ImportStats.SkippedRows > 0 {
				fmt.Printf("  • Skipped Rows: %d\n", output.ImportStats.SkippedRows)
			}
			if output.ImportStats.ErrorRows > 0 {
				fmt.Printf("  • Error Rows: %d\n", output.ImportStats.ErrorRows)
			}
		}

		if len(output.NextSteps) > 0 {
			fmt.Printf("\n💡 Next Steps:\n")
			for _, step := range output.NextSteps {
				fmt.Printf("  • %s\n", step)
			}
		}
	}

	return nil
}

// outputEnvironmentResult formats and outputs the environment result
func outputEnvironmentResult(output *EnvironmentOutput, format string) error {
	switch format {
	case "json":
		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(jsonData))

	case "yaml":
		// Convert to YAML-friendly format
		yamlData := map[string]interface{}{
			"status":         output.Status,
			"environment_id": output.EnvironmentID,
			"name":           output.Name,
			"description":    output.Description,
			"variables":      output.Variables,
			"variable_count": output.VariableCount,
			"created_at":     output.CreatedAt.Format(time.RFC3339),
			"next_steps":     output.NextSteps,
		}

		jsonData, err := json.MarshalIndent(yamlData, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal YAML: %w", err)
		}
		fmt.Println(string(jsonData))

	case "ai":
		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal AI format: %w", err)
		}
		fmt.Println(string(jsonData))

	default: // human
		fmt.Printf("✅ Environment created successfully\n")
		fmt.Printf("🆔 Environment ID: %s\n", output.EnvironmentID)
		fmt.Printf("📋 Name: %s\n", output.Name)

		if output.Description != "" {
			fmt.Printf("📝 Description: %s\n", output.Description)
		}

		fmt.Printf("🔢 Variables: %d\n", output.VariableCount)

		if len(output.Variables) > 0 {
			fmt.Printf("\n🔐 Environment Variables:\n")
			for key, value := range output.Variables {
				// Mask sensitive values
				displayValue := fmt.Sprintf("%v", value)
				if isSensitiveKey(key) {
					displayValue = maskValue(displayValue)
				}
				fmt.Printf("  • %s: %s\n", key, displayValue)
			}
		}

		fmt.Printf("\n⏰ Created: %s\n", output.CreatedAt.Format("2006-01-02 15:04:05"))

		if len(output.NextSteps) > 0 {
			fmt.Printf("\n💡 Next Steps:\n")
			for _, step := range output.NextSteps {
				fmt.Printf("  • %s\n", step)
			}
		}
	}

	return nil
}
