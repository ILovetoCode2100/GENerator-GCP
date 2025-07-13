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

.PHONY: all build build-test clean test generate docker help

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

build-test: ## Build the CLI binary for testing (fails fast on compilation errors)
	@echo "Building $(BINARY_NAME) for testing..."
	@mkdir -p $(BUILD_DIR)
	@cd $(SRC_DIR)/cmd && $(GOBUILD) -o ../../$(BUILD_DIR)/$(BINARY_NAME) || (echo "Build failed!" && exit 1)
	@echo "Test build complete: $(BUILD_DIR)/$(BINARY_NAME)"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@$(GOCLEAN)
	@echo "Clean complete"

test: ## Run Go unit tests
	@echo "Running Go unit tests..."
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

ci: ## Run CI pipeline locally
	@echo "Running CI pipeline locally..."
	make check
	make validate-spec
	make test-bats
	make docker-build
	@echo "CI pipeline completed!"

tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/cosmtrek/air@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install github.com/goreleaser/goreleaser@latest
	npm install -g @apidevtools/swagger-cli
	npm install -g bats
	@echo "Development tools installed!"

release-snapshot: ## Build release snapshot
	@which goreleaser > /dev/null || (echo "Installing goreleaser..." && go install github.com/goreleaser/goreleaser@latest)
	goreleaser release --snapshot --rm-dist

coverage: ## Generate test coverage report
	@echo "Generating coverage report..."
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-bats: build-test ## Run BATS integration tests with report
	@echo "Running BATS integration tests..."
	@if command -v bats >/dev/null 2>&1; then \
		./$(SRC_DIR)/cmd/tests/generate_report.sh; \
	else \
		echo "BATS not installed. Install with: npm install -g bats"; \
		exit 1; \
	fi

test-all: test test-bats ## Run all tests (Go unit tests and BATS integration tests)
	@echo "All tests completed!"
