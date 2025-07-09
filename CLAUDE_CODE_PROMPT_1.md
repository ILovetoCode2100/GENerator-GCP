# Claude Code Prompt: Build Virtuoso CLI - Create Project Command

## Context
You are building a CLI tool for the Virtuoso API. The project is located at `/Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator/`. 

The Virtuoso API credentials and configuration are already set up in:
- Config file: `config/virtuoso-config.yaml`
- Go config package: `pkg/config/virtuoso.go`
- Base client structure: `pkg/virtuoso/client.go`

## Task
Implement the first CLI command: `create-project`

## API Details
From the Postman collection (`postman_reference/collectionv1.json`):

**Endpoint**: POST {{baseURL}}/projects
**Headers**:
- Authorization: Bearer {{token}}
- X-Virtuoso-Client-ID: {{X-Virtuoso-Client-ID}}
- X-Virtuoso-Client-Name: {{X-Virtuoso-Client-Name}}
- Content-Type: application/json

**Request Body**:
```json
{
    "name": "Project Name",
    "organizationId": 2242
}
```

**Response** (201 Created):
```json
{
    "success": true,
    "item": {
        "id": 12345,
        "name": "Project Name",
        // other fields...
    }
}
```

## Implementation Requirements

1. **Update the Virtuoso client** (`pkg/virtuoso/client.go`):
   - The `CreateProject` method is already stubbed, ensure it works correctly
   - Response should match the actual API response structure

2. **Create a new command file** `src/cmd/create_project.go`:
   - Use Cobra for command structure
   - Command: `api-cli create-project --name "My Project"`
   - Use the organization ID from config (2242)
   - Support output formats: human (default), json, yaml, ai

3. **Update main.go** to register the new command properly

4. **Output formats**:
   - Human: `✅ Created project "My Project" (ID: 12345)`
   - JSON: Full response
   - AI: `{"status":"success","operation":"create_project","project_id":12345,"project_name":"My Project"}`

5. **Error handling**:
   - Check for duplicate names
   - Handle API errors gracefully
   - Show helpful error messages

6. **Testing**:
   - Create a simple test script to verify the command works
   - Test with: `./bin/api-cli create-project --name "Test Project"`

## File Structure
```
src/
├── cmd/
│   ├── main.go          # Update to add new command
│   └── create_project.go # New file
pkg/
└── virtuoso/
    └── client.go        # Update CreateProject method
```

## Example Usage
```bash
# Create project with default output
./bin/api-cli create-project --name "Q1 Testing Initiative"

# JSON output
./bin/api-cli create-project --name "Q1 Testing Initiative" -o json

# AI-friendly output
./bin/api-cli create-project --name "Q1 Testing Initiative" -o ai
```

Start by implementing the create-project command. Once this works, we'll add the remaining commands following the same pattern.
