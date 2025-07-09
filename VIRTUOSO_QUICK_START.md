# Virtuoso API CLI - Quick Start Guide

## Configuration Setup

Your Virtuoso API credentials are configured in `/config/virtuoso-config.yaml`:

```yaml
api:
  base_url: https://api-app2.virtuoso.qa/api
  auth_token: f7a55516-5cc4-4529-b2ae-8e106a7d164e
  
organization:
  id: "2242"
  
headers:
  X-Virtuoso-Client-ID: api-cli-generator
  X-Virtuoso-Client-Name: api-cli-generator
```

## Quick Implementation

### Step 1: Install Dependencies

```bash
cd /Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator
go mod tidy
```

### Step 2: Test Basic API Connection

Create a test file to verify the connection:

```go
// test/api_test.go
package main

import (
    "fmt"
    "log"
    "github.com/marklovelady/api-cli-generator/pkg/config"
    "github.com/marklovelady/api-cli-generator/pkg/virtuoso"
)

func main() {
    // Load config
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatal(err)
    }
    
    // Create client
    client := virtuoso.NewClient(cfg)
    
    // Test creating a project
    project, err := client.CreateProject("Test Project", "Testing API connection")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Created project: %s (ID: %s)\n", project.Name, project.ID)
}
```

## Phase 1: Structure Builder

### JSON Structure for Batch Creation

```json
{
  "project": {
    "name": "Q1 2025 Testing Initiative",
    "description": "Quarterly test automation"
  },
  "goals": [
    {
      "name": "Login Flow Testing",
      "description": "Test all login scenarios",
      "journeys": [
        {
          "name": "Happy Path Login",
          "checkpoints": [
            {
              "name": "Navigate to Login",
              "description": "Open login page"
            },
            {
              "name": "Submit Credentials",
              "description": "Enter and submit valid credentials"
            }
          ]
        }
      ]
    }
  ]
}
```

### CLI Command Implementation

```bash
# Build the CLI
make build

# Create entire structure
./bin/api-cli create-structure --file structure.json

# Output
{
  "status": "success",
  "project_id": "proj_123",
  "goals": [
    {
      "goal_id": "goal_456",
      "name": "Login Flow Testing",
      "initial_journey_id": "jrny_001",
      "initial_checkpoint_id": "chkp_001",
      "journeys": [...]
    }
  ]
}
```

## Phase 2: Step Commands

### Available Step Types

```bash
# Navigate to URL
./bin/api-cli add-step chkp_123 navigate --url "https://example.com" --name "Go to homepage"

# Click element
./bin/api-cli add-step chkp_123 click --selector "#login-button" --name "Click login"

# Fill input
./bin/api-cli add-step chkp_123 fill --selector "#username" --value "testuser" --name "Enter username"

# Wait
./bin/api-cli add-step chkp_123 wait --duration 2000 --name "Wait for page load"

# Assert text
./bin/api-cli add-step chkp_123 assert --selector ".welcome" --text "Welcome" --name "Verify welcome message"

# Take screenshot
./bin/api-cli add-step chkp_123 screenshot --name "Capture state"
```

## Complete Example Workflow

### 1. Create Project Structure

```bash
# structure.json
{
  "project": {
    "name": "E-commerce Test Suite",
    "description": "Full e-commerce flow testing"
  },
  "goals": [
    {
      "name": "Purchase Flow",
      "description": "Test complete purchase workflow",
      "journeys": [
        {
          "name": "Guest Checkout",
          "checkpoints": [
            {
              "name": "Product Selection",
              "description": "Browse and select product"
            },
            {
              "name": "Checkout Process",
              "description": "Complete checkout"
            }
          ]
        }
      ]
    }
  ]
}

# Create structure
./bin/api-cli create-structure --file structure.json
```

### 2. Add Steps to Checkpoints

```bash
# Add steps to Product Selection checkpoint
CHECKPOINT_ID="chkp_123"  # From previous output

./bin/api-cli add-step $CHECKPOINT_ID navigate --url "https://shop.example.com" --name "Go to shop"
./bin/api-cli add-step $CHECKPOINT_ID click --selector ".product-card:first" --name "Select first product"
./bin/api-cli add-step $CHECKPOINT_ID click --selector "#add-to-cart" --name "Add to cart"
./bin/api-cli add-step $CHECKPOINT_ID assert --selector ".cart-count" --text "1" --name "Verify cart updated"
```

## AI-Friendly Usage

### For Structure Creation
```bash
# AI can generate the JSON structure and run:
echo '{...json structure...}' > project.json
./bin/api-cli create-structure --file project.json --output json
```

### For Adding Steps
```bash
# AI can run sequences like:
./bin/api-cli add-step <checkpoint-id> navigate --url "<url>" --name "<description>"
./bin/api-cli add-step <checkpoint-id> click --selector "<css-selector>" --name "<description>"
./bin/api-cli add-step <checkpoint-id> assert --selector "<css-selector>" --text "<expected>" --name "<description>"
```

## Environment Variables (Alternative to Config File)

```bash
export VIRTUOSO_API_BASE_URL="https://api-app2.virtuoso.qa/api"
export VIRTUOSO_API_TOKEN="f7a55516-5cc4-4529-b2ae-8e106a7d164e"
export VIRTUOSO_ORGANIZATION_ID="2242"
export VIRTUOSO_HEADERS_X_VIRTUOSO_CLIENT_ID="api-cli-generator"
export VIRTUOSO_HEADERS_X_VIRTUOSO_CLIENT_NAME="api-cli-generator"
```

## Next Steps

1. **Provide API Endpoint Details**: I need the exact API paths and request/response formats for:
   - Projects endpoint
   - Goals endpoint  
   - Journeys endpoint
   - Checkpoints endpoint
   - Steps endpoint
   - Attach checkpoint endpoint (if separate)

2. **Clarify Business Rules**:
   - Does creating a goal automatically create an initial journey?
   - Is checkpoint attachment a separate API call?
   - What are the exact step types and their properties?

3. **Run Test**: Once we confirm the endpoints, you can test immediately:
   ```bash
   go run test/api_test.go
   ```

## Quick Debug Commands

```bash
# Test API connection
curl -H "Authorization: Bearer f7a55516-5cc4-4529-b2ae-8e106a7d164e" \
     -H "X-Virtuoso-Client-ID: api-cli-generator" \
     -H "X-Virtuoso-Client-Name: api-cli-generator" \
     https://api-app2.virtuoso.qa/api/projects

# View current config
cat config/virtuoso-config.yaml

# Run with verbose output
./bin/api-cli create-structure --file structure.json --verbose
```
