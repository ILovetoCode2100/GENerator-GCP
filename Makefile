# api-cli-generator Makefile

# Variables
BINARY_NAME=api-cli
BUILD_DIR=bin
SRC_DIR=src
SPEC_FILE=specs/api.yaml
DOCKER_IMAGE=api-cli-generator
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -s -w"

.PHONY: all build clean test generate docker help

all: clean generate build ## Clean, generate, and build

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

generate: ## Generate client code from OpenAPI spec
	@echo "Generating client code from OpenAPI spec..."
	@mkdir -p $(SRC_DIR)/api
	oapi-codegen -package api -generate types $(SPEC_FILE) > $(SRC_DIR)/api/types.gen.go
	oapi-codegen -package api -generate client $(SPEC_FILE) > $(SRC_DIR)/api/client.gen.go
	@echo "Code generation complete"

build: ## Build the CLI binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./$(SRC_DIR)/cmd
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

validate-spec: ## Validate OpenAPI specification
	@echo "Validating OpenAPI spec..."
	@if [ -f $(SPEC_FILE) ]; then \
		npx @apidevtools/swagger-cli validate $(SPEC_FILE) && echo "✅ Spec is valid"; \
	else \
		echo "❌ Spec file not found: $(SPEC_FILE)"; \
		exit 1; \
	fi

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(VERSION) .
	docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest

docker-run: ## Run Docker container
	docker run --rm -it $(DOCKER_IMAGE):latest --help

install: build ## Install binary to /usr/local/bin
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Installation complete"

dev: ## Run with hot reload (requires air)
	@which air > /dev/null || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	air -c .air.toml
