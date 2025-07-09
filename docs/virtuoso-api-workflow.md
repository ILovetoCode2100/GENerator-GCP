# Virtuoso API Workflow Summary

Based on the Postman collection analysis, here's the complete workflow for creating a test structure in Virtuoso:

## API Endpoints

1. **Projects**
   - POST /projects - Create project
   - GET /projects?organizationId={orgId} - List projects

2. **Goals**
   - POST /goals - Create goal (with createFirstJourney: true)
   - GET /projects/{projectId}/goals?archived=false - List goals
   - GET /goals/{goalId} - Get goal details
   - GET /goals/{goalId}/versions - Get snapshot IDs

3. **Journeys (TestSuites)**
   - POST /testsuites?envelope=false - Create journey
   - GET /testsuites/latest_status?versionId={snapshotId}&goalId={goalId} - Get journeys

4. **Checkpoints (TestCases)**
   - POST /testcases - Create checkpoint
   - POST /testsuites/{journeyId}/checkpoints/attach - Attach checkpoint to journey

5. **Steps**
   - POST /teststeps - Create test step

## Complete Workflow Sequence

```
1. Create Project
   └─> Returns: projectId

2. Create Goal (with createFirstJourney: true)
   └─> Input: projectId
   └─> Returns: goalId

3. Get Snapshot ID
   └─> Input: goalId
   └─> Returns: snapshotId (from snapshots[0].snapshotId)

4. Get Initial Journey (created automatically with goal)
   └─> Input: goalId, snapshotId
   └─> Returns: journeyId (from Object.keys(response.map)[0])

5. Create Additional Journey (if needed)
   └─> Input: goalId, snapshotId
   └─> Returns: new journeyId

6. Create Checkpoint
   └─> Input: goalId, snapshotId
   └─> Returns: checkpointId

7. Attach Checkpoint to Journey (CRITICAL STEP)
   └─> Input: journeyId, checkpointId, position (must be >= 2)
   └─> Returns: success confirmation

8. Create Steps
   └─> Input: checkpointId, stepIndex, parsedStep
   └─> Returns: stepId
```

## Key Business Rules

1. **Goals always create initial journey** when `createFirstJourney: true`
2. **Checkpoints MUST be attached** to a journey to be valid
3. **Position must be 2 or greater** when attaching checkpoints
4. **Snapshot ID is required** for creating journeys and checkpoints

## CLI Design Decisions

### Atomic Commands (Low-level)
```bash
api-cli create-project --name "Project"
api-cli create-goal --project-id 123 --name "Goal"
api-cli create-journey --goal-id 456 --snapshot-id 789
api-cli create-checkpoint --goal-id 456 --snapshot-id 789 --journey-id 101
```

### Orchestrated Commands (High-level)
```bash
# Create checkpoint automatically attaches it
api-cli create-checkpoint --goal-id 456 --journey-id 101 --name "Checkpoint" --position 2

# Create entire structure from JSON
api-cli create-structure --file structure.json
```

## Response Patterns

All API responses follow this pattern:
```json
{
    "success": true,
    "item": {
        "id": 12345,
        // other fields...
    }
}
```

For list operations:
```json
{
    "success": true,
    "items": [...] // or "map": {...}
}
```

## Headers Required for All Requests

```
Authorization: Bearer {token}
X-Virtuoso-Client-ID: api-cli-generator
X-Virtuoso-Client-Name: api-cli-generator
Content-Type: application/json
```

## Next Implementation Steps

1. ✅ Create Project command
2. Create Goal command (with automatic snapshot retrieval)
3. Create Journey command
4. Create Checkpoint command (with automatic attachment)
5. Add Step commands (navigate, click, fill, wait, assert)
6. Create Structure command (batch creation from JSON)
