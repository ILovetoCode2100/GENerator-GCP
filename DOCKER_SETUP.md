# Virtuoso CLI - Docker Setup Guide

This guide explains how to use the Virtuoso CLI in a Docker container for easy distribution and usage across teams.

## 🚀 Quick Start

### Option 1: Using the Wrapper Script (Recommended)
```bash
# Make the script executable
chmod +x scripts/virtuoso

# Run any command
./scripts/virtuoso --help
./scripts/virtuoso create-project "My Test Project"
./scripts/virtuoso list-projects
```

### Option 2: Direct Docker Commands
```bash
# Build the image
docker build -t virtuoso-cli:latest .

# Run commands
docker run --rm -v $(pwd):/workspace virtuoso-cli:latest --help
docker run --rm -v $(pwd):/workspace virtuoso-cli:latest create-project "Test"
```

### Option 3: Using docker-compose
```bash
# Build and run the container
docker-compose up -d

# Execute commands
docker-compose exec virtuoso-cli api-cli --help
docker-compose exec virtuoso-cli api-cli create-project "Test"

# Stop the container
docker-compose down
```

## 📦 Container Features

### Production-Ready Dockerfile
- **Multi-stage build** for minimal final image size
- **Security hardening** with non-root user (apiuser:1000)
- **Proper caching** with separate dependency download layer
- **Health checks** for container monitoring
- **Metadata labels** for image identification
- **Environment variables** for configuration

### Volume Mounts
- `/workspace` - Your current working directory
- `/config` - Configuration files
- `/home/apiuser/.virtuoso` - User configuration persistence

### Environment Variables
- `VIRTUOSO_API_TOKEN` - API authentication token
- `VIRTUOSO_BASE_URL` - API base URL (optional)
- `VIRTUOSO_ORG_ID` - Organization ID (optional)
- `VIRTUOSO_OUTPUT_FORMAT` - Output format (human, json, yaml, ai)

## 🛠️ Available Scripts

### Linux/macOS: `scripts/virtuoso`
```bash
# Show wrapper help
./scripts/virtuoso --wrapper-help

# Build/rebuild the image
./scripts/virtuoso --build

# Open interactive shell
./scripts/virtuoso --shell

# Run any CLI command
./scripts/virtuoso [command] [args...]
```

### Windows: `scripts/virtuoso.ps1`
```powershell
# Show wrapper help
.\scripts\virtuoso.ps1 -WrapperHelp

# Build/rebuild the image
.\scripts\virtuoso.ps1 -Build

# Open interactive shell
.\scripts\virtuoso.ps1 -Shell

# Run any CLI command
.\scripts\virtuoso.ps1 [command] [args...]
```

## 🔧 Configuration

### Method 1: Environment Variables
```bash
export VIRTUOSO_API_TOKEN="your-token-here"
export VIRTUOSO_ORG_ID="your-org-id"
export VIRTUOSO_BASE_URL="https://api-app2.virtuoso.qa/api"

./scripts/virtuoso create-project "Test Project"
```

### Method 2: Configuration File
Create `config/virtuoso-config.yaml`:
```yaml
api:
  base_url: "https://api-app2.virtuoso.qa/api"
  auth_token: "your-token-here"
organization:
  id: "your-org-id"
output:
  default_format: "human"
```

### Method 3: Docker Environment File
Create `.env` file:
```env
VIRTUOSO_API_TOKEN=your-token-here
VIRTUOSO_ORG_ID=your-org-id
VIRTUOSO_BASE_URL=https://api-app2.virtuoso.qa/api
```

Then use with docker-compose:
```bash
docker-compose --env-file .env up -d
```

## 📋 Common Use Cases

### 1. Creating Test Projects
```bash
# Create a simple project
./scripts/virtuoso create-project "E2E Tests"

# Create with custom configuration
VIRTUOSO_OUTPUT_FORMAT=json ./scripts/virtuoso create-project "API Tests"
```

### 2. Batch Operations
```bash
# Create from structure file
./scripts/virtuoso create-structure --file examples/test-structure.yaml

# List all projects
./scripts/virtuoso list-projects --output json
```

### 3. CI/CD Integration
```yaml
# GitHub Actions example
name: Create Virtuoso Tests
on: [push]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    
    - name: Create Test Structure
      run: |
        docker run --rm \
          -v ${{ github.workspace }}:/workspace \
          -e VIRTUOSO_API_TOKEN=${{ secrets.VIRTUOSO_TOKEN }} \
          -e VIRTUOSO_ORG_ID=${{ secrets.VIRTUOSO_ORG_ID }} \
          virtuoso-cli:latest \
          create-structure --file tests/e2e-structure.yaml
```

### 4. Team Distribution
```bash
# Tag and push to registry
docker tag virtuoso-cli:latest myregistry/virtuoso-cli:v1.0.0
docker push myregistry/virtuoso-cli:v1.0.0

# Team members can pull and use
docker pull myregistry/virtuoso-cli:v1.0.0
docker run --rm -v $(pwd):/workspace myregistry/virtuoso-cli:v1.0.0 --help
```

## 🏗️ Build Process

### Standard Build
```bash
# Build production image
docker build -t virtuoso-cli:latest .

# Build with custom tag
docker build -t virtuoso-cli:v1.0.0 .
```

### Development Build
```bash
# Build development image (includes build tools)
docker build --target builder -t virtuoso-cli:dev .

# Or use docker-compose
docker-compose build virtuoso-dev
```

### Build Arguments
```bash
# Build with custom Go version
docker build --build-arg GO_VERSION=1.21 -t virtuoso-cli:latest .
```

## 🔍 Debugging

### Container Debugging
```bash
# Open shell in container
./scripts/virtuoso --shell

# Or directly with docker
docker run -it --rm virtuoso-cli:latest /bin/sh
```

### Volume Debugging
```bash
# Check mounted volumes
docker run --rm -v $(pwd):/workspace virtuoso-cli:latest ls -la /workspace

# Check config directory
docker run --rm -v ~/.virtuoso:/home/apiuser/.virtuoso virtuoso-cli:latest ls -la /home/apiuser/.virtuoso
```

### Network Debugging
```bash
# Test API connectivity
docker run --rm -e VIRTUOSO_API_TOKEN=your-token virtuoso-cli:latest validate-config
```

## 🧪 Testing the Setup

### Basic Functionality Test
```bash
# Test build
docker build -t virtuoso-cli:test .

# Test help
docker run --rm virtuoso-cli:test --help

# Test version
docker run --rm virtuoso-cli:test --version
```

### Integration Test
```bash
# Test with wrapper
./scripts/virtuoso --help

# Test with docker-compose
docker-compose up -d
docker-compose exec virtuoso-cli api-cli --help
docker-compose down
```

## 📊 Image Information

### Size Optimization
The multi-stage build produces a minimal image:
- **Base image**: Alpine Linux (~5MB)
- **Final image**: ~15-20MB (including binary and dependencies)
- **Development image**: ~500MB (includes Go toolchain)

### Security Features
- Non-root user execution
- Minimal attack surface (Alpine base)
- No sensitive data in image
- Read-only file system support

## 🔐 Security Best Practices

### API Token Management
```bash
# Use environment variables, not config files
export VIRTUOSO_API_TOKEN="$(cat ~/.virtuoso/token)"

# Or use secret management
export VIRTUOSO_API_TOKEN="$(op read op://vault/virtuoso/token)"
```

### Container Security
```bash
# Run with read-only root filesystem
docker run --rm --read-only -v $(pwd):/workspace virtuoso-cli:latest --help

# Run with limited capabilities
docker run --rm --cap-drop=ALL virtuoso-cli:latest --help
```

## 🚨 Troubleshooting

### Common Issues

1. **Build fails with "go.sum not found"**
   - Check .dockerignore doesn't exclude go.sum
   - Ensure go.sum exists in project root

2. **Permission denied errors**
   - Check file permissions on mounted volumes
   - Ensure proper UID/GID mapping

3. **API connection fails**
   - Verify VIRTUOSO_API_TOKEN is set
   - Check network connectivity from container

4. **Config not found**
   - Ensure config directory is mounted
   - Check file paths in virtuoso-config.yaml

### Getting Help
```bash
# Show wrapper help
./scripts/virtuoso --wrapper-help

# Show CLI help
./scripts/virtuoso --help

# Debug mode
./scripts/virtuoso --verbose create-project "Debug Test"
```

## 📚 Additional Resources

- [Virtuoso API Documentation](https://docs.virtuoso.qa)
- [Docker Best Practices](https://docs.docker.com/develop/best-practices/)
- [Container Security Guide](https://docs.docker.com/engine/security/)

---

## 🎯 Team Onboarding

For new team members:

1. **Install Docker** on your system
2. **Clone this repository**
3. **Set environment variables** for your Virtuoso API token
4. **Run**: `./scripts/virtuoso --help`
5. **Create your first project**: `./scripts/virtuoso create-project "My First Test"`

That's it! No Go installation, no dependency management, just Docker and this repository.