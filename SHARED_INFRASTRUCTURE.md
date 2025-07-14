# Shared Infrastructure for Command Consolidation

## Overview

This document outlines the shared infrastructure components needed to support the consolidated command structure.

## Core Components

### 1. Base Command Interface

```go
// pkg/api-cli/commands/shared/interfaces.go
package shared

import (
    "github.com/spf13/cobra"
    "github.com/marklovelady/api-cli-generator/pkg/api-cli/config"
)

// StepCommand defines the interface for all step commands
type StepCommand interface {
    Execute(args []string, flags CommandFlags) (*StepOutput, error)
    Validate(args []string) error
    GetStepType() string
}

// CommandFlags holds common flags for all commands
type CommandFlags struct {
    CheckpointID int
    OutputFormat string
    Variables    map[string]string
    Options      map[string]interface{}
}

// SubCommandRegistry manages subcommands for consolidated commands
type SubCommandRegistry struct {
    commands map[string]StepCommand
    config   *config.VirtuosoConfig
}

func NewSubCommandRegistry(cfg *config.VirtuosoConfig) *SubCommandRegistry {
    return &SubCommandRegistry{
        commands: make(map[string]StepCommand),
        config:   cfg,
    }
}

func (r *SubCommandRegistry) Register(name string, cmd StepCommand) {
    r.commands[name] = cmd
}

func (r *SubCommandRegistry) Get(name string) (StepCommand, bool) {
    cmd, ok := r.commands[name]
    return cmd, ok
}
```

### 2. Common Context Resolution

```go
// pkg/api-cli/commands/shared/context.go
package shared

import (
    "fmt"
    "strconv"
    "github.com/marklovelady/api-cli-generator/pkg/api-cli/config"
)

// StepContext holds resolved checkpoint and position
type StepContext struct {
    CheckpointID int
    Position     int
    UsingContext bool
    AutoPosition bool
}

// ContextResolver handles checkpoint and position resolution
type ContextResolver struct {
    config *config.VirtuosoConfig
}

func NewContextResolver(cfg *config.VirtuosoConfig) *ContextResolver {
    return &ContextResolver{config: cfg}
}

func (r *ContextResolver) Resolve(args []string, checkpointFlag int, positionIndex int) (*StepContext, error) {
    ctx := &StepContext{}

    // Checkpoint resolution
    if checkpointFlag > 0 {
        ctx.CheckpointID = checkpointFlag
        ctx.UsingContext = false
    } else if checkpoint := r.config.GetCurrentCheckpoint(); checkpoint != nil {
        ctx.CheckpointID = *checkpoint
        ctx.UsingContext = true
    } else {
        return nil, fmt.Errorf("no checkpoint specified - use --checkpoint flag or set current checkpoint")
    }

    // Position resolution
    if positionIndex < len(args) && positionIndex >= 0 {
        pos, err := strconv.Atoi(args[positionIndex])
        if err != nil {
            return nil, fmt.Errorf("invalid position: %w", err)
        }
        ctx.Position = pos
        ctx.AutoPosition = false
    } else {
        ctx.Position = r.config.GetNextPosition()
        ctx.AutoPosition = true
    }

    return ctx, nil
}

func (r *ContextResolver) SaveContext(ctx *StepContext) error {
    if ctx.AutoPosition && r.config.Session.AutoIncrementPos {
        return r.config.SaveConfig()
    }
    return nil
}
```

### 3. Output Formatting

```go
// pkg/api-cli/commands/shared/output.go
package shared

import (
    "encoding/json"
    "fmt"
    "time"
    "gopkg.in/yaml.v2"
)

// StepOutput represents the output of a step command
type StepOutput struct {
    Status       string                 `json:"status" yaml:"status"`
    StepType     string                 `json:"step_type" yaml:"step_type"`
    CheckpointID int                    `json:"checkpoint_id" yaml:"checkpoint_id"`
    StepID       int                    `json:"step_id" yaml:"step_id"`
    Position     int                    `json:"position" yaml:"position"`
    ParsedStep   string                 `json:"parsed_step" yaml:"parsed_step"`
    UsingContext bool                   `json:"using_context" yaml:"using_context"`
    AutoPosition bool                   `json:"auto_position" yaml:"auto_position"`
    Extra        map[string]interface{} `json:"extra,omitempty" yaml:"extra,omitempty"`
    Timestamp    string                 `json:"timestamp" yaml:"timestamp"`
}

// OutputFormatter handles formatting step results
type OutputFormatter struct {
    format string
}

func NewOutputFormatter(format string) *OutputFormatter {
    return &OutputFormatter{format: format}
}

func (f *OutputFormatter) Format(output *StepOutput) error {
    // Set timestamp if not set
    if output.Timestamp == "" {
        output.Timestamp = time.Now().Format(time.RFC3339)
    }

    switch f.format {
    case "json":
        return f.formatJSON(output)
    case "yaml":
        return f.formatYAML(output)
    case "ai":
        return f.formatAI(output)
    default:
        return f.formatHuman(output)
    }
}

func (f *OutputFormatter) formatJSON(output *StepOutput) error {
    data, err := json.MarshalIndent(output, "", "  ")
    if err != nil {
        return err
    }
    fmt.Println(string(data))
    return nil
}

func (f *OutputFormatter) formatYAML(output *StepOutput) error {
    data, err := yaml.Marshal(output)
    if err != nil {
        return err
    }
    fmt.Print(string(data))
    return nil
}

func (f *OutputFormatter) formatHuman(output *StepOutput) error {
    fmt.Printf("✓ Step created successfully!\n")
    fmt.Printf("  Step Type: %s\n", output.StepType)
    fmt.Printf("  Step ID: %d\n", output.StepID)
    fmt.Printf("  Checkpoint: %d\n", output.CheckpointID)
    fmt.Printf("  Position: %d\n", output.Position)
    if output.ParsedStep != "" {
        fmt.Printf("  Action: %s\n", output.ParsedStep)
    }
    if output.UsingContext {
        fmt.Printf("  Using: Session context\n")
    }
    if output.AutoPosition {
        fmt.Printf("  Position: Auto-incremented\n")
    }
    return nil
}

func (f *OutputFormatter) formatAI(output *StepOutput) error {
    fmt.Printf("Created %s step (ID: %d) at position %d in checkpoint %d. ",
        output.StepType, output.StepID, output.Position, output.CheckpointID)
    if output.ParsedStep != "" {
        fmt.Printf("Action: %s", output.ParsedStep)
    }
    fmt.Println()
    return nil
}
```

### 4. Common Flag Handling

```go
// pkg/api-cli/commands/shared/flags.go
package shared

import (
    "github.com/spf13/cobra"
)

// CommonFlags adds standard flags to any command
type CommonFlags struct {
    Checkpoint   int
    OutputFormat string
}

// AddCommonFlags adds standard flags to a command
func AddCommonFlags(cmd *cobra.Command, flags *CommonFlags) {
    cmd.Flags().IntVar(&flags.Checkpoint, "checkpoint", 0, "Override checkpoint ID")
    cmd.Flags().StringVarP(&flags.OutputFormat, "output", "o", "human", "Output format (human, json, yaml, ai)")
}

// AddCheckpointFlag adds just the checkpoint flag
func AddCheckpointFlag(cmd *cobra.Command, checkpointFlag *int) {
    cmd.Flags().IntVar(checkpointFlag, "checkpoint", 0, "Override checkpoint ID from session")
}

// AddOutputFlag adds just the output format flag
func AddOutputFlag(cmd *cobra.Command, outputFlag *string) {
    cmd.Flags().StringVarP(outputFlag, "output", "o", "human", "Output format (human, json, yaml, ai)")
}

// ValidateOutputFormat validates the output format flag
func ValidateOutputFormat(format string) error {
    validFormats := map[string]bool{
        "human": true,
        "json":  true,
        "yaml":  true,
        "ai":    true,
    }

    if !validFormats[format] {
        return fmt.Errorf("invalid output format: %s (valid: human, json, yaml, ai)", format)
    }
    return nil
}
```

### 5. Command Builder Pattern

```go
// pkg/api-cli/commands/shared/builder.go
package shared

import (
    "github.com/spf13/cobra"
    "github.com/marklovelady/api-cli-generator/pkg/api-cli/config"
    "github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
)

// CommandBuilder helps build consistent commands
type CommandBuilder struct {
    config         *config.VirtuosoConfig
    client         *client.Client
    contextResolver *ContextResolver
    outputFormatter *OutputFormatter
}

func NewCommandBuilder(cfg *config.VirtuosoConfig) *CommandBuilder {
    return &CommandBuilder{
        config:          cfg,
        client:          client.NewClient(cfg),
        contextResolver: NewContextResolver(cfg),
    }
}

// BuildStepCommand creates a standard step command
func (b *CommandBuilder) BuildStepCommand(spec CommandSpec) *cobra.Command {
    var flags CommonFlags

    cmd := &cobra.Command{
        Use:   spec.Use,
        Short: spec.Short,
        Long:  spec.Long,
        Args:  spec.Args,
        RunE: func(cmd *cobra.Command, args []string) error {
            // Validate output format
            if err := ValidateOutputFormat(flags.OutputFormat); err != nil {
                return err
            }

            // Set output formatter
            b.outputFormatter = NewOutputFormatter(flags.OutputFormat)

            // Resolve context
            ctx, err := b.contextResolver.Resolve(args, flags.Checkpoint, spec.PositionIndex)
            if err != nil {
                return err
            }

            // Execute command
            output, err := spec.Execute(b, ctx, args)
            if err != nil {
                return err
            }

            // Save context if needed
            if err := b.contextResolver.SaveContext(ctx); err != nil {
                fmt.Fprintf(os.Stderr, "Warning: failed to save context: %v\n", err)
            }

            // Format and display output
            return b.outputFormatter.Format(output)
        },
    }

    // Add common flags
    AddCommonFlags(cmd, &flags)

    // Add any additional flags
    if spec.AddFlags != nil {
        spec.AddFlags(cmd)
    }

    return cmd
}

// CommandSpec defines the specification for a command
type CommandSpec struct {
    Use           string
    Short         string
    Long          string
    Args          cobra.PositionalArgs
    PositionIndex int
    Execute       func(*CommandBuilder, *StepContext, []string) (*StepOutput, error)
    AddFlags      func(*cobra.Command)
}
```

### 6. Error Handling

```go
// pkg/api-cli/commands/shared/errors.go
package shared

import (
    "fmt"
)

// CommandError represents a command execution error
type CommandError struct {
    Command string
    Phase   string
    Err     error
}

func (e CommandError) Error() string {
    return fmt.Sprintf("%s failed during %s: %v", e.Command, e.Phase, e.Err)
}

// WrapError wraps an error with command context
func WrapError(command, phase string, err error) error {
    if err == nil {
        return nil
    }
    return CommandError{
        Command: command,
        Phase:   phase,
        Err:     err,
    }
}

// Common error messages
var (
    ErrNoCheckpoint = fmt.Errorf("no checkpoint specified")
    ErrInvalidArgs  = fmt.Errorf("invalid arguments")
    ErrAPIFailure   = fmt.Errorf("API request failed")
)
```

## Usage Example

```go
// Example: Using the shared infrastructure in a consolidated command
package consolidated

import (
    "github.com/marklovelady/api-cli-generator/pkg/api-cli/commands/shared"
)

func NewInteractCommand(cfg *config.VirtuosoConfig) *cobra.Command {
    builder := shared.NewCommandBuilder(cfg)

    cmd := &cobra.Command{
        Use:   "interact",
        Short: "Perform interactions with elements",
    }

    // Add click subcommand using builder
    clickSpec := shared.CommandSpec{
        Use:           "click ELEMENT [POSITION]",
        Short:         "Click on an element",
        Long:          "Performs a click action on the specified element",
        Args:          cobra.RangeArgs(1, 2),
        PositionIndex: 1,
        Execute: func(b *shared.CommandBuilder, ctx *shared.StepContext, args []string) (*shared.StepOutput, error) {
            element := args[0]

            // Call API
            stepID, err := b.client.CreateStepClick(ctx.CheckpointID, element, ctx.Position)
            if err != nil {
                return nil, shared.WrapError("click", "api call", err)
            }

            // Build output
            return &shared.StepOutput{
                Status:       "success",
                StepType:     "CLICK",
                CheckpointID: ctx.CheckpointID,
                StepID:       stepID,
                Position:     ctx.Position,
                ParsedStep:   fmt.Sprintf("click on %s", element),
                UsingContext: ctx.UsingContext,
                AutoPosition: ctx.AutoPosition,
                Extra: map[string]interface{}{
                    "element": element,
                },
            }, nil
        },
        AddFlags: func(cmd *cobra.Command) {
            cmd.Flags().String("variable", "", "Store result in variable")
            cmd.Flags().String("position", "", "Click position (TOP_LEFT, CENTER, etc)")
        },
    }

    cmd.AddCommand(builder.BuildStepCommand(clickSpec))

    return cmd
}
```

## Benefits

1. **Consistency**: All commands use the same patterns
2. **Reusability**: Common logic is shared across commands
3. **Testability**: Each component can be tested independently
4. **Maintainability**: Changes to common behavior only need to be made once
5. **Extensibility**: Easy to add new commands following the pattern
