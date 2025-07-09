# Contributing to API CLI Generator

First off, thank you for considering contributing to API CLI Generator! It's people like you that make this tool better for everyone.

## Code of Conduct

This project and everyone participating in it is governed by our Code of Conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check existing issues as you might find out that you don't need to create one. When you are creating a bug report, please include as many details as possible:

- **Use a clear and descriptive title**
- **Describe the exact steps which reproduce the problem**
- **Provide specific examples to demonstrate the steps**
- **Describe the behavior you observed after following the steps**
- **Explain which behavior you expected to see instead and why**
- **Include logs and error messages**

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, please include:

- **Use a clear and descriptive title**
- **Provide a step-by-step description of the suggested enhancement**
- **Provide specific examples to demonstrate the steps**
- **Describe the current behavior and explain which behavior you expected to see instead**
- **Explain why this enhancement would be useful**

### Pull Requests

1. Fork the repo and create your branch from `main`
2. If you've added code that should be tested, add tests
3. If you've changed APIs, update the documentation
4. Ensure the test suite passes
5. Make sure your code follows the existing style
6. Issue that pull request!

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Make
- Git

### Setting Up Your Development Environment

```bash
# Clone your fork
git clone https://github.com/your-username/api-cli-generator.git
cd api-cli-generator

# Add upstream remote
git remote add upstream https://github.com/original/api-cli-generator.git

# Install dependencies
go mod download

# Run tests to ensure everything is working
make test
```

### Building the Project

```bash
# Build the CLI binary
make build

# Run the CLI
./bin/api-cli --help
```

## Coding Guidelines

### Go Code Style

- Follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` to format your code
- Use `golint` and `go vet` to check for common issues
- Write descriptive variable and function names
- Add comments for exported functions and types
- Keep functions small and focused

### Testing

- Write unit tests for new functionality
- Ensure all tests pass before submitting PR
- Aim for good test coverage
- Use table-driven tests where appropriate

Example test structure:

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:  "valid input",
            input: "test",
            want:  "TEST",
        },
        {
            name:    "empty input",
            input:   "",
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("FunctionName() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("FunctionName() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Commit Messages

- Use the present tense ("Add feature" not "Added feature")
- Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
- Limit the first line to 72 characters or less
- Reference issues and pull requests liberally after the first line

Example:

```
Add support for batch operations

- Implement create-structure command
- Add YAML/JSON parsing for structure files
- Include validation for structure format
- Add comprehensive tests

Fixes #123
```

### Documentation

- Update README.md if you change functionality
- Add or update command documentation in COMMANDS.md
- Include code comments for complex logic
- Add examples for new features

## Project Structure

```
api-cli-generator/
├── cmd/              # CLI commands (Cobra)
├── pkg/              # Core packages
│   └── virtuoso/     # API client implementation
├── config/           # Configuration files
├── docs/             # Documentation
├── examples/         # Example files
├── scripts/          # Utility scripts
├── test/             # Test files
└── specs/            # API specifications
```

## Making Changes

1. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**
   - Write code
   - Add tests
   - Update documentation

3. **Test your changes**
   ```bash
   make test
   make build
   ./bin/api-cli [your-new-command]
   ```

4. **Commit your changes**
   ```bash
   git add .
   git commit -m "Add your descriptive commit message"
   ```

5. **Push to your fork**
   ```bash
   git push origin feature/your-feature-name
   ```

6. **Create a Pull Request**
   - Go to your fork on GitHub
   - Click "New Pull Request"
   - Provide a clear description of your changes

## Review Process

Once you submit a PR:

1. The CI pipeline will run tests
2. A maintainer will review your code
3. You may be asked to make changes
4. Once approved, your PR will be merged

## Questions?

Feel free to open an issue with your question or reach out to the maintainers directly.

Thank you for contributing! 🎉
