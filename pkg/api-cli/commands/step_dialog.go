package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// DialogCommand implements the dialog command group using BaseCommand pattern
type DialogCommand struct {
	*BaseCommand
	dialogType string
}

// dialogV2Config contains configuration for each dialog type
type dialogV2Config struct {
	stepType     string
	description  string
	usage        string
	examples     []string
	requiredArgs int
	acceptFlag   bool // Whether this dialog type supports accept/reject flags
	buildMeta    func(args []string, accept bool) map[string]interface{}
}

// dialogV2Configs maps dialog operations to their configurations
var dialogV2Configs = map[string]dialogV2Config{
	"dismiss-alert": {
		stepType:    "DISMISS_ALERT",
		description: "Dismiss JavaScript alert dialog",
		usage:       "dialog dismiss-alert [checkpoint-id] [position]",
		examples: []string{
			`api-cli dialog dismiss-alert cp_12345 1`,
			`api-cli dialog dismiss-alert  # Uses session context and auto-increment`,
		},
		requiredArgs: 0,
		acceptFlag:   false,
		buildMeta: func(args []string, accept bool) map[string]interface{} {
			return map[string]interface{}{
				"action": "dismiss",
			}
		},
	},
	"dismiss-confirm": {
		stepType:    "DISMISS_CONFIRM",
		description: "Dismiss JavaScript confirmation dialog",
		usage:       "dialog dismiss-confirm [checkpoint-id] [position]",
		examples: []string{
			`api-cli dialog dismiss-confirm cp_12345 1`,
			`api-cli dialog dismiss-confirm --accept  # Accept and auto-increment position`,
			`api-cli dialog dismiss-confirm --reject  # Reject (default)`,
		},
		requiredArgs: 0,
		acceptFlag:   true,
		buildMeta: func(args []string, accept bool) map[string]interface{} {
			return map[string]interface{}{
				"action": map[string]bool{
					"accept": accept,
					"reject": !accept,
				},
			}
		},
	},
	"dismiss-prompt": {
		stepType:    "DISMISS_PROMPT",
		description: "Dismiss JavaScript prompt dialog without text",
		usage:       "dialog dismiss-prompt [checkpoint-id] [position]",
		examples: []string{
			`api-cli dialog dismiss-prompt cp_12345 1`,
			`api-cli dialog dismiss-prompt --accept  # Accept and auto-increment position`,
			`api-cli dialog dismiss-prompt --reject  # Reject (default)`,
		},
		requiredArgs: 0,
		acceptFlag:   true,
		buildMeta: func(args []string, accept bool) map[string]interface{} {
			return map[string]interface{}{
				"action": map[string]bool{
					"accept": accept,
					"reject": !accept,
				},
				"text": "",
			}
		},
	},
	"dismiss-prompt-with-text": {
		stepType:    "DISMISS_PROMPT_WITH_TEXT",
		description: "Dismiss JavaScript prompt dialog with text input",
		usage:       "dialog dismiss-prompt-with-text [checkpoint-id] <text> [position]",
		examples: []string{
			`api-cli dialog dismiss-prompt-with-text cp_12345 "My input text" 1`,
			`api-cli dialog dismiss-prompt-with-text "Answer" --accept  # Accept with text`,
			`api-cli dialog dismiss-prompt-with-text "Text" --reject  # Reject (text ignored)`,
		},
		requiredArgs: 1,
		acceptFlag:   true,
		buildMeta: func(args []string, accept bool) map[string]interface{} {
			return map[string]interface{}{
				"text": args[0],
				"action": map[string]bool{
					"accept": accept,
					"reject": !accept,
				},
			}
		},
	},
}

// newStepDialogCmd creates the new dialog command using BaseCommand pattern
func newStepDialogCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "step-dialog",
		Short: "Create dialog interaction steps in checkpoints",
		Long: `Create various types of dialog interaction steps in checkpoints.

This command uses the standardized positional argument pattern:
- Optional checkpoint ID as first argument (falls back to session context)
- Required dialog arguments (text for prompt-with-text)
- Optional position as last argument (auto-increments if not specified)

Available dialog operations:
  - dismiss-alert: Dismiss JavaScript alert dialogs
  - dismiss-confirm: Dismiss confirmation dialogs (with accept/reject option)
  - dismiss-prompt: Dismiss prompt dialogs without text (with accept/reject option)
  - dismiss-prompt-with-text: Dismiss prompt dialogs with text input`,
		Example: `  # Dismiss alert dialog (with explicit checkpoint)
  api-cli dialog dismiss-alert cp_12345 1

  # Dismiss alert dialog (using session context)
  api-cli dialog dismiss-alert

  # Accept confirmation dialog
  api-cli dialog dismiss-confirm --accept

  # Reject prompt dialog
  api-cli dialog dismiss-prompt --reject

  # Accept prompt with text
  api-cli dialog dismiss-prompt-with-text "My answer" --accept`,
	}

	// Add subcommands for each dialog type
	for dialogType, config := range dialogV2Configs {
		cmd.AddCommand(newDialogV2SubCmd(dialogType, config))
	}

	return cmd
}

// newDialogV2SubCmd creates a subcommand for a specific dialog type
func newDialogV2SubCmd(dialogType string, config dialogV2Config) *cobra.Command {
	var acceptFlag, rejectFlag bool

	cmd := &cobra.Command{
		Use:   dialogType + " " + extractDialogUsageArgs(config.usage),
		Short: config.description,
		Long: fmt.Sprintf(`%s

%s

Examples:
%s`, config.description, config.usage, strings.Join(config.examples, "\n")),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate accept/reject flags
			if acceptFlag && rejectFlag {
				return fmt.Errorf("cannot specify both --accept and --reject")
			}

			// Default to reject if neither specified and dialog supports it
			accept := acceptFlag
			if config.acceptFlag && !acceptFlag && !rejectFlag {
				accept = false // Default to reject/cancel
			}

			dc := &DialogCommand{
				BaseCommand: NewBaseCommand(),
				dialogType:  dialogType,
			}
			return dc.Execute(cmd, args, config, accept)
		},
	}

	// Add accept/reject flags for dialogs that support them
	if config.acceptFlag {
		cmd.Flags().BoolVar(&acceptFlag, "accept", false, "Accept the dialog (default is reject)")
		cmd.Flags().BoolVar(&rejectFlag, "reject", false, "Reject/cancel the dialog (default)")
	}

	return cmd
}

// extractDialogUsageArgs extracts the arguments portion from the usage string
func extractDialogUsageArgs(usage string) string {
	parts := strings.Fields(usage)
	if len(parts) > 2 {
		// Skip "dialog" and subcommand
		return strings.Join(parts[2:], " ")
	}
	return ""
}

// Execute runs the dialog command
func (dc *DialogCommand) Execute(cmd *cobra.Command, args []string, config dialogV2Config, accept bool) error {
	// Initialize base command
	if err := dc.Init(cmd); err != nil {
		return err
	}

	// Resolve checkpoint and position
	remainingArgs, err := dc.ResolveCheckpointAndPosition(args, config.requiredArgs)
	if err != nil {
		return fmt.Errorf("failed to resolve arguments: %w", err)
	}

	// Validate we have the required number of arguments
	if len(remainingArgs) != config.requiredArgs {
		return fmt.Errorf("expected %d arguments, got %d", config.requiredArgs, len(remainingArgs))
	}

	// Validate dialog-specific arguments
	if err := dc.validateDialogArgs(config, remainingArgs); err != nil {
		return err
	}

	// Build request metadata
	meta := config.buildMeta(remainingArgs, accept)

	// Create the step
	stepResult, err := dc.createDialogStep(config.stepType, meta, accept)
	if err != nil {
		return fmt.Errorf("failed to create %s step: %w", config.stepType, err)
	}

	// Format and output the result
	output, err := dc.FormatOutput(stepResult, dc.OutputFormat)
	if err != nil {
		return err
	}

	fmt.Print(output)
	return nil
}

// validateDialogArgs validates arguments for a specific dialog type
func (dc *DialogCommand) validateDialogArgs(config dialogV2Config, args []string) error {
	switch dc.dialogType {
	case "dismiss-prompt-with-text":
		if len(args) > 0 && args[0] == "" {
			return fmt.Errorf("text cannot be empty")
		}
	}
	return nil
}

// createDialogStep creates a dialog step via the API
func (dc *DialogCommand) createDialogStep(stepType string, meta map[string]interface{}, accept bool) (*StepResult, error) {
	// Convert checkpoint ID from string to int
	checkpointID, err := strconv.Atoi(dc.CheckpointID)
	if err != nil {
		return nil, fmt.Errorf("invalid checkpoint ID: %s", dc.CheckpointID)
	}

	// Build the request based on step type
	var stepID int

	// Use the client to create the appropriate step
	switch stepType {
	case "DISMISS_ALERT":
		stepID, err = dc.Client.CreateDismissAlertStep(checkpointID, dc.Position)
	case "DISMISS_CONFIRM":
		stepID, err = dc.Client.CreateDismissConfirmStep(checkpointID, accept, dc.Position)
	case "DISMISS_PROMPT":
		// Use empty text for simple prompt dismissal
		stepID, err = dc.Client.CreateDismissPromptStep(checkpointID, "", dc.Position)
	case "DISMISS_PROMPT_WITH_TEXT":
		// Pass the text from meta
		text := ""
		if textVal, ok := meta["text"].(string); ok {
			text = textVal
		}
		stepID, err = dc.Client.CreateDismissPromptStep(checkpointID, text, dc.Position)
	default:
		return nil, fmt.Errorf("unknown dialog type: %s", stepType)
	}

	if err != nil {
		return nil, err
	}

	// Build the result
	result := &StepResult{
		ID:           fmt.Sprintf("%d", stepID),
		CheckpointID: dc.CheckpointID,
		Type:         stepType,
		Position:     dc.Position,
		Description:  dc.buildDescription(stepType, meta, accept),
		Meta:         meta,
	}

	// Save session state if position was auto-incremented
	if dc.Position == -1 && cfg.Session.AutoIncrementPos {
		if err := cfg.SaveConfig(); err != nil {
			// Don't fail the command, just warn
			// Note: In production, this warning would be sent to stderr
		}
	}

	return result, nil
}

// buildDescription creates a human-readable description for the step
func (dc *DialogCommand) buildDescription(stepType string, meta map[string]interface{}, accept bool) string {
	switch stepType {
	case "DISMISS_ALERT":
		return "dismiss alert"
	case "DISMISS_CONFIRM":
		if accept {
			return "accept confirm dialog"
		}
		return "cancel confirm dialog"
	case "DISMISS_PROMPT":
		if accept {
			return "accept prompt dialog"
		}
		return "cancel prompt dialog"
	case "DISMISS_PROMPT_WITH_TEXT":
		text := ""
		if textVal, ok := meta["text"].(string); ok {
			text = textVal
		}
		if accept {
			return fmt.Sprintf("accept prompt dialog with text \"%s\"", text)
		}
		return fmt.Sprintf("cancel prompt dialog (text: \"%s\")", text)
	default:
		return stepType
	}
}
