# API CLI Generator - Quick Reference

## 🚀 Quick Start

```bash
# Navigate to project
cd /Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator

# Add your OpenAPI spec
cp /path/to/your/openapi.yaml specs/api.yaml

# Generate and build
make generate
make build

# Run CLI
./bin/api-cli --help
```

## 📋 Common Tasks

### Validate OpenAPI Spec
```bash
./scripts/validate-spec.sh
```

### Generate Client Code
```bash
./scripts/generate.sh
# or
make generate
```

### Build CLI
```bash
make build
```

### Run Tests
```bash
make test
```

### Build Docker Image
```bash
make docker-build
```

## 🔧 Configuration

### Via Environment Variables
```bash
export API_CLI_BASE_URL=https://api.example.com/v1
export API_CLI_API_KEY=your-key-here
export API_CLI_OUTPUT=table
```

### Via Config File
```bash
cat > ~/.api-cli.yaml << EOF
base_url: https://api.example.com/v1
api_key: your-key-here
output: json
verbose: false
EOF
```

## 📝 Generated Files

After running `make generate`:
- `src/api/types.gen.go` - Request/response types
- `src/api/client.gen.go` - API client methods
- `src/api/spec.gen.go` - Embedded OpenAPI spec

## 🛠️ Development Workflow

1. Edit OpenAPI spec
2. Run `make generate`
3. Implement CLI commands in `src/cmd/`
4. Run `make build`
5. Test with `./bin/api-cli`

## 📦 Distribution

### Local Install
```bash
make install  # Copies to /usr/local/bin
```

### Docker
```bash
docker run --rm your-registry/api-cli:latest --help
```

### Cross-Platform Build
```bash
# Linux
GOOS=linux GOARCH=amd64 make build

# macOS
GOOS=darwin GOARCH=amd64 make build

# Windows
GOOS=windows GOARCH=amd64 make build
```

## ❓ Troubleshooting

### Missing operationId in OpenAPI
Add unique operationId to each operation in your spec

### Authentication Issues
Check API_CLI_API_KEY environment variable or config file

### Generation Errors
Validate spec first: `./scripts/validate-spec.sh`

## 📚 More Info

- Full documentation: `/docs/implementation-guide.md`
- Usage examples: `/docs/usage.md`
- Example scripts: `/examples/example-calls.sh`
