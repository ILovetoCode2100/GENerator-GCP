# Step Variations and Merge Opportunities Analysis

## Executive Summary

After analyzing the API CLI Generator codebase, I've identified significant opportunities for consolidating and standardizing step commands. The current implementation has 40+ individual step commands that could be merged into more flexible, parameterized commands while maintaining backward compatibility.

## Current Step Command Structure

Currently, each step type has its own dedicated command:
- `create-step-navigate`
- `create-step-click`
- `create-step-write`
- `create-step-assert-equals`
- etc.

Each command follows a similar pattern with minor variations in:
1. Number of arguments (1-4)
2. Argument types (element, value, seconds, etc.)
3. Action type mapping
4. Metadata requirements

## Identified Step Variations

### 1. Navigation Variations
**Current Commands:**
- `create-step-navigate` - Navigate to URL
- `create-step-window` - Resize browser window

**Possible Variations:**
- Navigate with options: `--new-tab`, `--new-window`, `--incognito`
- Navigate with wait: `--wait-for-element`, `--wait-seconds`
- Navigate with validation: `--expect-url`, `--expect-title`
- Browser control: `--back`, `--forward`, `--refresh`

### 2. Wait Variations
**Current Commands:**
- `create-step-wait-time` - Wait for X seconds
- `create-step-wait-element` - Wait for element to appear

**Possible Variations:**
- Wait conditions: `--until-visible`, `--until-hidden`, `--until-clickable`
- Wait for text: `--contains-text "text"`, `--exact-text "text"`
- Wait for attribute: `--has-attribute "name=value"`
- Wait for URL change: `--url-contains`, `--url-matches`
- Custom timeout: `--timeout 30`

### 3. Click Variations
**Current Commands:**
- `create-step-click`
- `create-step-double-click`
- `create-step-right-click`
- `create-step-hover`

**Possible Variations:**
- Click with modifiers: `--ctrl`, `--shift`, `--alt`, `--meta`
- Click position: `--offset-x 10 --offset-y 20`
- Click and hold: `--hold-duration 2`
- Multiple clicks: `--count 3`
- Force click: `--force` (bypass visibility checks)

### 4. Input Variations
**Current Commands:**
- `create-step-write` - Type text
- `create-step-key` - Press key
- `create-step-pick` - Select from dropdown
- `create-step-upload` - Upload file

**Possible Variations:**
- Write options: `--clear-first`, `--append`, `--type-slowly`
- Key combinations: `--keys "ctrl+a"`, `--keys "shift+tab"`
- Multiple selections: `--multi-select`
- File upload: `--drag-drop`, `--accept-types ".pdf,.jpg"`
- Input validation: `--max-length`, `--pattern`

### 5. Assertion Variations
**Current Commands:**
- `create-step-assert-exists`
- `create-step-assert-not-exists`
- `create-step-assert-equals`
- `create-step-assert-checked`
- `create-step-assert-variable`

**Possible Variations:**
- Text assertions: `--contains`, `--starts-with`, `--ends-with`, `--matches-regex`
- Numeric assertions: `--greater-than`, `--less-than`, `--between`
- Attribute assertions: `--has-class`, `--has-attribute`, `--attribute-equals`
- Style assertions: `--css-property "color=red"`, `--is-visible`, `--is-enabled`
- Count assertions: `--count-equals`, `--count-greater-than`

### 6. Scroll Variations
**Current Commands:**
- `create-step-scroll-top`
- `create-step-scroll-bottom`
- `create-step-scroll-element`

**Possible Variations:**
- Scroll amount: `--pixels 500`, `--percentage 50`
- Scroll direction: `--horizontal`, `--vertical`
- Smooth scroll: `--smooth`, `--duration 2`
- Scroll into view: `--align-top`, `--align-center`, `--align-bottom`

### 7. Data Operations Variations
**Current Commands:**
- `create-step-store` - Store element value
- `create-step-execute-js` - Execute JavaScript
- `create-step-add-cookie` - Add cookie

**Possible Variations:**
- Store options: `--from-attribute "href"`, `--from-css "color"`
- Transform stored value: `--transform "toLowerCase"`, `--regex-extract`
- JS execution: `--async`, `--return-value`, `--inject-file`
- Cookie options: `--domain`, `--path`, `--expires`, `--secure`

## Merge Opportunities

### 1. Unified Assert Command
**Proposal:** `create-step-assert`

```bash
# Current separate commands
api-cli create-step-assert-exists CHECKPOINT_ID "element" POSITION
api-cli create-step-assert-equals CHECKPOINT_ID "element" "value" POSITION
api-cli create-step-assert-checked CHECKPOINT_ID "element" POSITION

# Merged command with condition argument
api-cli create-step-assert CHECKPOINT_ID POSITION \
  --element "element" \
  --condition "exists|not-exists|equals|contains|checked|..." \
  --value "expected value" \
  --timeout 10
```

### 2. Unified Click Command
**Proposal:** `create-step-click`

```bash
# Current separate commands
api-cli create-step-click CHECKPOINT_ID "element" POSITION
api-cli create-step-double-click CHECKPOINT_ID "element" POSITION
api-cli create-step-right-click CHECKPOINT_ID "element" POSITION

# Merged command with click type
api-cli create-step-click CHECKPOINT_ID "element" POSITION \
  --type "single|double|right|middle" \
  --modifiers "ctrl,shift" \
  --offset "10,20"
```

### 3. Unified Wait Command
**Proposal:** `create-step-wait`

```bash
# Current separate commands
api-cli create-step-wait-time CHECKPOINT_ID 5 POSITION
api-cli create-step-wait-element CHECKPOINT_ID "element" POSITION

# Merged command with wait type
api-cli create-step-wait CHECKPOINT_ID POSITION \
  --for "time|element|text|url" \
  --value "5|element-selector|expected-text|url-pattern" \
  --condition "appears|disappears|contains|matches" \
  --timeout 30
```

### 4. Unified Scroll Command
**Proposal:** `create-step-scroll`

```bash
# Current separate commands
api-cli create-step-scroll-top CHECKPOINT_ID POSITION
api-cli create-step-scroll-bottom CHECKPOINT_ID POSITION
api-cli create-step-scroll-element CHECKPOINT_ID "element" POSITION

# Merged command with scroll target
api-cli create-step-scroll CHECKPOINT_ID POSITION \
  --to "top|bottom|element|position" \
  --target "element-selector|500px|50%" \
  --smooth --duration 2
```

### 5. Unified Input Command
**Proposal:** `create-step-input`

```bash
# Current separate commands
api-cli create-step-write CHECKPOINT_ID "text" "element" POSITION
api-cli create-step-pick CHECKPOINT_ID "value" "element" POSITION
api-cli create-step-key CHECKPOINT_ID "Enter" POSITION

# Merged command with input type
api-cli create-step-input CHECKPOINT_ID POSITION \
  --type "text|select|key|file" \
  --element "element-selector" \
  --value "input-value" \
  --options "clear-first,type-slowly"
```

### 6. Universal Step Command
**Proposal:** `create-step`

```bash
# Ultimate merged command
api-cli create-step CHECKPOINT_ID POSITION \
  --action "navigate|click|wait|assert|scroll|input|store|execute" \
  --target "element-or-url" \
  --value "value-if-needed" \
  --options "key1=value1,key2=value2" \
  --meta "additional-metadata"
```

## Backward Compatibility Strategy

To maintain backward compatibility while introducing merged commands:

1. **Alias System**: Keep existing commands as aliases to new merged commands
2. **Adapter Layer**: Convert old command syntax to new internally
3. **Deprecation Warnings**: Gradually phase out old commands with warnings
4. **Migration Tool**: Provide script to update existing test files

## Implementation Benefits

### Reduced Code Duplication
- Single implementation for similar operations
- Shared validation logic
- Consistent error handling

### Enhanced Flexibility
- Easy to add new variations without new commands
- Complex scenarios handled with options
- Better composability

### Improved User Experience
- Fewer commands to remember
- Consistent syntax patterns
- Better documentation structure

### Easier Maintenance
- Single point of update for each action type
- Standardized testing approach
- Cleaner codebase

## Recommended Implementation Order

1. **Phase 1**: Implement merged commands alongside existing ones
   - Start with `create-step-assert` (most variations)
   - Add `create-step-wait` (clear distinction)
   - Implement `create-step-click` (related actions)

2. **Phase 2**: Add advanced options to merged commands
   - Implement all identified variations
   - Add comprehensive validation
   - Create migration documentation

3. **Phase 3**: Deprecation and migration
   - Add deprecation warnings to old commands
   - Provide migration tools
   - Update all documentation

4. **Phase 4**: Universal command
   - Implement `create-step` as ultimate merger
   - Support all action types
   - Maintain specialized commands for common cases

## Testing Considerations

### Validation Requirements
- Mutually exclusive options
- Required parameters per action type
- Value format validation
- Compatibility checks

### Test Coverage
- All variation combinations
- Error scenarios
- Migration paths
- Performance impact

## Conclusion

The current implementation has significant opportunities for consolidation. By merging related commands and adding flexible options, we can:
- Reduce the command surface area from 40+ to ~6-8 commands
- Increase flexibility and power
- Maintain backward compatibility
- Improve maintainability

The proposed merger maintains the simplicity of common cases while enabling complex scenarios through options.
