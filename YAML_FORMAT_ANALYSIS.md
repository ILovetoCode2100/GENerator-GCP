# Virtuoso API CLI YAML Format Analysis

## Executive Summary

After analyzing the Virtuoso API CLI codebase, I've identified three distinct YAML formats being used:

1. **Compact Format** - Expected by the yaml-layer validator
2. **Extended Format** - Used in examples directory
3. **Simplified Format** - Used in test-all-commands directory

The current `yaml` command group only supports the Compact Format. The other formats appear to be custom implementations that bypass the yaml-layer entirely.

## Format Comparison

### 1. Compact Format (yaml-layer)

**File Examples**: registration-demo.yaml, newsletter.yaml

```yaml
test: User Registration Flow
nav: https://example.com/register
data:
  email: test@example.com
  password: SecurePass123!
do:
  - c: "#email" # click
  - t: $email # type
  - wait: 1000 # wait milliseconds
  - ch: "Success" # check/assert
  - store: { .user-id: userId }
```

**Characteristics**:

- Root fields: `test`, `nav`, `data`, `do`, `setup`, `teardown`
- Ultra-compact action syntax (c, t, ch, etc.)
- Variable references with `$` prefix
- Supports advanced features (if/loop/store)

### 2. Extended Format (examples/)

**File Examples**: examples/minimal-test.yaml, examples/simple-login-test.yaml

```yaml
name: "Minimal Test"
description: "Test description"
steps:
  - type: navigate
    target: "https://example.com"
  - type: assert
    command: exists
    target: "body"
  - type: interact
    command: write
    target: "input#email"
    value: "test@example.com"
config:
  continue_on_error: false
  screenshot_on_error: true
```

**Characteristics**:

- Root fields: `name`, `description`, `steps`, `config`
- Verbose step structure with type/command/target
- Each step is an object with explicit fields
- No direct CLI command support

### 3. Simplified Format (test-all-commands/)

**File Examples**: test-all-commands/07-simple-working-test.yaml

```yaml
name: "Simple Working Test"
description: "Test description"
project: "Test Project"
steps:
  - navigate: "https://example.com"
  - wait: 2000
  - click: "button.submit"
  - write:
      selector: "input#email"
      text: "test@example.com"
  - assert: "Success message"
  - comment: "Test completed"
```

**Characteristics**:

- Root fields: `name`, `description`, `project`, `steps`
- Direct command names as keys
- Mixed simple and complex syntax
- More readable than compact format

## Format Conversion Mapping

| Compact Format | Extended Format                                          | Simplified Format | CLI Command           |
| -------------- | -------------------------------------------------------- | ----------------- | --------------------- |
| `nav: /path`   | `type: navigate`<br>`target: /path`                      | `navigate: /path` | `step-navigate to`    |
| `c: button`    | `type: interact`<br>`command: click`<br>`target: button` | `click: button`   | `step-interact click` |
| `t: text`      | `type: interact`<br>`command: write`<br>`value: text`    | `write: text`     | `step-interact write` |
| `ch: text`     | `type: assert`<br>`command: exists`<br>`target: text`    | `assert: text`    | `step-assert exists`  |
| `wait: 1000`   | `type: wait`<br>`command: time`<br>`value: 1000`         | `wait: 1000`      | `step-wait time`      |
| `store: {...}` | `type: data`<br>`command: store`                         | `store: {...}`    | `step-data store`     |

## Testing Results

### Compact Format Validation

```bash
./bin/api-cli yaml validate registration-demo.yaml
```

**Result**: Partial success with errors:

- Recognizes basic structure
- Has issues with advanced features (store syntax)
- Shows validator is working but strict

### Extended Format Validation

```bash
./bin/api-cli yaml validate examples/minimal-test.yaml
```

**Result**: Failed - not recognized format

- Missing required `test:` field
- Missing required `do:` section
- This format is incompatible with yaml-layer

### YAML Command Capabilities

The `yaml` command provides:

- `validate` - Validates compact format only
- `compile` - Converts to CLI commands
- `run` - Executes tests
- `generate` - Creates tests from prompts
- `convert` - Not implemented (needs ListSteps API)

## Command Support by Format

### Compact Format

- **Supported by**: `yaml validate`, `yaml compile`, `yaml run`, `yaml generate`
- **Validation**: Full validation with detailed error messages
- **Compilation**: Converts to CLI commands (e.g., `step-navigate to`, `step-interact click`)
- **Execution**: Can be executed through `yaml run`

### Simplified Format

- **Supported by**: `run-test` command
- **Validation**: Basic structure validation only
- **Compilation**: Direct step execution without conversion
- **Execution**: Creates full test infrastructure (project/goal/journey/checkpoint)

### Extended Format

- **Supported by**: None (appears to be documentation/example only)
- **Validation**: Not supported
- **Compilation**: Not supported
- **Execution**: Not supported

## Testing Results Summary

1. **Compact Format with `yaml compile`**:

   ```bash
   ./bin/api-cli yaml compile newsletter.yaml -o commands
   ```

   Successfully converts to CLI commands:

   - `step-navigate to https://api-app2.virtuoso.qa/api/`
   - `step-wait element body`
   - `step-interact click Start`
   - etc.

2. **Simplified Format with `run-test`**:

   ```bash
   ./bin/api-cli run-test test-all-commands/07-simple-working-test.yaml --dry-run
   ```

   Successfully parses and prepares for execution with full test infrastructure creation.

3. **Extended Format**: No command support found

## Recommendations

### 1. **Immediate Usage Guidelines**

For users of the Virtuoso API CLI:

- **For YAML validation and compilation**: Use the **Compact Format** with `yaml` commands
- **For quick test creation and execution**: Use the **Simplified Format** with `run-test` command
- **Avoid**: The Extended Format as it has no CLI support

### 2. **Format Selection Guide**

| Use Case                    | Recommended Format | Command to Use       |
| --------------------------- | ------------------ | -------------------- |
| AI-friendly test generation | Compact            | `yaml generate`      |
| Token-optimized tests       | Compact            | `yaml validate/run`  |
| Quick test prototyping      | Simplified         | `run-test`           |
| Full test lifecycle         | Simplified         | `run-test --execute` |
| Test validation             | Compact            | `yaml validate`      |
| Convert to CLI commands     | Compact            | `yaml compile`       |

### 3. **Short-term Actions**

1. Update documentation to clarify the two supported formats
2. Add format detection to provide better error messages
3. Consider adding a format converter between Compact and Simplified

### 4. **Long-term Vision**

Consider unifying around the Simplified format with enhancements:

- Add the compact syntax support from the Compact format
- Integrate with the yaml-layer validation
- Support both `yaml` and `run-test` commands
- Maintain backward compatibility

## Conclusion

The Virtuoso API CLI currently has three incompatible YAML formats, with only the compact format having full CLI support through the `yaml` command group. The other formats appear to be used for different purposes (examples, testing) but lack integration with the yaml validation and execution layer.

For users wanting to use YAML files with the CLI, they must use the compact format. The other formats would require custom processing or significant updates to the yaml-layer to support them.
