# API CLI Generator - Requirements Evolution

## Comparison: Original vs Enhanced Requirements

### Original Understanding
- Simple OpenAPI → CLI conversion
- 1:1 mapping of API endpoints to CLI commands
- Focus on immutable API definitions
- Basic templating for request bodies

### Enhanced Requirements
- **~100 endpoints** with complex orchestration needs
- **Multi-step workflows** with dependent API calls
- **Business rules enforcement** (auto-attach, initial resources)
- **Batch operations** from structured spec files
- **AI-optimised** interface and output
- **Two-tier commands** (atomic + orchestrated)

## Key Architectural Changes

### 1. Command Structure Evolution

**Before**:
```bash
api-cli users create --name "John"
api-cli projects get PROJECT_ID
```

**After**:
```bash
# Atomic (debug/development)
api-cli atomic create-project "My Project"
api-cli atomic create-goal PROJECT_ID "My Goal"

# Orchestrated (production)
api-cli orchestrate create-all "Project" "Goal" "Checkpoint"
api-cli orchestrate batch-create --spec-file=batch.json
```

### 2. New Orchestration Layer

**Original**: Direct API client calls
```
CLI Command → Generated Client → API
```

**Enhanced**: Orchestration layer manages workflows
```
CLI Command → Orchestrator → Workflow Logic → Generated Client → API
```

### 3. Business Rules Engine

**New Requirement**: Enforce rules like:
- Checkpoints must always be attached after creation
- Goals always create initial journey + checkpoint
- Certain resources have fixed naming conventions

**Implementation**:
```go
type BusinessRules struct {
    AutoAttachCheckpoints   bool
    CreateInitialResources  bool
    InitialCheckpointName   string
}
```

### 4. Batch Processing

**New Capability**: Process multiple resources from spec files

**Example Batch Spec**:
```json
{
  "project_name": "Q1 Project",
  "goals": [
    {
      "name": "Performance Goal",
      "journeys": [
        {
          "checkpoints": ["Milestone 1", "Milestone 2"]
        }
      ]
    }
  ]
}
```

## What Stays the Same

### 1. Core Security Model
- API definitions remain immutable
- Generated from OpenAPI spec
- No runtime modification of endpoints

### 2. Technology Stack
- **Go** as primary language
- **Cobra** for CLI framework
- **oapi-codegen** for client generation
- **Docker** support for containerisation

### 3. Configuration Approach
- YAML/JSON configuration files
- Environment variable support
- Secure credential management

## New Components to Build

### 1. Orchestrator Package (`pkg/orchestrator/`)
- Workflow execution engine
- Resource ID management
- Transaction-like operation handling

### 2. Rules Package (`pkg/rules/`)
- Business rule definitions
- Validation logic
- Default value application

### 3. Parser Package (`pkg/parser/`)
- Batch spec parsing (JSON/YAML)
- Schema validation
- Error reporting

### 4. Enhanced Output Package (`pkg/output/`)
- Multiple format support
- AI-optimised structured output
- Progress tracking for batch operations

## Implementation Priority

### Phase 1: Core Infrastructure (Must Have)
1. Restructure commands (atomic vs orchestrate)
2. Basic orchestrator with fixed workflows
3. Simple business rules (auto-attach)

### Phase 2: Enhanced Features (Should Have)
1. Batch processing from spec files
2. AI-optimised output formats
3. Progress tracking and logging

### Phase 3: Advanced Features (Nice to Have)
1. Dry-run mode
2. Rollback capabilities
3. Caching for performance
4. Concurrent batch execution

## Example Workflow Implementation

### Original Approach (User Manages Steps)
```bash
# User must execute each step
PROJECT_ID=$(api-cli projects create "My Project" | jq -r .id)
GOAL_ID=$(api-cli goals create $PROJECT_ID "My Goal" | jq -r .id)
SNAPSHOT_ID=$(api-cli goals get-snapshot $GOAL_ID | jq -r .id)
JOURNEY_ID=$(api-cli journeys create $GOAL_ID $SNAPSHOT_ID | jq -r .id)
CHECKPOINT_ID=$(api-cli checkpoints create $JOURNEY_ID "My Checkpoint" | jq -r .id)
api-cli checkpoints attach $JOURNEY_ID $CHECKPOINT_ID
```

### Enhanced Approach (Orchestrated)
```bash
# Single command handles everything
api-cli orchestrate create-all "My Project" "My Goal" "My Checkpoint"
```

**Output**:
```
✅ Created Project ID: proj_123
✅ Created Goal ID: goal_456
✅ Created Initial Journey ID: jrny_789
✅ Created Initial Checkpoint ID: chkp_001 (INITIAL_CHECKPOINT)
✅ Created User Journey ID: jrny_002
✅ Created User Checkpoint ID: chkp_002
✅ All resources created and linked successfully
```

## Benefits of Enhanced Architecture

### For Users
- Simple commands hide complexity
- Impossible to forget steps
- Batch operations save time
- Clear, structured output

### For AI Integration
- Self-descriptive commands
- Predictable output format
- No need to understand workflows
- Just provide data values

### For Maintenance
- Business rules in one place
- Testable workflow logic
- Clear separation of concerns
- Easy to add new workflows

## Next Steps

1. **Update Project Structure**
   - Create new package directories
   - Restructure command hierarchy

2. **Implement Core Orchestrator**
   - Start with one complete workflow
   - Add business rule enforcement

3. **Create Example Workflows**
   - Document common use cases
   - Create batch spec templates

4. **Update Documentation**
   - Reflect new architecture
   - Add workflow diagrams
   - Include AI integration guide

## Questions to Resolve

1. **Error Handling**: How to handle partial failures in batch operations?
2. **Idempotency**: Should workflows be idempotent (safe to retry)?
3. **Progress Tracking**: Real-time updates for long-running batches?
4. **Validation**: Pre-flight validation before executing workflows?
5. **Customisation**: Any workflow steps that users might need to customise?

## Conclusion

The enhanced requirements transform our CLI from a simple API wrapper into an intelligent orchestration tool. By hiding complexity behind clean interfaces and enforcing business rules automatically, we create a tool that's both powerful and easy to use - perfect for both human users and AI agents.
