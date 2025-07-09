# New Virtuoso CLI Commands Reference

## Overview
This document provides a quick reference for the newly implemented Virtuoso CLI commands that handle journey updates, step management, and structure creation with auto-creation awareness.

## Commands

### 1. update-journey
Update the name of an existing journey (testsuite).

```bash
# Basic usage
./bin/api-cli update-journey <journey-id> --name "New Journey Name"

# Example
./bin/api-cli update-journey 608048 --name "Updated Login Journey"

# With output format
./bin/api-cli update-journey 608048 --name "New Name" -o json
```

**Output**: Shows before/after names in human format

### 2. get-step
Retrieve detailed information about a test step, including the critical canonicalId.

```bash
# Basic usage
./bin/api-cli get-step <step-id>

# Example
./bin/api-cli get-step 19636330

# Get canonical ID programmatically
CANONICAL_ID=$(./bin/api-cli get-step 19636330 -o json | jq -r .canonical_id)
```

**Key fields returned**:
- `id`: Step ID
- `canonical_id`: Required for updates (highlighted in human output)
- `action`: Step type (NAVIGATE, CLICK, etc.)
- `value`: Step value
- `meta`: Additional metadata

### 3. update-navigation
Update the URL of a navigation step (requires both step ID and canonical ID).

```bash
# Basic usage
./bin/api-cli update-navigation <step-id> <canonical-id> --url "https://new-url.com"

# Example
./bin/api-cli update-navigation 19636330 "abc-def-123" --url "https://example.com"

# Open in new tab
./bin/api-cli update-navigation 19636330 "abc-def-123" --url "https://example.com" --new-tab
```

**Important**: 
- First get the canonical ID using `get-step`
- Command will fail if canonical ID doesn't match

### 4. list-checkpoints
List all checkpoints in a journey with their positions and step counts.

```bash
# Basic usage
./bin/api-cli list-checkpoints <journey-id>

# Example
./bin/api-cli list-checkpoints 608048

# Different output formats
./bin/api-cli list-checkpoints 608048 -o json
./bin/api-cli list-checkpoints 608048 -o yaml
./bin/api-cli list-checkpoints 608048 -o ai
```

**Human output example**:
```
Journey: Guest Checkout (ID: 608048)
Checkpoints:
1. Navigate to Site (ID: 1678320) [Navigation] - 1 step
2. Browse Products (ID: 1678321) - 3 steps
3. Add to Cart (ID: 1678322) - 2 steps
```

### 5. create-structure (Enhanced)
Create a complete test structure from YAML/JSON, properly handling Virtuoso's auto-creation behavior.

```bash
# Basic usage
./bin/api-cli create-structure --file structure.yaml

# Dry run (preview only)
./bin/api-cli create-structure --file structure.yaml --dry-run

# Verbose mode
./bin/api-cli create-structure --file structure.yaml --verbose

# Use existing project
./bin/api-cli create-structure --file structure.yaml --project-id 9056

# Combine options
./bin/api-cli create-structure --file structure.yaml --dry-run --verbose
```

**Key behaviors handled**:
- Auto-created journey is renamed (not recreated)
- First checkpoint navigation is updated (not recreated)
- Navigation step is shared across the goal

## YAML Structure Format

```yaml
project:
  name: "Project Name"
  id: 9056  # Optional - use existing project

goals:
  - name: "Goal Name"
    url: "https://example.com"
    journeys:
      - name: "Journey Name"  # First journey renames auto-created
        checkpoints:
          - name: "Navigation Checkpoint"  # First checkpoint updates existing
            navigation_url: "https://example.com/page"  # Updates navigation
            steps:
              - type: wait
                selector: ".element"
                timeout: 5000
          - name: "Action Checkpoint"  # Additional checkpoints created
            steps:
              - type: click
                selector: "#button"
              - type: fill
                selector: "#input"
                value: "test value"
```

## Supported Step Types

1. **navigate**: Navigation step (usually in first checkpoint)
   ```yaml
   type: navigate
   url: "https://example.com"
   ```

2. **click**: Click on an element
   ```yaml
   type: click
   selector: ".button-class"
   ```

3. **wait**: Wait for element to appear
   ```yaml
   type: wait
   selector: "#element-id"
   timeout: 5000  # milliseconds, optional
   ```

4. **fill**: Fill in a form field
   ```yaml
   type: fill
   selector: "input[name='username']"
   value: "testuser@example.com"
   ```

## Command Workflow Examples

### Update Navigation URL Workflow
```bash
# 1. Get step details
./bin/api-cli get-step 19636330

# 2. Extract canonical ID (note the output)
# Canonical ID: abc-def-123 ← Use this for updates

# 3. Update navigation
./bin/api-cli update-navigation 19636330 "abc-def-123" --url "https://new-site.com"
```

### Create Structure Workflow
```bash
# 1. Preview what will be created
./bin/api-cli create-structure --file my-tests.yaml --dry-run

# 2. Create with verbose output
./bin/api-cli create-structure --file my-tests.yaml --verbose

# 3. Check created resources
./bin/api-cli list-projects
./bin/api-cli list-goals <project-id>
./bin/api-cli list-journeys <goal-id> <snapshot-id>
```

## Error Handling

### Common Errors and Solutions

1. **Canonical ID mismatch**
   - Error: "update failed - canonical ID mismatch"
   - Solution: Re-fetch step details with `get-step` to get current canonical ID

2. **Invalid URL format**
   - Error: "URL must include scheme (http/https) and host"
   - Solution: Include full URL with protocol: `https://example.com`

3. **Journey rename fails**
   - Warning: "Failed to rename journey"
   - Note: Non-fatal, process continues with original name

4. **Navigation update fails**
   - Error: "failed to update navigation"
   - Note: Critical error, process stops

## Tips

1. Always use `get-step` before `update-navigation` to get the current canonical ID
2. Use `--dry-run` to preview structure creation before executing
3. Use `--verbose` to debug issues during structure creation
4. The first checkpoint in the first journey is special - it contains the shared navigation
5. When testing, start with small structures before creating large test suites

## Testing the Commands

Run the provided test script:
```bash
./test/test-new-commands.sh
```

Or test individually:
```bash
# Set test IDs (use your actual IDs)
export JOURNEY_ID=608048
export STEP_ID=19636330

# Run tests
./test/test-new-commands.sh
```