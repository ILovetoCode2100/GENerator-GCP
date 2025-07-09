# API CLI Generator - Enhanced Implementation Plan

## Executive Summary

Based on the refined requirements, we need to evolve our API CLI Generator from a simple OpenAPI-to-CLI converter into a sophisticated orchestration tool that:

1. Hides complex multi-step workflows behind simple commands
2. Enforces business rules automatically
3. Supports batch operations
4. Provides AI-friendly interfaces and output

## Key Architectural Decisions

### 1. Two-Tier Command Structure
- **Atomic Commands**: Direct 1:1 API mapping for debugging
- **Orchestrated Commands**: High-level workflows for production use

### 2. Orchestration Layer
- Encapsulates all workflow logic
- Enforces business rules (e.g., auto-attach checkpoints)
- Handles ID propagation between dependent calls

### 3. Immutable Workflow Definitions
- Workflows defined in Go code, not configuration
- Prevents AI/users from modifying business logic
- Only data values are parameterised

## Implementation Phases

### Phase 1: Foundation Enhancement (Week 1)

#### 1.1 Restructure Commands
```bash
# Current structure
api-cli <operation> <args>

# New structure
api-cli atomic <operation> <args>      # Low-level
api-cli orchestrate <workflow> <args>  # High-level
```

#### 1.2 Create Orchestrator Package
```go
// pkg/orchestrator/orchestrator.go
package orchestrator

type Orchestrator struct {
    client     *api.Client
    config     *Config
    rules      *BusinessRules
    output     *OutputFormatter
}

type WorkflowResult struct {
    Operations []Operation
    Resources  map[string]string // ID mappings
    Success    bool
    Error      error
}
```

#### 1.3 Define Workflow Interfaces
```go
// pkg/orchestrator/interfaces.go
type Workflow interface {
    Name() string
    Validate(inputs map[string]interface{}) error
    Execute(ctx context.Context) (*WorkflowResult, error)
}

type BatchExecutor interface {
    ParseSpec(path string) (*BatchSpec, error)
    Execute(spec *BatchSpec) (*BatchResult, error)
}
```

### Phase 2: Core Workflows (Week 1-2)

#### 2.1 Implement Fixed Workflows
```go
// pkg/orchestrator/workflows/create_goal.go
func (o *Orchestrator) CreateGoalWithInitials(
    projectID, goalName string,
) (*GoalResult, error) {
    // 1. Create goal
    goal, err := o.client.CreateGoal(projectID, goalName)
    
    // 2. Get snapshot ID
    snapshot, err := o.client.GetSnapshot(goal.ID)
    
    // 3. Create initial journey
    journey, err := o.client.CreateJourney(goal.ID, snapshot.ID)
    
    // 4. Create and attach initial checkpoint
    checkpoint, err := o.createAndAttachCheckpoint(
        journey.ID, 
        o.config.InitialCheckpointName,
    )
    
    return &GoalResult{
        GoalID:              goal.ID,
        SnapshotID:          snapshot.ID,
        InitialJourneyID:    journey.ID,
        InitialCheckpointID: checkpoint.ID,
    }, nil
}
```

#### 2.2 Implement Composite Commands
```go
// cmd/orchestrate/create_all.go
var createAllCmd = &cobra.Command{
    Use:   "create-all [project] [goal] [checkpoint]",
    Short: "Create complete workflow in one command",
    Args:  cobra.ExactArgs(3),
    RunE: func(cmd *cobra.Command, args []string) error {
        result, err := orchestrator.ExecuteFullWorkflow(
            args[0], args[1], args[2],
        )
        return outputFormatter.Display(result)
    },
}
```

### Phase 3: Batch Processing (Week 2)

#### 3.1 Spec Parser
```go
// pkg/parser/spec.go
type BatchSpec struct {
    ProjectName string `json:"project_name" yaml:"project_name"`
    Goals       []GoalSpec `json:"goals" yaml:"goals"`
}

type GoalSpec struct {
    Name     string       `json:"name" yaml:"name"`
    Journeys []JourneySpec `json:"journeys" yaml:"journeys"`
}

func ParseSpecFile(path string) (*BatchSpec, error) {
    // Support both JSON and YAML
}
```

#### 3.2 Batch Executor
```go
// pkg/orchestrator/batch.go
func (o *Orchestrator) ExecuteBatch(spec *BatchSpec) (*BatchResult, error) {
    result := &BatchResult{}
    
    // Create project
    project, err := o.client.CreateProject(spec.ProjectName)
    result.ProjectID = project.ID
    
    // Process each goal
    for _, goalSpec := range spec.Goals {
        goalResult, err := o.processGoalSpec(project.ID, goalSpec)
        result.Goals = append(result.Goals, goalResult)
    }
    
    return result, nil
}
```

### Phase 4: Output Formatting (Week 2-3)

#### 4.1 Multiple Output Formats
```go
// pkg/output/formatter.go
type Formatter interface {
    Format(result interface{}) (string, error)
}

type JSONFormatter struct{}
type YAMLFormatter struct{}
type HumanFormatter struct{}
type AIFormatter struct{} // Optimised for AI parsing
```

#### 4.2 AI-Optimised Output
```json
{
  "execution_id": "exec_12345",
  "timestamp": "2025-01-08T10:30:00Z",
  "status": "success",
  "operations": [
    {
      "sequence": 1,
      "type": "create_project",
      "resource_type": "project",
      "resource_id": "proj_abc123",
      "status": "success",
      "duration_ms": 245
    }
  ],
  "resource_map": {
    "project": "proj_abc123",
    "goal": "goal_def456",
    "initial_journey": "jrny_001",
    "initial_checkpoint": "chkp_001"
  }
}
```

### Phase 5: Testing & Documentation (Week 3)

#### 5.1 Integration Tests
```go
// tests/integration/workflows_test.go
func TestCompleteWorkflow(t *testing.T) {
    // Test with mock API server
    // Verify all steps execute in order
    // Verify business rules are enforced
}
```

#### 5.2 Documentation
- User guide for both atomic and orchestrated commands
- AI integration guide with examples
- Workflow documentation
- Batch spec examples

## Directory Structure Evolution

```
api-cli-generator/
├── cmd/
│   ├── api-cli/
│   │   └── main.go
│   ├── atomic/              # NEW: Atomic commands
│   │   ├── root.go
│   │   └── [generated commands]
│   └── orchestrate/         # NEW: Orchestrated commands
│       ├── root.go
│       ├── create_all.go
│       └── batch_create.go
├── pkg/                     # NEW: Package structure
│   ├── orchestrator/
│   │   ├── orchestrator.go
│   │   ├── workflows.go
│   │   └── batch.go
│   ├── rules/
│   │   └── business_rules.go
│   ├── parser/
│   │   └── spec_parser.go
│   └── output/
│       ├── formatter.go
│       ├── json.go
│       ├── yaml.go
│       └── human.go
├── internal/
│   └── api/                # Generated API client
├── specs/
│   ├── api.yaml
│   └── examples/
│       ├── simple-batch.json
│       └── complex-batch.yaml
└── tests/
    ├── unit/
    └── integration/
```

## Configuration Updates

```yaml
# Enhanced config.yaml
api:
  base_url: "https://api.example.com/v1"
  auth_token: "${API_AUTH_TOKEN}"
  
organization:
  id: "org123"
  
workflows:
  # Fixed workflow settings
  initial_checkpoint_name: "INITIAL_CHECKPOINT"
  auto_attach_checkpoints: true
  validate_before_execute: true
  
output:
  default_format: "human"
  ai_mode: false  # Enables structured output
  
logging:
  level: "info"
  file: "~/.api-cli/logs/api-cli.log"
```

## Key Implementation Notes

### 1. Maintain Immutability
- Workflows are defined in Go code, not config
- Only data values (names, IDs) are parameterised
- Business rules are compiled into the binary

### 2. Error Handling
- Each workflow step should be idempotent where possible
- Implement rollback for failed batch operations
- Clear error messages indicating which step failed

### 3. AI Integration
- Self-documenting commands with clear help text
- Structured output formats for easy parsing
- Dry-run mode for validation
- Progress indicators for long operations

### 4. Performance
- Concurrent execution for independent operations
- Connection pooling for API calls
- Caching for repeated lookups (e.g., snapshot IDs)

## Migration Strategy

1. **Week 1**: Implement orchestrator without breaking existing commands
2. **Week 2**: Add new command structure, keep old as aliases
3. **Week 3**: Update documentation and examples
4. **Week 4**: Deprecate old command structure

## Success Metrics

- All workflows complete without manual intervention
- Batch operations handle 100+ goals efficiently
- AI can use CLI with only command names and data inputs
- Zero exposure of internal workflow steps to users
