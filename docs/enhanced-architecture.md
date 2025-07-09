# Enhanced API CLI Architecture

## Overview

Building on our existing API CLI Generator, we need to add an orchestration layer that handles complex workflows while maintaining our security requirements of immutable API definitions.

## Architecture Layers

```
┌─────────────────────────────────────────┐
│          CLI Interface Layer            │
├─────────────────────────────────────────┤
│        Orchestration Layer              │
├─────────────────────────────────────────┤
│     Generated API Client Layer          │
├─────────────────────────────────────────┤
│         HTTP Transport Layer            │
└─────────────────────────────────────────┘
```

## 1. CLI Interface Layer

### Two-Tier Command Structure

```
api-cli/
├── atomic/                  # Low-level, one-to-one API commands
│   ├── create-project
│   ├── create-goal
│   ├── create-journey
│   ├── create-checkpoint
│   └── attach-checkpoint
└── orchestrate/            # High-level composite commands
    ├── create-all          # Single workflow execution
    └── batch-create        # Batch from spec file
```

### Command Examples

**Atomic (Debug/Dev)**:
```bash
api-cli atomic create-project "My Project"
api-cli atomic create-goal <project-id> "Goal Name"
api-cli atomic create-checkpoint <journey-id> "Checkpoint Name"
```

**Orchestrated (Production)**:
```bash
api-cli orchestrate create-all "Project" "Goal" "Checkpoint"
api-cli orchestrate batch-create --spec-file=batch.json
```

## 2. Orchestration Layer

### Core Components

```go
// pkg/orchestrator/workflows.go
type Orchestrator struct {
    client *api.Client
    config *Config
}

// High-level workflow functions
func (o *Orchestrator) CreateGoalWithInitials(projectID, goalName string) (*GoalResult, error)
func (o *Orchestrator) CreateAndAttachCheckpoint(journeyID, name string) (*CheckpointResult, error)
func (o *Orchestrator) ExecuteFullWorkflow(project, goal, checkpoint string) (*WorkflowResult, error)
func (o *Orchestrator) ExecuteBatchSpec(spec *BatchSpec) (*BatchResult, error)
```

### Business Rules Engine

```go
// pkg/rules/rules.go
type BusinessRules struct {
    InitialCheckpointName string
    RequireAttachment     bool
    AutoCreateInitials    bool
}

func (r *BusinessRules) ValidateWorkflow(spec *WorkflowSpec) error
func (r *BusinessRules) ApplyDefaults(spec *WorkflowSpec)
```

## 3. Enhanced Project Structure

```
api-cli-generator/
├── cmd/
│   ├── api-cli/          # Main CLI entry point
│   ├── atomic/           # Atomic commands
│   └── orchestrate/      # Composite commands
├── pkg/
│   ├── orchestrator/     # Workflow orchestration
│   │   ├── workflows.go  # Core workflow logic
│   │   ├── batch.go      # Batch processing
│   │   └── results.go    # Result formatting
│   ├── rules/            # Business rules
│   ├── parser/           # Spec file parsing
│   └── output/           # Output formatting
├── internal/
│   └── api/              # Generated API client
└── specs/
    ├── api.yaml          # OpenAPI spec
    └── examples/         # Example batch specs
```

## 4. Batch Specification Format

### JSON Schema for Batch Operations

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["project_name", "goals"],
  "properties": {
    "project_name": {
      "type": "string",
      "description": "Name of the project to create"
    },
    "goals": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["name"],
        "properties": {
          "name": {"type": "string"},
          "journeys": {
            "type": "array",
            "items": {
              "type": "object",
              "properties": {
                "checkpoints": {
                  "type": "array",
                  "items": {"type": "string"}
                }
              }
            }
          }
        }
      }
    }
  }
}
```

### YAML Alternative

```yaml
project_name: "My Project"
goals:
  - name: "Goal 1"
    journeys:
      - checkpoints:
          - "User Checkpoint A"
          - "User Checkpoint B"
  - name: "Goal 2"
    journeys:
      - checkpoints:
          - "User Checkpoint C"
```

## 5. Output Design

### Structured Output for AI Parsing

```go
type OutputFormatter struct {
    format OutputFormat // json, yaml, human
}

// Example JSON output
{
  "status": "success",
  "operations": [
    {"type": "create_project", "id": "abc123", "status": "✅"},
    {"type": "create_goal", "id": "def456", "status": "✅"},
    {"type": "create_journey", "id": "jny001", "status": "✅"}
  ],
  "summary": {
    "total_operations": 6,
    "successful": 6,
    "failed": 0
  }
}
```

### Human-Readable Output
```
✅ Created Project ID: abc123
✅ Created Goal ID: def456
✅ Snapshot ID: snap789
✅ Created Initial Journey ID: jny001
✅ Created Initial Checkpoint ID: chk001
✅ Created Journey ID: jny002
✅ Created Checkpoint ID: chk002
✅ Checkpoint attached to Journey

Summary: 8 operations completed successfully
```

## 6. Configuration Management

### Enhanced Config Structure

```yaml
# ~/.api-cli/config.yaml
api:
  base_url: "https://api.example.com/v1"
  auth_token: "${API_AUTH_TOKEN}"  # Env var support
  timeout: 30
  retries: 3

organization:
  id: "org123"
  
business_rules:
  initial_checkpoint_name: "INITIAL_CHECKPOINT"
  auto_attach_checkpoints: true
  create_initial_journey: true
  
output:
  default_format: "human"  # json, yaml, human
  verbose: false
  
workflows:
  validate_before_execute: true
  dry_run: false
```

## 7. Implementation Approach

### Phase 1: Enhance Existing Foundation
1. Add orchestrator package structure
2. Create workflow interfaces
3. Implement basic composite commands

### Phase 2: Business Rules Engine
1. Define rule interfaces
2. Implement validation logic
3. Add rule configuration

### Phase 3: Batch Processing
1. Create spec parsers (JSON/YAML)
2. Implement batch executor
3. Add progress tracking

### Phase 4: AI-Friendly Features
1. Structured output formats
2. Self-documenting commands
3. Validation and dry-run modes

## 8. Testing Strategy

```go
// pkg/orchestrator/workflows_test.go
func TestCreateGoalWithInitials(t *testing.T) {
    // Test that goal creation always includes:
    // - Initial journey creation
    // - Initial checkpoint creation and attachment
}

func TestBatchExecution(t *testing.T) {
    // Test batch spec parsing and execution
    // Verify all business rules are applied
}
```

## 9. Security Considerations

- All API definitions remain immutable (from OpenAPI)
- Orchestration logic is fixed in code, not configurable
- Input validation at every layer
- No dynamic workflow definition - only data values change

## 10. Migration Path

1. Keep existing atomic commands from generated code
2. Add orchestration layer on top
3. Gradually migrate users to composite commands
4. Maintain backwards compatibility
