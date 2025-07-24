package commands

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// NewExecuteGoalFixedCmd creates a fixed version of the execute-goal command
func NewExecuteGoalFixedCmd() *cobra.Command {
	var waitFlag bool
	var timeoutFlag int

	cmd := &cobra.Command{
		Use:   "execute-goal GOAL_ID [SNAPSHOT_ID]",
		Short: "Execute a goal (fixed version with robust response handling)",
		Long: `Execute a goal and optionally wait for completion.

This is the fixed version that handles both numeric and string execution IDs.`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse arguments
			goalID, err := parseIntArg(args[0], "goal ID")
			if err != nil {
				return err
			}

			snapshotID := ""
			if len(args) > 1 {
				snapshotID = args[1]
			}

			// Get API client
			apiClient, err := getAPIClient()
			if err != nil {
				return err
			}

			// Create context with timeout
			ctx, cancel := CommandContextWithTimeout(time.Duration(timeoutFlag) * time.Second)
			defer cancel()

			// Use the robust version that handles type mismatches
			execution, err := apiClient.ExecuteGoalRobustWithContext(ctx, goalID, snapshotID)
			if err != nil {
				return fmt.Errorf("failed to execute goal: %w", err)
			}

			// Handle output
			outputFormat, _ := cmd.Flags().GetString("output")
			switch outputFormat {
			case "json":
				return outputJSON(map[string]interface{}{
					"execution_id": execution.ID,
					"status":       execution.Status,
					"goal_id":      execution.GoalID,
					"snapshot_id":  execution.SnapshotID,
					"progress":     execution.Progress,
					"results_url":  execution.ResultsURL,
				})

			case "yaml":
				return outputYAML(map[string]interface{}{
					"execution_id": execution.ID,
					"status":       execution.Status,
					"goal_id":      execution.GoalID,
					"snapshot_id":  execution.SnapshotID,
					"progress":     execution.Progress,
					"results_url":  execution.ResultsURL,
				})

			case "ai":
				fmt.Printf("Goal execution started successfully!\n")
				fmt.Printf("Execution ID: %s\n", execution.ID)
				fmt.Printf("Status: %s\n", execution.Status)
				if execution.Progress > 0 {
					fmt.Printf("Progress: %d%%\n", execution.Progress)
				}
				if waitFlag {
					fmt.Printf("\nWaiting for execution to complete...\n")
				} else {
					fmt.Printf("\nTo monitor progress: api-cli monitor-execution %s\n", execution.ID)
				}

			default: // human
				fmt.Printf("Execution started!\n")
				fmt.Printf("ID: %s\n", execution.ID)
				fmt.Printf("Goal: %d\n", execution.GoalID)
				if snapshotID != "" {
					fmt.Printf("Snapshot: %s\n", execution.SnapshotID)
				}
				fmt.Printf("Status: %s\n", execution.Status)
			}

			// Wait for completion if requested
			if waitFlag {
				fmt.Printf("\nWaiting for execution to complete (timeout: %ds)...\n", timeoutFlag)

				// Monitor execution until complete or timeout
				ticker := time.NewTicker(5 * time.Second)
				defer ticker.Stop()

				startTime := time.Now()
				for {
					select {
					case <-ctx.Done():
						return fmt.Errorf("execution monitoring timed out after %ds", timeoutFlag)
					case <-ticker.C:
						// Get execution status
						status, err := apiClient.GetExecution(execution.ID)
						if err != nil {
							return fmt.Errorf("failed to get execution status: %w", err)
						}

						fmt.Printf("\rProgress: %d%% - Status: %s", status.Progress, status.Status)

						// Check if complete
						if status.Status == "COMPLETED" || status.Status == "FAILED" || status.Status == "CANCELLED" {
							fmt.Printf("\n\nExecution %s!\n", status.Status)
							if status.ResultsURL != "" {
								fmt.Printf("Results: %s\n", status.ResultsURL)
							}
							return nil
						}

						// Check timeout
						if time.Since(startTime) > time.Duration(timeoutFlag)*time.Second {
							return fmt.Errorf("execution did not complete within %ds", timeoutFlag)
						}
					}
				}
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().BoolVar(&waitFlag, "wait", false, "Wait for execution to complete")
	cmd.Flags().IntVar(&timeoutFlag, "timeout", 300, "Timeout in seconds when waiting (default: 300)")

	return cmd
}

// Example of how to register the fixed command
func init() {
	// In register.go, you would replace the old command:
	// rootCmd.AddCommand(NewExecuteGoalCmd())
	// With:
	// rootCmd.AddCommand(NewExecuteGoalFixedCmd())
}
