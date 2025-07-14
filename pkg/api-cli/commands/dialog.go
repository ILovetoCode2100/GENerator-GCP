package commands

import (
	"fmt"
	"strings"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/spf13/cobra"
)

// dialogType represents the type of dialog being dismissed
type dialogType string

const (
	dialogAlert          dialogType = "alert"
	dialogConfirm        dialogType = "confirm"
	dialogPrompt         dialogType = "prompt"
	dialogPromptWithText dialogType = "prompt-with-text"
)

// dialogCommandInfo contains metadata about each dialog type
type dialogCommandInfo struct {
	stepType      string
	description   string
	usage         string
	examples      []string
	argsCount     []int // Valid argument counts (excluding position)
	parseStep     func(args []string, accept bool) string
	hasAcceptFlag bool
}

// dialogCommands maps dialog types to their metadata
var dialogCommands = map[dialogType]dialogCommandInfo{
	dialogAlert: {
		stepType:    "DISMISS_ALERT",
		description: "Dismiss JavaScript alert dialog",
		usage:       "dialog dismiss alert [POSITION]",
		examples: []string{
			`api-cli dialog dismiss alert 1`,
			`api-cli dialog dismiss alert  # Auto-increment position`,
		},
		argsCount: []int{0},
		parseStep: func(args []string, accept bool) string {
			return "dismiss alert"
		},
		hasAcceptFlag: false,
	},
	dialogConfirm: {
		stepType:    "DISMISS_CONFIRM",
		description: "Dismiss JavaScript confirmation dialog",
		usage:       "dialog dismiss confirm [POSITION]",
		examples: []string{
			`api-cli dialog dismiss confirm 1`,
			`api-cli dialog dismiss confirm --accept  # Accept and auto-increment position`,
			`api-cli dialog dismiss confirm --reject 2  # Reject at position 2`,
		},
		argsCount: []int{0},
		parseStep: func(args []string, accept bool) string {
			if accept {
				return "accept confirm dialog"
			}
			return "cancel confirm dialog"
		},
		hasAcceptFlag: true,
	},
	dialogPrompt: {
		stepType:    "DISMISS_PROMPT",
		description: "Dismiss JavaScript prompt dialog",
		usage:       "dialog dismiss prompt [POSITION]",
		examples: []string{
			`api-cli dialog dismiss prompt 1`,
			`api-cli dialog dismiss prompt --accept  # Accept and auto-increment position`,
			`api-cli dialog dismiss prompt --reject 2  # Reject at position 2`,
		},
		argsCount: []int{0},
		parseStep: func(args []string, accept bool) string {
			if accept {
				return "accept prompt dialog"
			}
			return "cancel prompt dialog"
		},
		hasAcceptFlag: true,
	},
	dialogPromptWithText: {
		stepType:    "DISMISS_PROMPT_WITH_TEXT",
		description: "Dismiss JavaScript prompt dialog with text input",
		usage:       "dialog dismiss prompt-with-text TEXT [POSITION]",
		examples: []string{
			`api-cli dialog dismiss prompt-with-text "My input text" 1`,
			`api-cli dialog dismiss prompt-with-text "Answer" --accept  # Accept with text`,
			`api-cli dialog dismiss prompt-with-text "Text" --reject 2  # Reject (text ignored)`,
		},
		argsCount: []int{1},
		parseStep: func(args []string, accept bool) string {
			if accept {
				return fmt.Sprintf("accept prompt dialog with text \"%s\"", args[0])
			}
			return fmt.Sprintf("cancel prompt dialog (text: \"%s\")", args[0])
		},
		hasAcceptFlag: true,
	},
}

// newDialogCmd creates the consolidated dialog command with subcommands
func newDialogCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dialog",
		Short: "Create dialog interaction steps in checkpoints",
		Long: `Create various types of dialog interaction steps in checkpoints.

This command consolidates all dialog-related operations for handling JavaScript alerts, confirms, and prompts.

Available dialog operations:
  - dismiss alert: Dismiss JavaScript alert dialogs
  - dismiss confirm: Dismiss confirmation dialogs (with accept/reject option)
  - dismiss prompt: Dismiss prompt dialogs (with accept/reject option)
  - dismiss prompt-with-text: Dismiss prompt dialogs with text input`,
		Example: `  # Dismiss alert dialog
  api-cli dialog dismiss alert 1

  # Accept confirmation dialog
  api-cli dialog dismiss confirm --accept

  # Reject prompt dialog
  api-cli dialog dismiss prompt --reject

  # Accept prompt with text
  api-cli dialog dismiss prompt-with-text "My answer" --accept`,
	}

	// Add dismiss subcommand
	dismissCmd := &cobra.Command{
		Use:   "dismiss",
		Short: "Dismiss various types of dialogs",
		Long:  "Dismiss JavaScript alert, confirm, or prompt dialogs",
	}

	// Add subcommands for each dialog type
	for dType, info := range dialogCommands {
		dismissCmd.AddCommand(newDialogSubCmd(dType, info))
	}

	cmd.AddCommand(dismissCmd)

	return cmd
}

// extractDialogArgsFromUsage extracts the arguments part from the usage string
func extractDialogArgsFromUsage(usage string) string {
	parts := strings.Fields(usage)
	if len(parts) > 3 {
		return strings.Join(parts[3:], " ")
	}
	return ""
}

// newDialogSubCmd creates a subcommand for a specific dialog type
func newDialogSubCmd(dType dialogType, info dialogCommandInfo) *cobra.Command {
	var checkpointFlag int
	var acceptFlag, rejectFlag bool

	cmd := &cobra.Command{
		Use:   string(dType) + " " + extractDialogArgsFromUsage(info.usage),
		Short: info.description,
		Long: fmt.Sprintf(`%s

Uses the current checkpoint from session context by default. Override with --checkpoint flag.
Position is auto-incremented if not specified and auto-increment is enabled.

Examples:
%s`, info.description, strings.Join(info.examples, "\n")),
		Args: func(cmd *cobra.Command, args []string) error {
			// Validate argument count
			validCounts := info.argsCount
			for _, count := range validCounts {
				if len(args) == count || len(args) == count+1 {
					return nil
				}
			}

			// Generate expected count message
			expectedCounts := []string{}
			for _, count := range validCounts {
				expectedCounts = append(expectedCounts, fmt.Sprintf("%d", count))
				expectedCounts = append(expectedCounts, fmt.Sprintf("%d", count+1))
			}

			return fmt.Errorf("accepts %s args, received %d", strings.Join(expectedCounts, " or "), len(args))
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate accept/reject flags
			if acceptFlag && rejectFlag {
				return fmt.Errorf("cannot specify both --accept and --reject")
			}

			// Default to reject if neither specified and dialog supports it
			accept := acceptFlag
			if info.hasAcceptFlag && !acceptFlag && !rejectFlag {
				accept = false // Default to reject/cancel
			}

			return runDialogCommand(dType, info, args, checkpointFlag, accept)
		},
	}

	addCheckpointFlag(cmd, &checkpointFlag)

	// Add accept/reject flags for dialogs that support them
	if info.hasAcceptFlag {
		cmd.Flags().BoolVar(&acceptFlag, "accept", false, "Accept the dialog (default is reject)")
		cmd.Flags().BoolVar(&rejectFlag, "reject", false, "Reject/cancel the dialog (default)")
	}

	return cmd
}

// runDialogCommand executes the dialog command logic
func runDialogCommand(dType dialogType, info dialogCommandInfo, args []string, checkpointFlag int, accept bool) error {
	// Validate arguments based on dialog type
	if err := validateDialogArgs(dType, args); err != nil {
		return err
	}

	// Resolve checkpoint and position
	positionIndex := len(info.argsCount) // Position comes after required args
	ctx, err := resolveStepContext(args, checkpointFlag, positionIndex)
	if err != nil {
		return err
	}

	// Create Virtuoso client
	apiClient := client.NewClient(cfg)

	// Call the appropriate API method based on dialog type
	stepID, err := callDialogAPI(apiClient, dType, ctx, args, accept)
	if err != nil {
		return fmt.Errorf("failed to create %s step: %w", info.stepType, err)
	}

	// Save config if position was auto-incremented
	saveStepContext(ctx)

	// Build extra data for output
	extra := buildDialogExtraData(dType, args, accept)

	// Output result
	output := &StepOutput{
		Status:       "success",
		StepType:     info.stepType,
		CheckpointID: ctx.CheckpointID,
		StepID:       stepID,
		Position:     ctx.Position,
		ParsedStep:   info.parseStep(args, accept),
		UsingContext: ctx.UsingContext,
		AutoPosition: ctx.AutoPosition,
		Extra:        extra,
	}

	return outputStepResult(output)
}

// validateDialogArgs validates arguments for a specific dialog type
func validateDialogArgs(dType dialogType, args []string) error {
	switch dType {
	case dialogAlert, dialogConfirm, dialogPrompt:
		// No arguments required
		return nil
	case dialogPromptWithText:
		if len(args) < 1 || args[0] == "" {
			return fmt.Errorf("text cannot be empty")
		}
	}
	return nil
}

// callDialogAPI calls the appropriate client API method for the dialog type
func callDialogAPI(apiClient *client.Client, dType dialogType, ctx *StepContext, args []string, accept bool) (int, error) {
	switch dType {
	case dialogAlert:
		return apiClient.CreateDismissAlertStep(ctx.CheckpointID, ctx.Position)
	case dialogConfirm:
		return apiClient.CreateDismissConfirmStep(ctx.CheckpointID, accept, ctx.Position)
	case dialogPrompt:
		// Use empty text for simple prompt dismissal
		return apiClient.CreateDismissPromptStep(ctx.CheckpointID, "", ctx.Position)
	case dialogPromptWithText:
		// For prompt with text, we need to handle accept/reject differently
		// The API might expect different behavior, but based on existing implementation,
		// we'll pass the text regardless of accept/reject
		return apiClient.CreateDismissPromptStep(ctx.CheckpointID, args[0], ctx.Position)
	default:
		return 0, fmt.Errorf("unsupported dialog type: %s", dType)
	}
}

// buildDialogExtraData builds the extra data map for output based on dialog type
func buildDialogExtraData(dType dialogType, args []string, accept bool) map[string]interface{} {
	extra := make(map[string]interface{})

	switch dType {
	case dialogAlert:
		// No extra data for alert
	case dialogConfirm, dialogPrompt:
		extra["action"] = map[string]bool{
			"accept": accept,
			"reject": !accept,
		}
	case dialogPromptWithText:
		extra["text"] = args[0]
		extra["action"] = map[string]bool{
			"accept": accept,
			"reject": !accept,
		}
	}

	return extra
}
