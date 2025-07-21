# Data Command Group - Detailed Analysis

## Overview

The data command group in the Virtuoso API CLI handles two main categories of operations:

1. **Store Operations** - Storing values in variables for use in subsequent test steps
2. **Cookie Operations** - Managing browser cookies during test execution

## Command Structure

### Store Commands (3 types)

#### 1. `data store element-text`

- **Purpose**: Store element text content in a variable
- **Syntax**: `data store element-text SELECTOR VARIABLE_NAME [POSITION]`
- **Step Type**: STORE
- **API Method**: `CreateStepStoreElementText(checkpointID, selector, variableName, position)`
- **Parsed Step**: `store text from "SELECTOR" in $VARIABLE_NAME`
- **Parameter Order**:
  1. selector (string) - Element selector
  2. variable_name (string) - Variable name (letters, numbers, underscores only)
  3. position (optional int) - Step position
- **Example**: `api-cli data store element-text "div.username" "current_user" 5`

#### 2. `data store literal`

- **Purpose**: Store a literal value in a variable
- **Syntax**: `data store literal VALUE VARIABLE_NAME [POSITION]`
- **Step Type**: STORE
- **API Method**: `CreateStepStoreLiteralValue(checkpointID, value, variableName, position)`
- **Parsed Step**: `store "VALUE" in $VARIABLE_NAME`
- **Parameter Order**:
  1. value (string) - Literal value to store
  2. variable_name (string) - Variable name (letters, numbers, underscores only)
  3. position (optional int) - Step position
- **Example**: `api-cli data store literal "2024-01-01" "test_date" 3`

#### 3. `data store attribute` (Stage 3 Enhancement)

- **Purpose**: Store element attribute value in a variable
- **Syntax**: `data store attribute SELECTOR ATTRIBUTE_NAME VARIABLE_NAME [POSITION]`
- **Step Type**: STORE
- **API Method**: `CreateStepStoreAttribute(checkpointID, selector, attribute, variable, position)`
- **Parsed Step**: `store attribute "ATTRIBUTE_NAME" from "SELECTOR" in $VARIABLE_NAME`
- **Parameter Order**:
  1. selector (string) - Element selector
  2. attribute_name (string) - Attribute name (e.g., "href", "src", "value")
  3. variable_name (string) - Variable name (letters, numbers, underscores only)
  4. position (optional int) - Step position
- **Example**: `api-cli data store attribute "img.logo" "src" "logo_url" 2`

### Cookie Commands (3 types)

#### 1. `data cookie create`

- **Purpose**: Create a cookie with specified name and value
- **Syntax**: `data cookie create NAME VALUE [POSITION]`
- **Step Type**: ENVIRONMENT
- **API Methods**:
  - Basic: `CreateStepCookieCreate(checkpointID, name, value, position)`
  - With options: `CreateStepCookieCreateWithOptions(checkpointID, name, value, options, position)`
- **Cookie Options** (all optional):
  - `--domain` (string) - Cookie domain (e.g., ".example.com")
  - `--path` (string) - Cookie path (default: "/")
  - `--secure` (bool) - Set secure flag on cookie
  - `--http-only` (bool) - Set httpOnly flag on cookie
- **Parsed Step Examples**:
  - Basic: `create cookie "session" with value "abc123"`
  - With domain: `create cookie "session" with value "abc123" for domain .example.com`
  - Full options: `create cookie "auth" with value "token" for domain .site.com with path /api secure http-only`
- **Parameter Order**:
  1. name (string) - Cookie name
  2. value (string) - Cookie value
  3. position (optional int) - Step position
- **Examples**:
  - Basic: `api-cli data cookie create "user_id" "12345" 1`
  - With options: `api-cli data cookie create "auth" "xyz789" --domain ".api.com" --secure --http-only`

#### 2. `data cookie delete`

- **Purpose**: Delete a specific cookie by name
- **Syntax**: `data cookie delete NAME [POSITION]`
- **Step Type**: ENVIRONMENT
- **API Method**: `CreateStepDeleteCookie(checkpointID, name, position)`
- **Parsed Step**: `delete cookie "NAME"`
- **Parameter Order**:
  1. name (string) - Cookie name to delete
  2. position (optional int) - Step position
- **Example**: `api-cli data cookie delete "temp_session" 4`

#### 3. `data cookie clear-all`

- **Purpose**: Clear all browser cookies
- **Syntax**: `data cookie clear-all [POSITION]`
- **Step Type**: ENVIRONMENT
- **API Method**: `CreateStepCookieWipeAll(checkpointID, position)`
- **Parsed Step**: `clear all cookies`
- **Parameter Order**:
  1. position (optional int) - Step position
- **Example**: `api-cli data cookie clear-all 1`

## Common Features

### Position Handling

- All commands support optional position parameter
- If position is omitted and auto-increment is enabled, position is automatically incremented
- Position always comes as the last parameter (after all required arguments)

### Checkpoint Context

- All commands use session context by default
- Can override with `--checkpoint` flag
- Example: `api-cli data store literal "test" "var" --checkpoint 12345`

### Variable Name Validation

- Variable names must contain only:
  - Letters (a-z, A-Z)
  - Numbers (0-9) - but not as first character
  - Underscores (\_)
- Invalid examples: `my-var`, `123var`, `var$name`
- Valid examples: `myVar`, `user_name`, `test123`, `_private`

## Error Handling

### Common Validation Errors

1. Empty required fields (selector, value, variable name, cookie name)
2. Invalid variable names (special characters, starting with number)
3. Missing required arguments
4. Invalid argument count

### API Response Handling

- All commands return step ID on success
- Errors are wrapped with context (e.g., "failed to create STORE step: ...")
- Session context is saved after successful step creation with auto-increment

## Working Status

All 6 data commands are fully working and tested:

- ✅ data store element-text
- ✅ data store literal
- ✅ data store attribute
- ✅ data cookie create (with all options)
- ✅ data cookie delete
- ✅ data cookie clear-all

## Usage Patterns

### Store Operations Pattern

```bash
api-cli data store <type> <args...> [position]
```

Where type is: element-text, literal, or attribute

### Cookie Operations Pattern

```bash
api-cli data cookie <operation> <args...> [position]
```

Where operation is: create, delete, or clear-all

## Implementation Notes

1. **Client Methods**: The data commands use 6 different client methods, with cookie create having two variants based on whether options are provided.

2. **Step Types**: Store operations create "STORE" type steps, while cookie operations create "ENVIRONMENT" type steps.

3. **Flag Handling**: Only cookie create command has additional flags. These are collected into an options map and passed to the API.

4. **Argument Validation**: Each command type has specific validation rules implemented in `validateDataArgs()`.

5. **Output Format**: All commands support the standard output formats (human, json, yaml, ai) and include relevant extra data in the output.

## Testing

- All commands are tested in `test-stage3-features.sh`
- Store attribute and cookie create with options are Stage 3 enhancements
- Total of 5 different step types can be created (3 store + 2 cookie types)
