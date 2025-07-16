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