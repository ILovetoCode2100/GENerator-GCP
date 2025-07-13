package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newSetCheckpointCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-checkpoint CHECKPOINT_ID",
		Short: "Set the current checkpoint for session context",
		Long: `Set the current checkpoint ID to use as default for step commands.

This command:
- Sets the current checkpoint ID in the session context
- Resets the step position counter to 1
- Saves the session state to the configuration file
- Allows step commands to use the current checkpoint as default

Example:
  api-cli set-checkpoint 1678318
  api-cli create-step-navigate "https://example.com" 1  # Uses checkpoint 1678318
  api-cli create-step-click "Submit" 2                  # Uses checkpoint 1678318`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkpointID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid checkpoint ID: %w", err)
			}

			// Validate checkpoint exists by trying to get it
			client := virtuoso.NewClient(cfg)
			if err := client.ValidateCheckpoint(checkpointID); err != nil {
				return fmt.Errorf("checkpoint validation failed: %w", err)
			}

			// Set current checkpoint
			cfg.SetCurrentCheckpoint(checkpointID)

			// Save configuration
			if err := cfg.SaveConfig(); err != nil {
				return fmt.Errorf("failed to save session state: %w", err)
			}

			// Output results
			switch cfg.Output.DefaultFormat {
			case "json":
				result := map[string]interface{}{
					"status":          "success",
					"checkpoint_id":   checkpointID,
					"next_position":   cfg.Session.NextPosition,
					"auto_increment":  cfg.Session.AutoIncrementPos,
					"session_updated": true,
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				encoder.Encode(result)

			case "yaml":
				fmt.Println("status: success")
				fmt.Printf("checkpoint_id: %d\n", checkpointID)
				fmt.Printf("next_position: %d\n", cfg.Session.NextPosition)
				fmt.Printf("auto_increment: %t\n", cfg.Session.AutoIncrementPos)
				fmt.Println("session_updated: true")

			case "ai":
				fmt.Printf("Successfully set checkpoint %d as current checkpoint!\n", checkpointID)
				fmt.Printf("\nSession Context:\n")
				fmt.Printf("- Current checkpoint: %d\n", checkpointID)
				fmt.Printf("- Next step position: %d\n", cfg.Session.NextPosition)
				fmt.Printf("- Auto-increment position: %t\n", cfg.Session.AutoIncrementPos)
				fmt.Printf("\nNext steps:\n")
				fmt.Printf("1. Create steps without specifying checkpoint ID:\n")
				fmt.Printf("   api-cli create-step-navigate \"https://example.com\" 1\n")
				fmt.Printf("   api-cli create-step-click \"Submit\" 2\n")
				fmt.Printf("2. Override checkpoint for specific steps:\n")
				fmt.Printf("   api-cli create-step-click \"Submit\" 2 --checkpoint 1678319\n")

			default: // human
				fmt.Printf("✅ Current checkpoint set to: %d\n", checkpointID)
				fmt.Printf("📍 Next step position: %d\n", cfg.Session.NextPosition)
				fmt.Printf("🔄 Auto-increment position: %t\n", cfg.Session.AutoIncrementPos)
				fmt.Printf("💾 Session state saved to config file\n")
				fmt.Printf("\nYou can now create steps without specifying checkpoint ID:\n")
				fmt.Printf("  api-cli create-step-navigate \"https://example.com\" 1\n")
				fmt.Printf("  api-cli create-step-click \"Submit\" 2\n")
			}

			return nil
		},
	}

	return cmd
}
