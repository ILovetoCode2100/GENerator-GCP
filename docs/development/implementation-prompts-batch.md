# Implementation Prompts for Virtuoso CLI Enhancement

## Prompt 1: Update Journey Name Command

```
Please implement an 'update-journey' command for the Virtuoso CLI that:

1. Command structure: ./bin/api-cli update-journey <journey-id> --name "New Name"

2. API endpoint:
   PUT https://api-app2.virtuoso.qa/api/testsuites/{journey_id}
   Body: { "name": "Updated Journey Name" }

3. Features:
   - Accept journey ID as positional argument
   - Accept new name via --name flag
   - Support standard output formats (human, json, yaml, ai)
   - Include proper error handling
   - Validate journey ID format
   - Show before/after in human output

4. Integration:
   - Add to existing CLI command structure
   - Use the same auth/config as other commands
   - Follow the established pattern from create-journey

Location: /Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator
```

## Prompt 2: Get Step Details Command

```
Please implement a 'get-step' command for the Virtuoso CLI that:

1. Command structure: ./bin/api-cli get-step <step-id>

2. API endpoint:
   GET https://api-app2.virtuoso.qa/api/teststeps/{step_id}

3. Returns important fields:
   - id
   - canonicalId (critical for updates)
   - action
   - value
   - meta object
   - optional, ignoreOutcome, skip flags

4. Features:
   - Support standard output formats
   - Highlight canonicalId in human output (needed for updates)
   - Include all meta fields for navigation steps

5. Purpose: Users need this to get canonicalId before updating steps

Location: /Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator
```

## Prompt 3: Update Navigation Step Command

```
Please implement an 'update-navigation' command for the Virtuoso CLI that:

1. Command structure: 
   ./bin/api-cli update-navigation <step-id> <canonical-id> --url "https://new-url.com"

2. API endpoint:
   PUT https://api-app2.virtuoso.qa/api/teststeps/{step_id}/properties
   
3. Request body structure:
   {
     "id": {step_id},
     "canonicalId": "{canonical_id}",
     "action": "NAVIGATE",
     "value": "{url}",
     "meta": {
       "kind": "NAVIGATE",
       "type": "URL",
       "value": "{url}",
       "url": "{url}",
       "useNewTab": false
     },
     "optional": false,
     "ignoreOutcome": false,
     "skip": false
   }

4. Features:
   - Require both step-id and canonical-id (from get-step command)
   - URL validation
   - Option for --new-tab flag
   - Support standard output formats
   - Clear error if canonicalId doesn't match

5. Note: This is specifically for updating navigation steps only

Location: /Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator
```

## Prompt 4: List Checkpoints Command

```
Please implement a 'list-checkpoints' command for the Virtuoso CLI that:

1. Command structure: ./bin/api-cli list-checkpoints <journey-id>

2. Features:
   - List all checkpoints in a journey
   - Show checkpoint ID, name, and position
   - Identify which checkpoint has the shared navigation step
   - Support standard output formats
   - Include step count for each checkpoint

3. Output should help users understand:
   - Which checkpoint is first (has navigation)
   - Order of checkpoints
   - IDs needed for adding steps

4. Human format example:
   Journey: Guest Checkout (ID: 608048)
   Checkpoints:
   1. Navigate to Site (ID: 1678320) [Navigation] - 1 step
   2. Browse Products (ID: 1678321) - 3 steps
   3. Add to Cart (ID: 1678322) - 2 steps

Location: /Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator
```

## Prompt 5: Create Structure Command (Main Implementation)

```
Please implement a 'create-structure' command for the Virtuoso CLI that handles Virtuoso's auto-creation behavior:

1. Command: ./bin/api-cli create-structure --file structure.yaml [--dry-run] [--verbose]

2. CRITICAL BEHAVIORS TO HANDLE:
   - Goal creation auto-creates an initial journey - REUSE IT
   - First checkpoint is always navigation - UPDATE IT, don't create
   - Navigation step is shared across the goal
   
3. Implementation flow:
   a) Create project (unless --project-id specified)
   b) Create goal (captures auto-created journey ID)
   c) Rename auto-created journey to match first journey in structure
   d) For first checkpoint: update existing navigation URL if specified
   e) Add remaining steps to first checkpoint
   f) Create additional checkpoints normally
   g) Create additional journeys normally

4. YAML structure:
   project:
     name: "Test Suite"
     id: 9056  # Optional, use existing
   goals:
     - name: "User Flow"
       url: "https://app.com"
       journeys:
         - name: "Login Journey"  # Renames auto-created journey
           checkpoints:
             - name: "Go to Login"
               navigation_url: "https://app.com/login"  # Updates existing
               steps:
                 - type: wait
                   selector: "login-form"
             - name: "Submit Login"
               steps:
                 - type: fill
                   selector: "#username"
                   value: "testuser"

5. Features:
   - Transaction tracking (rollback on failure)
   - Progress indicator
   - Dry-run mode (show what would be created)
   - Verbose logging
   - Summary output with all created IDs

6. Error handling:
   - If rename fails, log warning but continue
   - If navigation update fails, stop (critical)
   - Track what was created for potential cleanup

Location: /Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator
```

## Integration Test for New Commands

```bash
#!/bin/bash
# test-journey-management.sh

# Test journey rename
JOURNEY_ID=608048
./bin/api-cli update-journey $JOURNEY_ID --name "Renamed Journey Test"

# Test get step details
STEP_ID=19636330
STEP_DETAILS=$(./bin/api-cli get-step $STEP_ID -o json)
CANONICAL_ID=$(echo $STEP_DETAILS | jq -r .canonicalId)

# Test navigation update
./bin/api-cli update-navigation $STEP_ID $CANONICAL_ID --url "https://updated.example.com"

# Test list checkpoints
./bin/api-cli list-checkpoints $JOURNEY_ID

# Test full structure creation
./bin/api-cli create-structure --file test-structure.yaml --dry-run
./bin/api-cli create-structure --file test-structure.yaml --verbose
```
