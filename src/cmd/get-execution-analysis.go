package main

import (
	"encoding/json"
	"fmt"

	"github.com/marklovelady/api-cli-generator/pkg/config"
	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

// AnalysisOutput represents the output structure for execution analysis
type AnalysisOutput struct {
	Status       string                        `json:"status"`
	ExecutionID  string                        `json:"execution_id"`
	Summary      *virtuoso.ExecutionSummary    `json:"summary"`
	Failures     []virtuoso.ExecutionFailure   `json:"failures"`
	Performance  *virtuoso.ExecutionPerformance `json:"performance,omitempty"`
	AIInsights   []string                      `json:"ai_insights,omitempty"`
	NextSteps    []string                      `json:"next_steps,omitempty"`
}

func newGetExecutionAnalysisCmd() *cobra.Command {
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
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			executionID := args[0]
			client := virtuoso.NewClient(cfg)

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

			return outputAnalysisResult(output, cfg.Output.Format)
		},
	}

	cmd.Flags().BoolVar(&includeAIFlag, "ai-insights", false, "Include AI-generated insights and suggestions")
	cmd.Flags().BoolVar(&failuresOnlyFlag, "failures-only", false, "Show only failure details")

	return cmd
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