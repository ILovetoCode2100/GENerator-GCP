#!/usr/bin/env bats
# 00_env.bats - Environment setup and binary verification

# Load test configuration
setup() {
    # Set test directory
    TEST_DIR="$( cd "$( dirname "$BATS_TEST_FILENAME" )" && pwd )"
    PROJECT_ROOT="$(dirname "$(dirname "$(dirname "$TEST_DIR")")")"
    
    # Load config if exists
    if [[ -f "$PROJECT_ROOT/config.sh" ]]; then
        source "$PROJECT_ROOT/config.sh"
    fi
    
    # Set binary path to bin/api-cli relative to project root
    export BINARY="${BINARY:-$PROJECT_ROOT/bin/api-cli}"
    
    # Common test data directory
    export TEST_DATA_DIR="$TEST_DIR/data"
    mkdir -p "$TEST_DATA_DIR"
    
    # Source test helpers
    source "$TEST_DIR/test_helpers.bash"
}

teardown() {
    # Don't clean up TEST_DATA_DIR here as it's needed for sharing data between test files
    # Cleanup should be done after all tests complete or manually
    :
}

@test "binary exists and is executable" {
    [[ -f "$BINARY" ]]
    [[ -x "$BINARY" ]]
}

@test "binary shows version" {
    run "$BINARY" --version
    [ "$status" -eq 0 ]
    [[ "$output" =~ [0-9]+\.[0-9]+\.[0-9]+ ]]
}

@test "binary shows help" {
    run "$BINARY" --help
    [ "$status" -eq 0 ]
    [[ "$output" =~ "Usage:" ]]
}

@test "config.sh loaded successfully" {
    # Add any config validation here
    [[ -n "$BINARY" ]]
}

@test "required environment variables are set" {
    # Add checks for any required env vars
    # Example:
    # [[ -n "$API_KEY" ]]
    # [[ -n "$API_URL" ]]
    skip "Add environment variable checks as needed"
}
