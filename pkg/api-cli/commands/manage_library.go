package commands

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

// LibraryCmd creates the library command with subcommands
func LibraryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "library",
		Short: "Manage library checkpoints",
		Long: `Manage library checkpoints for reusable test components.

This command allows you to:
- Add checkpoints to the library for reuse
- Get details of library checkpoints
- Attach library checkpoints to journeys
- Move test steps within library checkpoints
- Remove test steps from library checkpoints
- Update library checkpoint titles`,
	}

	// Add subcommands
	cmd.AddCommand(addSubCmd())
	cmd.AddCommand(getSubCmd())
	cmd.AddCommand(attachSubCmd())
	cmd.AddCommand(moveStepSubCmd())
	cmd.AddCommand(removeStepSubCmd())
	cmd.AddCommand(updateSubCmd())

	return cmd
}

// addSubCmd creates the add subcommand
func addSubCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <checkpoint-id>",
		Short: "Add a checkpoint to the library",
		Long: `Add an existing checkpoint to the library for reuse across journeys.

This converts a regular checkpoint into a library checkpoint that can be
reused in multiple test journeys.

Examples:
  # Add checkpoint to library
  api-cli library add 1680930
  api-cli library add cp_1680930`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLibraryAddCommand(cmd, args)
		},
	}

	return cmd
}

// getSubCmd creates the get subcommand
func getSubCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <library-checkpoint-id>",
		Short: "Get details of a library checkpoint",
		Long: `Retrieve details of a library checkpoint including its steps and metadata.

Examples:
  # Get library checkpoint details
  api-cli library get 7023
  api-cli library get lib_7023`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLibraryGetCommand(cmd, args)
		},
	}

	return cmd
}

// attachSubCmd creates the attach subcommand
func attachSubCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attach <journey-id> <library-checkpoint-id> <position>",
		Short: "Attach a library checkpoint to a journey",
		Long: `Attach a library checkpoint to a journey at a specific position.

This creates an instance of the library checkpoint in the specified journey.

Examples:
  # Attach library checkpoint to journey
  api-cli library attach 608926 7023 4
  api-cli library attach journey_608926 lib_7023 2`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLibraryAttachCommand(cmd, args)
		},
	}

	return cmd
}

// moveStepSubCmd creates the move-step subcommand
func moveStepSubCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "move-step <library-checkpoint-id> <test-step-id> <position>",
		Short: "Move a test step within a library checkpoint",
		Long: `Move a test step to a new position within a library checkpoint.

The position is 1-based, where 1 is the first position.

Examples:
  # Move step to position 2
  api-cli library move-step 7023 19660498 2
  api-cli library move-step lib_7023 step_19660498 1`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLibraryMoveStepCommand(cmd, args)
		},
	}

	return cmd
}

// removeStepSubCmd creates the remove-step subcommand
func removeStepSubCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-step <library-checkpoint-id> <test-step-id>",
		Short: "Remove a test step from a library checkpoint",
		Long: `Remove a test step from a library checkpoint.

This permanently removes the step from the library checkpoint.

Examples:
  # Remove step from library checkpoint
  api-cli library remove-step 7023 19660498
  api-cli library remove-step lib_7023 step_19660498`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLibraryRemoveStepCommand(cmd, args)
		},
	}

	return cmd
}

// updateSubCmd creates the update subcommand
func updateSubCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <library-checkpoint-id> <new-title>",
		Short: "Update a library checkpoint's title",
		Long: `Update the title (name) of a library checkpoint.

Examples:
  # Update library checkpoint title
  api-cli library update 7023 "New Checkpoint Title"
  api-cli library update lib_7023 "Updated Test Flow"`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLibraryUpdateCommand(cmd, args)
		},
	}

	return cmd
}

// runLibraryAddCommand executes the library add command
func runLibraryAddCommand(cmd *cobra.Command, args []string) error {
	base := NewBaseCommand()
	if err := base.Init(cmd); err != nil {
		return err
	}

	// Parse checkpoint ID
	checkpointIDStr := stripPrefix(args[0], "cp_")
	checkpointID, err := strconv.Atoi(checkpointIDStr)
	if err != nil {
		return fmt.Errorf("invalid checkpoint ID: %w", err)
	}

	// Add checkpoint to library
	libraryCheckpoint, err := base.Client.AddCheckpointToLibrary(checkpointID)
	if err != nil {
		return fmt.Errorf("failed to add checkpoint to library: %w", err)
	}

	// Format result
	result := map[string]interface{}{
		"library_checkpoint_id":  libraryCheckpoint.ID,
		"original_checkpoint_id": checkpointID,
		"name":                   libraryCheckpoint.Name,
		"description":            libraryCheckpoint.Description,
		"createdAt":              libraryCheckpoint.CreatedAt,
		"message":                fmt.Sprintf("✅ Added checkpoint %d to library as library checkpoint %d", checkpointID, libraryCheckpoint.ID),
		"next_steps": []string{
			fmt.Sprintf("Use 'library get %d' to view details", libraryCheckpoint.ID),
			fmt.Sprintf("Use 'library attach <journey-id> %d <position>' to attach to a journey", libraryCheckpoint.ID),
		},
	}

	output, err := base.FormatOutput(result, base.OutputFormat)
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	fmt.Println(output)
	return nil
}

// runLibraryGetCommand executes the library get command
func runLibraryGetCommand(cmd *cobra.Command, args []string) error {
	base := NewBaseCommand()
	if err := base.Init(cmd); err != nil {
		return err
	}

	// Parse library checkpoint ID
	libraryCheckpointIDStr := stripPrefix(args[0], "lib_")
	libraryCheckpointID, err := strconv.Atoi(libraryCheckpointIDStr)
	if err != nil {
		return fmt.Errorf("invalid library checkpoint ID: %w", err)
	}

	// Get library checkpoint details
	libraryCheckpoint, err := base.Client.GetLibraryCheckpoint(libraryCheckpointID)
	if err != nil {
		return fmt.Errorf("failed to get library checkpoint %d: %w\nNote: Make sure the checkpoint has been added to the library using 'library add'", libraryCheckpointID, err)
	}

	// Format result
	result := map[string]interface{}{
		"id":          libraryCheckpoint.ID,
		"name":        libraryCheckpoint.Name,
		"description": libraryCheckpoint.Description,
		"steps":       libraryCheckpoint.Steps,
		"createdAt":   libraryCheckpoint.CreatedAt,
		"updatedAt":   libraryCheckpoint.UpdatedAt,
		"stepCount":   len(libraryCheckpoint.Steps),
	}

	output, err := base.FormatOutput(result, base.OutputFormat)
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	fmt.Println(output)
	return nil
}

// runLibraryAttachCommand executes the library attach command
func runLibraryAttachCommand(cmd *cobra.Command, args []string) error {
	base := NewBaseCommand()
	if err := base.Init(cmd); err != nil {
		return err
	}

	// Parse journey ID
	journeyIDStr := stripPrefix(args[0], "journey_")
	journeyID, err := strconv.Atoi(journeyIDStr)
	if err != nil {
		return fmt.Errorf("invalid journey ID: %w", err)
	}

	// Parse library checkpoint ID
	libraryCheckpointIDStr := stripPrefix(args[1], "lib_")
	libraryCheckpointID, err := strconv.Atoi(libraryCheckpointIDStr)
	if err != nil {
		return fmt.Errorf("invalid library checkpoint ID: %w", err)
	}

	// Parse position
	position, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid position: %w", err)
	}

	if position < 1 {
		return fmt.Errorf("position must be 1 or greater")
	}

	// Attach library checkpoint to journey
	checkpoint, err := base.Client.AttachLibraryCheckpoint(journeyID, libraryCheckpointID, position)
	if err != nil {
		return fmt.Errorf("failed to attach library checkpoint %d to journey %d: %w\nNote: Make sure the checkpoint exists in the library (use 'library get' to verify)", libraryCheckpointID, journeyID, err)
	}

	// Format result
	result := map[string]interface{}{
		"checkpointId":        checkpoint.ID,
		"journeyId":           journeyID,
		"libraryCheckpointId": libraryCheckpointID,
		"position":            position,
		"title":               checkpoint.Title,
		"message":             fmt.Sprintf("✅ Attached library checkpoint %d to journey %d at position %d", libraryCheckpointID, journeyID, position),
	}

	output, err := base.FormatOutput(result, base.OutputFormat)
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	fmt.Println(output)
	return nil
}

// runLibraryMoveStepCommand executes the library move-step command
func runLibraryMoveStepCommand(cmd *cobra.Command, args []string) error {
	base := NewBaseCommand()
	if err := base.Init(cmd); err != nil {
		return err
	}

	// Parse library checkpoint ID
	libraryCheckpointIDStr := stripPrefix(args[0], "lib_")
	libraryCheckpointID, err := strconv.Atoi(libraryCheckpointIDStr)
	if err != nil {
		return fmt.Errorf("invalid library checkpoint ID: %w", err)
	}

	// Parse test step ID
	testStepIDStr := stripPrefix(args[1], "step_")
	testStepID, err := strconv.Atoi(testStepIDStr)
	if err != nil {
		return fmt.Errorf("invalid test step ID: %w", err)
	}

	// Parse position
	position, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid position: %w", err)
	}

	if position < 1 {
		return fmt.Errorf("position must be 1 or greater")
	}

	// Move the step
	err = base.Client.MoveLibraryCheckpointStep(libraryCheckpointID, testStepID, position)
	if err != nil {
		return fmt.Errorf("failed to move library checkpoint step: %w", err)
	}

	// Format result
	result := map[string]interface{}{
		"libraryCheckpointId": libraryCheckpointID,
		"testStepId":          testStepID,
		"position":            position,
		"message":             fmt.Sprintf("✅ Moved step %d to position %d in library checkpoint %d", testStepID, position, libraryCheckpointID),
	}

	output, err := base.FormatOutput(result, base.OutputFormat)
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	fmt.Println(output)
	return nil
}

// runLibraryRemoveStepCommand executes the library remove-step command
func runLibraryRemoveStepCommand(cmd *cobra.Command, args []string) error {
	base := NewBaseCommand()
	if err := base.Init(cmd); err != nil {
		return err
	}

	// Parse library checkpoint ID
	libraryCheckpointIDStr := stripPrefix(args[0], "lib_")
	libraryCheckpointID, err := strconv.Atoi(libraryCheckpointIDStr)
	if err != nil {
		return fmt.Errorf("invalid library checkpoint ID: %w", err)
	}

	// Parse test step ID
	testStepIDStr := stripPrefix(args[1], "step_")
	testStepID, err := strconv.Atoi(testStepIDStr)
	if err != nil {
		return fmt.Errorf("invalid test step ID: %w", err)
	}

	// Remove the step
	err = base.Client.RemoveLibraryCheckpointStep(libraryCheckpointID, testStepID)
	if err != nil {
		return fmt.Errorf("failed to remove library checkpoint step: %w", err)
	}

	// Format result
	result := map[string]interface{}{
		"libraryCheckpointId": libraryCheckpointID,
		"testStepId":          testStepID,
		"message":             fmt.Sprintf("✅ Removed step %d from library checkpoint %d", testStepID, libraryCheckpointID),
	}

	output, err := base.FormatOutput(result, base.OutputFormat)
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	fmt.Println(output)
	return nil
}

// runLibraryUpdateCommand executes the library update command
func runLibraryUpdateCommand(cmd *cobra.Command, args []string) error {
	base := NewBaseCommand()
	if err := base.Init(cmd); err != nil {
		return err
	}

	// Parse library checkpoint ID
	libraryCheckpointIDStr := stripPrefix(args[0], "lib_")
	libraryCheckpointID, err := strconv.Atoi(libraryCheckpointIDStr)
	if err != nil {
		return fmt.Errorf("invalid library checkpoint ID: %w", err)
	}

	// Get new title
	newTitle := args[1]
	if newTitle == "" {
		return fmt.Errorf("title cannot be empty")
	}

	// Update the library checkpoint
	updatedCheckpoint, err := base.Client.UpdateLibraryCheckpoint(libraryCheckpointID, newTitle)
	if err != nil {
		return fmt.Errorf("failed to update library checkpoint: %w", err)
	}

	// Format result
	result := map[string]interface{}{
		"id":          updatedCheckpoint.ID,
		"name":        updatedCheckpoint.Name,
		"description": updatedCheckpoint.Description,
		"updatedAt":   updatedCheckpoint.UpdatedAt,
		"message":     fmt.Sprintf("✅ Updated library checkpoint %d with new title: %s", libraryCheckpointID, newTitle),
	}

	output, err := base.FormatOutput(result, base.OutputFormat)
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	fmt.Println(output)
	return nil
}

// stripPrefix removes a prefix from a string if present
func stripPrefix(s, prefix string) string {
	if len(s) > len(prefix) && s[:len(prefix)] == prefix {
		return s[len(prefix):]
	}
	return s
}
