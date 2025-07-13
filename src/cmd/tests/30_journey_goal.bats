#!/usr/bin/env bats
# 30_journey_goal.bats - Journey and goal operations tests

load 00_env

# Variables that will be captured during tests
JOURNEY_ID=""
GOAL_ID=""

setup() {
    load 00_env
    setup
    
    # Load PROJECT_ID from shared test data (saved by 20_project.bats)
    if [ -f "$TEST_DATA_DIR/project_id" ]; then
        PROJECT_ID=$(cat "$TEST_DATA_DIR/project_id")
    fi
    
    # If not available, we need to fail early
    if [ -z "${PROJECT_ID}" ]; then
        skip "PROJECT_ID not available from previous tests"
    fi
}

teardown() {
    # Clean up is handled by later test files or manually
    :
}

@test "journey create - creates 'Smoke Journey' and captures JOURNEY_ID" {
    run "$BINARY" journey create "$PROJECT_ID" "Smoke Journey"
    [ "$status" -eq 0 ]
    
    # Capture JOURNEY_ID from output - assuming JSON output
    if command -v jq &> /dev/null; then
        JOURNEY_ID=$(echo "$output" | jq -r '.id // .journey_id // .journeyId // empty')
    fi
    
    # If no JSON parsing available or ID not found, try to extract from plain text
    if [ -z "$JOURNEY_ID" ]; then
        # Try to extract ID from output like "Journey created: <id>"
        JOURNEY_ID=$(echo "$output" | grep -oE '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}' | head -1)
    fi
    
    # Save to shared test data for use in subsequent tests
    echo "$JOURNEY_ID" > "$TEST_DATA_DIR/journey_id"
    
    # Verify we got an ID
    [ -n "$JOURNEY_ID" ]
}

@test "journey list - verifies 'Smoke Journey' exists in project" {
    # Load JOURNEY_ID if needed
    if [ -z "$JOURNEY_ID" ] && [ -f "$TEST_DATA_DIR/journey_id" ]; then
        JOURNEY_ID=$(cat "$TEST_DATA_DIR/journey_id")
    fi
    
    run "$BINARY" journey list "$PROJECT_ID"
    [ "$status" -eq 0 ]
    
    # Verify the journey we created appears in the list
    [[ "$output" =~ "Smoke Journey" ]] || [[ "$output" =~ "$JOURNEY_ID" ]]
}

@test "goal create - creates 'Arrival Goal' and captures GOAL_ID" {
    # Load JOURNEY_ID from shared test data if not already loaded
    if [ -z "$JOURNEY_ID" ] && [ -f "$TEST_DATA_DIR/journey_id" ]; then
        JOURNEY_ID=$(cat "$TEST_DATA_DIR/journey_id")
    fi
    
    # Ensure we have JOURNEY_ID
    [ -n "$JOURNEY_ID" ] || skip "JOURNEY_ID not available"
    
    run "$BINARY" goal create "$JOURNEY_ID" "Arrival Goal"
    [ "$status" -eq 0 ]
    
    # Capture GOAL_ID from output - assuming JSON output
    if command -v jq &> /dev/null; then
        GOAL_ID=$(echo "$output" | jq -r '.id // .goal_id // .goalId // empty')
    fi
    
    # If no JSON parsing available or ID not found, try to extract from plain text
    if [ -z "$GOAL_ID" ]; then
        # Try to extract ID from output like "Goal created: <id>"
        GOAL_ID=$(echo "$output" | grep -oE '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}' | head -1)
    fi
    
    # Save to shared test data for use in subsequent tests
    echo "$GOAL_ID" > "$TEST_DATA_DIR/goal_id"
    
    # Verify we got an ID
    [ -n "$GOAL_ID" ]
}

@test "goal list - verifies 'Arrival Goal' exists in journey" {
    # Load JOURNEY_ID from shared test data if not already loaded
    if [ -z "$JOURNEY_ID" ] && [ -f "$TEST_DATA_DIR/journey_id" ]; then
        JOURNEY_ID=$(cat "$TEST_DATA_DIR/journey_id")
    fi
    
    # Load GOAL_ID if needed
    if [ -z "$GOAL_ID" ] && [ -f "$TEST_DATA_DIR/goal_id" ]; then
        GOAL_ID=$(cat "$TEST_DATA_DIR/goal_id")
    fi
    
    # Ensure we have JOURNEY_ID
    [ -n "$JOURNEY_ID" ] || skip "JOURNEY_ID not available"
    
    run "$BINARY" goal list "$JOURNEY_ID"
    [ "$status" -eq 0 ]
    
    # Verify the goal we created appears in the list
    [[ "$output" =~ "Arrival Goal" ]] || [[ "$output" =~ "$GOAL_ID" ]]
}

# Additional tests for verification and edge cases

@test "verify saved JOURNEY_ID is available" {
    # This test verifies that JOURNEY_ID was properly saved
    [ -f "$TEST_DATA_DIR/journey_id" ] || skip "journey_id file not found"
    JOURNEY_ID=$(cat "$TEST_DATA_DIR/journey_id")
    [ -n "$JOURNEY_ID" ] || skip "JOURNEY_ID not set in previous tests"
    echo "JOURNEY_ID: $JOURNEY_ID"
}

@test "verify saved GOAL_ID is available" {
    # This test verifies that GOAL_ID was properly saved
    [ -f "$TEST_DATA_DIR/goal_id" ] || skip "goal_id file not found"
    GOAL_ID=$(cat "$TEST_DATA_DIR/goal_id")
    [ -n "$GOAL_ID" ] || skip "GOAL_ID not set in previous tests"
    echo "GOAL_ID: $GOAL_ID"
}
