# Virtuoso CLI - Complete Command Reference

## Setup (One Time)
```bash
cd /Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator
source ./scripts/setup-virtuoso.sh
```

## Test Commands
```bash
# Test build
./scripts/test-build.sh

# Test API connection  
./scripts/test-virtuoso-api.sh
```

## Project Management
```bash
# List all projects
./bin/api-cli list-projects

# Create a new project
./bin/api-cli create-project "My Test Project"
```

## Goal Management
```bash
# List goals in a project
./bin/api-cli list-goals <project-id>

# Create a new goal (auto-creates first journey)
./bin/api-cli create-goal <project-id> "Homepage Tests" --url "https://example.com"
```

## Journey Management
```bash
# List journeys for a goal
./bin/api-cli list-journeys <goal-id> <snapshot-id>

# Create additional journey
./bin/api-cli create-journey <goal-id> <snapshot-id> "Alternative Flow"

# Update journey title (rename)
./bin/api-cli update-journey <journey-id> --name "New Journey Title"

# List checkpoints in a journey
./bin/api-cli list-checkpoints <journey-id>
```

## Checkpoint Management
```bash
# Create checkpoint with navigation
./bin/api-cli create-checkpoint <journey-id> "Login Page" --navigation "https://example.com/login"

# Create checkpoint without navigation
./bin/api-cli create-checkpoint <journey-id> "Submit Form"
```

## Step Management
```bash
# Get step details (including canonical ID for navigation steps)
./bin/api-cli get-step <step-id>

# Add various step types
./bin/api-cli add-step <checkpoint-id> click --selector ".button"
./bin/api-cli add-step <checkpoint-id> fill --selector "#email" --value "test@example.com"
./bin/api-cli add-step <checkpoint-id> wait --selector ".loading" --timeout 3000
./bin/api-cli add-step <checkpoint-id> press --key "Enter"

# Update navigation step URL
./bin/api-cli update-navigation <step-id> <canonical-id> --url "https://new-url.com"
./bin/api-cli update-navigation <step-id> <canonical-id> --url "https://new-url.com" --new-tab
```

## Batch Structure Creation
```bash
# Preview what will be created (dry run)
./bin/api-cli create-structure --file structure.yaml --dry-run

# Create structure with verbose output
./bin/api-cli create-structure --file structure.yaml --verbose

# Use existing project
./bin/api-cli create-structure --file structure.yaml --project-id 1234

# Different output formats
./bin/api-cli create-structure --file structure.yaml -o json
./bin/api-cli create-structure --file structure.yaml -o ai
```

## Example Structure File
```yaml
project:
  name: "E-Commerce Test Suite"
  # Or use existing: id: 1234
  
goals:
  - name: "Homepage Tests"
    url: "https://shop.example.com"
    journeys:
      - name: "Main User Flow"  # Will rename auto-created journey
        checkpoints:
          - name: "Landing Page"
            navigation_url: "https://shop.example.com"
            steps:
              - type: wait
                selector: ".hero-banner"
                timeout: 3000
              - type: click
                selector: ".shop-now"
                
      - name: "Search Flow"  # Additional journey
        checkpoints:
          - name: "Search Products"
            navigation_url: "https://shop.example.com/search"
            steps:
              - type: fill
                selector: "#search"
                value: "laptop"
              - type: press
                key: "Enter"
```

## Output Formats
```bash
# Human readable (default)
./bin/api-cli list-projects

# JSON output
./bin/api-cli list-projects -o json

# YAML output
./bin/api-cli list-projects -o yaml

# AI-optimized output
./bin/api-cli list-projects -o ai
```

## Key Features
- **Auto-created journeys**: First goal creates "Suite 1" journey automatically
- **Journey renaming**: Update journey titles while keeping system names
- **Navigation updates**: Modify navigation URLs for existing checkpoints
- **Batch creation**: Create entire test structures from YAML/JSON files
- **Dry run mode**: Preview changes before applying them

## Configuration
- Config file: `config/virtuoso-config.yaml`
- Environment prefix: `VIRTUOSO_`
- Headers automatically included

## Notes
- Goals automatically create a first journey named "Suite 1"
- Journey `name` field is system-generated (Suite 1, Suite 2, etc.)
- Journey `title` field is the user-friendly display name
- First checkpoint in each journey is always navigation
- Use `--verbose` flag for detailed operation logging