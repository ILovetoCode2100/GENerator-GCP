# API CLI Generator

A powerful command-line tool for generating OpenAPI-based CLIs with intelligent orchestration capabilities. Currently configured for Virtuoso API testing, but adaptable to any OpenAPI specification.

## 🚀 Features

- **OpenAPI-Driven**: Generate CLI from any OpenAPI 3.0 specification
- **Intelligent Orchestration**: Handle complex multi-step workflows automatically
- **Batch Operations**: Create complete structures from YAML/JSON files
- **Multiple Output Formats**: Human-readable, JSON, YAML, and AI-friendly outputs
- **Type-Safe**: Fully typed Go code generation from OpenAPI specs
- **Extensible**: Easy to add new commands and features

## 📋 Prerequisites

- Go 1.21 or higher
- Make (for build automation)
- Docker (optional, for containerized deployment)

## 🛠️ Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/yourusername/api-cli-generator.git
cd api-cli-generator

# Install dependencies
go mod download

# Build the CLI
make build

# Binary will be available at ./bin/api-cli
```

### Using Docker

```bash
# Build the Docker image
docker build -t api-cli-generator .

# Run the CLI in a container
docker run -it api-cli-generator --help
```

## ⚡ Quick Start

### 1. Basic Configuration

```bash
# Set up environment (for Virtuoso API)
source ./scripts/setup-virtuoso.sh

# Or configure your own API
cp config/example-config.yaml config/config.yaml
# Edit config/config.yaml with your API details
```

### 2. Create Test Structure

```bash
# Create a complete test structure from YAML
./bin/api-cli create-structure --file examples/generic-test-structure.yaml

# Or create individual components
./bin/api-cli create-project "My Test Project"
./bin/api-cli create-goal <project-id> "Test Goal" "https://example.com"
./bin/api-cli create-journey <goal-id> <snapshot-id> "User Journey"
```

### 3. Add Test Steps

```bash
# Add various step types
./bin/api-cli create-step-navigate <checkpoint-id> "https://example.com" 1
./bin/api-cli create-step-click <checkpoint-id> "Submit button" 2
./bin/api-cli create-step-write <checkpoint-id> "test@example.com" "Email field" 3
./bin/api-cli create-step-assert-exists <checkpoint-id> "Success message" 4
```

## 📚 Documentation

- [Command Reference](COMMANDS.md) - Complete list of available commands
- [Quick Reference](QUICK_REFERENCE.md) - Common usage patterns
- [API Structure](docs/architecture/virtuoso-api-structure.md) - Understanding the API
- [Testing Guide](docs/testing-guide.md) - How to test the CLI
- [Examples](examples/) - Sample configuration and structure files

## 🏗️ Architecture

The CLI follows a clean architecture pattern:

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  OpenAPI Spec   │────▶│  Code Generator │────▶│   Go CLI App    │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                                                          │
                              ┌───────────────────────────┼───────────────────────────┐
                              │                           │                           │
                        ┌─────▼─────┐            ┌────────▼────────┐         ┌───────▼───────┐
                        │  Commands │            │  API Client     │         │  Templates    │
                        │  (Cobra)  │            │  (Generated)    │         │  (Validated)  │
                        └───────────┘            └─────────────────┘         └───────────────┘
```

## 🔧 Configuration

The CLI can be configured through:

1. **Configuration file** (`config/config.yaml`)
2. **Environment variables** (prefix: `VIRTUOSO_` or your custom prefix)
3. **Command-line flags**

Example configuration:

```yaml
api:
  base_url: "https://api.example.com"
  headers:
    X-API-Key: "your-api-key"
    X-Client-ID: "your-client-id"
  
output:
  default_format: "human" # human, json, yaml, ai
  
retry:
  max_attempts: 3
  backoff_seconds: 2
```

## 🧪 Testing

```bash
# Run unit tests
make test

# Run integration tests
make test-integration

# Run all tests with coverage
make test-coverage
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [oapi-codegen](https://github.com/deepmap/oapi-codegen) for OpenAPI code generation
- CLI framework powered by [Cobra](https://github.com/spf13/cobra)
- HTTP client enhanced with [go-resty](https://github.com/go-resty/resty)

## 📞 Support

- 📧 Email: support@example.com
- 💬 Discord: [Join our community](https://discord.gg/example)
- 📖 Documentation: [docs.example.com](https://docs.example.com)
- 🐛 Issues: [GitHub Issues](https://github.com/yourusername/api-cli-generator/issues)
