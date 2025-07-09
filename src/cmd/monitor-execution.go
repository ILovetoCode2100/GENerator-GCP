package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/marklovelady/api-cli-generator/pkg/config"
	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

// MonitorOutput represents the output structure for execution monitoring
type MonitorOutput struct {
	Status       string                      `json:"status"`
	ExecutionID  string                      `json:"execution_id"`
	CurrentState string                      `json:"current_state"`
	Progress     *virtuoso.ExecutionProgress `json:"progress,omitempty"`
	Duration     string                      `json:"duration,omitempty"`
	EstimatedEnd *time.Time                  `json:"estimated_end,omitempty"`
	LastUpdated  time.Time                   `json:"last_updated"`
	NextSteps    []string                    `json:"next_steps,omitempty"`
}

func newMonitorExecutionCmd() *cobra.Command {
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
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			executionID := args[0]
			client := virtuoso.NewClient(cfg)

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

// followExecution follows an execution with real-time updates
func followExecution(client *virtuoso.Client, executionID string, interval, timeout int, format string) error {
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
func displayExecutionStatus(execution *virtuoso.Execution) {
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
			fmt.Printf("\n💡 Next steps:\n")
			for _, step := range output.NextSteps {
				fmt.Printf("  • %s\n", step)
			}
		}
	}

	return nil
}