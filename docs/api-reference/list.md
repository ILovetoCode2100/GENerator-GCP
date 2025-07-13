# List Commands

The Virtuoso API CLI provides powerful list commands to explore and navigate your test structure. These commands support multiple output formats (JSON, YAML, human-readable, and AI-friendly) and include pagination for large result sets.

## Overview

List commands follow a hierarchical structure:
- **Projects** → Top-level organizational units
- **Goals** → Test environments within projects  
- **Journeys** → Test suites within goals
- **Checkpoints** → Test cases within journeys

## Commands

### list-projects

Lists all projects in your organization.

```bash
api-cli list-projects [flags]
```

**Flags:**
- `--limit int` - Maximum number of projects to return (default: 50)
- `--offset int` - Number of projects to skip for pagination (default: 0)
- `-o, --output string` - Output format: json, yaml, human, ai (default: human)

**Example:**
```bash
# List all projects in human-readable format
api-cli list-projects

# Get first 10 projects as JSON
api-cli list-projects --limit 10 -o json

# Get projects 20-30 for pagination
api-cli list-projects --offset 20 --limit 10
```

**Output (human format):**
```
ID    NAME                 DESCRIPTION                      CREATED
──    ────                 ───────────                      ───────
123   E-commerce Tests     Full e-commerce test suite       2024-01-15
124   Mobile App Tests     iOS and Android test coverage    2024-01-20
125   API Integration      Backend API testing              2024-02-01

Total: 3 projects
```

### list-goals

Lists all goals (test environments) within a project.

```bash
api-cli list-goals PROJECT_ID [flags]
```

**Arguments:**
- `PROJECT_ID` - The ID of the project to list goals from

**Flags:**
- `--include-archived` - Include archived goals in results
- `-o, --output string` - Output format: json, yaml, human, ai

**Example:**
```bash
# List active goals for project 123
api-cli list-goals 123

# Include archived goals
api-cli list-goals 123 --include-archived

# Get as JSON for automation
api-cli list-goals 123 -o json
```

**Output (AI format):**
```
Found 3 goals in project 123:

1. Goal: Production Tests
   - ID: 456
   - URL: https://www.example.com
   - Snapshot ID: 789

2. Goal: Staging Tests
   - ID: 457
   - URL: https://staging.example.com
   - Snapshot ID: 790

3. Goal: Development Tests
   - ID: 458
   - URL: https://dev.example.com
   - Snapshot ID: 791

Next steps:
1. Get goal snapshot: api-cli get-goal-snapshot 456
2. List journeys: api-cli list-journeys 456 789
3. Create a new journey: api-cli create-journey 456 789 "New Journey"
```

### list-journeys

Lists all journeys (test suites) within a goal.

```bash
api-cli list-journeys GOAL_ID SNAPSHOT_ID [flags]
```

**Arguments:**
- `GOAL_ID` - The ID of the goal
- `SNAPSHOT_ID` - The snapshot (version) ID of the goal

**Flags:**
- `-o, --output string` - Output format: json, yaml, human, ai

**Example:**
```bash
# List journeys for goal 456, snapshot 789
api-cli list-journeys 456 789

# Get detailed JSON output
api-cli list-journeys 456 789 -o json
```

**Output (human format):**
```
ID      NAME                    TITLE                           STATUS
──      ────                    ─────                           ──────
1001    login_flow_e7a9b2      User Login Flow                 Active
1002    checkout_process_3c4    Complete Checkout Process       Draft
1003    search_feature_9f1      Search Functionality Tests      Active
1004    user_profile_2d8        User Profile Management         Archived

Total: 4 journeys
```

### list-checkpoints

Lists all checkpoints (test cases) within a journey, showing their execution order.

```bash
api-cli list-checkpoints JOURNEY_ID [flags]
```

**Arguments:**
- `JOURNEY_ID` - The ID of the journey

**Flags:**
- `--show-steps` - Include step count for each checkpoint
- `-o, --output string` - Output format: json, yaml, human, ai

**Example:**
```bash
# List checkpoints for journey 1001
api-cli list-checkpoints 1001

# Show with step counts
api-cli list-checkpoints 1001 --show-steps
```

**Output (human format):**
```
Journey: User Login Flow (ID: 1001)
Status: Active

Checkpoints:
1. Navigate to Login Page (ID: 5001) [Navigation] - 1 step
2. Enter Valid Credentials (ID: 5002) - 3 steps
3. Verify Dashboard Access (ID: 5003) - 2 steps
4. Test Remember Me Feature (ID: 5004) - 4 steps
```

## Output Formats

### JSON Format

Structured data ideal for automation and integration:

```json
{
  "status": "success",
  "count": 3,
  "projects": [
    {
      "id": 123,
      "name": "E-commerce Tests",
      "description": "Full e-commerce test suite",
      "organizationId": 456,
      "createdAt": "2024-01-15T10:30:00Z"
    }
  ]
}
```

### YAML Format

Human-readable structured data:

```yaml
status: success
count: 3
projects:
  - id: 123
    name: E-commerce Tests
    description: Full e-commerce test suite
    organization_id: 456
    created_at: 2024-01-15T10:30:00Z
```

### AI Format

Optimized for AI assistants and automation tools, includes context and next steps:

```
Found 3 projects in organization 456:

1. Project: E-commerce Tests
   - ID: 123
   - Description: Full e-commerce test suite
   - Created: Jan 15, 2024

Next steps:
1. List goals for a project: api-cli list-goals 123
2. Create a new goal: api-cli create-goal 123 "Goal Name"
```

### Human Format (Default)

Tabular format optimized for terminal display with automatic column width adjustment.

## Pagination

For large result sets, use pagination flags to retrieve data in chunks:

```bash
# Get first 20 projects
api-cli list-projects --limit 20

# Get next 20 projects
api-cli list-projects --offset 20 --limit 20

# Get projects 40-60
api-cli list-projects --offset 40 --limit 20
```

## Integration Examples

### Bash Script - Find Project by Name
```bash
#!/bin/bash
PROJECT_NAME="E-commerce Tests"
PROJECT_ID=$(api-cli list-projects -o json | jq -r ".projects[] | select(.name==\"$PROJECT_NAME\") | .id")
echo "Project ID: $PROJECT_ID"
```

### Python - List All Active Journeys
```python
import subprocess
import json

def get_active_journeys(goal_id, snapshot_id):
    cmd = f"api-cli list-journeys {goal_id} {snapshot_id} -o json"
    result = subprocess.run(cmd.split(), capture_output=True, text=True)
    data = json.loads(result.stdout)
    
    active_journeys = [j for j in data['journeys'] 
                      if not j['archived'] and not j['draft']]
    return active_journeys
```

### Node.js - Navigate Test Hierarchy
```javascript
const { execSync } = require('child_process');

async function getTestHierarchy(projectId) {
  // Get goals
  const goalsJson = execSync(`api-cli list-goals ${projectId} -o json`);
  const goals = JSON.parse(goalsJson).goals;
  
  for (const goal of goals) {
    console.log(`Goal: ${goal.name}`);
    
    // Get journeys for each goal
    const journeysJson = execSync(
      `api-cli list-journeys ${goal.id} ${goal.snapshotId} -o json`
    );
    const journeys = JSON.parse(journeysJson).journeys;
    
    for (const journey of journeys) {
      console.log(`  Journey: ${journey.title}`);
    }
  }
}
```

## Best Practices

1. **Use JSON output for automation** - Parse structured data instead of screen-scraping human output
2. **Implement pagination** - Don't assume all results fit in one response
3. **Check status in responses** - Verify `"status": "success"` before processing data
4. **Cache IDs** - Store frequently used IDs to reduce API calls
5. **Use AI format for troubleshooting** - It includes helpful context and next steps

## Common Workflows

### 1. Navigate from Project to Checkpoints
```bash
# List projects
api-cli list-projects

# List goals in project 123
api-cli list-goals 123

# Get snapshot for goal 456
SNAPSHOT=$(api-cli get-goal-snapshot 456)

# List journeys
api-cli list-journeys 456 $SNAPSHOT

# List checkpoints in journey 1001
api-cli list-checkpoints 1001
```

### 2. Find All Draft Journeys
```bash
# Using jq to filter JSON output
api-cli list-journeys 456 789 -o json | jq '.journeys[] | select(.draft==true)'
```

### 3. Count Total Tests
```bash
# Count total projects
PROJECT_COUNT=$(api-cli list-projects -o json | jq '.count')

# Count goals in a project
GOAL_COUNT=$(api-cli list-goals 123 -o json | jq '.count')

echo "Total projects: $PROJECT_COUNT"
echo "Goals in project 123: $GOAL_COUNT"
```

## Error Handling

List commands handle errors gracefully:

- **Invalid IDs**: Clear error message with the invalid value
- **Missing arguments**: Usage help with required arguments
- **API errors**: Descriptive error messages with status codes
- **Network issues**: Timeout and retry information

Example error:
```
Error: invalid project ID: strconv.Atoi: parsing "abc": invalid syntax
```

## Related Commands

- [`create-project`](create.md#create-project) - Create a new project
- [`create-goal`](create.md#create-goal) - Create a new goal
- [`create-journey`](create.md#create-journey) - Create a new journey
- [`create-checkpoint`](create.md#create-checkpoint) - Create a new checkpoint
- [`get-goal-snapshot`](get.md#get-goal-snapshot) - Get snapshot ID for a goal
- [`execute-goal`](execute.md#execute-goal) - Run tests in a goal
