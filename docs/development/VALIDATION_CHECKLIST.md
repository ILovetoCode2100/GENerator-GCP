# Step Command Consolidation - Validation Checklist

## Summary of Findings

I've analyzed the API CLI Generator project and identified opportunities to consolidate 40+ individual step commands into 6-8 flexible, parameterized commands while maintaining backward compatibility.

## Step Variations to Validate

### ✅ Navigation & Browser Control
Please confirm these variations make sense for your use cases:

- [ ] Navigate with `--new-tab` option
- [ ] Navigate with `--new-window` option  
- [ ] Navigate with `--wait-for-element "selector"` (navigate then wait)
- [ ] Navigate with `--expect-url "pattern"` (navigate then assert)
- [ ] Browser history navigation: `--action "back|forward|refresh"`
- [ ] Window control with `--maximize`, `--minimize`, `--fullscreen`

### ✅ Wait Conditions
Current implementation only has time and element waits. Validate these additions:

- [ ] Wait `--until-visible` (element exists AND is visible)
- [ ] Wait `--until-hidden` (element becomes hidden/removed)
- [ ] Wait `--until-clickable` (element is visible AND enabled)
- [ ] Wait `--contains-text "text"` (element contains specific text)
- [ ] Wait `--url-contains "pattern"` (URL changes to match)
- [ ] Wait `--url-changes` (any URL change from current)
- [ ] Custom timeout with `--timeout 60` (override default)

### ✅ Click Modifiers
Currently only basic clicks exist. Validate these enhancements:

- [ ] Click with keyboard modifiers: `--modifiers "ctrl,shift"`
- [ ] Click at offset: `--offset-x 10 --offset-y 20`
- [ ] Click and hold: `--hold-duration 2`
- [ ] Multiple clicks: `--count 3` (for triple-click selection)
- [ ] Force click: `--force` (bypass visibility checks)
- [ ] Click and wait: `--wait-after 2` (click then wait X seconds)

### ✅ Text Input Options
Current write command is basic. Validate these options:

- [ ] Clear before typing: `--clear-first`
- [ ] Append to existing: `--append`
- [ ] Type slowly: `--type-delay 100` (ms between keystrokes)
- [ ] Type with key events: `--use-key-events` (vs paste)
- [ ] Input validation: `--expect-max-length 50`
- [ ] Masked input: `--mask` (for passwords, no logging)

### ✅ Assertion Conditions
Many assertion types are missing. Validate this comprehensive list:

**Text Assertions:**
- [ ] `--condition "contains"` with partial text match
- [ ] `--condition "starts-with"` for text beginning
- [ ] `--condition "ends-with"` for text ending  
- [ ] `--condition "matches-regex"` with regex pattern
- [ ] `--condition "has-length"` for text length

**Numeric Assertions:**
- [ ] `--condition "greater-than"` with numeric value
- [ ] `--condition "less-than"` with numeric value
- [ ] `--condition "between"` with `--min X --max Y`
- [ ] `--condition "equals-number"` for numeric equality

**Element State Assertions:**
- [ ] `--condition "visible"` (exists and visible)
- [ ] `--condition "hidden"` (exists but not visible)
- [ ] `--condition "enabled"` (not disabled)
- [ ] `--condition "disabled"` (is disabled)
- [ ] `--condition "focused"` (has focus)
- [ ] `--condition "selected"` (option is selected)

**Attribute Assertions:**
- [ ] `--condition "has-class"` with `--class "active"`
- [ ] `--condition "has-attribute"` with `--attribute "data-id"`
- [ ] `--condition "attribute-equals"` with `--attribute "href" --value "/home"`
- [ ] `--condition "css-property"` with `--property "color" --value "red"`

**Count Assertions:**
- [ ] `--condition "count-equals"` for element count
- [ ] `--condition "count-greater-than"` with minimum
- [ ] `--condition "count-less-than"` with maximum

### ✅ Scroll Options
Basic scroll commands exist. Validate these enhancements:

- [ ] Scroll by pixels: `--pixels 500`
- [ ] Scroll by percentage: `--percentage 50`
- [ ] Horizontal scroll: `--direction "horizontal"`
- [ ] Smooth scrolling: `--smooth --duration 2`
- [ ] Scroll alignment: `--align "top|center|bottom"`
- [ ] Scroll and wait: `--wait-after 1`

### ✅ Data Operations
Limited data operations currently. Validate these additions:

- [ ] Store from attribute: `--from-attribute "href"`
- [ ] Store from CSS property: `--from-css "background-color"`
- [ ] Transform stored value: `--transform "toLowerCase|toUpperCase|trim"`
- [ ] Extract with regex: `--extract-pattern "[0-9]+"`
- [ ] Store multiple values: `--store-all` (for multiple elements)

## Merge Opportunities to Validate

### 1. ✅ Unified Assert Command
**Current State:** 8+ separate assert commands  
**Proposed:** Single `create-step-assert` with `--condition` parameter

**Please validate this makes sense:**
```bash
# Instead of:
api-cli create-step-assert-exists CHECKPOINT "element" POSITION
api-cli create-step-assert-equals CHECKPOINT "element" "value" POSITION
api-cli create-step-assert-checked CHECKPOINT "element" POSITION

# Use:
api-cli create-step-assert CHECKPOINT POSITION \
  --element "element" \
  --condition "exists|equals|checked|..." \
  --value "expected"
```

### 2. ✅ Unified Click Command  
**Current State:** click, double-click, right-click, hover (4 commands)  
**Proposed:** Single `create-step-click` with `--type` parameter

**Please validate:**
```bash
# Instead of separate commands, use:
api-cli create-step-click CHECKPOINT POSITION \
  --element "button" \
  --type "single|double|right|hover" \
  --modifiers "ctrl"
```

### 3. ✅ Unified Wait Command
**Current State:** wait-time, wait-element (2 commands)  
**Proposed:** Single `create-step-wait` with `--for` parameter

**Please validate:**
```bash
# Flexible wait command:
api-cli create-step-wait CHECKPOINT POSITION \
  --for "time|element|text|url" \
  --value "5|.spinner|Loading...|/success" \
  --condition "appears|disappears|contains"
```

### 4. ✅ Unified Scroll Command
**Current State:** scroll-top, scroll-bottom, scroll-element (3 commands)  
**Proposed:** Single `create-step-scroll` with `--to` parameter

**Please validate:**
```bash
# Flexible scroll:
api-cli create-step-scroll CHECKPOINT POSITION \
  --to "top|bottom|element|position" \
  --target "#footer|500px|75%"
```

### 5. ✅ Universal Step Command
**Most flexible option:** Single `create-step` for everything

**Please validate if this is too complex or useful:**
```bash
api-cli create-step CHECKPOINT POSITION \
  --action "click|wait|assert|navigate|..." \
  --target "element-or-url" \
  --value "input-value" \
  --options "key1=val1,key2=val2"
```

## Questions for Validation

1. **Backward Compatibility**: Should we keep ALL old commands as aliases, or is it OK to deprecate some?

2. **Naming Preferences**: Do you prefer:
   - `create-step-assert` or `assert-step`?
   - `--condition` or `--type` for assertion types?
   - `--element` or `--target` for selectors?

3. **Default Behaviors**: What should be the defaults for:
   - Wait timeout (currently 20 seconds)?
   - Scroll behavior (smooth vs instant)?
   - Click type (single click)?

4. **Priority Features**: Which variations are MUST-HAVE vs NICE-TO-HAVE?

5. **Complex Scenarios**: Do you need support for:
   - Chained actions (click then wait automatically)?
   - Conditional steps (only execute if condition met)?
   - Retry logic (retry failed steps X times)?

6. **Output Format**: Is the current 4-format output (human, json, yaml, ai) sufficient?

## Next Steps

Please review and mark which variations and merges you want to proceed with. I can then create:

1. Detailed implementation plan for approved changes
2. Migration guide for existing scripts
3. Test suite for new functionality
4. Documentation updates

The goal is to reduce complexity while increasing power and flexibility. Let me know what makes sense for your use cases!
