# Virtuoso API CLI Migration Guide: v3 to v4

## Overview

The Virtuoso API CLI v4 introduces a more consistent command syntax by moving from flag-based checkpoint specification to positional arguments. This change improves command readability and consistency across all command groups.

**Current Version:** v3.1 (with backward compatibility)
**Target Version:** v4.0
**Deprecation Date:** July 2025
**Compatibility Period:** 6 months

## Why Migrate?

1. **Consistency**: All commands follow the same pattern
2. **Simplicity**: Fewer flags to remember
3. **Performance**: Slightly faster command parsing
4. **Future Features**: New features will only support the new syntax

## Migration Timeline

- **January 2025**: v3.1 released with backward compatibility layer
- **January-June 2025**: Both syntaxes supported, deprecation warnings shown
- **July 2025**: v4.0 release, old syntax removed
- **Post-July 2025**: Only new syntax supported

## Command Migration Examples

### Assert Commands

**Old Syntax:**

```bash
api-cli assert exists "button.submit" --checkpoint 12345
api-cli assert equals "#title" "Welcome" --checkpoint 12345
api-cli assert not-exists ".error" --checkpoint 12345 --position 3
```

**New Syntax:**

```bash
api-cli assert exists 12345 "button.submit"
api-cli assert equals 12345 "#title" "Welcome"
api-cli assert not-exists 12345 ".error" 3
```

**Pattern:** `assert [type] [checkpoint] [selector] [value?] [position?]`

### Wait Commands

**Old Syntax:**

```bash
api-cli wait element "div.loaded" --checkpoint 12345
api-cli wait element-not-visible ".spinner" --checkpoint 12345
api-cli wait time 3000 --checkpoint 12345 --position 2
```

**New Syntax:**

```bash
api-cli wait element 12345 "div.loaded"
api-cli wait element-not-visible 12345 ".spinner"
api-cli wait time 12345 3000 2
```

**Pattern:** `wait [type] [checkpoint] [argument] [position?]`

### Mouse Commands

**Old Syntax:**

```bash
api-cli mouse move-to "100,200" --checkpoint 12345
api-cli mouse down --checkpoint 12345
api-cli mouse move-by "50,50" --checkpoint 12345 --position 4
```

**New Syntax:**

```bash
api-cli mouse move-to 12345 "100,200"
api-cli mouse down 12345
api-cli mouse move-by 12345 "50,50" 4
```

**Pattern:** `mouse [action] [checkpoint] [coordinates?] [position?]`

### Data Commands

**Old Syntax:**

```bash
api-cli data store-text "h1" "pageTitle" 3 --checkpoint 12345
api-cli data store-value "input#email" "userEmail" --checkpoint 12345
api-cli data cookie-create "session" "abc123" 5 --checkpoint 12345
```

**New Syntax:**

```bash
api-cli data store-text 12345 "h1" "pageTitle" 3
api-cli data store-value 12345 "input#email" "userEmail"
api-cli data cookie-create 12345 "session" "abc123" 5
```

**Pattern:** `data [action] [checkpoint] [args...] [position?]`

### Window Commands

**Old Syntax:**

```bash
api-cli window resize "1024x768" 2 --checkpoint 12345
api-cli window maximize --checkpoint 12345
api-cli window switch-tab "next" 4 --checkpoint 12345
api-cli window switch-iframe "#payment" --checkpoint 12345 --position 5
```

**New Syntax:**

```bash
api-cli window resize 12345 "1024x768" 2
api-cli window maximize 12345
api-cli window switch-tab 12345 "next" 4
api-cli window switch-iframe 12345 "#payment" 5
```

**Pattern:** `window [action] [checkpoint] [args...] [position?]`

### Interact Commands (No Change)

Interact commands already use positional arguments:

```bash
# These remain the same
api-cli interact click 12345 "button" 1
api-cli interact write 12345 "input" "text" 2
api-cli interact hover 12345 ".menu" 3
```

### Navigate Commands (No Change)

Navigate commands already use positional arguments:

```bash
# These remain the same
api-cli navigate to 12345 "https://example.com" 1
api-cli navigate scroll-by 12345 "0,500" 2
api-cli navigate scroll-top 12345 3
```

## Migration Script

Here's a simple bash script to help migrate your scripts:

```bash
#!/bin/bash
# migrate-virtuoso-commands.sh

# Function to migrate a single command
migrate_command() {
    local cmd="$1"

    # Extract checkpoint ID
    if [[ "$cmd" =~ --checkpoint[[:space:]]+([0-9]+) ]]; then
        checkpoint="${BASH_REMATCH[1]}"

        # Remove --checkpoint flag and value
        cmd_clean=$(echo "$cmd" | sed -E 's/--checkpoint[[:space:]]+[0-9]+//')

        # Extract command parts
        if [[ "$cmd_clean" =~ ^api-cli[[:space:]]+([a-z-]+)[[:space:]]+([a-z-]+)[[:space:]]+(.*) ]]; then
            main_cmd="${BASH_REMATCH[1]}"
            sub_cmd="${BASH_REMATCH[2]}"
            args="${BASH_REMATCH[3]}"

            # Reconstruct based on command type
            case "$main_cmd" in
                assert|wait|mouse)
                    echo "api-cli $main_cmd $sub_cmd $checkpoint $args"
                    ;;
                data|window)
                    echo "api-cli $main_cmd $sub_cmd $checkpoint $args"
                    ;;
                *)
                    echo "$cmd"  # No change needed
                    ;;
            esac
        else
            echo "$cmd"  # Return original if pattern doesn't match
        fi
    else
        echo "$cmd"  # No --checkpoint flag found
    fi
}

# Example usage:
# migrate_command "api-cli assert exists \"button\" --checkpoint 12345"
```

## Automated Migration Tools

### Using sed for Batch Migration

```bash
# Backup your scripts first!
cp script.sh script.sh.backup

# Migrate assert commands
sed -i 's/\(api-cli assert \w\+\) \(.*\) --checkpoint \([0-9]\+\)/\1 \3 \2/g' script.sh

# Migrate wait commands
sed -i 's/\(api-cli wait \w\+\) \(.*\) --checkpoint \([0-9]\+\)/\1 \3 \2/g' script.sh

# Migrate mouse commands
sed -i 's/\(api-cli mouse \w\+\) \(.*\) --checkpoint \([0-9]\+\)/\1 \3 \2/g' script.sh
```

### Using the CLI's Built-in Migration Helper

```bash
# Check which commands need migration
api-cli migrate check script.sh

# Show migration suggestions
api-cli migrate suggest script.sh

# Apply migrations (creates backup)
api-cli migrate apply script.sh
```

## Common Migration Patterns

### Pattern 1: Simple Flag Replacement

```bash
# Old
command subcommand args --checkpoint ID

# New
command subcommand ID args
```

### Pattern 2: With Position

```bash
# Old
command subcommand args --checkpoint ID --position N

# New
command subcommand ID args N
```

### Pattern 3: Data Commands with Position

```bash
# Old
data store-text "selector" "variable" 3 --checkpoint ID

# New
data store-text ID "selector" "variable" 3
```

## Testing Your Migration

1. **Run in Dry Mode**: Most commands support `-o json` to see what would happen
2. **Use Test Checkpoints**: Create test checkpoints to validate migrated commands
3. **Compare Outputs**: Old and new syntax should produce identical results

```bash
# Test old syntax (with deprecation warning)
api-cli assert exists "button" --checkpoint 12345 -o json > old.json

# Test new syntax
api-cli assert exists 12345 "button" -o json > new.json

# Compare results (should be identical except for timing)
diff old.json new.json
```

## Troubleshooting

### Command Not Working After Migration?

1. Check command order - checkpoint ID should come after the command type
2. Verify position is at the end (if used)
3. Ensure quotes are preserved around selectors and values

### Getting Deprecation Warnings?

The warnings show:

- The old command you used
- The correct new syntax
- When support will be removed

### Need Help?

- Run `api-cli help [command]` for new syntax examples
- Check test files for working examples
- Refer to `CLAUDE.md` for comprehensive documentation

## Benefits After Migration

1. **Cleaner Scripts**: Less verbose, easier to read
2. **Better Performance**: Faster argument parsing
3. **Future Ready**: Access to new v4 features
4. **Consistency**: All commands follow the same pattern

## Summary

The migration from v3 to v4 syntax is straightforward:

1. Move `--checkpoint ID` to positional argument after command type
2. Keep other arguments in the same relative order
3. Position (if used) stays at the end

Take advantage of the 6-month compatibility period to migrate at your own pace. The backward compatibility layer ensures your existing scripts continue to work while showing helpful migration hints.
