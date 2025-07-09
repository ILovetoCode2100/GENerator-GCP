package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	
	"github.com/spf13/cobra"
)

// StepContext holds the resolved checkpoint and position for a step command
type StepContext struct {
	CheckpointID   int
	Position       int
	UsingContext   bool
	AutoPosition   bool
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
}

// outputStepResult outputs the step creation result in the specified format
func outputStepResult(output *StepOutput) error {
	switch cfg.Output.DefaultFormat {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(output); err != nil {
			return fmt.Errorf("failed to encode JSON output: %w", err)
		}
	case "yaml":
		fmt.Printf("status: %s\n", output.Status)
		fmt.Printf("step_type: %s\n", output.StepType)
		fmt.Printf("checkpoint_id: %d\n", output.CheckpointID)
		fmt.Printf("step_id: %d\n", output.StepID)
		fmt.Printf("position: %d\n", output.Position)
		fmt.Printf("parsed_step: %s\n", output.ParsedStep)
		fmt.Printf("using_context: %t\n", output.UsingContext)
		fmt.Printf("auto_position: %t\n", output.AutoPosition)
		if output.Extra != nil {
			fmt.Printf("extra: %v\n", output.Extra)
		}
	case "ai":
		fmt.Printf("Successfully created %s step:\n", output.StepType)
		fmt.Printf("- Step ID: %d\n", output.StepID)
		fmt.Printf("- Step Type: %s\n", output.StepType)
		fmt.Printf("- Checkpoint ID: %d\n", output.CheckpointID)
		fmt.Printf("- Position: %d\n", output.Position)
		fmt.Printf("- Parsed Step: %s\n", output.ParsedStep)
		if output.UsingContext {
			fmt.Printf("- Used session context checkpoint\n")
		}
		if output.AutoPosition {
			fmt.Printf("- Auto-incremented position\n")
		}
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("1. Add another step: api-cli create-step-* (uses checkpoint %d)\n", output.CheckpointID)
		fmt.Printf("2. Execute the test journey\n")
	default: // human
		fmt.Printf("✅ Created %s step at position %d in checkpoint %d\n", output.StepType, output.Position, output.CheckpointID)
		fmt.Printf("   Step ID: %d\n", output.StepID)
		fmt.Printf("   Parsed Step: %s\n", output.ParsedStep)
		if output.UsingContext {
			fmt.Printf("   🎯 Used session context checkpoint\n")
		}
		if output.AutoPosition {
			fmt.Printf("   🔄 Auto-incremented position\n")
		}
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