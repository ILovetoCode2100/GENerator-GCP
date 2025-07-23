# Commands Package File Organization

## Overview

The commands package has been significantly consolidated and reorganized for better maintainability and consistency. This document outlines the current file structure and organization principles.

**Latest Update:** January 2025
**Files:** 19 (reduced from 27)
**Reduction:** 30% fewer files with improved naming consistency

## File Structure

### Core Infrastructure (6 files)

These files provide the foundation for all commands:

1. **`base.go`** - Base command functionality with session support
2. **`config.go`** - Global configuration management
3. **`context_helpers.go`** - Context utility functions
4. **`register.go`** - Command registration with rootCmd
5. **`types.go`** - Shared type definitions
6. **`validate_config.go`** - Configuration validation

### Step Commands (8 files)

All commands that create test steps follow the `step_*.go` naming pattern:

1. **`step_assert.go`** - Assertion commands (12 types)

   - exists, not-exists, equals, not-equals
   - checked, selected, variable
   - gt, gte, lt, lte, matches

2. **`step_browser.go`** - Browser operations (consolidated)

   - Navigation: to, scroll operations
   - Window: resize, maximize, switch tab/iframe

3. **`step_data.go`** - Data management (6 types)

   - Store: element-text, element-value, attribute
   - Cookies: create, delete, clear

4. **`step_dialog.go`** - Dialog handling (4 types)

   - dismiss-alert, dismiss-confirm
   - dismiss-prompt, dismiss-prompt-with-text

5. **`step_file.go`** - File operations (2 types)

   - upload, upload-url (URL only)

6. **`step_interact.go`** - User interactions (consolidated)

   - Basic: click, double-click, right-click, hover, write, key
   - Mouse: move-to, move-by, move, down, up, enter
   - Select: option, index, last

7. **`step_misc.go`** - Miscellaneous (2 types)

   - comment, execute (JavaScript)

8. **`step_wait.go`** - Wait operations (2 types)
   - element, time

### Management Commands (5 files)

Commands for managing projects, executions, and resources:

1. **`manage_projects.go`** - Project hierarchy CRUD (consolidated)

   - Create: project, goal, journey, checkpoint
   - Update: journey, navigation
   - Get: step details

2. **`manage_executions.go`** - Execution workflow

   - Create environment
   - Manage test data
   - Execute goals
   - Monitor execution
   - Get analysis

3. **`manage_library.go`** - Library checkpoint operations

   - add, get, attach
   - move-step, remove-step, update

4. **`manage_lists.go`** - List operations for all entities

   - List projects, goals, journeys, checkpoints
   - Generic framework with pagination

5. **`manage_templates.go`** - Test template integration
   - Load templates
   - Generate commands
   - Get available templates

## Key Design Principles

### 1. Consistent Naming

- **Step commands**: `step_*.go` pattern aligns with CLI `step-*` commands
- **Management commands**: `manage_*.go` pattern for CRUD and control operations
- **Core files**: Simple names for infrastructure components

### 2. Functional Cohesion

- Related commands are grouped together
- Shared functionality reduces code duplication
- Each file has a clear, single purpose

### 3. Maintainability

- 30% fewer files to navigate
- Logical grouping makes changes easier
- Consistent patterns across all files

### 4. Single Package Structure

- All files remain in the commands package
- Avoids circular dependency issues
- Simplifies imports and references

## Major Consolidations

### 1. User Interactions (3 → 1 file)

- `interact.go` + `mouse.go` + `select.go` → **`step_interact.go`**
- Shares common interaction helpers
- Reduces duplication in element targeting

### 2. Browser Operations (2 → 1 file)

- `navigate.go` + `window.go` → **`step_browser.go`**
- Combined navigation and window management
- Shared browser state handling

### 3. Project Management (7 → 1 file)

- 7 individual CRUD files → **`manage_projects.go`**
- Shared output formatting functions
- Consistent error handling
- ~30% code reduction through consolidation

### 4. Execution Management (5 → 1 file)

- Multiple execution files → **`manage_executions.go`**
- Unified execution workflow
- Shared monitoring utilities

## Benefits Achieved

1. **File Reduction**: 27 → 19 files (30% reduction)
2. **Code Reduction**: ~30% less code through shared functions
3. **Better Organization**: Clear separation between step commands and management
4. **Improved Naming**: Consistent patterns make purpose obvious
5. **Easier Navigation**: Related functionality in single files
6. **AI-Friendly**: Fewer files with clearer structure

## Command Count

- **Total Commands**: 70+
- **Step Commands**: ~50 (across 8 files)
- **Management Commands**: ~20 (across 5 files)
- **Core Commands**: 2 (validate-config, session management)

## Future Considerations

1. Further consolidation should maintain balance between file size and functionality
2. New commands should follow established naming patterns
3. Shared utilities should be extracted to reduce duplication
4. Test coverage should be maintained through integration tests
