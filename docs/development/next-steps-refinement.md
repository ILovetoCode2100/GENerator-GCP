# Next Steps: Refine and Test Virtuoso CLI

## 1. Add List/Read Commands
Create commands to view existing resources:
- `list-projects` - List all projects in the organization
- `list-goals <project-id>` - List goals in a project
- `list-journeys <goal-id> <snapshot-id>` - List journeys in a goal
- `list-checkpoints <journey-id>` - List checkpoints in a journey

## 2. Create Batch/Structure Command
Implement the `create-structure` command that takes a JSON/YAML file:
```yaml
project:
  name: "E2E Test Suite"
goals:
  - name: "User Authentication"
    url: "https://app.example.com"
    journeys:
      - name: "Login Flow"
        checkpoints:
          - name: "Navigate to Login"
            steps:
              - type: navigate
                url: "https://app.example.com/login"
              - type: wait
                selector: "Login Form"
                timeout: 5000
          - name: "Submit Login"
            steps:
              - type: click
                selector: "Submit Button"
```

## 3. Add Validation and Error Recovery
- Pre-validate IDs before making API calls
- Add retry logic for transient failures
- Implement rollback for partial failures in batch operations
- Add dry-run mode to preview what will be created

## 4. Improve Output and Logging
- Add progress bars for batch operations
- Create detailed logs for debugging
- Add export functionality to save created structure as JSON/YAML
- Add quiet mode for CI/CD pipelines

## 5. Create Integration Tests
```bash
#!/bin/bash
# test/integration_test.sh

# Test 1: Create and verify project
PROJECT_ID=$(./bin/api-cli create-project "Test-$(date +%s)" -o json | jq -r .project_id)
assert_not_empty "$PROJECT_ID" "Project creation failed"

# Test 2: Create goal and verify snapshot
GOAL_RESULT=$(./bin/api-cli create-goal $PROJECT_ID "Test Goal" -o json)
GOAL_ID=$(echo $GOAL_RESULT | jq -r .goal_id)
SNAPSHOT_ID=$(echo $GOAL_RESULT | jq -r .snapshot_id)
assert_not_empty "$GOAL_ID" "Goal creation failed"
assert_not_empty "$SNAPSHOT_ID" "Snapshot retrieval failed"

# Test 3: Verify journey creation
JOURNEY_ID=$(./bin/api-cli create-journey $GOAL_ID $SNAPSHOT_ID "Test Journey" -o json | jq -r .journey_id)
assert_not_empty "$JOURNEY_ID" "Journey creation failed"

# Test 4: Test checkpoint creation and attachment
CHECKPOINT_ID=$(./bin/api-cli create-checkpoint $JOURNEY_ID $GOAL_ID $SNAPSHOT_ID "Test CP" -o json | jq -r .checkpoint_id)
assert_not_empty "$CHECKPOINT_ID" "Checkpoint creation failed"

# Test 5: Verify all step types work
./bin/api-cli add-step navigate $CHECKPOINT_ID --url "https://example.com" || exit 1
./bin/api-cli add-step wait $CHECKPOINT_ID --selector "body" --timeout 1000 || exit 1
./bin/api-cli add-step click $CHECKPOINT_ID --selector "button" || exit 1

echo "✅ All integration tests passed!"
```

## 6. Add Helper Commands
- `validate-config` - Check if config file is valid and API is reachable
- `whoami` - Show current organization and user info
- `delete-project <id>` - Clean up test projects
- `export-structure <project-id>` - Export project as JSON/YAML

## 7. Create Documentation
- **User Guide**: Step-by-step tutorials
- **API Reference**: Document all commands and options
- **Examples**: Common workflows and patterns
- **Troubleshooting**: Common issues and solutions

## 8. Performance Optimizations
- Implement concurrent operations for batch creation
- Add caching for frequently accessed data (like snapshot IDs)
- Optimize API calls by batching where possible

## 9. CI/CD Integration
```yaml
# .github/workflows/test-creation.yml
name: Create Virtuoso Tests
on:
  push:
    branches: [main]
jobs:
  create-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Create test structure
        run: |
          ./bin/api-cli create-structure --file tests/structure.yaml
        env:
          API_CLI_CONFIG: ./config/virtuoso-config.yaml
```

## 10. Package and Distribute
- Create releases with pre-built binaries
- Add Homebrew formula for macOS
- Create Docker image
- Add shell completion scripts

## Quick Refinements to Start With

### 1. Add a validate command:
```bash
./bin/api-cli validate-config
# Output: ✅ Configuration valid, API reachable
```

### 2. Add list commands:
```bash
./bin/api-cli list-projects
./bin/api-cli list-goals 9061
```

### 3. Create the batch structure command:
```bash
./bin/api-cli create-structure --file test-suite.yaml
```

Which of these would you like to implement first?
