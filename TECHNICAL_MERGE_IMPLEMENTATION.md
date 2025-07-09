# Technical Implementation Analysis for Step Command Consolidation

## Current Implementation Pattern Analysis

### Common Pattern Across All Step Commands

After analyzing the codebase, every step command follows this exact pattern:

```go
1. Parse arguments (1-4 args depending on command)
2. Convert string IDs to integers
3. Validate inputs
4. Create Virtuoso client
5. Call specific client method (e.g., CreateClickStep)
6. Format output based on output flag
```

### Client Implementation Pattern

In `pkg/virtuoso/client.go`, each step type has:

```go
func (c *Client) CreateXXXStep(checkpointID int, args..., position int) (int, error) {
    parsedStep := map[string]interface{}{
        "action": "ACTION_TYPE",
        "target": targetConfig,  // Optional
        "value":  value,         // Optional
        "meta":   metadata,      // Optional
    }
    return c.addStep(checkpointID, position, parsedStep)
}
```

## Detailed Merge Implementation Proposal

### 1. Core Step Structure

All steps can be represented with this unified structure:

```go
type StepConfig struct {
    CheckpointID int
    Position     int
    Action       string
    Target       *TargetConfig
    Value        string
    Meta         map[string]interface{}
}

type TargetConfig struct {
    Type     string // GUESS, CSS, XPATH, etc.
    Selector string
    Options  map[string]interface{}
}
```

### 2. Action Type Mapping

Create a comprehensive action registry:

```go
var ActionRegistry = map[string]ActionConfig{
    // Navigation
    "navigate": {
        Action: "NAVIGATE",
        RequiresTarget: true,
        RequiresValue: false,
        ParseTarget: func(input string) *TargetConfig {
            return &TargetConfig{
                Type: "GUESS",
                Selector: fmt.Sprintf(`{"clue":"%s"}`, input),
            }
        },
    },
    
    // Clicks
    "click": {
        Action: "CLICK",
        RequiresTarget: true,
        ParseTarget: guessSelector,
    },
    "double-click": {
        Action: "DOUBLE_CLICK",
        RequiresTarget: true,
        ParseTarget: guessSelector,
    },
    
    // Assertions
    "assert-exists": {
        Action: "SEE",
        RequiresTarget: true,
        ParseTarget: guessSelector,
    },
    "assert-equals": {
        Action: "EXPECT",
        RequiresTarget: true,
        RequiresValue: true,
        ParseTarget: guessSelector,
        Meta: map[string]interface{}{
            "type": "EQUALS",
        },
    },
    
    // Waits
    "wait-time": {
        Action: "WAIT",
        RequiresValue: true,
        ParseValue: func(input string) (string, map[string]interface{}) {
            seconds, _ := strconv.Atoi(input)
            return fmt.Sprintf("%d", seconds*1000), map[string]interface{}{
                "kind": "WAIT",
                "type": "TIME",
                "duration": seconds * 1000,
                "poll": 100,
            }
        },
    },
    
    // ... continue for all action types
}
```

### 3. Unified Command Implementation

```go
func newCreateStepCmd() *cobra.Command {
    var (
        action      string
        target      string
        value       string
        condition   string
        timeout     int
        modifiers   []string
        options     map[string]string
    )
    
    cmd := &cobra.Command{
        Use:   "create-step CHECKPOINT_ID POSITION",
        Short: "Create any type of test step",
        Args:  cobra.ExactArgs(2),
        RunE: func(cmd *cobra.Command, args []string) error {
            checkpointID, _ := strconv.Atoi(args[0])
            position, _ := strconv.Atoi(args[1])
            
            // Get action configuration
            actionConfig, exists := ActionRegistry[action]
            if !exists {
                return fmt.Errorf("unknown action type: %s", action)
            }
            
            // Build step configuration
            stepConfig := StepConfig{
                CheckpointID: checkpointID,
                Position:     position,
                Action:       actionConfig.Action,
            }
            
            // Handle target
            if actionConfig.RequiresTarget {
                if target == "" {
                    return fmt.Errorf("action %s requires --target", action)
                }
                stepConfig.Target = actionConfig.ParseTarget(target)
            }
            
            // Handle value
            if actionConfig.RequiresValue {
                if value == "" {
                    return fmt.Errorf("action %s requires --value", action)
                }
                if actionConfig.ParseValue != nil {
                    stepConfig.Value, stepConfig.Meta = actionConfig.ParseValue(value)
                } else {
                    stepConfig.Value = value
                }
            }
            
            // Apply modifiers and options
            applyModifiers(&stepConfig, modifiers, options)
            
            // Create step
            client := virtuoso.NewClient(cfg)
            stepID, err := client.CreateUnifiedStep(stepConfig)
            
            // Output formatting (same as current)
            return formatOutput(stepConfig, stepID, err)
        },
    }
    
    // Add flags
    cmd.Flags().StringVarP(&action, "action", "a", "", "Step action type")
    cmd.Flags().StringVarP(&target, "target", "t", "", "Target element or URL")
    cmd.Flags().StringVarP(&value, "value", "v", "", "Value for the action")
    cmd.Flags().StringVar(&condition, "condition", "", "Condition for assertions")
    cmd.Flags().IntVar(&timeout, "timeout", 0, "Timeout in seconds")
    cmd.Flags().StringSliceVar(&modifiers, "modifiers", nil, "Action modifiers")
    
    return cmd
}
```

### 4. Specialized Merged Commands

#### Assert Command
```go
func newCreateStepAssertCmd() *cobra.Command {
    var (
        element   string
        condition string
        value     string
        timeout   int
        attribute string
    )
    
    cmd := &cobra.Command{
        Use:   "create-step-assert CHECKPOINT_ID POSITION",
        Short: "Create any type of assertion step",
        Args:  cobra.ExactArgs(2),
        RunE: func(cmd *cobra.Command, args []string) error {
            // Map conditions to actions
            actionMap := map[string]string{
                "exists":       "assert-exists",
                "not-exists":   "assert-not-exists",
                "equals":       "assert-equals",
                "contains":     "assert-contains",
                "checked":      "assert-checked",
                "selected":     "assert-selected",
                "visible":      "assert-visible",
                "enabled":      "assert-enabled",
                "has-class":    "assert-class",
                "has-attribute":"assert-attribute",
            }
            
            action, exists := actionMap[condition]
            if !exists {
                return fmt.Errorf("unknown assertion condition: %s", condition)
            }
            
            // Delegate to unified implementation
            return createUnifiedStep(args[0], args[1], action, element, value, map[string]string{
                "attribute": attribute,
                "timeout": strconv.Itoa(timeout),
            })
        },
    }
    
    cmd.Flags().StringVarP(&element, "element", "e", "", "Target element")
    cmd.Flags().StringVarP(&condition, "condition", "c", "", "Assertion condition")
    cmd.Flags().StringVarP(&value, "value", "v", "", "Expected value")
    cmd.Flags().IntVar(&timeout, "timeout", 20, "Timeout in seconds")
    cmd.Flags().StringVar(&attribute, "attribute", "", "Attribute name for attribute assertions")
    
    return cmd
}
```

#### Click Command with Variations
```go
func newCreateStepClickCmd() *cobra.Command {
    var (
        element    string
        clickType  string
        modifiers  []string
        offsetX    int
        offsetY    int
        force      bool
        count      int
    )
    
    cmd := &cobra.Command{
        Use:   "create-step-click CHECKPOINT_ID POSITION",
        Short: "Create any type of click step",
        Args:  cobra.ExactArgs(2),
        RunE: func(cmd *cobra.Command, args []string) error {
            // Build meta based on options
            meta := map[string]interface{}{
                "kind": "CLICK",
                "type": strings.ToUpper(clickType),
            }
            
            if len(modifiers) > 0 {
                meta["modifiers"] = modifiers
            }
            
            if offsetX != 0 || offsetY != 0 {
                meta["offset"] = map[string]int{
                    "x": offsetX,
                    "y": offsetY,
                }
            }
            
            if force {
                meta["force"] = true
            }
            
            if count > 1 {
                meta["count"] = count
            }
            
            // Map click types to actions
            actionMap := map[string]string{
                "single": "CLICK",
                "double": "DOUBLE_CLICK",
                "right":  "RIGHT_CLICK",
                "middle": "MIDDLE_CLICK",
                "hover":  "HOVER",
            }
            
            action := actionMap[clickType]
            
            return createStepWithMeta(args[0], args[1], action, element, "", meta)
        },
    }
    
    cmd.Flags().StringVarP(&element, "element", "e", "", "Target element")
    cmd.Flags().StringVar(&clickType, "type", "single", "Click type")
    cmd.Flags().StringSliceVar(&modifiers, "modifiers", nil, "Modifier keys")
    cmd.Flags().IntVar(&offsetX, "offset-x", 0, "X offset from element center")
    cmd.Flags().IntVar(&offsetY, "offset-y", 0, "Y offset from element center")
    cmd.Flags().BoolVar(&force, "force", false, "Force click even if element not visible")
    cmd.Flags().IntVar(&count, "count", 1, "Number of clicks")
    
    return cmd
}
```

### 5. Migration Aliases

Create backward-compatible aliases:

```go
func createAliasCommand(oldName, newAction string, extraArgs ...string) *cobra.Command {
    return &cobra.Command{
        Use:   oldName + " [args...]",
        Short: fmt.Sprintf("Deprecated: Use create-step with --action %s", newAction),
        RunE: func(cmd *cobra.Command, args []string) error {
            // Show deprecation warning
            fmt.Fprintf(os.Stderr, "⚠️  Warning: %s is deprecated. Use 'create-step' instead.\n", oldName)
            
            // Build new command arguments
            newArgs := []string{"create-step", args[0], args[len(args)-1]} // checkpoint, position
            newArgs = append(newArgs, "--action", newAction)
            
            // Map old arguments to new flags
            switch newAction {
            case "click":
                if len(args) > 2 {
                    newArgs = append(newArgs, "--target", args[1])
                }
            case "assert-equals":
                if len(args) > 3 {
                    newArgs = append(newArgs, "--target", args[1], "--value", args[2])
                }
            // ... handle other cases
            }
            
            // Execute new command
            rootCmd := cmd.Root()
            rootCmd.SetArgs(newArgs)
            return rootCmd.Execute()
        },
    }
}
```

### 6. Complex Step Examples

#### Multi-condition Wait
```bash
api-cli create-step CHECKPOINT_ID POSITION \
  --action wait \
  --target ".loading-spinner" \
  --condition "not-visible" \
  --timeout 30 \
  --fallback "continue-on-timeout"
```

#### Click with Complex Options
```bash
api-cli create-step CHECKPOINT_ID POSITION \
  --action click \
  --target "#submit-button" \
  --modifiers "ctrl,shift" \
  --offset-x 10 \
  --offset-y -5 \
  --force \
  --wait-after 2
```

#### Advanced Assertion
```bash
api-cli create-step CHECKPOINT_ID POSITION \
  --action assert \
  --target ".price-display" \
  --condition "matches-regex" \
  --value "^\$[0-9]+\.[0-9]{2}$" \
  --timeout 10 \
  --screenshot-on-fail
```

### 7. Benefits of This Approach

1. **Minimal Breaking Changes**: Existing commands continue working
2. **Progressive Enhancement**: Add features without new commands
3. **Consistent Interface**: All steps follow same pattern
4. **Extensible**: Easy to add new action types or options
5. **Discoverable**: Users can explore options with --help

### 8. Implementation Priority

1. **Core Infrastructure** (Week 1)
   - Unified step structure
   - Action registry
   - Base create-step command

2. **High-Value Mergers** (Week 2)
   - Assert variations (highest complexity reduction)
   - Click variations (most common usage)
   - Wait variations (clear benefits)

3. **Migration Support** (Week 3)
   - Alias commands
   - Deprecation warnings
   - Migration documentation

4. **Advanced Features** (Week 4)
   - Complex conditions
   - Chained actions
   - Conditional steps

## Conclusion

This implementation approach provides a clear path to consolidating 40+ commands into a flexible, maintainable system while preserving backward compatibility. The unified structure makes it easier to add new capabilities and reduces code duplication significantly.
