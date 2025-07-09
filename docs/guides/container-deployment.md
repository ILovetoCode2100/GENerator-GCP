# Virtuoso CLI - Container Deployment Guide

## Why Containerize?

✅ **No Go installation required** - Users don't need Go
✅ **No dependencies** - Everything bundled
✅ **Version control** - Pin specific CLI versions
✅ **Easy distribution** - Just pull and run
✅ **CI/CD friendly** - Perfect for pipelines
✅ **Cross-platform** - Works on any OS with Docker

## Quick Start

### 1. Build Container
```bash
cd /Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator
docker build -t virtuoso-cli:latest .
```

### 2. Run Commands
```bash
# Using config file
docker run -v $(pwd)/config:/config virtuoso-cli:latest \
  create-project "Test Project" --config /config/virtuoso-config.yaml

# Using environment variables
docker run \
  -e VIRTUOSO_API_BASE_URL="https://api-app2.virtuoso.qa/api" \
  -e VIRTUOSO_API_TOKEN="f7a55516-5cc4-4529-b2ae-8e106a7d164e" \
  -e VIRTUOSO_ORGANIZATION_ID="2242" \
  virtuoso-cli:latest create-project "Test Project"
```

### 3. Create Structure from YAML
```bash
# Mount both config and examples
docker run \
  -v $(pwd)/config:/config \
  -v $(pwd)/examples:/examples \
  virtuoso-cli:latest \
  create-structure --file /examples/test-structure.yaml --config /config/virtuoso-config.yaml
```

## Enhanced Dockerfile

Create this enhanced Dockerfile:

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o virtuoso-cli ./src/cmd

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates bash

# Create non-root user
RUN adduser -D -g '' virtuoso

# Create config directory
RUN mkdir -p /config /examples /workspace && \
    chown -R virtuoso:virtuoso /config /examples /workspace

# Copy binary from builder
COPY --from=builder /app/virtuoso-cli /usr/local/bin/

# Copy example files
COPY --from=builder /app/examples /examples/

# Switch to non-root user
USER virtuoso

# Set working directory
WORKDIR /workspace

# Set entrypoint
ENTRYPOINT ["virtuoso-cli"]

# Default command (show help)
CMD ["--help"]
```

## Docker Compose for Teams

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  virtuoso-cli:
    image: virtuoso-cli:latest
    build: .
    volumes:
      - ./config:/config:ro
      - ./tests:/workspace
    environment:
      - VIRTUOSO_API_BASE_URL=${VIRTUOSO_API_BASE_URL:-https://api-app2.virtuoso.qa/api}
      - VIRTUOSO_API_TOKEN=${VIRTUOSO_API_TOKEN}
      - VIRTUOSO_ORGANIZATION_ID=${VIRTUOSO_ORGANIZATION_ID:-2242}
    working_dir: /workspace
```

## Usage Examples for Other Projects

### 1. Shell Script Wrapper
Create `virtuoso-cli.sh`:
```bash
#!/bin/bash
docker run --rm \
  -v $(pwd):/workspace \
  -v ~/.virtuoso:/config:ro \
  virtuoso-cli:latest "$@"
```

### 2. CI/CD Integration
```yaml
# .github/workflows/create-tests.yml
name: Create Virtuoso Tests
on:
  push:
    branches: [main]
jobs:
  create-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Create Virtuoso Tests
        run: |
          docker run --rm \
            -v ${{ github.workspace }}/tests:/workspace \
            -e VIRTUOSO_API_TOKEN=${{ secrets.VIRTUOSO_TOKEN }} \
            -e VIRTUOSO_ORGANIZATION_ID=${{ secrets.VIRTUOSO_ORG }} \
            virtuoso-cli:latest \
            create-structure --file /workspace/e2e-tests.yaml
```

### 3. Make it Available to Teams

#### Option A: Docker Hub
```bash
# Tag and push to Docker Hub
docker tag virtuoso-cli:latest yourorg/virtuoso-cli:latest
docker push yourorg/virtuoso-cli:latest

# Teams can then use:
docker run yourorg/virtuoso-cli:latest --help
```

#### Option B: OrbStack Registry
```bash
# Use OrbStack's local registry
docker tag virtuoso-cli:latest orbstack.local/virtuoso-cli:latest

# Other devs on same network can pull
docker pull orbstack.local/virtuoso-cli:latest
```

## Best Practices

### 1. Version Tags
```bash
# Tag with version
docker build -t virtuoso-cli:1.0.0 -t virtuoso-cli:latest .

# Users can pin versions
docker run virtuoso-cli:1.0.0 --help
```

### 2. Config Management
```bash
# Create standard config location
mkdir -p ~/.virtuoso
cp config/virtuoso-config.yaml ~/.virtuoso/

# Use in container
alias virtuoso='docker run --rm -v ~/.virtuoso:/config:ro -v $(pwd):/workspace virtuoso-cli:latest --config /config/virtuoso-config.yaml'
```

### 3. Example Project Structure
```
my-project/
├── tests/
│   ├── e2e-tests.yaml
│   ├── smoke-tests.yaml
│   └── regression-tests.yaml
├── .virtuoso-config.yaml
└── Makefile
```

With Makefile:
```makefile
test-create:
	docker run --rm \
		-v $(PWD):/workspace \
		virtuoso-cli:latest \
		create-structure --file /workspace/tests/e2e-tests.yaml

test-validate:
	docker run --rm \
		-v $(PWD)/.virtuoso-config.yaml:/config.yaml:ro \
		virtuoso-cli:latest \
		validate-config --config /config.yaml
```

## Quick OrbStack Deployment

```bash
# Build with OrbStack
orb build -t virtuoso-cli .

# Run with OrbStack
orb run -v $(pwd)/config:/config virtuoso-cli create-project "Test"

# Create a Linux VM with the CLI
orb create ubuntu virtuoso-vm
orb shell virtuoso-vm
docker run virtuoso-cli --help
```

## Summary

Containerizing makes your CLI:
- **Portable**: No installation required
- **Versioned**: Teams can use specific versions
- **Isolated**: No dependency conflicts
- **CI/CD Ready**: Perfect for automation
- **Shareable**: Easy distribution to teams

Would you like me to create the enhanced Dockerfile and wrapper scripts?
