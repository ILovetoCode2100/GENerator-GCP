# API CLI Quick Start - Rapid Implementation Guide

## Overview

Build a CLI that creates hierarchical structures (Project → Goals → Journeys → Checkpoints) from a single JSON file, then add steps individually.

## Phase 1: Structure Builder (Batch Creation)

### 1.1 JSON Structure Definition

```json
{
  "project": {
    "name": "Q1 2025 Initiative",
    "description": "Quarterly planning"
  },
  "goals": [
    {
      "name": "Performance Optimization",
      "description": "Improve system performance",
      "journeys": [
        {
          "name": "Database Optimization",
          "checkpoints": [
            {
              "name": "Baseline Metrics",
              "description": "Establish current performance"
            },
            {
              "name": "Index Optimization",
              "description": "Optimize database indexes"
            }
          ]
        },
        {
          "name": "API Response Time",
          "checkpoints": [
            {
              "name": "Current State Analysis",
              "description": "Measure current response times"
            }
          ]
        }
      ]
    },
    {
      "name": "Security Improvements",
      "description": "Enhance security posture",
      "journeys": [
        {
          "name": "Authentication Upgrade",
          "checkpoints": [
            {
              "name": "MFA Implementation",
              "description": "Add multi-factor auth"
            }
          ]
        }
      ]
    }
  ]
}
```

### 1.2 CLI Command

```bash
# One command to build entire structure
api-cli create-structure --file structure.json

# AI-friendly output
{
  "status": "success",
  "created": {
    "project_id": "proj_123",
    "goals": [
      {
        "goal_id": "goal_456",
        "name": "Performance Optimization",
        "journeys": [
          {
            "journey_id": "jrny_789",
            "name": "Database Optimization",
            "checkpoints": [
              {
                "checkpoint_id": "chkp_012",
                "name": "Baseline Metrics"
              }
            ]
          }
        ]
      }
    ]
  }
}
```

## Phase 2: Step Builder (Individual Commands)

### 2.1 Step Commands

```bash
# Add step to checkpoint
api-cli add-step <checkpoint-id> <step-type> --name "Step Name" [options]

# Examples:
api-cli add-step chkp_012 navigate --url "https://example.com"
api-cli add-step chkp_012 click --selector "#submit-button"
api-cli add-step chkp_012 fill --selector "#username" --value "testuser"
api-cli add-step chkp_012 wait --duration 2000
api-cli add-step chkp_012 assert --selector ".success" --text "Complete"
```

## Minimal Implementation Plan

### Step 1: Core Structure (2-3 hours)

```go
// cmd/create-structure.go
package cmd

import (
    "encoding/json"
    "github.com/spf13/cobra"
)

type Structure struct {
    Project Project `json:"project"`
    Goals   []Goal  `json:"goals"`
}

type Project struct {
    Name        string `json:"name"`
    Description string `json:"description"`
}

type Goal struct {
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Journeys    []Journey `json:"journeys"`
}

type Journey struct {
    Name        string       `json:"name"`
    Checkpoints []Checkpoint `json:"checkpoints"`
}

type Checkpoint struct {
    Name        string `json:"name"`
    Description string `json:"description"`
}

var createStructureCmd = &cobra.Command{
    Use:   "create-structure",
    Short: "Create entire project structure from JSON",
    RunE: func(cmd *cobra.Command, args []string) error {
        // 1. Parse JSON file
        // 2. Create project
        // 3. For each goal: create goal, initial journey, initial checkpoint
        // 4. For each user journey: create journey
        // 5. For each checkpoint: create and attach
        // 6. Output results in AI-friendly format
        return nil
    },
}
```

### Step 2: Step Management (1-2 hours)

```go
// cmd/add-step.go
var stepTypes = map[string]bool{
    "navigate": true,
    "click":    true,
    "fill":     true,
    "wait":     true,
    "assert":   true,
}

var addStepCmd = &cobra.Command{
    Use:   "add-step <checkpoint-id> <step-type>",
    Short: "Add a step to a checkpoint",
    Args:  cobra.ExactArgs(2),
    RunE: func(cmd *cobra.Command, args []string) error {
        checkpointID := args[0]
        stepType := args[1]
        
        // Validate step type
        if !stepTypes[stepType] {
            return fmt.Errorf("invalid step type: %s", stepType)
        }
        
        // Build step based on type and flags
        // Call API to add step
        // Return result
        return nil
    },
}
```

## What I Need From You

To implement this quickly, I need:

### 1. API Endpoints
Please provide the OpenAPI spec or documentation for these endpoints:

```
POST /projects
POST /goals
GET  /goals/{id}/snapshot
POST /journeys
POST /checkpoints
POST /checkpoints/{id}/attach
POST /steps (or whatever adds steps to checkpoints)
```

### 2. Step Types
What step types do you support? Common ones might be:
- navigate (URL)
- click (selector)
- fill/type (selector, value)
- wait (duration)
- assert (selector, expected)
- screenshot
- custom?

### 3. Required Fields
For each resource, what fields are:
- Required vs optional?
- System-generated vs user-provided?
- Any special validation rules?

## Quick Start Prompt

**Please provide:**

1. **API Documentation** - Either:
   - OpenAPI/Swagger spec file
   - Postman collection
   - Or just paste the relevant endpoints

2. **Step Types** - List of supported step types with their parameters

3. **Business Rules** - Any specific rules like:
   - Do goals always need initial journeys?
   - Are checkpoints always attached?
   - Any naming conventions?

4. **Authentication** - How does your API handle auth?
   - Bearer token?
   - API key?
   - OAuth?

Once you provide this, I can generate the exact Go code you need to get running immediately.

## Example Response Format I Need

```yaml
# Example endpoint documentation
create_project:
  method: POST
  path: /api/v1/projects
  body:
    name: string (required)
    description: string (optional)
    org_id: string (required)
  response:
    id: string
    name: string
    created_at: timestamp

create_goal:
  method: POST
  path: /api/v1/goals
  body:
    project_id: string (required)
    name: string (required)
    description: string (optional)
  response:
    id: string
    name: string
    snapshot_id: string  # Is this returned immediately?
    
# ... continue for other endpoints
```

This will let me build you a working CLI in the next response!
