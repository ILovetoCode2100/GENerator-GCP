# 🧪 Virtuoso CLI Testing Guide

## Overview
This guide covers comprehensive testing of the enhanced Virtuoso CLI with all its new features.

## ✅ What to Check After Testing

### 1. **Configuration Validation**
- ✅ Config file is validated
- ✅ API connectivity is confirmed
- ✅ Authentication works
- ✅ Organization ID is shown

### 2. **Batch Structure Creation**
- ✅ Projects are created with correct names
- ✅ Goals are created with URLs
- ✅ Journeys are attached to goals
- ✅ Checkpoints are created and attached to journeys
- ✅ Steps are added to checkpoints
- ✅ All IDs are returned properly

### 3. **List Commands**
- ✅ Commands execute without errors
- ✅ Output formats work (JSON, YAML, Human)
- Note: List APIs may return empty results - this is a known issue

### 4. **Output Formats**
- ✅ Human-readable (default) - Tables and friendly messages
- ✅ JSON - Structured data for automation
- ✅ YAML - Clean YAML output
- ✅ AI - Detailed explanations with next steps

## 🚀 Quick Start Testing

### Run Automated Tests
```bash
cd /Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator
./test/test-enhanced-cli.sh
```

### Run Quick Manual Test
```bash
./test/quick-test.sh
```

## 📋 Manual Testing Steps

### 1. Validate Your Configuration
```bash
./bin/api-cli validate-config --config ./config/virtuoso-config.yaml
```

**Expected Output:**
- Green checkmarks for each validation step
- Shows your base URL and organization ID
- Confirms API connection and authentication

### 2. Test Batch Creation (The Killer Feature\!)

#### Preview Mode (Dry Run)
```bash
./bin/api-cli create-structure --file examples/test-structure.yaml --config ./config/virtuoso-config.yaml --dry-run
```

**Expected Output:**
- Shows hierarchical preview of what will be created
- Lists total counts for goals, journeys, checkpoints, and steps
- No actual resources are created

#### Actual Creation
```bash
./bin/api-cli create-structure --file examples/test-small.yaml --config ./config/virtuoso-config.yaml
```

**Expected Output:**
- Progress messages for each resource created
- Green checkmarks as each item succeeds
- Summary with all created IDs

### 3. Test Individual Commands

#### Create Resources
```bash
# Create a project
./bin/api-cli create-project "My Test Project" --config ./config/virtuoso-config.yaml

# Create a goal (need project ID from above)
./bin/api-cli create-goal PROJECT_ID "My Goal" --url "https://example.com" --config ./config/virtuoso-config.yaml

# Create a journey (need goal ID and snapshot ID from above)
./bin/api-cli create-journey GOAL_ID SNAPSHOT_ID "My Journey" --config ./config/virtuoso-config.yaml
```

#### List Resources
```bash
# List all projects
./bin/api-cli list-projects --config ./config/virtuoso-config.yaml

# List goals in a project
./bin/api-cli list-goals PROJECT_ID --config ./config/virtuoso-config.yaml

# List journeys in a goal
./bin/api-cli list-journeys GOAL_ID SNAPSHOT_ID --config ./config/virtuoso-config.yaml
```

### 4. Test Output Formats

```bash
# JSON output
./bin/api-cli create-project "JSON Test" --config ./config/virtuoso-config.yaml -o json

# YAML output
./bin/api-cli validate-config --config ./config/virtuoso-config.yaml -o yaml

# AI-friendly output
./bin/api-cli list-projects --config ./config/virtuoso-config.yaml -o ai
```

## 🎯 Creating Your Own Test Structure

### 1. Create a YAML file
```yaml
project:
  name: "My E2E Tests"
  description: "Complete test suite"
goals:
  - name: "Login Tests"
    url: "https://myapp.com/login"
    journeys:
      - name: "Valid Login"
        checkpoints:
          - name: "Navigate to Login"
            steps:
              - type: navigate
                url: "https://myapp.com/login"
              - type: wait
                selector: "form.login"
                timeout: 5000
          - name: "Click Login"
            steps:
              - type: click
                selector: "button.submit"
```

### 2. Run it
```bash
./bin/api-cli create-structure --file my-tests.yaml --config ./config/virtuoso-config.yaml
```

## 🔍 What Success Looks Like

### Configuration Validation Success
```
✅ Configuration is valid\!
📋 Configuration:
   Base URL: https://api-app2.virtuoso.qa/api
   Organization ID: 2242
✅ API connection successful
✅ Authentication valid
```

### Structure Creation Success
```
Creating project: E2E Test Suite...
  ✓ Created project ID: 9071
Creating goal: User Authentication...
  ✓ Created goal ID: 13788
  Creating journey: Login Flow...
    ✓ Created journey ID: 608061
    Creating checkpoint: Navigate to Login...
      ✓ Created checkpoint ID: 1678343
      ✓ Added 2 steps
✅ Created test structure successfully\!
```

### JSON Output Success
```json
{
  "project_id": 9071,
  "goals": [{
    "id": 13788,
    "name": "User Authentication",
    "snapshot_id": "43817",
    "journeys": [{
      "id": 608061,
      "name": "Login Flow",
      "checkpoints": [...]
    }]
  }],
  "total_steps": 14
}
```

## 🐛 Troubleshooting

### Common Issues

1. **"Project already exists" error**
   - Add a timestamp to make names unique
   - Or use a different project name

2. **401 Unauthorized**
   - Check your API token in config/virtuoso-config.yaml
   - Run validate-config to verify authentication

3. **Empty list results**
   - This is a known issue with the list APIs
   - Resources are still being created successfully

4. **Fill step errors**
   - Fill steps have been removed due to API issues
   - Use click steps instead

## 📊 Test Coverage

| Feature | Command | Status |
|---------|---------|--------|
| Config Validation | `validate-config` | ✅ Working |
| Project Creation | `create-project` | ✅ Working |
| Goal Creation | `create-goal` | ✅ Working |
| Journey Creation | `create-journey` | ✅ Working |
| Checkpoint Creation | `create-checkpoint` | ✅ Working |
| Add Steps | `add-step navigate/click/wait` | ✅ Working |
| List Projects | `list-projects` | ⚠️ Returns empty |
| List Goals | `list-goals` | ⚠️ Returns empty |
| List Journeys | `list-journeys` | ⚠️ Returns empty |
| Batch Creation | `create-structure` | ✅ Working |
| Dry Run | `--dry-run` | ✅ Working |
| JSON Output | `-o json` | ✅ Working |
| YAML Output | `-o yaml` | ✅ Working |
| AI Output | `-o ai` | ✅ Working |

## 🎉 Success Criteria

Your CLI is working correctly if:
1. ✅ Config validation passes
2. ✅ You can create projects, goals, journeys, and checkpoints
3. ✅ Batch structure creation works from YAML files
4. ✅ Dry run shows correct preview
5. ✅ All output formats produce expected results
6. ✅ Created resources appear in Virtuoso UI

## 📚 Next Steps

After successful testing:
1. Check created resources in Virtuoso UI: https://app2.virtuoso.qa
2. Run the test journeys you created
3. Create more complex test structures
4. Integrate the CLI into your CI/CD pipeline

---
*Last updated: December 2024*
EOF < /dev/null