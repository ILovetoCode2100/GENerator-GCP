#!/usr/bin/env bats
# 10_auth.bats - Authentication, health check and connectivity tests

load 00_env

# Load test configuration for Virtuoso credentials
setup() {
    # Call parent setup
    setup
    
    # Source config to get credentials
    if [[ -f "$PROJECT_ROOT/scripts/test/config.sh" ]]; then
        source "$PROJECT_ROOT/scripts/test/config.sh"
    fi
    
    # Ensure we have required environment variables
    if [[ -z "$VIRTUOSO_BASE_URL" ]]; then
        skip "VIRTUOSO_BASE_URL not set"
    fi
}

# Test 1: Raw curl to /health endpoint to isolate network/auth problems
@test "raw curl to /health endpoint works" {
    # Skip if curl is not available
    if ! command -v curl &> /dev/null; then
        skip "curl command not found"
    fi
    
    # Make raw curl request to health endpoint
    if [[ -n "$VIRTUOSO_AUTH_TOKEN" ]]; then
        run curl -s -o /dev/null -w "%{http_code}" \
            -H "Authorization: Bearer $VIRTUOSO_AUTH_TOKEN" \
            "${VIRTUOSO_BASE_URL}/health"
    else
        run curl -s -o /dev/null -w "%{http_code}" \
            "${VIRTUOSO_BASE_URL}/health"
    fi
    
    # Check that we get a successful HTTP status code
    [ "$status" -eq 0 ]
    [[ "$output" =~ ^(200|204)$ ]] || {
        echo "Expected HTTP 200 or 204, got: $output"
        false
    }
}

# Test 2: Execute bin/api-cli status and expect HTTP 200/204
@test "api-cli status command returns success" {
    # Check if binary exists
    if [[ ! -x "$BINARY" ]]; then
        skip "Binary not found at: $BINARY"
    fi
    
    # Run status command
    run "$BINARY" status
    
    # Check exit code
    [ "$status" -eq 0 ] || {
        echo "Command failed with exit code: $status"
        echo "Output: $output"
        false
    }
    
    # Check output indicates success
    [[ "$output" =~ "OK" ]] || \
    [[ "$output" =~ "ok" ]] || \
    [[ "$output" =~ "Success" ]] || \
    [[ "$output" =~ "success" ]] || \
    [[ "$output" =~ "200" ]] || \
    [[ "$output" =~ "204" ]] || {
        echo "Expected success indication in output, got: $output"
        false
    }
}

# Test 3: Time request and ensure it's less than configured timeout
@test "api-cli status completes within configured timeout" {
    # Check if binary exists
    if [[ ! -x "$BINARY" ]]; then
        skip "Binary not found at: $BINARY"
    fi
    
    # Get configured timeout (default to 30 seconds if not set)
    local timeout_seconds="${VIRTUOSO_TEST_TIMEOUT:-30}"
    
    # Record start time
    local start_time=$(date +%s)
    
    # Run status command with timeout
    run timeout "$timeout_seconds" "$BINARY" status
    
    # Record end time
    local end_time=$(date +%s)
    local elapsed=$((end_time - start_time))
    
    # Check that command succeeded
    [ "$status" -eq 0 ] || {
        if [ "$status" -eq 124 ]; then
            echo "Command timed out after $timeout_seconds seconds"
        else
            echo "Command failed with exit code: $status"
        fi
        echo "Output: $output"
        false
    }
    
    # Verify it completed within timeout
    [ "$elapsed" -lt "$timeout_seconds" ] || {
        echo "Request took $elapsed seconds, exceeding timeout of $timeout_seconds seconds"
        false
    }
    
    echo "Request completed in $elapsed seconds (timeout: $timeout_seconds seconds)"
}

# Additional connectivity tests

@test "health check command exists" {
    run "$BINARY" health --help
    [ "$status" -eq 0 ]
}

@test "health check returns success when API is available" {
    run "$BINARY" health
    [ "$status" -eq 0 ]
    [[ "$output" =~ "healthy" ]] || [[ "$output" =~ "ok" ]] || [[ "$output" =~ "success" ]]
}

@test "authentication with valid credentials succeeds" {
    # Check if we have auth token
    if [[ -z "$VIRTUOSO_AUTH_TOKEN" ]]; then
        skip "VIRTUOSO_AUTH_TOKEN not set"
    fi
    
    # Test authentication by making an authenticated request
    run "$BINARY" status --auth-token="$VIRTUOSO_AUTH_TOKEN"
    [ "$status" -eq 0 ]
}

@test "authentication with invalid credentials fails" {
    # Test with invalid auth token
    run "$BINARY" status --auth-token="invalid-token-12345"
    [ "$status" -ne 0 ]
    [[ "$output" =~ "unauthorized" ]] || [[ "$output" =~ "invalid" ]] || [[ "$output" =~ "401" ]]
}

@test "handles network timeout gracefully" {
    # Test with unreachable endpoint
    export VIRTUOSO_TEST_TIMEOUT=5
    run timeout 10 "$BINARY" status --base-url="http://localhost:99999"
    [ "$status" -ne 0 ]
    [[ "$output" =~ "timeout" ]] || [[ "$output" =~ "connection" ]] || [[ "$output" =~ "refused" ]]
}
