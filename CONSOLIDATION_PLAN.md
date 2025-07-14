# CLI Command Consolidation Implementation Plan

## Executive Summary

This plan outlines the consolidation of 54 separate CLI commands into 10 unified commands for better maintainability and usability. The consolidation will maintain full backward compatibility while providing a cleaner, more intuitive interface.

## Current State Analysis

### Command Categories (54 commands total)

1. **Assertions (12 commands)**

   - assert-checked, assert-equals, assert-exists, assert-greater-than, assert-greater-than-or-equal
   - assert-less-than, assert-less-than-or-equal, assert-matches, assert-not-equals
   - assert-not-exists, assert-selected, assert-variable

2. **Navigation (5 commands)**

   - navigate, scroll-bottom, scroll-element, scroll-position, scroll-top
   - scroll-to-position, scroll-by-offset, scroll-to-top

3. **Interactions (5 commands)**

   - click, double-click, hover, right-click, write, key

4. **Window Operations (6 commands)**

   - window, window-resize, switch-iframe, switch-next-tab
   - switch-parent-frame, switch-prev-tab

5. **Mouse Operations (6 commands)**

   - mouse-down, mouse-enter, mouse-move, mouse-move-by
   - mouse-move-to, mouse-up

6. **Data Operations (5 commands)**

   - store-element-text, store-literal-value, store, store-value
   - cookie-create, cookie-wipe-all, add-cookie, delete-cookie, clear-cookies

7. **Dialog Operations (4 commands)**

   - dismiss-alert, dismiss-confirm, dismiss-prompt, dismiss-prompt-with-text

8. **Wait Operations (4 commands)**

   - wait-element, wait-time, wait-for-element-default, wait-for-element-timeout

9. **File Operations (2 commands)**

   - upload, upload-url

10. **Selection Operations (5 commands)**

    - pick, pick-index, pick-last, pick-text, pick-value

11. **Other Operations (3 commands)**
    - comment, execute-script, execute-js

## Proposed Consolidated Structure

### 1. `api-cli assert [type]` - Assertion Commands

```bash
api-cli assert equals ELEMENT VALUE [POSITION] [--checkpoint ID]
api-cli assert not-equals ELEMENT VALUE [POSITION] [--checkpoint ID]
api-cli assert exists ELEMENT [POSITION] [--checkpoint ID]
api-cli assert not-exists ELEMENT [POSITION] [--checkpoint ID]
api-cli assert checked ELEMENT [POSITION] [--checkpoint ID]
api-cli assert selected ELEMENT VALUE [POSITION] [--checkpoint ID]
api-cli assert variable NAME VALUE [POSITION] [--checkpoint ID]
api-cli assert greater-than ELEMENT VALUE [POSITION] [--checkpoint ID]
api-cli assert greater-than-or-equal ELEMENT VALUE [POSITION] [--checkpoint ID]
api-cli assert less-than ELEMENT VALUE [POSITION] [--checkpoint ID]
api-cli assert less-than-or-equal ELEMENT VALUE [POSITION] [--checkpoint ID]
api-cli assert matches ELEMENT PATTERN [POSITION] [--checkpoint ID]
```

### 2. `api-cli interact [action]` - Interaction Commands

```bash
api-cli interact click ELEMENT [POSITION] [--checkpoint ID] [--variable VAR] [--position POS] [--element-type TYPE]
api-cli interact double-click ELEMENT [POSITION] [--checkpoint ID]
api-cli interact right-click ELEMENT [POSITION] [--checkpoint ID]
api-cli interact hover ELEMENT [POSITION] [--checkpoint ID]
api-cli interact write ELEMENT TEXT [POSITION] [--checkpoint ID] [--variable VAR]
api-cli interact key KEY [POSITION] [--checkpoint ID] [--target ELEMENT]
```

### 3. `api-cli navigate [action]` - Navigation Commands

```bash
api-cli navigate url URL [POSITION] [--checkpoint ID] [--new-tab]
api-cli navigate scroll-to X Y [POSITION] [--checkpoint ID]
api-cli navigate scroll-by X Y [POSITION] [--checkpoint ID]
api-cli navigate scroll-top [POSITION] [--checkpoint ID]
api-cli navigate scroll-bottom [POSITION] [--checkpoint ID]
api-cli navigate scroll-element ELEMENT [POSITION] [--checkpoint ID]
```

### 4. `api-cli window [action]` - Window/Frame Commands

```bash
api-cli window resize WIDTH HEIGHT [POSITION] [--checkpoint ID]
api-cli window switch-tab next [POSITION] [--checkpoint ID]
api-cli window switch-tab prev [POSITION] [--checkpoint ID]
api-cli window switch-frame IFRAME [POSITION] [--checkpoint ID]
api-cli window switch-frame parent [POSITION] [--checkpoint ID]
```

### 5. `api-cli mouse [action]` - Mouse Commands

```bash
api-cli mouse move-to X Y [POSITION] [--checkpoint ID]
api-cli mouse move-by X Y [POSITION] [--checkpoint ID]
api-cli mouse move ELEMENT [POSITION] [--checkpoint ID]
api-cli mouse down ELEMENT [POSITION] [--checkpoint ID]
api-cli mouse up ELEMENT [POSITION] [--checkpoint ID]
api-cli mouse enter ELEMENT [POSITION] [--checkpoint ID]
```

### 6. `api-cli data [action]` - Data Management Commands

```bash
api-cli data store-text ELEMENT VARIABLE [POSITION] [--checkpoint ID]
api-cli data store-value VALUE VARIABLE [POSITION] [--checkpoint ID]
api-cli data cookie-create NAME VALUE [POSITION] [--checkpoint ID]
api-cli data cookie-delete NAME [POSITION] [--checkpoint ID]
api-cli data cookie-clear [POSITION] [--checkpoint ID]
```

### 7. `api-cli dialog [action]` - Dialog Commands

```bash
api-cli dialog dismiss-alert [POSITION] [--checkpoint ID]
api-cli dialog dismiss-confirm [POSITION] [--checkpoint ID] [--accept]
api-cli dialog dismiss-prompt [TEXT] [POSITION] [--checkpoint ID]
```

### 8. `api-cli wait [type]` - Wait Commands

```bash
api-cli wait element ELEMENT [TIMEOUT] [POSITION] [--checkpoint ID]
api-cli wait time MILLISECONDS [POSITION] [--checkpoint ID]
```

### 9. `api-cli file [action]` - File Commands

```bash
api-cli file upload URL ELEMENT [POSITION] [--checkpoint ID]
```

### 10. `api-cli select [action]` - Selection Commands

```bash
api-cli select option ELEMENT VALUE [POSITION] [--checkpoint ID]
api-cli select index ELEMENT INDEX [POSITION] [--checkpoint ID]
api-cli select last ELEMENT [POSITION] [--checkpoint ID]
api-cli select text ELEMENT TEXT [POSITION] [--checkpoint ID]
```

### 11. `api-cli misc [action]` - Miscellaneous Commands

```bash
api-cli misc comment TEXT [POSITION] [--checkpoint ID]
api-cli misc execute-script SCRIPT [POSITION] [--checkpoint ID]
```

## Implementation Strategy

### Phase 1: Create Shared Infrastructure (Week 1)

1. **Create base command structure**

   ```go
   // pkg/api-cli/commands/base_command.go
   type BaseStepCommand struct {
       CheckpointFlag int
       OutputFormat   string
       // Common fields
   }
   ```

2. **Extract common functionality**

   - Move `resolveStepContext()` to shared package
   - Move `outputStepResult()` to shared package
   - Create command registry pattern

3. **Create command router**
   ```go
   // pkg/api-cli/commands/consolidated/router.go
   type CommandRouter struct {
       subcommands map[string]SubCommand
   }
   ```

### Phase 2: Implement Consolidated Commands (Week 2-3)

1. **Create consolidated command files**

   ```
   pkg/api-cli/commands/consolidated/
   ├── assert.go
   ├── interact.go
   ├── navigate.go
   ├── window.go
   ├── mouse.go
   ├── data.go
   ├── dialog.go
   ├── wait.go
   ├── file.go
   ├── select.go
   └── misc.go
   ```

2. **Implement subcommand routing**
   - Each consolidated command will parse its subcommand
   - Route to appropriate handler function
   - Maintain consistent argument parsing

### Phase 3: Backward Compatibility Layer (Week 3)

1. **Create legacy command wrappers**

   ```go
   // pkg/api-cli/commands/legacy/wrapper.go
   func CreateLegacyWrapper(newCmd string, subCmd string) *cobra.Command {
       // Wrapper that translates old command to new format
   }
   ```

2. **Update register.go**

   - Register both new consolidated commands
   - Register legacy commands with deprecation warnings

3. **Add deprecation notices**
   ```go
   cmd.Deprecated = "Use 'api-cli assert equals' instead"
   ```

### Phase 4: Test Migration (Week 4)

1. **Create new test suite**

   - Test consolidated commands
   - Test legacy compatibility
   - Test argument parsing

2. **Update existing tests**

   - Add parallel tests for new commands
   - Verify output compatibility

3. **Create migration scripts**
   - Script to update user scripts from old to new format

## File Structure

```
pkg/api-cli/
├── commands/
│   ├── consolidated/         # New consolidated commands
│   │   ├── base.go          # Base command structure
│   │   ├── router.go        # Command routing
│   │   ├── assert.go        # Assert subcommands
│   │   ├── interact.go      # Interact subcommands
│   │   └── ...
│   ├── legacy/              # Legacy command wrappers
│   │   └── wrapper.go
│   ├── shared/              # Shared functionality
│   │   ├── context.go       # Step context resolution
│   │   ├── output.go        # Output formatting
│   │   └── helpers.go       # Common helpers
│   └── register.go          # Command registration
```

## Migration Guide for Users

### Before (Old Format)

```bash
api-cli create-step-assert-equals "Username" "john@example.com" 1
api-cli create-step-click "Submit" 1 --variable "result"
api-cli create-step-navigate "https://example.com" 1 --new-tab
```

### After (New Format)

```bash
api-cli assert equals "Username" "john@example.com" 1
api-cli interact click "Submit" 1 --variable "result"
api-cli navigate url "https://example.com" 1 --new-tab
```

## Benefits

1. **Improved Discoverability**

   - Users can explore subcommands with `api-cli assert --help`
   - Logical grouping makes commands easier to find

2. **Reduced Maintenance**

   - Shared code reduces duplication
   - Easier to add new subcommands

3. **Better User Experience**

   - Consistent command structure
   - Cleaner command namespace

4. **Future Extensibility**
   - Easy to add new subcommands
   - Plugin architecture possible

## Risk Mitigation

1. **Backward Compatibility**

   - All old commands continue to work
   - Deprecation warnings guide users
   - Migration period of 6+ months

2. **Testing Coverage**

   - Comprehensive test suite
   - Automated compatibility testing
   - Performance benchmarks

3. **Documentation**
   - Update all documentation
   - Create migration guides
   - Video tutorials for new structure

## Timeline

- **Week 1**: Infrastructure and planning
- **Week 2-3**: Implementation of consolidated commands
- **Week 3**: Backward compatibility layer
- **Week 4**: Testing and documentation
- **Week 5**: Release candidate and user testing
- **Week 6**: Production release with migration support

## Success Criteria

1. All 54 existing commands work via legacy layer
2. All new consolidated commands pass tests
3. No performance regression
4. Documentation fully updated
5. User migration guide available
6. Positive feedback from beta users

## Next Steps

1. Review and approve this plan
2. Create feature branch for consolidation
3. Begin Phase 1 implementation
4. Set up CI/CD for compatibility testing
