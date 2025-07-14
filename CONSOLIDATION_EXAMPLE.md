# Command Consolidation Implementation Example

## Example: Consolidating Assert Commands

This example demonstrates how to consolidate the 12 assert commands into a single `api-cli assert` command with subcommands.

### Current Implementation (12 separate files)

```go
// create-step-assert-equals.go
func newCreateStepAssertEqualsCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "create-step-assert-equals ELEMENT VALUE [POSITION]",
        Short: "Create an assertion step that verifies an element has a specific text value",
        // ... implementation
    }
}

// create-step-assert-exists.go
func newCreateStepAssertExistsCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "create-step-assert-exists ELEMENT [POSITION]",
        Short: "Create an assertion step that verifies an element exists",
        // ... implementation
    }
}

// ... 10 more similar files
```

### New Consolidated Implementation

```go
// pkg/api-cli/commands/consolidated/assert.go
package consolidated

import (
    "fmt"
    "github.com/spf13/cobra"
    "github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
    "github.com/marklovelady/api-cli-generator/pkg/api-cli/commands/shared"
)

// AssertCommand creates the consolidated assert command with subcommands
func NewAssertCommand(cfg *config.VirtuosoConfig) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "assert",
        Short: "Create assertion steps",
        Long:  "Create various types of assertion steps to verify element states and values",
    }

    // Add all assert subcommands
    cmd.AddCommand(newAssertEqualsCmd(cfg))
    cmd.AddCommand(newAssertNotEqualsCmd(cfg))
    cmd.AddCommand(newAssertExistsCmd(cfg))
    cmd.AddCommand(newAssertNotExistsCmd(cfg))
    cmd.AddCommand(newAssertCheckedCmd(cfg))
    cmd.AddCommand(newAssertSelectedCmd(cfg))
    cmd.AddCommand(newAssertVariableCmd(cfg))
    cmd.AddCommand(newAssertGreaterThanCmd(cfg))
    cmd.AddCommand(newAssertGreaterThanOrEqualCmd(cfg))
    cmd.AddCommand(newAssertLessThanCmd(cfg))
    cmd.AddCommand(newAssertLessThanOrEqualCmd(cfg))
    cmd.AddCommand(newAssertMatchesCmd(cfg))

    return cmd
}

// Example: Assert Equals subcommand
func newAssertEqualsCmd(cfg *config.VirtuosoConfig) *cobra.Command {
    var checkpointFlag int

    cmd := &cobra.Command{
        Use:   "equals ELEMENT VALUE [POSITION]",
        Short: "Assert an element has a specific text value",
        Long: `Create an assertion step that verifies an element has a specific text value.

Examples:
  # Using current checkpoint context
  api-cli assert equals "Username field" "john@example.com" 1
  api-cli assert equals "Username field" "john@example.com"  # Auto-increment position

  # Override checkpoint explicitly
  api-cli assert equals "Username field" "john@example.com" 1 --checkpoint 1678318`,
        Args: cobra.RangeArgs(2, 3),
        RunE: func(cmd *cobra.Command, args []string) error {
            return executeAssertCommand(cfg, "equals", args, checkpointFlag)
        },
    }

    shared.AddCheckpointFlag(cmd, &checkpointFlag)
    shared.AddOutputFlags(cmd)

    return cmd
}

// Shared execution logic for all assert commands
func executeAssertCommand(cfg *config.VirtuosoConfig, assertType string, args []string, checkpointFlag int) error {
    // Common validation
    element := args[0]
    if element == "" {
        return fmt.Errorf("element cannot be empty")
    }

    // Resolve checkpoint and position
    ctx, err := shared.ResolveStepContext(args, checkpointFlag, getPositionIndex(assertType))
    if err != nil {
        return err
    }

    // Create client
    client := client.NewClient(cfg)

    // Execute appropriate assert based on type
    var stepID int
    switch assertType {
    case "equals":
        value := args[1]
        stepID, err = client.CreateAssertEqualsStep(ctx.CheckpointID, element, value, ctx.Position)
    case "not-equals":
        value := args[1]
        stepID, err = client.CreateAssertNotEqualsStep(ctx.CheckpointID, element, value, ctx.Position)
    case "exists":
        stepID, err = client.CreateAssertExistsStep(ctx.CheckpointID, element, ctx.Position)
    case "not-exists":
        stepID, err = client.CreateAssertNotExistsStep(ctx.CheckpointID, element, ctx.Position)
    // ... other assert types
    }

    if err != nil {
        return fmt.Errorf("failed to create assert %s step: %w", assertType, err)
    }

    // Save context and output result
    shared.SaveStepContext(ctx)

    output := &shared.StepOutput{
        Status:       "success",
        StepType:     fmt.Sprintf("ASSERT_%s", strings.ToUpper(assertType)),
        CheckpointID: ctx.CheckpointID,
        StepID:       stepID,
        Position:     ctx.Position,
        ParsedStep:   formatAssertDescription(assertType, args),
        UsingContext: ctx.UsingContext,
        AutoPosition: ctx.AutoPosition,
    }

    return shared.OutputStepResult(output)
}
```

### Shared Infrastructure

```go
// pkg/api-cli/commands/shared/context.go
package shared

import (
    "fmt"
    "strconv"
)

// ResolveStepContext resolves checkpoint and position for any step command
func ResolveStepContext(args []string, checkpointFlag int, positionIndex int) (*StepContext, error) {
    ctx := &StepContext{}

    // Checkpoint resolution logic (same as before)
    if checkpointFlag > 0 {
        ctx.CheckpointID = checkpointFlag
        ctx.UsingContext = false
    } else if cfg.GetCurrentCheckpoint() != nil {
        ctx.CheckpointID = *cfg.GetCurrentCheckpoint()
        ctx.UsingContext = true
    } else {
        return nil, fmt.Errorf("no checkpoint specified")
    }

    // Position resolution logic (same as before)
    if positionIndex < len(args) {
        var err error
        ctx.Position, err = strconv.Atoi(args[positionIndex])
        if err != nil {
            return nil, fmt.Errorf("invalid position: %w", err)
        }
        ctx.AutoPosition = false
    } else {
        ctx.Position = cfg.GetNextPosition()
        ctx.AutoPosition = true
    }

    return ctx, nil
}
```

### Legacy Wrapper for Backward Compatibility

```go
// pkg/api-cli/commands/legacy/assert_wrappers.go
package legacy

import (
    "github.com/spf13/cobra"
    "os"
)

// CreateAssertEqualsLegacyWrapper creates a backward-compatible wrapper
func CreateAssertEqualsLegacyWrapper() *cobra.Command {
    cmd := &cobra.Command{
        Use:        "create-step-assert-equals ELEMENT VALUE [POSITION]",
        Short:      "Create an assertion step that verifies an element has a specific text value",
        Deprecated: "Use 'api-cli assert equals' instead",
        Hidden:     false, // Keep visible during migration period
        RunE: func(cmd *cobra.Command, args []string) error {
            // Translate to new command format
            newArgs := append([]string{"assert", "equals"}, args...)

            // Execute new command
            rootCmd := cmd.Root()
            rootCmd.SetArgs(newArgs)
            return rootCmd.Execute()
        },
    }

    // Copy all flags from new command
    // This ensures complete compatibility
    return cmd
}
```

### Migration Helper Script

```bash
#!/bin/bash
# migrate-commands.sh - Helper script to update user scripts

# Function to migrate commands in a file
migrate_file() {
    local file=$1

    # Assert commands
    sed -i.bak 's/api-cli create-step-assert-equals/api-cli assert equals/g' "$file"
    sed -i.bak 's/api-cli create-step-assert-not-equals/api-cli assert not-equals/g' "$file"
    sed -i.bak 's/api-cli create-step-assert-exists/api-cli assert exists/g' "$file"
    sed -i.bak 's/api-cli create-step-assert-not-exists/api-cli assert not-exists/g' "$file"

    # Interact commands
    sed -i.bak 's/api-cli create-step-click/api-cli interact click/g' "$file"
    sed -i.bak 's/api-cli create-step-double-click/api-cli interact double-click/g' "$file"
    sed -i.bak 's/api-cli create-step-write/api-cli interact write/g' "$file"

    # Navigate commands
    sed -i.bak 's/api-cli create-step-navigate/api-cli navigate url/g' "$file"
    sed -i.bak 's/api-cli create-step-scroll-top/api-cli navigate scroll-top/g' "$file"

    echo "Migrated: $file"
}

# Check if file provided
if [ $# -eq 0 ]; then
    echo "Usage: $0 <script-file>"
    exit 1
fi

migrate_file "$1"
```

### Test Suite Updates

```go
// pkg/api-cli/commands/consolidated/assert_test.go
package consolidated_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestAssertCommands(t *testing.T) {
    tests := []struct {
        name     string
        args     []string
        expected string
    }{
        {
            name:     "assert equals with position",
            args:     []string{"assert", "equals", "Username", "john@example.com", "1"},
            expected: "ASSERT_EQUALS",
        },
        {
            name:     "assert exists without position (auto-increment)",
            args:     []string{"assert", "exists", "div.content"},
            expected: "ASSERT_EXISTS",
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}

// Test backward compatibility
func TestLegacyCommands(t *testing.T) {
    tests := []struct {
        oldCmd string
        newCmd string
    }{
        {
            oldCmd: "create-step-assert-equals Username john@example.com 1",
            newCmd: "assert equals Username john@example.com 1",
        },
        // ... more compatibility tests
    }
}
```

## Benefits of This Approach

1. **Code Reuse**: The `executeAssertCommand` function handles all assert types
2. **Consistency**: All assert commands share the same flag handling and output formatting
3. **Maintainability**: Adding a new assert type only requires adding a case to the switch statement
4. **Discoverability**: `api-cli assert --help` shows all available assertions
5. **Backward Compatibility**: Old commands continue to work with deprecation notices

## Implementation Checklist

- [ ] Create shared infrastructure package
- [ ] Implement base command structure
- [ ] Create consolidated assert command
- [ ] Add all assert subcommands
- [ ] Create legacy wrappers
- [ ] Update command registration
- [ ] Write comprehensive tests
- [ ] Update documentation
- [ ] Create migration script
- [ ] Test with real API
- [ ] Performance benchmarks
- [ ] User acceptance testing
