# Virtuoso API CLI Makefile

# Variables
BINARY_NAME=api-cli
BUILD_DIR=bin
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -s -w"

.PHONY: all build clean test help

all: clean build ## Clean and build

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the CLI binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/api-cli
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@$(GOCLEAN)
	@echo "Clean complete"

test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v ./...

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

install: build ## Install binary to /usr/local/bin
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Installation complete"

lint: ## Run linting
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

fmt: ## Format code
	gofmt -s -w .
	goimports -w .

check: ## Run all checks
	@echo "Running all checks..."
	make fmt
	make lint
	make test
	make build
	@echo "All checks passed!"

# Integrated test scripts
test-commands: build ## Test all CLI commands
	@echo "Testing all commands..."
	@chmod +x ./test-consolidated-commands-final.sh 2>/dev/null || true
	@./test-consolidated-commands-final.sh

test-library: build ## Test library commands
	@echo "Testing library commands..."
	@chmod +x ./test-library-commands.sh 2>/dev/null || true
	@./test-library-commands.sh

# AI documentation generation
ai-generate-docs: build ## Generate AI-friendly command documentation
	@echo "Generating AI-optimized documentation..."
	@echo "## Command Reference (Auto-generated)" > docs/AI_COMMANDS.md
	@echo "" >> docs/AI_COMMANDS.md
	@echo "Generated on: $$(date)" >> docs/AI_COMMANDS.md
	@echo "" >> docs/AI_COMMANDS.md
	@./bin/api-cli help --format json 2>/dev/null | jq -r '.commands[]' | while read cmd; do \
		echo "### $$cmd" >> docs/AI_COMMANDS.md; \
		./bin/api-cli $$cmd --help 2>/dev/null | grep -A 100 "Usage:" >> docs/AI_COMMANDS.md || true; \
		echo "" >> docs/AI_COMMANDS.md; \
	done
	@echo "AI documentation generated in docs/AI_COMMANDS.md"

ai-schema-export: ## Export command schemas for AI parsing
	@echo "Exporting command schemas..."
	@mkdir -p schemas
	@echo '{"version": "2.0", "commands": {' > schemas/commands.json
	@first=true; \
	for group in assert interact navigate data dialog wait window mouse select file misc library; do \
		if [ "$$first" = "true" ]; then first=false; else echo -n "," >> schemas/commands.json; fi; \
		echo -n "\"$$group\": {" >> schemas/commands.json; \
		echo -n "\"description\": \"$$(./bin/api-cli $$group --help 2>/dev/null | head -1)\"," >> schemas/commands.json; \
		echo -n "\"subcommands\": []" >> schemas/commands.json; \
		echo -n "}" >> schemas/commands.json; \
	done
	@echo "}}" >> schemas/commands.json
	@jq . schemas/commands.json > schemas/commands_formatted.json
	@mv schemas/commands_formatted.json schemas/commands.json
	@echo "Command schemas exported to schemas/commands.json"