# AI Implementation Brief: API CLI with Orchestration

## Project Context

You are implementing an API CLI tool that wraps ~100 API endpoints with complex orchestration logic. The CLI must hide multi-step workflows behind simple commands and enforce business rules automatically.

## Core Requirements

### 1. Two-Tier Command Structure
```bash
# Atomic (low-level, for debugging)
api-cli atomic create-project "Name"
api-cli atomic create-goal <project-id> "Name"

# Orchestrated (high-level, for production)
api-cli orchestrate create-all "Project" "Goal" "Checkpoint"
api-cli orchestrate batch-create --spec-file=batch.json
```

### 2. Workflow Example
When user runs: `api-cli orchestrate create-all "Project" "Goal" "Checkpoint"`

The CLI must internally execute:
1. Create Project → get project ID
2. Create Goal (using project ID) → get goal ID  
3. Get Goal Snapshot ID (using goal ID)
4. Create Initial Journey (using goal ID, snapshot ID) → get journey ID
5. Create Initial Checkpoint (using journey ID, name="INITIAL_CHECKPOINT")
6. Attach Initial Checkpoint to Journey
7. Create User Journey (for user's checkpoint)
8. Create User Checkpoint (using user-provided name)
9. Attach User Checkpoint to Journey

User sees only: "✅ Created all resources successfully"

### 3. Business Rules (MUST enforce)
- **Every checkpoint** must be attached after creation (no exceptions)
- **Every goal** automatically gets initial journey + initial checkpoint
- **Initial checkpoint** always named "INITIAL_CHECKPOINT"
- Users **never** manually attach checkpoints or create initial resources

### 4. Batch Operations
Support spec files for creating multiple resources:

```json
{
  "project_name": "Q1 Planning",
  "goals": [
    {
      "name": "Performance Goal",
      "journeys": [
        {
          "checkpoints": ["Milestone 1", "Milestone 2"]
        }
      ]
    },
    {
      "name": "Security Goal",
      "journeys": [
        {
          "checkpoints": ["Audit Complete"]
        }
      ]
    }
  ]
}
```

## Technical Requirements

### Architecture
```
CLI Layer (Cobra commands)
    ↓
Orchestration Layer (workflows)
    ↓
Business Rules Layer (validation)
    ↓
Generated API Client (from OpenAPI)
    ↓
HTTP Layer
```

### Technology Stack
- **Language**: Go 1.21+
- **CLI**: Cobra + Viper
- **API Client**: Generated via oapi-codegen from OpenAPI spec
- **Config**: YAML with env var support
- **Output**: JSON/YAML/Human-readable formats

### Project Structure
```
cmd/
├── api-cli/          # Main entry
├── atomic/           # Low-level commands
└── orchestrate/      # High-level commands

pkg/
├── orchestrator/     # Workflow logic
├── rules/           # Business rules
├── parser/          # Batch spec parsing
└── output/          # Formatters

internal/
└── api/             # Generated client
```

## Implementation Constraints

### Must Have
1. Workflows defined in Go code (not config files)
2. All business rules enforced automatically
3. Clean output suitable for AI parsing
4. Comprehensive error handling
5. Atomic operations where possible

### Must NOT Have
1. No manual checkpoint attachment commands
2. No manual initial resource creation
3. No workflow modification via config
4. No skipping of business rules

## Output Requirements

### Human Format
```
✅ Created Project ID: proj_123
✅ Created Goal ID: goal_456
✅ Created Initial Journey ID: jrny_001
✅ Created Initial Checkpoint ID: chkp_001
✅ Created User Journey ID: jrny_002
✅ Created User Checkpoint ID: chkp_002

Summary: 6 resources created successfully
```

### AI Format (--output=ai)
```json
{
  "status": "success",
  "execution_id": "exec_12345",
  "operations": [
    {
      "type": "create_project",
      "resource_id": "proj_123",
      "status": "success"
    }
  ],
  "resource_map": {
    "project": "proj_123",
    "goal": "goal_456",
    "journeys": ["jrny_001", "jrny_002"],
    "checkpoints": ["chkp_001", "chkp_002"]
  }
}
```

## Example Implementation Pattern

```go
// pkg/orchestrator/workflows.go
func (o *Orchestrator) CreateGoalWithInitials(projectID, goalName string) (*GoalResult, error) {
    // Step 1: Create goal
    goal, err := o.apiClient.CreateGoal(projectID, goalName)
    if err != nil {
        return nil, fmt.Errorf("create goal: %w", err)
    }
    
    // Step 2: Get snapshot (required for journey)
    snapshot, err := o.apiClient.GetGoalSnapshot(goal.ID)
    if err != nil {
        return nil, fmt.Errorf("get snapshot: %w", err)
    }
    
    // Step 3: Create initial journey (business rule)
    journey, err := o.apiClient.CreateJourney(goal.ID, snapshot.ID)
    if err != nil {
        return nil, fmt.Errorf("create initial journey: %w", err)
    }
    
    // Step 4: Create and attach initial checkpoint (business rule)
    checkpoint, err := o.createAndAttachCheckpoint(
        journey.ID, 
        o.config.InitialCheckpointName, // Always "INITIAL_CHECKPOINT"
    )
    if err != nil {
        return nil, fmt.Errorf("create initial checkpoint: %w", err)
    }
    
    return &GoalResult{
        GoalID:              goal.ID,
        SnapshotID:          snapshot.ID,
        InitialJourneyID:    journey.ID,
        InitialCheckpointID: checkpoint.ID,
    }, nil
}

// Helper that ALWAYS attaches after creation
func (o *Orchestrator) createAndAttachCheckpoint(journeyID, name string) (*Checkpoint, error) {
    // Create
    checkpoint, err := o.apiClient.CreateCheckpoint(journeyID, name)
    if err != nil {
        return nil, err
    }
    
    // ALWAYS attach (business rule)
    if err := o.apiClient.AttachCheckpoint(journeyID, checkpoint.ID); err != nil {
        return nil, fmt.Errorf("attach checkpoint: %w", err)
    }
    
    return checkpoint, nil
}
```

## Testing Requirements

1. Unit tests for each orchestrator method
2. Integration tests for complete workflows
3. Verify business rules cannot be bypassed
4. Test batch operations with 50+ goals
5. Verify output formats are consistent

## Success Criteria

✅ User can create complex resource hierarchies with one command
✅ Impossible to create checkpoint without attachment
✅ Impossible to create goal without initial resources
✅ AI can use CLI without understanding workflows
✅ Batch operations complete successfully for large specs

## Additional Notes

- Prioritise clarity over cleverness
- Every workflow step should log progress
- Error messages must indicate which step failed
- Consider implementing dry-run mode for validation
- All IDs should be preserved in output for reference

Remember: The goal is to make the complex simple. Users should never need to understand the underlying API workflow - just what they want to create.
