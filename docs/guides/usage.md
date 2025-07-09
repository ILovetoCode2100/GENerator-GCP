# API CLI Usage Guide

## Installation

### From Source
```bash
# Clone and build
git clone <repository>
cd api-cli-generator
make build

# Install globally (optional)
make install
```

### Using Docker
```bash
# Build image
make docker-build

# Run with Docker
docker run --rm api-cli:latest <command>
```

## Configuration

The CLI can be configured via:
1. Command-line flags
2. Environment variables (prefix: `API_CLI_`)
3. Configuration file (`~/.api-cli.yaml`)

### Configuration File Example
```yaml
base_url: https://api.example.com/v1
api_key: your-api-key-here
timeout: 30
output: json
verbose: false
```

## Authentication

### API Key
```bash
# Via flag
api-cli --api-key YOUR_KEY <command>

# Via environment variable
export API_CLI_API_KEY=YOUR_KEY
api-cli <command>
```

## Commands

### List Users
```bash
# Basic usage
api-cli users list

# With pagination
api-cli users list --limit 20 --offset 40

# With output formatting
api-cli users list -o table
api-cli users list -o yaml
```

### Get User
```bash
# Get specific user
api-cli users get USER_ID

# With verbose output
api-cli users get USER_ID -v
```

### Create User
```bash
# Basic creation
api-cli users create --name "John Doe" --email john@example.com

# With role
api-cli users create \
  --name "Admin User" \
  --email admin@example.com \
  --role admin
```

## Output Formats

### JSON (default)
```bash
api-cli users list -o json
```

### YAML
```bash
api-cli users list -o yaml
```

### Table
```bash
api-cli users list -o table
```

## Advanced Usage

### Using Different Servers
```bash
# Production
api-cli --base-url https://api.example.com/v1 users list

# Staging
api-cli --base-url https://staging-api.example.com/v1 users list
```

### Debugging
```bash
# Verbose output
api-cli -v users list

# Debug HTTP requests
export API_CLI_DEBUG=true
api-cli users list
```

### Timeout Control
```bash
# Set 60 second timeout
api-cli --timeout 60 users create --name "Test" --email test@example.com
```

## Error Handling

The CLI provides clear error messages:
- HTTP errors show status codes and messages
- Validation errors indicate which fields failed
- Network errors include retry information

## Best Practices

1. **Use configuration files** for common settings
2. **Set API keys via environment** variables for security
3. **Use verbose mode** when debugging issues
4. **Choose appropriate timeouts** for your operations
5. **Use output formats** that suit your workflow (JSON for scripts, table for humans)
