# Virtuoso API CLI - Command Syntax Fixes

This document shows the correct syntax for commands that were previously failing, based on analysis of successful test logs from previous commits.

## Key Syntax Patterns

### 1. **Nested Subcommands**

Many commands use a nested structure with spaces between levels:

- ❌ `data store-text`
- ✅ `data store element-text`

- ❌ `dialog dismiss-alert`
- ✅ `dialog dismiss alert`

- ❌ `window switch-tab`
- ✅ `window switch tab`

### 2. **Format Requirements**

Some commands have specific format requirements:

- ❌ `window resize 1280 720`
- ✅ `window resize 1280x720` (must use WIDTHxHEIGHT format)

- ❌ `file upload '#file-input' 'file.pdf'`
- ✅ `file upload 'https://example.com/file.pdf' '#file-input'` (URL first, then selector)

## Complete Command Reference

### Assert Commands

```bash
assert exists 'element'
assert not-exists 'element'
assert equals 'element' 'value'
assert not-equals 'element' 'value'
assert checked 'element'
assert selected '#dropdown' [position]  # Note: needs selector AND position
assert variable 'varName' 'value'
assert gt 'element' 'value'
assert gte 'element' 'value'
assert lt 'element' 'value'
assert lte 'element' 'value'
assert matches 'element' 'regex'
```

### Data Commands

```bash
# Store subcommands
data store element-text 'selector' 'variable'
data store literal 'value' 'variable'

# Cookie subcommands
data cookie create 'name' 'value'
data cookie delete 'name'
data cookie clear-all
```

### Dialog Commands

```bash
dialog dismiss alert
dialog dismiss confirm [--accept|--reject]
dialog dismiss prompt [--accept|--reject]
dialog dismiss prompt-with-text 'text' [--accept|--reject]
```

### Navigate Commands

```bash
navigate to 'url' [--new-tab]
navigate scroll-top [--smooth]
navigate scroll-bottom [--smooth]
navigate scroll-position X,Y [--smooth]
# navigate scroll-element (NOT IMPLEMENTED)
```

### Window Commands

```bash
window resize WIDTHxHEIGHT  # e.g., 1024x768
window switch tab next
window switch tab prev
window switch iframe 'selector'
window switch parent-frame
```

### Mouse Commands

```bash
mouse move-to X Y
mouse move-by DX DY
mouse move 'selector'
mouse down 'selector'  # Note: needs selector
mouse up 'selector'    # Note: needs selector
mouse enter 'selector'
```

### Select Commands

```bash
select index 'selector' INDEX  # 0-based index
select last 'selector'
# Note: 'select option' by text is not supported
```

### File Commands

```bash
file upload 'URL' 'selector'  # URL first, then selector
```

### Misc Commands

```bash
misc comment 'text'
misc execute 'javascript'
# Note: 'misc key' is not implemented, use 'interact key' instead
```

### Wait Commands

```bash
wait element 'selector' [--timeout MS]
wait time MS
```

### Library Commands

```bash
library add CHECKPOINT_ID
library get LIBRARY_ID
library attach JOURNEY_ID LIBRARY_ID POSITION
```

## Common Mistakes to Avoid

1. **Using hyphens instead of spaces in nested commands**

   - Wrong: `data store-text`, `dialog dismiss-alert`
   - Right: `data store element-text`, `dialog dismiss alert`

2. **Wrong parameter order**

   - Wrong: `file upload '#input' 'file.pdf'`
   - Right: `file upload 'https://example.com/file.pdf' '#input'`

3. **Missing required parameters**

   - Wrong: `assert selected 'USA'`
   - Right: `assert selected '#dropdown' 1`

4. **Wrong format for specific commands**

   - Wrong: `window resize 1024 768`
   - Right: `window resize 1024x768`

5. **Using non-existent subcommands**
   - Wrong: `select option '#dropdown' 'USA'`
   - Right: `select index '#dropdown' 2`

## Output Formats

All commands support these output formats:

```bash
--output human  # Default, human-readable
--output json   # JSON format
--output yaml   # YAML format
--output ai     # AI-optimized format
```

## Session Context

Commands can use session context to auto-increment position:

```bash
export VIRTUOSO_SESSION_ID=CHECKPOINT_ID
# Now commands will auto-increment position
```

Or specify checkpoint and position explicitly:

```bash
command --checkpoint CHECKPOINT_ID POSITION
```
