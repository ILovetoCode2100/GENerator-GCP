#!/usr/bin/env bats
# 80_errors.bats - Error handling and negative scenario tests

load 00_env

setup() {
    load 00_env
    setup
}

# Missing Required Arguments
@test "error when required arguments are missing" {
    run "$BINARY" project create
    [ "$status" -ne 0 ]
    [[ "$output" =~ "required" ]] || [[ "$output" =~ "missing" ]]
}

@test "error when project name is missing" {
    run "$BINARY" project create --description "No name project"
    [ "$status" -ne 0 ]
    [[ "$output" =~ "name" ]]
}

# Invalid Input Validation
@test "error on invalid project name characters" {
    run "$BINARY" project create --name "invalid/project*name" --description "Test"
    [ "$status" -ne 0 ]
    [[ "$output" =~ "invalid" ]] || [[ "$output" =~ "character" ]]
}

@test "error on excessively long names" {
    LONG_NAME=$(printf 'a%.0s' {1..500})
    run "$BINARY" project create --name "$LONG_NAME" --description "Test"
    [ "$status" -ne 0 ]
    [[ "$output" =~ "long" ]] || [[ "$output" =~ "limit" ]] || [[ "$output" =~ "exceed" ]]
}

# Non-Existent Resources
@test "error when accessing non-existent project" {
    run "$BINARY" project get --name "non-existent-project-12345"
    [ "$status" -ne 0 ]
    [[ "$output" =~ "not found" ]] || [[ "$output" =~ "does not exist" ]]
}

@test "error when accessing non-existent journey" {
    TEST_PROJECT="error-test-project-$(date +%s)"
    run "$BINARY" project create --name "$TEST_PROJECT" --description "Test"
    
    run "$BINARY" journey get --project "$TEST_PROJECT" --journey "non-existent-journey"
    [ "$status" -ne 0 ]
    [[ "$output" =~ "not found" ]] || [[ "$output" =~ "does not exist" ]]
}

@test "creating journey with nonexistent project ID returns 404" {
    # Use a UUID that doesn't exist
    NONEXISTENT_PROJECT_ID="00000000-0000-0000-0000-000000000000"
    
    run "$BINARY" journey create "$NONEXISTENT_PROJECT_ID" "Test Journey"
    [ "$status" -ne 0 ]
    [[ "$output" =~ "404" ]] || [[ "$output" =~ "not found" ]] || [[ "$output" =~ "does not exist" ]]
}

# Authentication Errors
@test "error on invalid authentication" {
    # Save current auth
    OLD_API_KEY="$API_KEY"
    export API_KEY="invalid-key-12345"
    
    run "$BINARY" project list
    [ "$status" -ne 0 ]
    [[ "$output" =~ "unauthorized" ]] || [[ "$output" =~ "authentication" ]] || [[ "$output" =~ "401" ]]
    
    # Restore auth
    export API_KEY="$OLD_API_KEY"
}

@test "invalid token returns 401" {
    # Save current auth token/key
    OLD_API_KEY="${API_KEY:-}"
    OLD_AUTH_TOKEN="${AUTH_TOKEN:-}"
    
    # Set invalid token
    export API_KEY="invalid-token-test-12345"
    export AUTH_TOKEN="invalid-token-test-12345"
    
    run "$BINARY" project list
    [ "$status" -ne 0 ]
    [[ "$output" =~ "401" ]] || [[ "$output" =~ "Unauthorized" ]] || [[ "$output" =~ "unauthorized" ]]
    
    # Restore auth
    export API_KEY="$OLD_API_KEY"
    export AUTH_TOKEN="$OLD_AUTH_TOKEN"
}

# Network/Connection Errors
@test "error on network timeout" {
    skip "Implement network timeout simulation"
    # export API_TIMEOUT=1
    # export API_URL="http://10.255.255.1"  # Non-routable IP
    # run timeout 5 "$BINARY" project list
    # [ "$status" -ne 0 ]
    # [[ "$output" =~ "timeout" ]] || [[ "$output" =~ "connection" ]]
}

@test "error on invalid API endpoint" {
    OLD_API_URL="$API_URL"
    export API_URL="http://invalid-endpoint-12345.local"
    
    run "$BINARY" health
    [ "$status" -ne 0 ]
    
    export API_URL="$OLD_API_URL"
}

# Permission/Access Errors
@test "error on insufficient permissions" {
    skip "Implement based on your permission model"
    # run "$BINARY" admin-only-command
    # [ "$status" -ne 0 ]
    # [[ "$output" =~ "permission" ]] || [[ "$output" =~ "forbidden" ]] || [[ "$output" =~ "403" ]]
}

# Invalid Operations
@test "error creating duplicate resources" {
    TEST_PROJECT="dup-test-$(date +%s)"
    
    run "$BINARY" project create --name "$TEST_PROJECT" --description "First"
    [ "$status" -eq 0 ]
    
    run "$BINARY" project create --name "$TEST_PROJECT" --description "Duplicate"
    [ "$status" -ne 0 ]
    [[ "$output" =~ "exists" ]] || [[ "$output" =~ "duplicate" ]]
}

@test "error on circular dependencies" {
    skip "Implement if your model has dependency constraints"
    # Test creating circular references between resources
}

# Data Validation Errors
@test "supplying bad JSON in --data flags returns 400 with helpful message" {
    # Skip if PROJECT_ID is not available
    if [ -f "$TEST_DATA_DIR/project_id" ]; then
        PROJECT_ID=$(cat "$TEST_DATA_DIR/project_id")
    fi
    [ -n "${PROJECT_ID}" ] || skip "PROJECT_ID not available from previous tests"
    
    # Test with invalid JSON syntax
    INVALID_JSON='{"invalid JSON}'
    
    run "$BINARY" journey create "$PROJECT_ID" "Test Journey" --data "$INVALID_JSON"
    [ "$status" -ne 0 ]
    [[ "$output" =~ "400" ]] || [[ "$output" =~ "Bad Request" ]] || [[ "$output" =~ "invalid JSON" ]]
}

@test "malformed JSON with unclosed brackets returns 400" {
    # Skip if PROJECT_ID is not available
    if [ -f "$TEST_DATA_DIR/project_id" ]; then
        PROJECT_ID=$(cat "$TEST_DATA_DIR/project_id")
    fi
    [ -n "${PROJECT_ID}" ] || skip "PROJECT_ID not available from previous tests"
    
    INVALID_JSON='{"title":"test","description":"missing closing bracket"'
    
    run "$BINARY" journey create "$PROJECT_ID" "Test Journey" --data "$INVALID_JSON"
    [ "$status" -ne 0 ]
    [[ "$output" =~ "400" ]] || [[ "$output" =~ "JSON" ]] || [[ "$output" =~ "parse" ]]
}

@test "invalid JSON data types returns 400" {
    # Skip if PROJECT_ID is not available
    if [ -f "$TEST_DATA_DIR/project_id" ]; then
        PROJECT_ID=$(cat "$TEST_DATA_DIR/project_id")
    fi
    [ -n "${PROJECT_ID}" ] || skip "PROJECT_ID not available from previous tests"
    
    # If the API expects certain fields to be strings but we provide numbers
    INVALID_JSON='{"name": 12345, "description": null}'
    
    run "$BINARY" project create --data "$INVALID_JSON"
    [ "$status" -ne 0 ]
    [[ "$output" =~ "400" ]] || [[ "$output" =~ "invalid" ]] || [[ "$output" =~ "type" ]]
}

@test "error on invalid date format" {
    skip "Implement if CLI accepts date inputs"
    # run "$BINARY" task create --due-date "not-a-date"
    # [ "$status" -ne 0 ]
    # [[ "$output" =~ "date" ]] || [[ "$output" =~ "format" ]]
}

# Command Parsing Errors
@test "error on unknown command" {
    run "$BINARY" unknown-command
    [ "$status" -ne 0 ]
    [[ "$output" =~ "unknown" ]] || [[ "$output" =~ "command not found" ]]
}

@test "error on unknown flag" {
    run "$BINARY" project list --unknown-flag
    [ "$status" -ne 0 ]
    [[ "$output" =~ "unknown" ]] || [[ "$output" =~ "flag" ]]
}

# Resource State Errors
@test "error on invalid state transition" {
    skip "Implement based on your state model"
    # run "$BINARY" project complete --name "non-started-project"
    # [ "$status" -ne 0 ]
    # [[ "$output" =~ "state" ]] || [[ "$output" =~ "cannot" ]]
}

# Error Message Quality
@test "error messages include helpful context" {
    run "$BINARY" project get --name "non-existent"
    [ "$status" -ne 0 ]
    # Should include the resource name in error
    [[ "$output" =~ "non-existent" ]]
}

@test "error messages suggest corrections" {
    skip "Implement if your CLI has suggestion feature"
    # run "$BINARY" projeckt list  # Typo
    # [ "$status" -ne 0 ]
    # [[ "$output" =~ "Did you mean" ]] || [[ "$output" =~ "project" ]]
}

# Graceful Degradation
@test "partial success is reported correctly" {
    skip "Implement for batch operations"
    # run "$BINARY" project delete --names "exists,not-exists,also-exists"
    # Check that partial success is reported
}

@test "CLI doesn't crash on unexpected errors" {
    # Send interrupt signal
    skip "Complex signal handling test"
    # Test various signal handling scenarios
}
