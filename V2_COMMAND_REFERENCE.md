# Virtuoso API CLI v2 Command Reference

## Overview

This document provides a complete reference for all v2 commands in the Virtuoso API CLI. The v2 syntax standardizes all commands to use consistent positional arguments, making the CLI more intuitive and easier to use.

## v2 Syntax Pattern

**Standard Pattern:** `api-cli <command> <subcommand> [checkpoint-id] <args...> [position]`

- `<command>` - The command group (e.g., assert, wait, interact)
- `<subcommand>` - The specific operation (e.g., exists, click, element)
- `[checkpoint-id]` - Optional when using session context
- `<args...>` - Command-specific arguments
- `[position]` - Optional step position, auto-increments in session

## Session Context

Set the checkpoint ID once and all commands will use it:

```bash
export VIRTUOSO_SESSION_ID=cp_12345

# Now commands automatically use this checkpoint
api-cli navigate to "https://example.com"
api-cli interact click "button"
```

## Command Groups

### 1. Assert Commands (v2)

Assert commands validate elements and conditions on the page.

#### assert exists

```bash
# With session context
api-cli assert exists "Login button"

# With explicit checkpoint
api-cli assert exists cp_12345 "Login button" 1

# Examples
api-cli assert exists "button.submit"
api-cli assert exists "div#success-message"
api-cli assert exists "Welcome to our site"
```

#### assert not-exists

```bash
api-cli assert not-exists "Error message"
api-cli assert not-exists cp_12345 "div.error" 2
```

#### assert equals

```bash
api-cli assert equals "h1" "Welcome"
api-cli assert equals cp_12345 "span.username" "John Doe" 3
```

#### assert not-equals

```bash
api-cli assert not-equals "div.status" "Error"
api-cli assert not-equals cp_12345 "input#email" "" 4
```

#### assert checked

```bash
api-cli assert checked "input#terms"
api-cli assert checked cp_12345 "input[type='checkbox']" 5
```

#### assert selected

```bash
api-cli assert selected "select#country" "USA"
api-cli assert selected cp_12345 "select.dropdown" "Option 1" 6
```

#### assert variable

```bash
api-cli assert variable "username" "testuser"
api-cli assert variable cp_12345 "total" "100.00" 7
```

#### assert gt/gte/lt/lte

```bash
api-cli assert gt "span.price" "50"
api-cli assert gte cp_12345 "div.count" "10" 8
api-cli assert lt "input.quantity" "100"
api-cli assert lte cp_12345 "span.remaining" "5" 9
```

#### assert matches

```bash
api-cli assert matches "div.email" "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$"
api-cli assert matches cp_12345 "span.phone" "^\d{3}-\d{3}-\d{4}$" 10
```

### 2. Wait Commands (v2)

Wait commands pause execution until conditions are met.

#### wait element

```bash
# Basic wait
api-cli wait element "div.loaded"

# With timeout (milliseconds)
api-cli wait element "Success message" --timeout 5000

# With explicit checkpoint
api-cli wait element cp_12345 "button.continue" 1
api-cli wait element cp_12345 "#spinner" 2 --timeout 10000
```

#### wait element-not-visible

```bash
api-cli wait element-not-visible "div.loading"
api-cli wait element-not-visible cp_12345 ".spinner" 3 --timeout 3000
```

#### wait time

```bash
# Wait in milliseconds
api-cli wait time 1000  # 1 second
api-cli wait time cp_12345 2500 4  # 2.5 seconds
```

### 3. Mouse Commands (v2)

Mouse commands control advanced mouse operations.

#### mouse move-to

```bash
api-cli mouse move-to "button.hover"
api-cli mouse move-to cp_12345 "#menu-item" 1
```

#### mouse move-by

```bash
api-cli mouse move-by "100,50"  # Move 100px right, 50px down
api-cli mouse move-by cp_12345 "-50,0" 2  # Move 50px left
```

#### mouse move

```bash
api-cli mouse move "500,300"  # Move to absolute position
api-cli mouse move cp_12345 "0,0" 3  # Move to top-left
```

#### mouse down/up

```bash
api-cli mouse down
api-cli mouse up
api-cli mouse down cp_12345 1
api-cli mouse up cp_12345 2
```

#### mouse enter

```bash
api-cli mouse enter
api-cli mouse enter cp_12345 1
```

### 4. Data Commands (v2)

Data commands store values and manage cookies.

#### data store element-text

```bash
api-cli data store element-text "h1" "pageTitle"
api-cli data store element-text cp_12345 "span.price" "productPrice" 1
```

#### data store element-value

```bash
api-cli data store element-value "input#username" "savedUsername"
api-cli data store element-value cp_12345 "select#country" "selectedCountry" 2
```

#### data store attribute

```bash
api-cli data store attribute "a.link" "href" "linkUrl"
api-cli data store attribute cp_12345 "img.logo" "src" "logoSource" 3
```

#### data cookie create

```bash
# Basic cookie
api-cli data cookie create "session" "abc123"

# With options
api-cli data cookie create "auth" "token123" --domain ".example.com" --path "/" --secure --httpOnly

# With checkpoint
api-cli data cookie create cp_12345 "user" "john" 4
```

#### data cookie delete

```bash
api-cli data cookie delete "session"
api-cli data cookie delete cp_12345 "tracking" 5
```

#### data cookie clear

```bash
api-cli data cookie clear
api-cli data cookie clear cp_12345 6
```

### 5. Window Commands (v2)

Window commands manage browser windows and frames.

#### window resize

```bash
api-cli window resize 1024x768
api-cli window resize cp_12345 1920x1080 1
```

#### window maximize

```bash
api-cli window maximize
api-cli window maximize cp_12345 2
```

#### window switch tab

```bash
# Switch by direction
api-cli window switch tab NEXT
api-cli window switch tab PREVIOUS

# Switch by index
api-cli window switch tab INDEX 0
api-cli window switch tab cp_12345 INDEX 2 3
```

#### window switch iframe

```bash
api-cli window switch iframe "#payment-frame"
api-cli window switch iframe cp_12345 "iframe.embedded" 4
```

#### window switch parent-frame

```bash
api-cli window switch parent-frame
api-cli window switch parent-frame cp_12345 5
```

### 6. Dialog Commands (v2)

Dialog commands handle browser dialogs.

#### dialog alert accept/dismiss

```bash
api-cli dialog alert accept
api-cli dialog alert dismiss
api-cli dialog alert accept cp_12345 1
```

#### dialog confirm accept/dismiss

```bash
api-cli dialog confirm accept
api-cli dialog confirm dismiss cp_12345 2
```

#### dialog prompt

```bash
# Accept with text
api-cli dialog prompt "My answer"
api-cli dialog prompt cp_12345 "User input" 3

# Dismiss prompt
api-cli dialog prompt dismiss
api-cli dialog prompt dismiss cp_12345 4
```

### 7. Select Commands (v2)

Select commands handle dropdown menus.

#### select option

```bash
api-cli select option "select#country" "United States"
api-cli select option cp_12345 ".dropdown" "Option 2" 1
```

#### select index

```bash
api-cli select index "select#country" 0  # First option
api-cli select index cp_12345 "#category" 3 2  # Fourth option
```

#### select last

```bash
api-cli select last "select#country"
api-cli select last cp_12345 ".dropdown" 3
```

### 8. Commands Already Using v2-Compatible Syntax

These commands already follow the v2 positional pattern:

#### Navigate Commands

```bash
api-cli navigate to "https://example.com"
api-cli navigate scroll-by "0,500"
api-cli navigate scroll-up
api-cli navigate scroll-down
api-cli navigate scroll-top
api-cli navigate scroll-bottom
api-cli navigate scroll-element "div.section"
api-cli navigate scroll-position "div.content" "0,100"
```

#### Interact Commands

```bash
api-cli interact click "button"
api-cli interact click "button" --position TOP_LEFT
api-cli interact double-click "div.item"
api-cli interact right-click "td.cell"
api-cli interact hover "nav.menu"
api-cli interact write "input" "text to type"
api-cli interact key "Tab"
api-cli interact key "a" --modifiers ctrl
```

#### File Commands

```bash
api-cli file upload "input[type=file]" "https://example.com/file.pdf"
api-cli file upload-url "input.uploader" "https://example.com/doc.docx"
```

#### Misc Commands

```bash
api-cli misc comment "This is a test comment"
api-cli misc execute "return document.title"
```

## Flags and Options

### Global Flags

- `--output, -o` - Output format: human, json, yaml, ai (default: human)
- `--help, -h` - Show help for command

### Command-Specific Flags

#### Wait Commands

- `--timeout` - Timeout in milliseconds (default varies by command)

#### Interact Commands

- `--position` - Click position: TOP_LEFT, TOP_CENTER, TOP_RIGHT, CENTER_LEFT, CENTER, CENTER_RIGHT, BOTTOM_LEFT, BOTTOM_CENTER, BOTTOM_RIGHT
- `--modifiers` - Key modifiers: ctrl, shift, alt, meta (can combine with commas)

#### Data Cookie Commands

- `--domain` - Cookie domain
- `--path` - Cookie path
- `--secure` - Secure cookie flag
- `--httpOnly` - HTTP only cookie flag

#### Navigate Scroll Commands

- `--x` - X offset for scroll-by
- `--y` - Y offset for scroll-by

## Output Formats

### Human Format (Default)

```
✓ Step created successfully
  Checkpoint: cp_12345
  Position: 1
  Type: CLICK
  Details: Click on "button"
```

### JSON Format

```json
{
  "success": true,
  "checkpoint_id": "cp_12345",
  "position": 1,
  "step_type": "CLICK",
  "details": {
    "selector": "button"
  }
}
```

### YAML Format

```yaml
success: true
checkpoint_id: cp_12345
position: 1
step_type: CLICK
details:
  selector: button
```

### AI Format

```
Step created: Click on "button"
Context: This step will click on the element matching "button"
Test Structure: checkpoint cp_12345 → step 1
Next Steps: Consider adding assertions to verify the click result
```

## Error Handling

Commands provide clear error messages:

```bash
# Missing required arguments
$ api-cli interact click
Error: insufficient arguments: expected selector

# Invalid checkpoint
$ api-cli navigate to invalid_checkpoint "https://example.com"
Error: checkpoint not found: invalid_checkpoint

# No checkpoint or session
$ api-cli assert exists "button"
Error: checkpoint ID required (use --checkpoint flag or set VIRTUOSO_SESSION_ID)
```

## Best Practices

1. **Use Session Context for Scripts:**

   ```bash
   #!/bin/bash
   export VIRTUOSO_SESSION_ID=cp_12345

   api-cli navigate to "https://app.example.com"
   api-cli interact click "button#login"
   api-cli wait element "div.dashboard"
   ```

2. **Explicit Checkpoint for One-off Commands:**

   ```bash
   api-cli assert exists cp_12345 "Login button" 1
   ```

3. **Use Meaningful Variable Names:**

   ```bash
   api-cli data store element-text "h1.title" "pageTitle"
   api-cli data store element-text "span.price" "productPrice"
   ```

4. **Add Wait Commands for Dynamic Content:**

   ```bash
   api-cli interact click "button.load"
   api-cli wait element "div.results" --timeout 5000
   api-cli assert exists "Results loaded"
   ```

5. **Use Comments for Documentation:**
   ```bash
   api-cli misc comment "Testing login flow"
   api-cli interact write "input#username" "testuser"
   api-cli misc comment "Credentials entered, submitting form"
   ```

## Migration from Legacy Syntax

If you're using the old syntax, here's how to migrate:

### Old Flag-Based Syntax

```bash
# Old
api-cli assert exists "button" --checkpoint cp_12345 --position 1
api-cli wait element "div" --checkpoint cp_12345 --timeout 5000

# New v2
api-cli assert exists cp_12345 "button" 1
api-cli wait element cp_12345 "div" 2 --timeout 5000
```

### Using Session Context

```bash
# Old (always need --checkpoint)
api-cli assert exists "button" --checkpoint cp_12345
api-cli wait element "div" --checkpoint cp_12345

# New v2 (with session)
export VIRTUOSO_SESSION_ID=cp_12345
api-cli assert exists "button"
api-cli wait element "div"
```

## Complete Example Script

```bash
#!/bin/bash
# Example: Complete user registration flow

# Setup
PROJECT_ID=$(api-cli create-project "User Registration Test" -o json | jq -r '.project_id')
GOAL_JSON=$(api-cli create-goal $PROJECT_ID "Test Registration" -o json)
GOAL_ID=$(echo "$GOAL_JSON" | jq -r '.goal_id')
SNAPSHOT_ID=$(echo "$GOAL_JSON" | jq -r '.snapshot_id')
JOURNEY_ID=$(api-cli create-journey $GOAL_ID $SNAPSHOT_ID "Registration Flow" -o json | jq -r '.journey_id')
CHECKPOINT_ID=$(api-cli create-checkpoint $JOURNEY_ID $GOAL_ID $SNAPSHOT_ID "Register New User" -o json | jq -r '.checkpoint_id')

# Set session context
export VIRTUOSO_SESSION_ID=$CHECKPOINT_ID

# Test steps using v2 syntax
api-cli navigate to "https://example.com/register"
api-cli wait element "form#registration"
api-cli misc comment "Starting registration process"

# Fill form
api-cli interact write "input#firstName" "John"
api-cli interact write "input#lastName" "Doe"
api-cli interact write "input#email" "john.doe@example.com"
api-cli interact write "input#password" "SecurePass123!"
api-cli interact write "input#confirmPassword" "SecurePass123!"

# Accept terms
api-cli interact click "input#terms"
api-cli assert checked "input#terms"

# Submit
api-cli interact click "button#submit"
api-cli wait element "div.success-message" --timeout 5000
api-cli assert exists "Registration successful"

# Store confirmation number
api-cli data store element-text "span.confirmation-number" "confirmationNumber"

echo "Registration test completed successfully!"
```
