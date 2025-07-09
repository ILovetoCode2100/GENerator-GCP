# Virtuoso API CLI - Structure Creation Overview

## API Workflow from Postman Collection

Based on the Postman collection analysis, here's the complete workflow for creating a Virtuoso test structure:

### 1. Create Project
- **Endpoint**: POST `/projects`
- **Body**: `{ "name": "Project Name", "organizationId": 2242 }`
- **Returns**: Project ID

### 2. Create Goal
- **Endpoint**: POST `/goals`
- **Body**: 
  ```json
  {
    "projectId": 12345,
    "name": "Goal Name",
    "url": "https://www.example.com",
    "deviceSize": { "width": 1280, "height": 800 },
    "createFirstJourney": true,
    "meta": { ... }
  }
  ```
- **Returns**: Goal ID
- **Note**: `createFirstJourney: true` creates initial journey automatically

### 3. Get Snapshot ID
- **Endpoint**: GET `/goals/{goalId}/versions`
- **Returns**: Array of snapshots, use first one's `snapshotId`

### 4. Get/Create Journeys
- **Get Existing**: GET `/testsuites/latest_status?versionId={snapshotId}&goalId={goalId}`
- **Create New**: POST `/testsuites`
  ```json
  {
    "snapshotId": 123,
    "goalId": 456,
    "name": "Journey Name",
    "title": "Journey Title",
    "archived": false,
    "draft": true
  }
  ```

### 5. Create & Attach Checkpoint (COMBINED)
**Important**: The CLI must combine these two operations into one command

#### 5a. Create Checkpoint
- **Endpoint**: POST `/testcases`
- **Body**:
  ```json
  {
    "snapshotId": 123,
    "goalId": 456,
    "title": "Checkpoint Name"
  }
  ```
- **Returns**: Checkpoint ID

#### 5b. Attach Checkpoint (AUTOMATICALLY)
- **Endpoint**: POST `/testsuites/{journeyId}/checkpoints/attach`
- **Body**:
  ```json
  {
    "checkpointId": 789,
    "position": 2
  }
  ```
- **Note**: Position must be 2 or greater

### 6. Add Steps to Checkpoint
- **Endpoint**: POST `/teststeps`
- **Body Format**:
  ```json
  {
    "checkpointId": 789,
    "stepIndex": 0,
    "parsedStep": {
      "action": "NAVIGATE",
      "value": "https://example.com",
      "meta": { ... }
    }
  }
  ```

## CLI Command Structure

### Phase 1: Individual Commands (Building Blocks)
```bash
# Create project
api-cli create-project "Project Name"

# Create goal (with initial journey)
api-cli create-goal <project-id> "Goal Name" --url "https://example.com"

# Create journey
api-cli create-journey <goal-id> <snapshot-id> "Journey Name"

# Create and attach checkpoint (COMBINED)
api-cli create-checkpoint <journey-id> "Checkpoint Name" [--position 2]
```

### Phase 2: Structure Creation from JSON
```bash
# Create entire structure
api-cli create-structure --file structure.json
```

**Example structure.json**:
```json
{
  "project": {
    "name": "E-commerce Testing"
  },
  "goals": [
    {
      "name": "Login Flow",
      "url": "https://example.com",
      "journeys": [
        {
          "name": "Happy Path",
          "checkpoints": [
            {
              "name": "Navigate to Login",
              "position": 2
            },
            {
              "name": "Submit Credentials",
              "position": 3
            }
          ]
        }
      ]
    }
  ]
}
```

## Implementation Priority

1. **First**: Implement `create-project` command (simplest)
2. **Second**: Implement `create-goal` with snapshot retrieval
3. **Third**: Implement `create-checkpoint` that combines creation + attachment
4. **Fourth**: Implement `create-structure` for batch creation
5. **Last**: Implement step commands

## Key Business Rules

1. **Checkpoint Attachment**: Always attach checkpoints after creation - never expose as separate command
2. **Initial Journey**: Goals with `createFirstJourney: true` get automatic journey
3. **Position Rule**: Checkpoint position must be 2 or greater
4. **Snapshot Required**: All operations need valid snapshot ID from goal

## Next Steps

Start with implementing the `create-project` command as outlined in the Claude Code prompt, then proceed with the other commands in order.
