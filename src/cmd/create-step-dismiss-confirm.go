package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepDismissConfirmCmd() *cobra.Command {
	var accept bool
	
	cmd := &cobra.Command{
		Use:   "create-step-dismiss-confirm CHECKPOINT_ID POSITION",
		Short: "Create a dismiss confirm dialog step at a specific position in a checkpoint",
		Long: `Create a dismiss confirm dialog step that handles a JavaScript confirm dialog at the specified position in the checkpoint.
		
Example:
  api-cli create-step-dismiss-confirm 1678318 1 --accept
  api-cli create-step-dismiss-confirm 1678318 2 --cancel -o json`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkpointIDStr := args[0]
			positionStr := args[1]
			
			// Convert IDs to int
			checkpointID, err := strconv.Atoi(checkpointIDStr)
			if err != nil {
				return fmt.Errorf("invalid checkpoint ID: %w", err)
			}
			
			position, err := strconv.Atoi(positionStr)
			if err != nil {
				return fmt.Errorf("invalid position: %w", err)
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create dismiss confirm step using the enhanced client
			stepID, err := client.CreateDismissConfirmStep(checkpointID, accept, position)
			if err != nil {
				return fmt.Errorf("failed to create dismiss confirm step: %w", err)
			}
			
			action := "cancel"
			if accept {
				action = "accept"
			}
			
			// Format output based on the format flag
			switch cfg.Output.DefaultFormat {
			case "json":
				output := map[string]interface{}{
					"status":        "success",
					"step_type":     "DISMISS_CONFIRM",
					"checkpoint_id": checkpointID,
					"step_id":       stepID,
					"action":        action,
					"position":      position,
					"parsed_step":   fmt.Sprintf("dismiss confirm dialog (%s)", action),
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(output); err != nil {
					return fmt.Errorf("failed to encode JSON output: %w", err)
				}
			case "yaml":
				fmt.Printf("status: success\n")
				fmt.Printf("step_type: DISMISS_CONFIRM\n")
				fmt.Printf("checkpoint_id: %d\n", checkpointID)
				fmt.Printf("step_id: %d\n", stepID)
				fmt.Printf("action: %s\n", action)
				fmt.Printf("position: %d\n", position)
				fmt.Printf("parsed_step: dismiss confirm dialog (%s)\n", action)
			case "ai":
				fmt.Printf("Successfully created dismiss confirm step:\n")
				fmt.Printf("- Step ID: %d\n", stepID)
				fmt.Printf("- Step Type: DISMISS_CONFIRM\n")
				fmt.Printf("- Checkpoint ID: %d\n", checkpointID)
				fmt.Printf("- Action: %s\n", action)
				fmt.Printf("- Position: %d\n", position)
				fmt.Printf("- Parsed Step: dismiss confirm dialog (%s)\n", action)
				fmt.Printf("\nNext steps:\n")
				fmt.Printf("1. Add another step: api-cli create-step-* %d <options>\n", checkpointID)
				fmt.Printf("2. Execute the test journey\n")
			default: // human
				fmt.Printf("✅ Created dismiss confirm step at position %d in checkpoint %d\n", position, checkpointID)
				fmt.Printf("   Action: %s\n", action)
				fmt.Printf("   Step ID: %d\n", stepID)
			}
			
			return nil
		},
	}
	
	cmd.Flags().BoolVar(&accept, "accept", false, "Accept the confirm dialog (default is cancel)")
	cmd.Flags().Bool("cancel", false, "Cancel the confirm dialog (default)")
	
	return cmd
}
