# Interaction Commands Consolidation

## Overview

The `interact.go`, `mouse.go`, and `select.go` commands have been consolidated into a single `interaction_commands.go` file to reduce code duplication and improve maintainability.

## Changes Made

### 1. File Consolidation

**Before:**

- `interact.go` - Basic interactions (click, write, hover, etc.)
- `mouse.go` - Mouse operations (move, drag, etc.)
- `select.go` - Dropdown selection

**After:**

- `interaction_commands.go` - All interaction commands consolidated

### 2. Command Structure Update

**Before:**

```bash
api-cli interact click "button"
api-cli mouse move-to "element"
api-cli select option "dropdown" "value"
```

**After:**

```bash
api-cli interact click "button"
api-cli interact mouse move-to "element"
api-cli interact select option "dropdown" "value"
```

### 3. Code Organization

The consolidated file is organized into logical sections:

1. **Main Command** - The root `interact` command
2. **Click-based Interactions** - click, double-click, right-click
3. **Text and Keyboard** - write, key press
4. **Mouse Operations** - hover, mouse movements
5. **Dropdown Selection** - select by value, index, or last
6. **Shared Functions** - Common execution and validation logic

### 4. Benefits

- **Reduced Code Duplication**: Shared validation and execution functions
- **Consistent Interface**: All interactions under single `interact` command
- **Easier Maintenance**: Single file to update for interaction logic
- **Better Organization**: Clear sections for different interaction types

### 5. Backward Compatibility

To maintain backward compatibility, you could add aliases in the register.go file:

```go
// Add these for backward compatibility if needed
rootCmd.AddCommand(newMouseAliasCmd())   // Alias to interact mouse
rootCmd.AddCommand(newSelectAliasCmd())  // Alias to interact select
```

However, it's recommended to update scripts to use the new consolidated command structure.

## Testing

A new test script has been created at `test-commands/test-consolidated-interact.sh` to verify all consolidated commands work correctly.

## Migration Guide

For users updating their scripts:

1. Replace `api-cli mouse <action>` with `api-cli interact mouse <action>`
2. Replace `api-cli select <action>` with `api-cli interact select <action>`
3. No changes needed for existing `api-cli interact` commands

## Files to Remove

After verifying the consolidation works correctly, these files can be removed:

- `pkg/api-cli/commands/interact.go`
- `pkg/api-cli/commands/mouse.go`
- `pkg/api-cli/commands/select.go`
