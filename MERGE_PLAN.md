# Version B to Version A Merge Plan

## Overview
This document outlines the comprehensive merge plan for integrating Version B's enhancements into Version A of the virtuoso-api-cli-generator project.

## Version B Enhancements to Merge

### 1. New Command Files (15 files)
These commands are unique to Version B and should be copied to Version A:

```
src/cmd/create-step-cookie-create.go
src/cmd/create-step-cookie-wipe-all.go
src/cmd/create-step-dismiss-prompt-with-text.go
src/cmd/create-step-execute-script.go
src/cmd/create-step-mouse-move-by.go
src/cmd/create-step-mouse-move-to.go
src/cmd/create-step-pick-index.go
src/cmd/create-step-pick-last.go
src/cmd/create-step-scroll.go
src/cmd/create-step-store-element-text.go
src/cmd/create-step-store-literal-value.go
src/cmd/create-step-upload-url.go
src/cmd/create-step-wait-for-element-default.go
src/cmd/create-step-wait-for-element-timeout.go
src/cmd/create-step-window-resize.go
```

### 2. Enhanced Commands (7 additional files)
Commands that exist in both but Version B has enhanced implementations:

```
src/cmd/create-step-navigate.go (with --new-tab option)
src/cmd/create-step-click.go (with --variable and positioning options)
src/cmd/create-step-write.go (with --variable option)
src/cmd/create-step-key.go (with --target option)
src/cmd/create-step-comment.go
src/cmd/create-step-switch-iframe.go
src/cmd/create-step-switch-parent-frame.go
```

### 3. Client API Methods
Version B's client.go includes these specialized methods that need to be added to Version A's client:

- Cookie Management Methods:
  - `CreateStepCookieCreate()`
  - `CreateStepCookieWipeAll()`

- Enhanced Navigation Methods:
  - `CreateStepNavigate()` with new-tab support
  - `CreateStepExecuteScript()`

- Mouse Movement Methods:
  - `CreateStepMouseMoveTo()`
  - `CreateStepMouseMoveBy()`

- Element Selection Methods:
  - `CreateStepPickIndex()`
  - `CreateStepPickLast()`

- Wait Methods:
  - `CreateStepWaitForElementTimeout()`
  - `CreateStepWaitForElementDefault()`

- Storage Methods:
  - `CreateStepStoreElementText()`
  - `CreateStepStoreLiteralValue()`

- Scroll Methods:
  - `CreateStepScrollToPosition()`
  - `CreateStepScrollByOffset()`
  - `CreateStepScrollToTop()`

- Window Methods:
  - `CreateStepWindowResize()`

- Assertion Methods:
  - `CreateStepAssertNotEquals()`
  - `CreateStepAssertGreaterThan()`
  - `CreateStepAssertGreaterThanOrEqual()`
  - `CreateStepAssertMatches()`

- Interaction Methods:
  - `CreateStepClick()` with advanced options
  - `CreateStepWrite()` with variable support
  - `CreateStepKey()` with targeting

### 4. Documentation Files
- `CLAUDE.md` - Comprehensive project documentation
- `NEW_COMMANDS_SUMMARY.md` - New commands documentation
- `COMPREHENSIVE_TEST_RESULTS.md` - Test results
- `IMPLEMENTATION_SUMMARY.md` - Implementation details

### 5. Test Scripts
- `test-all-commands-variations.sh`
- `test-new-commands.sh`
- `test-fixed-commands.sh`
- `test-cookie-commands.sh`
- `test-runtime-failures.sh`

### 6. Command Registration Updates
The main.go init() function needs to include all Version B command registrations.

## Merge Strategy

### Phase 1: File Copying
1. Copy all unique command files from Version B to Version A
2. Copy documentation files
3. Copy test scripts

### Phase 2: Client Enhancement
1. Merge Version B's client methods into Version A's client.go
2. Ensure all helper methods are included
3. Maintain Version A's existing functionality

### Phase 3: Command Registration
1. Update Version A's main.go to register all Version B commands
2. Organize commands by category as in Version B

### Phase 4: Testing & Validation
1. Run all Version B test scripts in Version A
2. Verify existing Version A functionality still works
3. Test integration between old and new features

## File Mapping

### Direct Copies (No conflicts)
- All 15 new command files
- All documentation files
- All test scripts

### Files Requiring Merge
- `src/cmd/main.go` - Add Version B command registrations
- `pkg/virtuoso/client.go` - Add Version B methods
- `README.md` - Update with Version B features

### Files to Preserve from Version A
- All project/goal/journey management commands
- Configuration system
- Build infrastructure (.dockerignore, .goreleaser.yml, etc.)