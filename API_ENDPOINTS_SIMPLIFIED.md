# Virtuoso API Gateway - Simplified Endpoint Specifications

## Section 1: Extracted Endpoint List with Simplifications

| Method | Path | Original Description | Simplified Request | Simplified Response |
|--------|------|---------------------|-------------------|-------------------|
| GET | /api/user | Retrieve current user details | No params | `{ id, name, email }` |
| GET | /api/projects | List accessible projects | Query: `organizationId?` | `[{ id, name, status }]` |
| GET | /api/projects/{project_id}/goals | List goals in a project | Path: `project_id` | `[{ id, name, lastRun }]` |
| POST | /api/projects | Create a new project | Body: `{ name, organizationId }` | `{ id, name }` |
| POST | /api/goals | Create a new goal | Body: `{ name, projectId }` | `{ id, name }` |
| GET | /api/goals/{goal_id}/versions | Get goal versions/snapshots | Path: `goal_id` | `[{ id, version, createdAt }]` |
| POST | /api/goals/{goal_id}/execute | Execute journeys in a goal | Body: `{ startingUrl? }` | `{ jobId, status }` |
| POST | /api/goals/{goal_id}/snapshots/{snapshot_id}/execute | Execute from snapshot | Paths: `goal_id, snapshot_id` | `{ jobId, status }` |
| POST | /api/journeys | Create a new journey | Body: `{ name, goalId }` | `{ id, name }` |
| POST | /api/checkpoints | Create a new checkpoint | Body: `{ name, journeyId }` | `{ id, name }` |
| GET | /api/checkpoints/{checkpoint_id}/steps | Get test steps | Path: `checkpoint_id` | `[{ id, action, target }]` |
| POST | /api/steps | Create a test step | Body: `{ action, target, value?, checkpointId }` | `{ id, action }` |
| POST | /api/executions | Start an execution | Body: `{ goalId, environment? }` | `{ id, status }` |
| GET | /api/executions/{execution_id} | Get execution status | Path: `execution_id` | `{ id, status, progress }` |
| GET | /api/executions/{execution_id}/analysis | Get execution analysis | Path: `execution_id` | `{ passed, failed, errors }` |
| POST | /api/library/checkpoints | Create library checkpoint | Body: `{ name, steps }` | `{ id, name }` |
| GET | /api/library/checkpoints | List library checkpoints | No params | `[{ id, name, stepCount }]` |
| POST | /api/testdata/tables | Create test data table | Body: `{ name, columns }` | `{ id, name }` |
| POST | /api/environments | Create environment | Body: `{ name, variables }` | `{ id, name }` |

## Key Simplifications Applied:

1. **Request Simplifications:**
   - Removed nested objects and complex configurations
   - Made most fields optional with sensible defaults
   - Converted body params to query/path params where appropriate
   - Reduced required fields to bare minimum

2. **Response Simplifications:**
   - Removed verbose metadata and timestamps (except where essential)
   - Flattened nested structures
   - Returned only IDs and essential status fields
   - Simplified error responses to `{ error: "message" }`

3. **Authentication:**
   - Single header: `Authorization: Bearer {token}`
   - Token forwarded to Virtuoso backend unchanged