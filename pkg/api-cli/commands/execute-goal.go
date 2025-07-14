package commands

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/marklovelady/api-cli-generator/pkg/api-cli/config"
	"github.com/spf13/cobra"
)

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

func newExecuteGoalCmd() *cobra.Command {
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
			cfg, err := config.LoadConfig()
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
