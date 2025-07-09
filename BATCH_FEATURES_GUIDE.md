# Virtuoso CLI - Batch Features Quick Guide

## 🚀 Quick Start

```bash
# Build the CLI
make build

# Set configuration
export API_CLI_CONFIG="./config/virtuoso-config.yaml"

# Test connection
./bin/api-cli list-projects
```

## 📋 New Commands

### 1. List Journeys (Fixed)
```bash
# Now correctly shows auto-created journeys
./bin/api-cli list-journeys <goal-id> <snapshot-id>
```

### 2. Update Journey Title
```bash
# Renames journey display title (not system name)
./bin/api-cli update-journey <journey-id> --name "New Title"
```

### 3. Update Navigation URL
```bash
# Get step details first
STEP_ID=$(./bin/api-cli get-steps <checkpoint-id> -o json | jq -r '.[0].id')
CANONICAL_ID=$(./bin/api-cli get-step $STEP_ID -o json | jq -r '.canonicalId')

# Update the URL
./bin/api-cli update-navigation $STEP_ID $CANONICAL_ID --url "https://new-url.com"
```

### 4. Create Structure (Batch)
```bash
# Preview mode
./bin/api-cli create-structure --file structure.yaml --dry-run

# Create with verbose output  
./bin/api-cli create-structure --file structure.yaml --verbose

# Use existing project
./bin/api-cli create-structure --file structure.yaml --project-id 1234
```

## 📄 Structure File Format

```yaml
project:
  name: "My Test Project"    # Creates new project
  # id: 1234                # Or use existing
  
goals:
  - name: "Homepage Tests"
    url: "https://example.com"
    journeys:
      # First journey uses auto-created one
      - name: "Main Flow"    
        checkpoints:
          # First checkpoint updates existing navigation
          - name: "Landing"
            navigation_url: "https://example.com"
            steps:
              - type: wait
                selector: "body"
                timeout: 2000
              - type: click
                selector: ".button"
                
          # Additional checkpoints
          - name: "Submit Form"
            steps:
              - type: fill
                selector: "#email"
                value: "test@example.com"
                
      # Second journey is created new
      - name: "Alternative Flow"
        checkpoints:
          - name: "Login"
            navigation_url: "https://example.com/login"
```

## 🎯 How It Works

1. **Goal Creation**: Automatically creates first journey "Suite 1"
2. **First Journey**: Renames auto-created journey to your title
3. **Additional Journeys**: Creates new journeys as needed
4. **Navigation**: Updates existing navigation checkpoint URL
5. **Steps**: Adds all defined test steps

## 💡 Tips

- Goals always create a "Suite 1" journey automatically
- Use `--dry-run` to preview changes first
- Use `--verbose` to see detailed progress
- Journey titles are displayed in UI, names are system-internal
- First checkpoint in each journey is always navigation

## 🔧 Troubleshooting

**Can't see journeys?**
- Make sure you're using the correct snapshot ID
- Check that the goal exists: `./bin/api-cli list-goals <project-id>`

**Journey not renamed?**
- API uses `title` field for display name
- System `name` field stays as "Suite 1", "Suite 2", etc.

**Navigation not updating?**
- Need both step ID and canonical ID
- Use `get-step` command to find canonical ID first

## 📊 Example Output

```
Creating project: E-Commerce Tests...
  ✓ Created project ID: 9089

Creating goal: Homepage Tests...
  ✓ Created goal ID: 13811
  Using auto-created journey: Suite 1 (ID: 608100)
    Renaming to: Main Flow...
      ✓ Journey renamed successfully
    Updating navigation URL...
      ✓ Navigation updated
    Adding test steps...
      ✓ Added 2 steps
      
✅ Structure created successfully!
```