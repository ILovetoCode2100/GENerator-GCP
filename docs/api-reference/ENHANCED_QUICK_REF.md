# API CLI Generator - Enhanced Requirements Quick Reference

## What We're Building

An intelligent CLI that wraps ~100 API endpoints with automatic orchestration, hiding complex multi-step workflows behind simple commands.

## Key Features

### 1. Two-Tier Commands
```bash
# Debug/Development (atomic)
api-cli atomic create-project "Name"
api-cli atomic create-goal PROJECT_ID "Name"

# Production (orchestrated)
api-cli orchestrate create-all "Project" "Goal" "Checkpoint"
api-cli orchestrate batch-create --spec-file=goals.json
```

### 2. Automatic Workflow Management
**User types**: One simple command  
**CLI executes**: 6-10 API calls in correct sequence  
**User sees**: Clean success summary

### 3. Business Rules Enforcement
- ✅ Checkpoints always auto-attached
- ✅ Goals always get initial journey + checkpoint
- ✅ Fixed naming for system resources
- ❌ No manual attachment commands
- ❌ No skipping required steps

### 4. Batch Operations
```json
{
  "project_name": "Q1 Goals",
  "goals": [
    {
      "name": "Performance",
      "journeys": [{
        "checkpoints": ["Sprint 1", "Sprint 2"]
      }]
    }
  ]
}
```

### 5. AI-Friendly Output
```json
{
  "status": "success",
  "operations": [...],
  "resource_map": {
    "project": "proj_123",
    "goal": "goal_456"
  }
}
```

## Architecture

```
User Command
    ↓
Orchestration Layer (handles workflows)
    ↓
Business Rules (enforces requirements)
    ↓
API Client (generated from OpenAPI)
    ↓
HTTP Calls
```

## Example Workflow

**Goal Creation** (what happens internally):
1. Create Goal
2. Get Snapshot ID
3. Create Initial Journey
4. Create Initial Checkpoint (name="INITIAL_CHECKPOINT")
5. Attach Initial Checkpoint
6. Create User Journey  
7. Create User Checkpoint
8. Attach User Checkpoint

**What user types**:
```bash
api-cli orchestrate create-all "Project" "Goal" "Checkpoint"
```

## Implementation Status

- [x] Basic project structure
- [x] OpenAPI code generation setup
- [ ] Orchestration layer
- [ ] Business rules engine
- [ ] Batch processing
- [ ] AI-optimised output
- [ ] Comprehensive testing

## Next Steps

1. Implement orchestrator package
2. Add workflow definitions
3. Create batch parser
4. Build output formatters
5. Write integration tests

## Key Documents

- **Enhanced Architecture**: `/docs/enhanced-architecture.md`
- **Implementation Plan**: `/docs/implementation-plan-enhanced.md`
- **Requirements Evolution**: `/docs/requirements-evolution.md`
- **AI Implementation Brief**: `/docs/ai-implementation-brief.md`

## Success Metrics

- Zero manual workflow steps
- 100+ batch operations supported
- AI-usable with minimal context
- Business rules 100% enforced
- Clean, structured output
