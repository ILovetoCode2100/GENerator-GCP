# Virtuoso API CLI - File Index

## Configuration Files
- [`/config/virtuoso-config.yaml`](../config/virtuoso-config.yaml) - Main configuration with credentials
- [`/scripts/setup-virtuoso.sh`](../scripts/setup-virtuoso.sh) - Environment variable setup script

## Implementation Files
- [`/pkg/config/virtuoso.go`](../pkg/config/virtuoso.go) - Configuration loader
- [`/pkg/virtuoso/client.go`](../pkg/virtuoso/client.go) - API client implementation

## Documentation
- [`/VIRTUOSO_QUICK_START.md`](../VIRTUOSO_QUICK_START.md) - Complete usage guide
- [`/docs/api-reference/list.md`](list.md) - List commands reference (list-projects, list-goals, list-journeys, list-checkpoints)
- [`/documents/virtuoso-config-summary.md`](../../documents/virtuoso-config-summary.md) - Configuration summary
- [`/documents/virtuoso-ready-for-endpoints.md`](../../documents/virtuoso-ready-for-endpoints.md) - What's needed next

## Test Files
- [`/examples/test-structure.json`](../examples/test-structure.json) - Example project structure
- [`/scripts/test-virtuoso-api.sh`](../scripts/test-virtuoso-api.sh) - API connection test

## Quick Commands

```bash
# Set environment
source ./scripts/setup-virtuoso.sh

# Test API
./scripts/test-virtuoso-api.sh

# Build CLI
make build

# Create structure
./bin/api-cli create-structure --file examples/test-structure.json

# List commands
./bin/api-cli list-projects                                    # List all projects
./bin/api-cli list-goals <project-id>                         # List goals in project
./bin/api-cli list-journeys <goal-id> <snapshot-id>          # List journeys in goal
./bin/api-cli list-checkpoints <journey-id>                   # List checkpoints in journey

# Add steps
./bin/api-cli add-step <checkpoint-id> navigate --url "https://example.com"
```

## Current Status

✅ **Complete**:
- Configuration management
- Authentication setup
- Client structure
- Error handling
- List commands (list-projects, list-goals, list-journeys, list-checkpoints)
- Pagination support for list commands
- Multiple output formats (JSON, YAML, human, AI)
- Rich documentation with examples

🔴 **Needed**:
- Additional API endpoint paths
- Complex request/response formats
- Advanced step type definitions
- Business rule clarifications
