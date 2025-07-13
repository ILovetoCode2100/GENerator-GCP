#!/bin/bash
# test_helpers.bash - Common helper functions for BATS tests

# Get test tag prefix from environment or generate a default
get_test_tag() {
    if [[ -n "$TEST_TAG_PREFIX" ]]; then
        echo "$TEST_TAG_PREFIX"
    else
        # Generate a unique tag for local testing
        echo "test-local-$(date +%s)"
    fi
}

# Create a resource with test tags
# Usage: create_with_tag "project" "create" "--name" "test-project"
create_with_tag() {
    local resource_type="$1"
    local action="$2"
    shift 2
    
    local tag_prefix=$(get_test_tag)
    
    # Add tag option to the command
    "$BINARY" "$resource_type" "$action" "$@" --tag "$tag_prefix"
}

# Helper to extract ID from command output
# Usage: extract_id "$output"
extract_id() {
    local output="$1"
    
    # Try JSON parsing first
    if command -v jq &> /dev/null; then
        local id=$(echo "$output" | jq -r '.id // .resource_id // empty' 2>/dev/null)
        if [[ -n "$id" && "$id" != "null" ]]; then
            echo "$id"
            return
        fi
    fi
    
    # Fallback to regex for UUID pattern
    echo "$output" | grep -oE '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}' | head -1
}

# Save test data for cross-test sharing
# Usage: save_test_data "project_id" "$PROJECT_ID"
save_test_data() {
    local key="$1"
    local value="$2"
    echo "$value" > "$TEST_DATA_DIR/$key"
}

# Load test data saved by another test
# Usage: PROJECT_ID=$(load_test_data "project_id")
load_test_data() {
    local key="$1"
    if [[ -f "$TEST_DATA_DIR/$key" ]]; then
        cat "$TEST_DATA_DIR/$key"
    fi
}

# Clean up test data directory
cleanup_test_data() {
    rm -rf "$TEST_DATA_DIR"
    mkdir -p "$TEST_DATA_DIR"
}

# Log test information
log_test_info() {
    echo "# TEST: $1" >&3
    echo "# TAG: $(get_test_tag)" >&3
}
