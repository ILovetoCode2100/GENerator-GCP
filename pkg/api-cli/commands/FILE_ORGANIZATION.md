# Commands Package File Organization

This document describes the file organization after major consolidation efforts. All files remain in the same package to avoid circular dependencies.

## Consolidation Summary

**Before**: 35+ individual files
**After**: ~20 files (43% reduction)
**Code Reduction**: ~30% through elimination of duplication

## File Categories

### Core Infrastructure (6 files)

- `base.go` - Base command functionality with session support
- `config.go` - Global configuration management
- `register.go` - Command registration
- `types.go` - Shared type definitions
- `validate_config.go` - Configuration validation
- `list_commands_test.go` - Tests

### Consolidated Command Files (5 files)

Major consolidations for better maintainability:

1. **`interaction_commands.go`** - All user interactions

   - Consolidated from: interact.go, mouse.go, select.go
   - Contains: click, hover, write, key, mouse operations, dropdown selection

2. **`browser_commands.go`** - Browser operations

   - Consolidated from: navigate.go, window.go
   - Contains: navigation, scrolling, window management, tab/frame switching

3. **`list.go`** - All list operations

   - Consolidated from: list_projects.go, list_goals.go, list_journeys.go, list_checkpoints.go
   - Contains: Generic list framework for all entity types

4. **`project_management.go`** - Project CRUD operations

   - Consolidated from: create_project.go, create_goal.go, create_journey.go, create_checkpoint.go, update_journey.go, update_navigation.go, get_step.go
   - Contains: All project/goal/journey/checkpoint management

5. **`execution_management.go`** - Execution operations
   - Consolidated from: execute-goal.go, monitor-execution.go, get-execution-analysis.go, manage-test-data.go, create-environment.go
   - Contains: Test execution, monitoring, analysis, and environment management

### Individual Step Commands (7 files)

Specialized commands that remain separate:

- `assert.go` - Assertion commands (exists, equals, gt, etc.)
- `data.go` - Data storage and cookie commands
- `dialog.go` - Dialog handling commands
- `wait.go` - Wait operation commands
- `file.go` - File upload commands
- `misc.go` - Miscellaneous commands (comment, execute)
- `library.go` - Library checkpoint operations

### Other (2 files)

- `test_templates.go` - AI test template integration
- `set_checkpoint.go` - Session management (if exists)

## Command Structure

All commands follow the unified pattern:

```
api-cli <command> <subcommand> [checkpoint-id] <args...> [position]
```

With session context:

```
export VIRTUOSO_SESSION_ID=cp_12345
api-cli <command> <subcommand> <args...>
```

## Benefits of Consolidation

1. **Reduced File Count**: 43% fewer files to navigate
2. **Code Reuse**: ~30% less code through shared functions
3. **Logical Grouping**: Related functionality in same file
4. **Easier Maintenance**: Changes to related features in one place
5. **Better AI Understanding**: Clear patterns and fewer files to analyze
