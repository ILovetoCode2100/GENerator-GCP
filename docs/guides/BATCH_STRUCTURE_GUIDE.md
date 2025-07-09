# Virtuoso CLI - Batch Structure Quick Reference

## 🚀 New Commands Overview

### Journey Management
```bash
# Update journey name
./bin/api-cli update-journey <journey-id> --name "New Journey Name"

# Example
./bin/api-cli update-journey 608048 --name "Updated Login Flow"
```

### Step Management
```bash
# Get step details (includes canonicalId)
./bin/api-cli get-step <step-id>

# Update navigation step
./bin/api-cli update-navigation <step-id> <canonical-id> --url "https://new-url.com"

# Example workflow
STEP_DETAILS=$(./bin/api-cli get-step 19636330 -o json)
CANONICAL_ID=$(echo $STEP_DETAILS | jq -r .canonicalId)
./bin/api-cli update-navigation 19636330 $CANONICAL_ID --url "https://updated.example.com"
```

### Checkpoint Management
```bash
# List all checkpoints in a journey
./bin/api-cli list-checkpoints <journey-id>

# Example output
Journey: Guest Checkout (ID: 608048)
Checkpoints:
1. Navigate to Site (ID: 1678320) [Navigation] - 1 step
2. Browse Products (ID: 1678321) - 3 steps
3. Add to Cart (ID: 1678322) - 2 steps
```

### Batch Structure Creation
```bash
# Create complete test structure from YAML
./bin/api-cli create-structure --file structure.yaml

# Dry run (preview only)
./bin/api-cli create-structure --file structure.yaml --dry-run

# Use existing project
./bin/api-cli create-structure --file structure.yaml --project-id 9056

# Verbose output
./bin/api-cli create-structure --file structure.yaml --verbose
```

## 📝 Structure File Format

```yaml
project:
  name: "Test Suite Name"
  # Optional: use existing project
  id: 9056
  
goals:
  - name: "Goal Name"
    url: "https://start-url.com"
    journeys:
      # First journey reuses auto-created journey
      - name: "Journey Name"
        checkpoints:
          # First checkpoint updates existing navigation
          - name: "First Checkpoint"
            navigation_url: "https://navigate-to.com"
            steps:
              - type: wait
                selector: ".element"
                timeout: 5000
              - type: click
                selector: ".button"
          
          # Additional checkpoints created normally
          - name: "Second Checkpoint"
            steps:
              - type: fill
                selector: "#input"
                value: "test value"
```

## ⚠️ Important Behaviors

1. **Auto-Created Journey**: When creating a goal, Virtuoso automatically creates a journey. The `create-structure` command **renames and reuses** this journey instead of creating a duplicate.

2. **First Navigation Step**: The first checkpoint always has a navigation step that's shared across the goal. The command **updates** this existing step rather than creating a new one.

3. **Order Matters**: Process resources in this order:
   - Project → Goal → Journey (rename first) → Checkpoints → Steps

## 🔄 Complete Workflow Example

```bash
# 1. Create structure file
cat > test-suite.yaml << EOF
project:
  name: "E2E Tests"
goals:
  - name: "User Flow"
    url: "https://app.com"
    journeys:
      - name: "Login Journey"
        checkpoints:
          - name: "Go to Login"
            navigation_url: "https://app.com/login"
            steps:
              - type: wait
                selector: ".login-form"
                timeout: 3000
          - name: "Submit Login"
            steps:
              - type: fill
                selector: "#username"
                value: "testuser"
              - type: click
                selector: ".submit"
EOF

# 2. Preview what will be created
./bin/api-cli create-structure --file test-suite.yaml --dry-run

# 3. Create the structure
./bin/api-cli create-structure --file test-suite.yaml --verbose

# 4. List created resources
./bin/api-cli list-projects
./bin/api-cli list-goals <project-id>
./bin/api-cli list-journeys <goal-id>
./bin/api-cli list-checkpoints <journey-id>
```

## 🎯 Tips

- Always use `--dry-run` first to preview changes
- Use `--verbose` to see detailed progress
- Save structure files in version control
- The first journey in each goal reuses the auto-created one
- Navigation URLs in first checkpoints update existing navigation
- Use meaningful names for easy identification later
