# Implementation Guide

## Converting Your Postman Collection to CLI

This guide walks through implementing the API CLI generator with your actual Postman collection.

## Step 1: Export OpenAPI from Postman

### Option A: Using Postman UI
1. Open your collection in Postman
2. Click the three dots menu → "Export"
3. Choose "OpenAPI 3.0" format
4. Save as `specs/api.yaml`

### Option B: Using Postman API
```bash
# Get your collection ID and API key from Postman
COLLECTION_ID="your-collection-id"
POSTMAN_API_KEY="your-postman-api-key"

# Convert to OpenAPI
curl -X POST https://api.getpostman.com/collections/$COLLECTION_ID/transformations \
  -H "X-Api-Key: $POSTMAN_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"format": "openapi3"}' \
  > specs/api.yaml
```

## Step 2: Validate and Clean the Spec

```bash
# Validate the exported spec
./scripts/validate-spec.sh

# Common issues to fix:
# - Missing operationId: Add unique IDs to each operation
# - Server URLs: Ensure correct base URLs
# - Security schemes: Add authentication definitions
```

## Step 3: Generate Client Code

```bash
# Run code generation
./scripts/generate.sh

# This creates:
# - src/api/types.gen.go (request/response types)
# - src/api/client.gen.go (API client)
# - src/api/spec.gen.go (embedded spec)
```

## Step 4: Create CLI Commands

For each API operation, create a corresponding CLI command:

```go
// src/cmd/commands.go
package main

import (
    "github.com/spf13/cobra"
    "github.com/marklovelady/api-cli-generator/src/api"
    "github.com/marklovelady/api-cli-generator/src/client"
)

func init() {
    // Add command for each operation
    rootCmd.AddCommand(newUsersCommand())
}

func newUsersCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "users",
        Short: "Manage users",
    }
    
    cmd.AddCommand(
        newListUsersCommand(),
        newGetUserCommand(),
        newCreateUserCommand(),
    )
    
    return cmd
}
```

## Step 5: Implement Command Handlers

```go
// src/cmd/users.go
func newListUsersCommand() *cobra.Command {
    var limit, offset int
    
    cmd := &cobra.Command{
        Use:   "list",
        Short: "List users",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Create client
            cfg := client.Config{
                BaseURL: viper.GetString("base_url"),
                APIKey:  viper.GetString("api_key"),
            }
            c, err := client.NewEnhancedClient(cfg)
            if err != nil {
                return err
            }
            
            // Call API
            resp, err := c.GetAPIClient().ListUsers(cmd.Context(), &api.ListUsersParams{
                Limit:  &limit,
                Offset: &offset,
            })
            if err != nil {
                return c.HandleError(err, "list users")
            }
            
            // Output results
            return outputResults(resp)
        },
    }
    
    cmd.Flags().IntVar(&limit, "limit", 10, "Maximum results")
    cmd.Flags().IntVar(&offset, "offset", 0, "Skip results")
    
    return cmd
}
```

## Step 6: Handle Request Bodies

For operations with request bodies, use templates:

```go
// src/cmd/create_user.go
func newCreateUserCommand() *cobra.Command {
    var name, email, role string
    
    cmd := &cobra.Command{
        Use:   "create",
        Short: "Create a new user",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Build request
            req := api.CreateUserRequest{
                Name:  name,
                Email: email,
            }
            if role != "" {
                req.Role = &role
            }
            
            // Call API
            resp, err := c.GetAPIClient().CreateUser(cmd.Context(), req)
            // ... handle response
        },
    }
    
    cmd.Flags().StringVar(&name, "name", "", "User name (required)")
    cmd.Flags().StringVar(&email, "email", "", "User email (required)")
    cmd.Flags().StringVar(&role, "role", "", "User role")
    
    cmd.MarkFlagRequired("name")
    cmd.MarkFlagRequired("email")
    
    return cmd
}
```

## Step 7: Add Authentication

Configure authentication based on your API:

```go
// For Bearer token
httpClient.SetAuthToken(apiKey)

// For API key header
httpClient.SetHeader("X-API-Key", apiKey)

// For basic auth
httpClient.SetBasicAuth(username, password)
```

## Step 8: Test the CLI

```bash
# Build
make build

# Test commands
./bin/api-cli users list
./bin/api-cli users get USER_ID
./bin/api-cli users create --name "Test" --email test@example.com
```

## Step 9: Package for Distribution

### Binary Distribution
```bash
# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o dist/api-cli-linux-amd64
GOOS=darwin GOARCH=amd64 go build -o dist/api-cli-darwin-amd64
GOOS=windows GOARCH=amd64 go build -o dist/api-cli-windows-amd64.exe
```

### Docker Distribution
```bash
# Build and push
docker build -t your-registry/api-cli:latest .
docker push your-registry/api-cli:latest
```

### Web Service Wrapper (Optional)
```go
// src/server/main.go
func main() {
    r := gin.Default()
    
    r.POST("/execute", func(c *gin.Context) {
        var req struct {
            Command string   `json:"command"`
            Args    []string `json:"args"`
        }
        
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }
        
        // Execute CLI command internally
        output, err := executeCommand(req.Command, req.Args)
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }
        
        c.JSON(200, gin.H{"output": output})
    })
    
    r.Run(":8080")
}
```

## Security Considerations

1. **Input Validation**: All inputs are validated through Cobra flags
2. **Template Safety**: Only pre-defined templates, no user templates
3. **API Key Storage**: Use environment variables or secure config
4. **Network Security**: HTTPS only, certificate validation
5. **Container Security**: Run as non-root user

## Next Steps

1. Replace `specs/api.yaml` with your actual OpenAPI spec
2. Run code generation
3. Implement command handlers for your operations
4. Add any custom authentication logic
5. Test thoroughly with your API
6. Deploy as needed (binary/container/service)
