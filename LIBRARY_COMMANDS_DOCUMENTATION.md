# Virtuoso API CLI - Library Checkpoint Commands

## Overview

The library commands allow you to manage library checkpoints for reusable test components. Library checkpoints are test sequences that can be reused across multiple journeys.

## Available Commands

### 1. Add Checkpoint to Library

Convert a regular checkpoint into a library checkpoint.

```bash
api-cli library add <checkpoint-id>
```

**Examples:**

```bash
api-cli library add 1680930
api-cli library add cp_1680930
```

**Output:**

- Creates a library checkpoint from the specified checkpoint
- Returns the new library checkpoint ID

### 2. Get Library Checkpoint Details

Retrieve details of a library checkpoint including its steps and metadata.

```bash
api-cli library get <library-checkpoint-id>
```

**Examples:**

```bash
api-cli library get 7023
api-cli library get lib_7023
api-cli library get 7023 --output json
```

**Output:**

- Library checkpoint ID, name, description
- List of all steps in the checkpoint
- Creation and update timestamps

### 3. Attach Library Checkpoint to Journey

Attach a library checkpoint to a journey at a specific position.

```bash
api-cli library attach <journey-id> <library-checkpoint-id> <position>
```

**Examples:**

```bash
api-cli library attach 608926 7023 4
api-cli library attach journey_608926 lib_7023 2
```

**Output:**

- Creates an instance of the library checkpoint in the journey
- Returns the checkpoint ID in the journey

### 4. Move Test Step (NEW)

Move a test step to a new position within a library checkpoint.

```bash
api-cli library move-step <library-checkpoint-id> <test-step-id> <position>
```

**Examples:**

```bash
api-cli library move-step 7023 19660498 2
api-cli library move-step lib_7023 step_19660498 1
```

**Parameters:**

- `library-checkpoint-id`: The ID of the library checkpoint
- `test-step-id`: The ID of the step to move
- `position`: The new position (1-based, where 1 is first)

**Output:**

- Confirmation message with the new position
- Returns 204 No Content on success

### 5. Remove Test Step (NEW)

Remove a test step from a library checkpoint.

```bash
api-cli library remove-step <library-checkpoint-id> <test-step-id>
```

**Examples:**

```bash
api-cli library remove-step 7023 19660498
api-cli library remove-step lib_7023 step_19660498
```

**Parameters:**

- `library-checkpoint-id`: The ID of the library checkpoint
- `test-step-id`: The ID of the step to remove

**Output:**

- Confirmation message
- Returns 204 No Content on success

**Warning:** This permanently removes the step from the library checkpoint.

### 6. Update Library Checkpoint Title (NEW)

Update the title (name) of a library checkpoint.

```bash
api-cli library update <library-checkpoint-id> <new-title>
```

**Examples:**

```bash
api-cli library update 7023 "New Checkpoint Title"
api-cli library update lib_7023 "Updated Test Flow"
api-cli library update 7023 "Login Flow v2" --output json
```

**Parameters:**

- `library-checkpoint-id`: The ID of the library checkpoint
- `new-title`: The new title for the checkpoint

**Output:**

- Updated library checkpoint details
- Confirmation message

## Output Formats

All library commands support multiple output formats:

```bash
--output human  # Default, human-readable format
--output json   # JSON format for scripting
--output yaml   # YAML format
--output ai     # AI-optimized format
```

## Common Use Cases

### 1. Creating Reusable Test Components

```bash
# Create a checkpoint in a journey
api-cli create-checkpoint 608926 13961 43992 "Login Flow"

# Add it to the library for reuse
api-cli library add 1680930

# Use it in other journeys
api-cli library attach 608927 7023 1
```

### 2. Managing Library Checkpoint Steps

```bash
# Get checkpoint details to see step IDs
api-cli library get 7023 --output json

# Reorder steps
api-cli library move-step 7023 19660498 1
api-cli library move-step 7023 19660499 2

# Remove unnecessary steps
api-cli library remove-step 7023 19660500
```

### 3. Updating Library Checkpoints

```bash
# Update the title
api-cli library update 7023 "Login Flow - Updated"

# Get updated details
api-cli library get 7023
```

## Error Handling

Common errors and their meanings:

- **404 Not Found**: Library checkpoint or step ID doesn't exist
- **400 Bad Request**: Invalid parameters (e.g., position < 1)
- **401 Unauthorized**: Invalid or missing API token
- **500 Internal Server Error**: Server-side issue

## Configuration

The library commands use the standard Virtuoso API configuration:

```yaml
api:
  auth_token: your-api-key-here
  base_url: https://api-app2.virtuoso.qa/api
organization:
  id: "2242"
```

## Notes

- Library checkpoints are organization-wide and can be shared across projects
- Position values are 1-based (1 is the first position)
- The `remove-step` command is permanent and cannot be undone
- Library checkpoint IDs can be prefixed with `lib_` for clarity
- Test step IDs can be prefixed with `step_` for clarity
- Journey IDs can be prefixed with `journey_` for clarity

## Implementation Details

These commands follow the project's standards:

- Use BaseCommand for consistent initialization
- Support all output formats (human, json, yaml, ai)
- Proper error handling with descriptive messages
- Parameter validation (e.g., position >= 1)
- Strip optional prefixes from IDs
- Use the configured base URL and authentication
