# API CLI Command Migration Guide

## Overview

This guide helps users transition from the old command format to the new consolidated command structure. The new structure provides better organization, discoverability, and consistency while maintaining full backward compatibility.

## Migration Tools

### 1. Automatic Migration Script

The `scripts/migrate-commands.sh` script helps you automatically update your scripts:

```bash
# Generate a migration report for all shell scripts
./scripts/migrate-commands.sh -r *.sh

# Dry run to see what would change
./scripts/migrate-commands.sh -d my-script.sh

# Auto-update with backups
./scripts/migrate-commands.sh -a -b *.sh

# Update specific files
./scripts/migrate-commands.sh -a script1.sh script2.sh
```

### 2. Legacy Command Support

All old commands continue to work but will show deprecation warnings:

```bash
# Old command (deprecated)
api-cli create-step-assert-equals 1680449 "Username" "john@example.com" 1

# Shows warning:
⚠️  DEPRECATION WARNING
The command 'create-step-assert-equals' is deprecated and will be removed in a future version.
Please use: api-cli assert equals
```

### 3. Testing Tools

Verify your migration with included test scripts:

```bash
# Test both old and new formats
./test-all-commands.sh

# Test only consolidated commands
./test-consolidated-commands.sh

# Verify command equivalence
./test-migration-equivalence.sh
```

## Command Mapping Reference

### Assert Commands

| Old Command                                | New Command         | Example                                                   |
| ------------------------------------------ | ------------------- | --------------------------------------------------------- |
| `create-step-assert-equals`                | `assert equals`     | `api-cli assert equals "Username" "john@example.com" 1`   |
| `create-step-assert-not-equals`            | `assert not-equals` | `api-cli assert not-equals "Status" "Error" 1`            |
| `create-step-assert-exists`                | `assert exists`     | `api-cli assert exists "Login button" 1`                  |
| `create-step-assert-not-exists`            | `assert not-exists` | `api-cli assert not-exists "Error message" 1`             |
| `create-step-assert-checked`               | `assert checked`    | `api-cli assert checked "Terms checkbox" 1`               |
| `create-step-assert-selected`              | `assert selected`   | `api-cli assert selected "Country" "USA" 1`               |
| `create-step-assert-variable`              | `assert variable`   | `api-cli assert variable "username" "testuser" 1`         |
| `create-step-assert-greater-than`          | `assert gt`         | `api-cli assert gt "Count" "10" 1`                        |
| `create-step-assert-greater-than-or-equal` | `assert gte`        | `api-cli assert gte "Count" "10" 1`                       |
| `create-step-assert-less-than`             | `assert lt`         | `api-cli assert lt "Count" "100" 1`                       |
| `create-step-assert-less-than-or-equal`    | `assert lte`        | `api-cli assert lte "Count" "100" 1`                      |
| `create-step-assert-matches`               | `assert matches`    | `api-cli assert matches "Email" "^[a-z]+@[a-z]+\.com$" 1` |

### Interact Commands

| Old Command                | New Command             | Example                                                     |
| -------------------------- | ----------------------- | ----------------------------------------------------------- |
| `create-step-click`        | `interact click`        | `api-cli interact click "Submit button" 1`                  |
| `create-step-double-click` | `interact double-click` | `api-cli interact double-click "Icon" 1`                    |
| `create-step-right-click`  | `interact right-click`  | `api-cli interact right-click "Menu item" 1`                |
| `create-step-hover`        | `interact hover`        | `api-cli interact hover "Tooltip trigger" 1`                |
| `create-step-write`        | `interact write`        | `api-cli interact write "Email field" "test@example.com" 1` |
| `create-step-key`          | `interact key`          | `api-cli interact key "Enter" 1`                            |

### Navigate Commands

| Old Command                   | New Command               | Example                                          |
| ----------------------------- | ------------------------- | ------------------------------------------------ |
| `create-step-navigate`        | `navigate url`            | `api-cli navigate url "https://example.com" 1`   |
| `create-step-scroll-position` | `navigate scroll-to`      | `api-cli navigate scroll-to 0 500 1`             |
| `create-step-scroll-top`      | `navigate scroll-top`     | `api-cli navigate scroll-top 1`                  |
| `create-step-scroll-bottom`   | `navigate scroll-bottom`  | `api-cli navigate scroll-bottom 1`               |
| `create-step-scroll-element`  | `navigate scroll-element` | `api-cli navigate scroll-element "div#footer" 1` |

### Window Commands

| Old Command                       | New Command                  | Example                                          |
| --------------------------------- | ---------------------------- | ------------------------------------------------ |
| `create-step-window-resize`       | `window resize`              | `api-cli window resize 1024 768 1`               |
| `create-step-switch-next-tab`     | `window switch-tab next`     | `api-cli window switch-tab next 1`               |
| `create-step-switch-prev-tab`     | `window switch-tab prev`     | `api-cli window switch-tab prev 1`               |
| `create-step-switch-iframe`       | `window switch-frame`        | `api-cli window switch-frame "iframe#content" 1` |
| `create-step-switch-parent-frame` | `window switch-frame parent` | `api-cli window switch-frame parent 1`           |

### Mouse Commands

| Old Command                 | New Command     | Example                                  |
| --------------------------- | --------------- | ---------------------------------------- |
| `create-step-mouse-move-to` | `mouse move-to` | `api-cli mouse move-to 100 200 1`        |
| `create-step-mouse-move-by` | `mouse move-by` | `api-cli mouse move-by 50 50 1`          |
| `create-step-mouse-move`    | `mouse move`    | `api-cli mouse move "div.target" 1`      |
| `create-step-mouse-down`    | `mouse down`    | `api-cli mouse down "div.draggable" 1`   |
| `create-step-mouse-up`      | `mouse up`      | `api-cli mouse up "div.drop-zone" 1`     |
| `create-step-mouse-enter`   | `mouse enter`   | `api-cli mouse enter "div.hover-area" 1` |

### Data Commands

| Old Command                       | New Command          | Example                                             |
| --------------------------------- | -------------------- | --------------------------------------------------- |
| `create-step-store-element-text`  | `data store-text`    | `api-cli data store-text "h1.title" "pageTitle" 1`  |
| `create-step-store-literal-value` | `data store-value`   | `api-cli data store-value "test123" "myVar" 1`      |
| `create-step-cookie-create`       | `data cookie-create` | `api-cli data cookie-create "sessionId" "abc123" 1` |
| `create-step-delete-cookie`       | `data cookie-delete` | `api-cli data cookie-delete "sessionId" 1`          |
| `create-step-cookie-wipe-all`     | `data cookie-clear`  | `api-cli data cookie-clear 1`                       |

### Dialog Commands

| Old Command                            | New Command              | Example                                  |
| -------------------------------------- | ------------------------ | ---------------------------------------- |
| `create-step-dismiss-alert`            | `dialog dismiss-alert`   | `api-cli dialog dismiss-alert 1`         |
| `create-step-dismiss-confirm`          | `dialog dismiss-confirm` | `api-cli dialog dismiss-confirm 1`       |
| `create-step-dismiss-prompt`           | `dialog dismiss-prompt`  | `api-cli dialog dismiss-prompt 1`        |
| `create-step-dismiss-prompt-with-text` | `dialog dismiss-prompt`  | `api-cli dialog dismiss-prompt "text" 1` |

### Wait Commands

| Old Command                            | New Command    | Example                                    |
| -------------------------------------- | -------------- | ------------------------------------------ |
| `create-step-wait-element`             | `wait element` | `api-cli wait element "div.loaded" 1`      |
| `create-step-wait-for-element-default` | `wait element` | `api-cli wait element "div.loaded" 1`      |
| `create-step-wait-for-element-timeout` | `wait element` | `api-cli wait element "div.loaded" 5000 1` |
| `create-step-wait-time`                | `wait time`    | `api-cli wait time 3000 1`                 |

### File Commands

| Old Command              | New Command   | Example                                                                   |
| ------------------------ | ------------- | ------------------------------------------------------------------------- |
| `create-step-upload`     | `file upload` | `api-cli file upload "https://example.com/file.pdf" "input[type=file]" 1` |
| `create-step-upload-url` | `file upload` | `api-cli file upload "https://example.com/file.pdf" "input[type=file]" 1` |

### Select Commands

| Old Command              | New Command     | Example                                          |
| ------------------------ | --------------- | ------------------------------------------------ |
| `create-step-pick`       | `select option` | `api-cli select option "select#country" "USA" 1` |
| `create-step-pick-index` | `select index`  | `api-cli select index "select#country" 2 1`      |
| `create-step-pick-last`  | `select last`   | `api-cli select last "select#country" 1`         |

### Miscellaneous Commands

| Old Command                  | New Command           | Example                                               |
| ---------------------------- | --------------------- | ----------------------------------------------------- |
| `create-step-comment`        | `misc comment`        | `api-cli misc comment "Test comment" 1`               |
| `create-step-execute-script` | `misc execute-script` | `api-cli misc execute-script "console.log('test')" 1` |

## Migration Best Practices

### 1. Gradual Migration

- Start by running the migration script in report mode to understand the scope
- Use dry-run mode to preview changes before applying them
- Migrate one script at a time if you have many

### 2. Testing

- Always test migrated scripts in a non-production environment first
- Use the equivalence test script to verify commands produce the same results
- Keep backups of original scripts until migration is complete

### 3. CI/CD Updates

If your CI/CD pipeline uses these commands:

1. Update your pipeline scripts using the migration tool
2. Test thoroughly in a staging environment
3. Consider running both old and new commands in parallel initially
4. Remove old commands once confident in the migration

### 4. Team Communication

- Inform your team about the migration
- Share this guide and the new command structure
- Update internal documentation and wikis
- Consider team training sessions for the new structure

## Benefits of Migration

### 1. Better Organization

Commands are logically grouped:

- `assert` - All assertion types
- `interact` - User interactions
- `navigate` - Navigation and scrolling
- `window` - Window and frame management
- etc.

### 2. Improved Discoverability

```bash
# See all assertion options
api-cli assert --help

# See all interaction options
api-cli interact --help
```

### 3. Consistent Patterns

All commands follow the same pattern:

```bash
api-cli [category] [action] [args...] [position] [--flags]
```

### 4. Future-Proof

The new structure makes it easier to:

- Add new commands without cluttering the namespace
- Maintain consistent behavior across related commands
- Provide better help and documentation

## Support

### Getting Help

1. Use `--help` with any command for detailed information
2. Check the test scripts for examples
3. Review the migration report for specific changes

### Reporting Issues

If you encounter any issues during migration:

1. Check if the issue is already known in the migration report
2. Verify you're using the latest version of the CLI
3. Test with both old and new commands to isolate the issue
4. Report with specific examples and error messages

### Timeline

- **Current**: Both old and new commands work (old shows deprecation warnings)
- **6 months**: Deprecation warnings become more prominent
- **12 months**: Old commands removed in major version update

Plan your migration accordingly to avoid disruption.

## Examples

### Before Migration

```bash
#!/bin/bash
# Old script using legacy commands

CHECKPOINT=1680449

./bin/api-cli create-step-navigate $CHECKPOINT "https://example.com" 1
./bin/api-cli create-step-assert-exists $CHECKPOINT "Login form" 2
./bin/api-cli create-step-write $CHECKPOINT "input#username" "testuser" 3
./bin/api-cli create-step-write $CHECKPOINT "input#password" "testpass" 4
./bin/api-cli create-step-click $CHECKPOINT "button.submit" 5
./bin/api-cli create-step-wait-element $CHECKPOINT "div.dashboard" 6
./bin/api-cli create-step-assert-equals $CHECKPOINT "h1.welcome" "Welcome, testuser" 7
```

### After Migration

```bash
#!/bin/bash
# New script using consolidated commands

CHECKPOINT=1680449

./bin/api-cli navigate url "https://example.com" 1 --checkpoint $CHECKPOINT
./bin/api-cli assert exists "Login form" 2 --checkpoint $CHECKPOINT
./bin/api-cli interact write "input#username" "testuser" 3 --checkpoint $CHECKPOINT
./bin/api-cli interact write "input#password" "testpass" 4 --checkpoint $CHECKPOINT
./bin/api-cli interact click "button.submit" 5 --checkpoint $CHECKPOINT
./bin/api-cli wait element "div.dashboard" 6 --checkpoint $CHECKPOINT
./bin/api-cli assert equals "h1.welcome" "Welcome, testuser" 7 --checkpoint $CHECKPOINT
```

### Using Session Context

```bash
#!/bin/bash
# Even cleaner with session context

# Set checkpoint once
./bin/api-cli set-checkpoint 1680449

# All subsequent commands use the session checkpoint
./bin/api-cli navigate url "https://example.com" 1
./bin/api-cli assert exists "Login form" 2
./bin/api-cli interact write "input#username" "testuser" 3
./bin/api-cli interact write "input#password" "testpass" 4
./bin/api-cli interact click "button.submit" 5
./bin/api-cli wait element "div.dashboard" 6
./bin/api-cli assert equals "h1.welcome" "Welcome, testuser" 7
```

---

**Last Updated**: 2025-01-14
**Migration Tool Version**: 1.0.0
**Status**: Active Migration Period
