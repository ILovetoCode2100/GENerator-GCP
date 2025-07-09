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

### Navigation and Control Steps
```bash
# Create navigation step
./bin/api-cli create-step-navigate <checkpoint-id> "https://example.com" <position>

# Create wait time step (seconds)
./bin/api-cli create-step-wait-time <checkpoint-id> 5 <position>

# Create wait for element step
./bin/api-cli create-step-wait-element <checkpoint-id> "Loading Complete" <position>

# Create window resize step
./bin/api-cli create-step-window <checkpoint-id> 1920 1080 <position>
```

### Mouse Action Steps
```bash
# Create click step
./bin/api-cli create-step-click <checkpoint-id> "Sign in button" <position>

# Create double-click step
./bin/api-cli create-step-double-click <checkpoint-id> "Element" <position>

# Create hover step
./bin/api-cli create-step-hover <checkpoint-id> "Menu item" <position>

# Create right-click step
./bin/api-cli create-step-right-click <checkpoint-id> "Context menu" <position>
```

### Input and Form Steps
```bash
# Create text input step
./bin/api-cli create-step-write <checkpoint-id> "user@example.com" "Email field" <position>

# Create keyboard press step
./bin/api-cli create-step-key <checkpoint-id> "Enter" <position>

# Create dropdown selection step
./bin/api-cli create-step-pick <checkpoint-id> "United States" "Country dropdown" <position>

# Create file upload step
./bin/api-cli create-step-upload <checkpoint-id> "document.pdf" "File input" <position>
```

### Scroll Steps
```bash
# Create scroll to top step
./bin/api-cli create-step-scroll-top <checkpoint-id> <position>

# Create scroll to bottom step
./bin/api-cli create-step-scroll-bottom <checkpoint-id> <position>

# Create scroll to element step
./bin/api-cli create-step-scroll-element <checkpoint-id> "Submit button" <position>
```

### Assertion Steps
```bash
# Create element exists assertion
./bin/api-cli create-step-assert-exists <checkpoint-id> "Welcome message" <position>

# Create element not exists assertion
./bin/api-cli create-step-assert-not-exists <checkpoint-id> "Error message" <position>

# Create value equals assertion
./bin/api-cli create-step-assert-equals <checkpoint-id> "Total price" "$99.99" <position>

# Create checkbox/radio checked assertion
./bin/api-cli create-step-assert-checked <checkpoint-id> "Terms checkbox" <position>
```

### Data and Browser Management Steps
```bash
# Create store variable step
./bin/api-cli create-step-store <checkpoint-id> "Order ID" "orderId" <position>

# Create execute JavaScript step
./bin/api-cli create-step-execute-js <checkpoint-id> "console.log('test')" <position>

# Create add cookie step
./bin/api-cli create-step-add-cookie <checkpoint-id> "session" "abc123" <position>

# Create dismiss alert step
./bin/api-cli create-step-dismiss-alert <checkpoint-id> <position>

# Create comment step
./bin/api-cli create-step-comment <checkpoint-id> "This validates the login flow" <position>
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