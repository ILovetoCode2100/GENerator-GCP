package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// StepContext holds the resolved checkpoint and position for a step command
type StepContext struct {
	CheckpointID int
	Position     int
	UsingContext bool
	AutoPosition bool
}

// resolveStepContext resolves the checkpoint ID and position for step commands
// It handles both explicit arguments and session context
func resolveStepContext(args []string, checkpointFlag int, positionIndex int) (*StepContext, error) {
	ctx := &StepContext{}

	// Determine checkpoint ID
	if checkpointFlag > 0 {
		// Use explicit checkpoint from flag
		ctx.CheckpointID = checkpointFlag
		ctx.UsingContext = false
	} else if cfg.GetCurrentCheckpoint() != nil {
		// Use current checkpoint from session
		ctx.CheckpointID = *cfg.GetCurrentCheckpoint()
		ctx.UsingContext = true
	} else {
		return nil, fmt.Errorf("no checkpoint specified - use --checkpoint flag or set current checkpoint with 'api-cli set-checkpoint CHECKPOINT_ID'")
	}

	// Determine position
	if positionIndex < len(args) {
		// Position provided as argument
		var err error
		ctx.Position, err = parseIntArg(args[positionIndex], "position")
		if err != nil {
			return nil, err
		}
		ctx.AutoPosition = false
	} else {
		// Use auto-increment from session
		ctx.Position = cfg.GetNextPosition()
		ctx.AutoPosition = true
	}

	return ctx, nil
}

// saveStepContext saves the session state if position was auto-incremented
func saveStepContext(ctx *StepContext) {
	if ctx.AutoPosition && cfg.Session.AutoIncrementPos {
		if err := cfg.SaveConfig(); err != nil {
			// Don't fail the command, just warn
			fmt.Fprintf(os.Stderr, "Warning: failed to save session state: %v\n", err)
		}
	}
}

// StepOutput holds the output data for step commands
type StepOutput struct {
	Status       string      `json:"status"`
	StepType     string      `json:"step_type"`
	CheckpointID int         `json:"checkpoint_id"`
	StepID       int         `json:"step_id"`
	Position     int         `json:"position"`
	ParsedStep   string      `json:"parsed_step"`
	UsingContext bool        `json:"using_context"`
	AutoPosition bool        `json:"auto_position"`
	Extra        interface{} `json:"extra,omitempty"`
	Timestamp    string      `json:"timestamp"`
}

// validateOutputFormat validates that the output format is supported
func validateOutputFormat(format string) error {
	supportedFormats := []string{"human", "json", "yaml", "ai"}
	for _, supported := range supportedFormats {
		if format == supported {
			return nil
		}
	}
	return fmt.Errorf("unsupported output format '%s'. Supported formats: %s", format, strings.Join(supportedFormats, ", "))
}

// outputStepResult outputs the step creation result in the specified format
func outputStepResult(output *StepOutput) error {
	// Set timestamp if not already set
	if output.Timestamp == "" {
		output.Timestamp = time.Now().Format(time.RFC3339)
	}

	// Validate output format
	if err := validateOutputFormat(cfg.Output.DefaultFormat); err != nil {
		return err
	}

	switch cfg.Output.DefaultFormat {
	case "json":
		return outputStepResultJSON(output)
	case "yaml":
		return outputStepResultYAML(output)
	case "ai":
		return outputStepResultAI(output)
	default: // human
		return outputStepResultHuman(output)
	}
}

// outputStepResultJSON outputs step result in JSON format with rich metadata
func outputStepResultJSON(output *StepOutput) error {
	// Create enhanced JSON structure
	result := map[string]interface{}{
		"metadata": map[string]interface{}{
			"timestamp": output.Timestamp,
			"format":    "json",
			"version":   "1.0",
		},
		"step": map[string]interface{}{
			"id":            output.StepID,
			"type":          output.StepType,
			"position":      output.Position,
			"checkpoint_id": output.CheckpointID,
			"parsed_step":   output.ParsedStep,
			"status":        output.Status,
		},
		"context": map[string]interface{}{
			"using_session_context": output.UsingContext,
			"auto_position":         output.AutoPosition,
		},
	}

	if output.Extra != nil {
		result["extra"] = output.Extra
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		return fmt.Errorf("failed to encode JSON output: %w", err)
	}
	return nil
}

// outputStepResultYAML outputs step result in YAML format with structured data
func outputStepResultYAML(output *StepOutput) error {
	fmt.Printf("# Virtuoso API CLI - Step Creation Result\n")
	fmt.Printf("metadata:\n")
	fmt.Printf("  timestamp: %s\n", output.Timestamp)
	fmt.Printf("  format: yaml\n")
	fmt.Printf("  version: \"1.0\"\n")
	fmt.Printf("\n")
	fmt.Printf("step:\n")
	fmt.Printf("  id: %d\n", output.StepID)
	fmt.Printf("  type: %s\n", output.StepType)
	fmt.Printf("  position: %d\n", output.Position)
	fmt.Printf("  checkpoint_id: %d\n", output.CheckpointID)
	fmt.Printf("  parsed_step: \"%s\"\n", output.ParsedStep)
	fmt.Printf("  status: %s\n", output.Status)
	fmt.Printf("\n")
	fmt.Printf("context:\n")
	fmt.Printf("  using_session_context: %t\n", output.UsingContext)
	fmt.Printf("  auto_position: %t\n", output.AutoPosition)

	if output.Extra != nil {
		fmt.Printf("\n")
		fmt.Printf("extra:\n")
		fmt.Printf("  data: %v\n", output.Extra)
	}

	return nil
}

// outputStepResultAI outputs step result in AI-friendly conversational format
func outputStepResultAI(output *StepOutput) error {
	fmt.Printf("🎯 Step Creation Summary\n")
	fmt.Printf("========================\n\n")

	// Main result
	fmt.Printf("✅ Successfully created a %s step!\n\n", strings.ToUpper(output.StepType))

	// Key details
	fmt.Printf("📋 Step Details:\n")
	fmt.Printf("   • Step ID: %d\n", output.StepID)
	fmt.Printf("   • Type: %s\n", output.StepType)
	fmt.Printf("   • Position: %d (execution order)\n", output.Position)
	fmt.Printf("   • Checkpoint: %d\n", output.CheckpointID)
	fmt.Printf("   • Description: %s\n", output.ParsedStep)
	fmt.Printf("   • Created: %s\n", output.Timestamp)

	// Context information
	fmt.Printf("\n🔄 Context Information:\n")
	if output.UsingContext {
		fmt.Printf("   • ✅ Used session context checkpoint (%d)\n", output.CheckpointID)
	} else {
		fmt.Printf("   • ⚙️  Used explicit checkpoint specification\n")
	}

	if output.AutoPosition {
		fmt.Printf("   • ✅ Auto-incremented position to %d\n", output.Position)
	} else {
		fmt.Printf("   • ⚙️  Used explicit position %d\n", output.Position)
	}

	// Next steps suggestions
	fmt.Printf("\n🚀 What's Next?\n")
	fmt.Printf("   1. Add another step: `api-cli create-step-[type] [args]`\n")
	fmt.Printf("      (will use checkpoint %d and position %d)\n", output.CheckpointID, output.Position+1)
	fmt.Printf("   2. View all steps: `api-cli list-checkpoints [journey_id]`\n")
	fmt.Printf("   3. Execute the test: `api-cli execute-goal [goal_id]`\n")

	// Extra data if available
	if output.Extra != nil {
		fmt.Printf("\n📊 Additional Information:\n")
		fmt.Printf("   %v\n", output.Extra)
	}

	fmt.Printf("\n💡 Tip: Use `--checkpoint [id]` to override session context for specific steps\n")

	return nil
}

// outputStepResultHuman outputs step result in human-friendly format
func outputStepResultHuman(output *StepOutput) error {
	// Status icon and main message
	statusIcon := "✅"
	if output.Status != "success" {
		statusIcon = "❌"
	}

	fmt.Printf("%s Created %s step at position %d\n", statusIcon, output.StepType, output.Position)

	// Core details with visual hierarchy
	fmt.Printf("   📍 Step ID: %d\n", output.StepID)
	fmt.Printf("   🎯 Checkpoint: %d\n", output.CheckpointID)
	fmt.Printf("   📝 Description: %s\n", output.ParsedStep)

	// Context indicators
	contextIndicators := []string{}
	if output.UsingContext {
		contextIndicators = append(contextIndicators, "🔗 session context")
	}
	if output.AutoPosition {
		contextIndicators = append(contextIndicators, "🔄 auto-position")
	}

	if len(contextIndicators) > 0 {
		fmt.Printf("   ⚙️  Context: %s\n", strings.Join(contextIndicators, ", "))
	}

	// Extra information if available
	if output.Extra != nil {
		fmt.Printf("   📊 Extra: %v\n", output.Extra)
	}

	return nil
}

// addCheckpointFlag adds the standard --checkpoint flag to a step command
func addCheckpointFlag(cmd *cobra.Command, checkpointFlag *int) {
	cmd.Flags().IntVar(checkpointFlag, "checkpoint", 0, "Checkpoint ID (overrides session context)")
}

// parseIntArg safely parses an integer argument, handling negative numbers
func parseIntArg(arg string, fieldName string) (int, error) {
	val, err := strconv.Atoi(arg)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", fieldName, err)
	}
	return val, nil
}

// enableNegativeNumbers configures a command to properly handle negative number arguments
func enableNegativeNumbers(cmd *cobra.Command) {
	cmd.FParseErrWhitelist = cobra.FParseErrWhitelist{
		UnknownFlags: true,
	}
}
