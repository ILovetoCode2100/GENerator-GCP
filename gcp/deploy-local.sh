#!/bin/bash
# Local Development Setup Script for Virtuoso API CLI
# This script sets up a local development environment with GCP emulators

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Default values
FIRESTORE_PORT=8080
PUBSUB_PORT=8085
STORAGE_PORT=9000
CLOUD_TASKS_PORT=8123
API_PORT=8000
FRONTEND_PORT=3000

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check prerequisites
check_prerequisites() {
    print_info "Checking prerequisites..."

    local missing_tools=()

    # Check required tools
    command -v docker >/dev/null 2>&1 || missing_tools+=("docker")
    command -v docker-compose >/dev/null 2>&1 || missing_tools+=("docker-compose")
    command -v gcloud >/dev/null 2>&1 || missing_tools+=("gcloud")
    command -v go >/dev/null 2>&1 || missing_tools+=("go")
    command -v python3 >/dev/null 2>&1 || missing_tools+=("python3")

    if [ ${#missing_tools[@]} -ne 0 ]; then
        print_error "Missing required tools: ${missing_tools[*]}"
        exit 1
    fi

    # Check if GCP SDK components are installed
    if ! gcloud components list 2>/dev/null | grep -q "cloud-firestore-emulator.*Installed"; then
        print_warning "Firestore emulator not installed"
        print_info "Installing GCP emulators..."
        gcloud components install cloud-firestore-emulator pubsub-emulator beta --quiet
    fi

    print_success "All prerequisites met"
}

# Function to start GCP emulators
start_emulators() {
    print_info "Starting GCP emulators..."

    # Create emulator data directory
    mkdir -p "$PROJECT_ROOT/.emulator-data"

    # Start Firestore emulator
    print_info "Starting Firestore emulator on port $FIRESTORE_PORT..."
    gcloud emulators firestore start \
        --port=$FIRESTORE_PORT \
        --host-port=0.0.0.0:$FIRESTORE_PORT \
        > "$PROJECT_ROOT/.emulator-data/firestore.log" 2>&1 &

    # Start Pub/Sub emulator
    print_info "Starting Pub/Sub emulator on port $PUBSUB_PORT..."
    gcloud emulators pubsub start \
        --port=$PUBSUB_PORT \
        --host-port=0.0.0.0:$PUBSUB_PORT \
        > "$PROJECT_ROOT/.emulator-data/pubsub.log" 2>&1 &

    # Wait for emulators to start
    sleep 5

    # Export emulator environment variables
    export FIRESTORE_EMULATOR_HOST="localhost:$FIRESTORE_PORT"
    export PUBSUB_EMULATOR_HOST="localhost:$PUBSUB_PORT"

    print_success "Emulators started"
}

# Function to create local configuration
create_local_config() {
    print_info "Creating local configuration..."

    # Create local config directory
    mkdir -p "$PROJECT_ROOT/.local"

    # Create environment file for API
    cat > "$PROJECT_ROOT/.local/.env" <<EOF
# Environment
ENVIRONMENT=local
DEBUG=true

# GCP Emulators
FIRESTORE_EMULATOR_HOST=localhost:$FIRESTORE_PORT
PUBSUB_EMULATOR_HOST=localhost:$PUBSUB_PORT
STORAGE_EMULATOR_HOST=localhost:$STORAGE_PORT

# API Configuration
API_PORT=$API_PORT
API_HOST=0.0.0.0

# Virtuoso Configuration (use test credentials)
VIRTUOSO_API_KEY=${VIRTUOSO_API_KEY:-test-api-key}
VIRTUOSO_ORG_ID=${VIRTUOSO_ORG_ID:-test-org-id}
VIRTUOSO_BASE_URL=${VIRTUOSO_BASE_URL:-https://api-app2.virtuoso.qa/api}

# Security (local development only)
JWT_SECRET=local-development-secret-key
API_KEY=local-api-key

# Cloud Tasks emulator
CLOUD_TASKS_EMULATOR_HOST=localhost:$CLOUD_TASKS_PORT

# Disable authentication for local development
DISABLE_AUTH=true
EOF

    # Create local Virtuoso config
    mkdir -p "$HOME/.api-cli"
    cat > "$HOME/.api-cli/virtuoso-config.yaml" <<EOF
# Local development configuration
api:
  auth_token: ${VIRTUOSO_API_KEY:-test-api-key}
  base_url: ${VIRTUOSO_BASE_URL:-https://api-app2.virtuoso.qa/api}
organization:
  id: "${VIRTUOSO_ORG_ID:-2242}"
test:
  output_format: human
  auto_validate: true
EOF

    # Create docker-compose override for local development
    cat > "$PROJECT_ROOT/docker-compose.override.yml" <<EOF
version: '3.8'

services:
  api:
    build:
      context: .
      dockerfile: api/Dockerfile.api
    ports:
      - "$API_PORT:8000"
    environment:
      - ENVIRONMENT=local
      - FIRESTORE_EMULATOR_HOST=host.docker.internal:$FIRESTORE_PORT
      - PUBSUB_EMULATOR_HOST=host.docker.internal:$PUBSUB_PORT
      - DISABLE_AUTH=true
    volumes:
      - ./api:/app
      - ./bin:/app/bin
      - ./.local/.env:/app/.env
    extra_hosts:
      - "host.docker.internal:host-gateway"

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  # Local storage emulator
  storage:
    image: fsouza/fake-gcs-server
    ports:
      - "$STORAGE_PORT:9000"
    command: -scheme http -public-host localhost:$STORAGE_PORT
    volumes:
      - ./.emulator-data/storage:/data
EOF

    print_success "Local configuration created"
}

# Function to build CLI binary
build_cli() {
    print_info "Building CLI binary..."

    cd "$PROJECT_ROOT"
    make build

    if [ ! -f "bin/api-cli" ]; then
        print_error "Failed to build CLI binary"
        exit 1
    fi

    print_success "CLI binary built"
}

# Function to initialize test data
initialize_test_data() {
    print_info "Initializing test data..."

    # Wait for services to be ready
    sleep 5

    # Create test project using the CLI
    export VIRTUOSO_SESSION_ID=""

    # Use the CLI to create test infrastructure
    print_info "Creating test project..."
    if PROJECT_RESULT=$("$PROJECT_ROOT/bin/api-cli" create-project "Local Test Project" -o json 2>/dev/null); then
        PROJECT_ID=$(echo "$PROJECT_RESULT" | jq -r '.project_id')
        print_success "Test project created: $PROJECT_ID"

        # Create test goal
        if GOAL_RESULT=$("$PROJECT_ROOT/bin/api-cli" create-goal "$PROJECT_ID" "Test Goal" -o json 2>/dev/null); then
            GOAL_ID=$(echo "$GOAL_RESULT" | jq -r '.goal_id')
            SNAPSHOT_ID=$(echo "$GOAL_RESULT" | jq -r '.snapshot_id')
            print_success "Test goal created: $GOAL_ID"

            # Create test journey
            if JOURNEY_RESULT=$("$PROJECT_ROOT/bin/api-cli" create-journey "$GOAL_ID" "$SNAPSHOT_ID" "Test Journey" -o json 2>/dev/null); then
                JOURNEY_ID=$(echo "$JOURNEY_RESULT" | jq -r '.journey_id')
                print_success "Test journey created: $JOURNEY_ID"

                # Save test IDs
                cat > "$PROJECT_ROOT/.local/test-data.env" <<EOF
export TEST_PROJECT_ID=$PROJECT_ID
export TEST_GOAL_ID=$GOAL_ID
export TEST_SNAPSHOT_ID=$SNAPSHOT_ID
export TEST_JOURNEY_ID=$JOURNEY_ID
EOF

                print_info "Test data saved to .local/test-data.env"
            fi
        fi
    else
        print_warning "Could not create test data (API might not be accessible in local mode)"
    fi

    # Create sample test YAML files
    mkdir -p "$PROJECT_ROOT/.local/tests"

    cat > "$PROJECT_ROOT/.local/tests/local-test.yaml" <<EOF
name: "Local Development Test"
steps:
  - navigate: "https://example.com"
  - assert: "Example Domain"
  - click: "More information..."
  - wait: 1000
  - comment: "Test completed"
EOF

    print_success "Test data initialized"
}

# Function to start services
start_services() {
    print_info "Starting services..."

    cd "$PROJECT_ROOT"

    # Start Docker services
    docker-compose up -d

    # Wait for services to be ready
    print_info "Waiting for services to start..."
    sleep 10

    # Check API health
    if curl -s "http://localhost:$API_PORT/health" | grep -q "healthy"; then
        print_success "API service is healthy"
    else
        print_warning "API service health check failed (might need more time to start)"
    fi

    print_success "Services started"
}

# Function to setup port forwarding
setup_port_forwarding() {
    print_info "Setting up port forwarding..."

    # Create convenience script for port forwarding
    cat > "$PROJECT_ROOT/.local/port-forward.sh" <<'EOF'
#!/bin/bash
# Port forwarding convenience script

echo "Port forwarding active for:"
echo "  - API: http://localhost:8000"
echo "  - Firestore UI: http://localhost:4000"
echo "  - Redis Commander: http://localhost:8081"
echo ""
echo "Press Ctrl+C to stop"

# Keep script running
while true; do
    sleep 1
done
EOF

    chmod +x "$PROJECT_ROOT/.local/port-forward.sh"

    print_success "Port forwarding configured"
}

# Function to print local development info
print_dev_info() {
    print_info "Local Development Environment Ready!"
    echo -e "${BLUE}===================================================${NC}"
    echo -e "\nServices running:"
    echo -e "  - API: ${GREEN}http://localhost:$API_PORT${NC}"
    echo -e "  - Firestore Emulator: ${GREEN}localhost:$FIRESTORE_PORT${NC}"
    echo -e "  - Pub/Sub Emulator: ${GREEN}localhost:$PUBSUB_PORT${NC}"
    echo -e "  - Storage Emulator: ${GREEN}http://localhost:$STORAGE_PORT${NC}"
    echo -e "  - Redis: ${GREEN}localhost:6379${NC}"
    echo -e "\nUseful commands:"
    echo -e "  - View logs: ${YELLOW}docker-compose logs -f${NC}"
    echo -e "  - Stop services: ${YELLOW}docker-compose down${NC}"
    echo -e "  - Rebuild: ${YELLOW}docker-compose build${NC}"
    echo -e "  - Test CLI: ${YELLOW}./bin/api-cli run-test .local/tests/local-test.yaml${NC}"
    echo -e "\nEnvironment variables:"
    echo -e "  ${YELLOW}source .local/test-data.env${NC} (load test project IDs)"
    echo -e "\nEmulator logs:"
    echo -e "  - Firestore: ${YELLOW}tail -f .emulator-data/firestore.log${NC}"
    echo -e "  - Pub/Sub: ${YELLOW}tail -f .emulator-data/pubsub.log${NC}"
    echo -e "${BLUE}===================================================${NC}"
}

# Function to cleanup on exit
cleanup() {
    print_info "Cleaning up..."

    # Stop emulators
    pkill -f "cloud-firestore-emulator" || true
    pkill -f "pubsub-emulator" || true

    # Stop Docker services
    cd "$PROJECT_ROOT"
    docker-compose down

    print_success "Cleanup completed"
}

# Set trap for cleanup
trap cleanup EXIT

# Parse command line arguments
SKIP_EMULATORS=false
SKIP_BUILD=false
SKIP_SERVICES=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-emulators)
            SKIP_EMULATORS=true
            shift
            ;;
        --skip-build)
            SKIP_BUILD=true
            shift
            ;;
        --skip-services)
            SKIP_SERVICES=true
            shift
            ;;
        --api-port)
            API_PORT="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo "Options:"
            echo "  --skip-emulators    Skip starting GCP emulators"
            echo "  --skip-build        Skip building CLI binary"
            echo "  --skip-services     Skip starting Docker services"
            echo "  --api-port PORT     API port (default: 8000)"
            echo "  --help              Show this help message"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Main flow
print_info "Setting up local development environment"
echo -e "${BLUE}===================================================${NC}\n"

check_prerequisites

if [ "$SKIP_EMULATORS" = false ]; then
    start_emulators
fi

create_local_config

if [ "$SKIP_BUILD" = false ]; then
    build_cli
fi

if [ "$SKIP_SERVICES" = false ]; then
    start_services
fi

initialize_test_data
setup_port_forwarding
print_dev_info

print_success "Local development environment is ready!"
print_info "Press Ctrl+C to stop all services"

# Keep script running
while true; do
    sleep 1
done
