#!/usr/bin/env bash
# 00_env.bash - Shared environment setup for all tests

# Load test helpers
load test_helpers

# Source project configuration if available
if [ -f "src/cmd/tests/config.sh" ]; then
    source src/cmd/tests/config.sh
fi

# Export common test variables
export TEST_TAG_PREFIX="${TEST_TAG_PREFIX:-test-$$}"
export API_CLI_BIN="${API_CLI_BIN:-./bin/api-cli}"

# Setup function that can be called by other tests
setup() {
    # Ensure binary exists
    [ -f "$API_CLI_BIN" ] || skip "API CLI binary not found at $API_CLI_BIN"
}