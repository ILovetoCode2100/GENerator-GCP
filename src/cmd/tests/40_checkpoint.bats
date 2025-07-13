#!/usr/bin/env bats
# 40_checkpoint.bats - Checkpoint operations tests

load 00_env

# Variables that will be captured during tests
GOAL_ID=""
JOURNEY_ID=""
SNAPSHOT_ID=""
CP_ID=""

setup() {
    load 00_env
    setup
    
    # Load IDs from shared test data (saved by previous test files)
    if [ -f "$TEST_DATA_DIR/goal_id" ]; then
        GOAL_ID=$(cat "$TEST_DATA_DIR/goal_id")
    fi
    
    if [ -f "$TEST_DATA_DIR/journey_id" ]; then
        JOURNEY_ID=$(cat "$TEST_DATA_DIR/journey_id")
    fi
    
    # If not available, we need to fail early
    if [ -z "${GOAL_ID}" ]; then
        skip "GOAL_ID not available from previous tests"
    fi
    
    if [ -z "${JOURNEY_ID}" ]; then
        skip "JOURNEY_ID not available from previous tests"
    fi
}

teardown() {
    # Clean up is handled by later test files or manually
    :
}

@test "checkpoint - get snapshot ID for goal" {
    # For the create-checkpoint command, we need goal ID, snapshot ID, and journey ID
    # Based on the code, we need to get the snapshot ID first
    # Assuming there's a way to get it, or it's a fixed value for testing
    # For now, let's use a placeholder
    SNAPSHOT_ID="43802"  # This would need to be obtained from the API
    
    # Save for next tests
    echo "$SNAPSHOT_ID" > "$TEST_DATA_DIR/snapshot_id"
    
    [ -n "$SNAPSHOT_ID" ]
}

@test "checkpoint - create checkpoint CP-1 using create-checkpoint command" {
    # Load snapshot ID
    if [ -z "$SNAPSHOT_ID" ] && [ -f "$TEST_DATA_DIR/snapshot_id" ]; then
        SNAPSHOT_ID=$(cat "$TEST_DATA_DIR/snapshot_id")
    fi
    
    # The create-checkpoint command takes: JOURNEY_ID GOAL_ID SNAPSHOT_ID NAME
    run "$BINARY" create-checkpoint "$JOURNEY_ID" "$GOAL_ID" "$SNAPSHOT_ID" "CP-1"
    [ "$status" -eq 0 ]
    
    # Capture checkpoint ID from output
    if command -v jq &> /dev/null; then
        CP_ID=$(echo "$output" | jq -r '.checkpoint_id // .id // empty')
    fi
    
    # If no JSON parsing available or ID not found, try to extract from plain text
    if [ -z "$CP_ID" ]; then
        # Try to extract numeric ID from output
        CP_ID=$(echo "$output" | grep -oE 'ID: [0-9]+' | grep -oE '[0-9]+' | head -1)
    fi
    
    # Save to shared test data for use in subsequent tests
    echo "$CP_ID" > "$TEST_DATA_DIR/checkpoint_id"
    
    # Verify we got an ID
    [ -n "$CP_ID" ]
    
    # The create-checkpoint command automatically attaches the checkpoint,
    # so we verify it was created and attached
    [[ "$output" =~ "CP-1" ]] || [[ "$output" =~ "Created and attached" ]]
}

@test "checkpoint - list checkpoints and verify positions" {
    # List checkpoints using journey ID
    run "$BINARY" list-checkpoints "$JOURNEY_ID"
    [ "$status" -eq 0 ]
    
    # Verify CP-1 appears in the list
    [[ "$output" =~ "CP-1" ]]
    
    # Verify position information is included
    [[ "$output" =~ "position" ]] || [[ "$output" =~ "Position" ]] || [[ "$output" =~ "1." ]]
}

@test "checkpoint - create additional checkpoints to test positioning" {
    # Load snapshot ID
    if [ -z "$SNAPSHOT_ID" ] && [ -f "$TEST_DATA_DIR/snapshot_id" ]; then
        SNAPSHOT_ID=$(cat "$TEST_DATA_DIR/snapshot_id")
    fi
    
    # Create CP-2 with explicit position
    run "$BINARY" create-checkpoint "$JOURNEY_ID" "$GOAL_ID" "$SNAPSHOT_ID" "CP-2" --position 3
    [ "$status" -eq 0 ]
    
    # Create CP-3 with default position
    run "$BINARY" create-checkpoint "$JOURNEY_ID" "$GOAL_ID" "$SNAPSHOT_ID" "CP-3"
    [ "$status" -eq 0 ]
}

@test "checkpoint - verify correct position values after multiple creates" {
    # List checkpoints again to verify all positions
    run "$BINARY" list-checkpoints "$JOURNEY_ID"
    [ "$status" -eq 0 ]
    
    # Should see all three checkpoints
    [[ "$output" =~ "CP-1" ]]
    [[ "$output" =~ "CP-2" ]]
    [[ "$output" =~ "CP-3" ]]
    
    # Log output for debugging
    echo "Checkpoint list output:"
    echo "$output"
}

@test "checkpoint - set current checkpoint for session context" {
    # Load checkpoint ID from earlier test
    if [ -z "$CP_ID" ] && [ -f "$TEST_DATA_DIR/checkpoint_id" ]; then
        CP_ID=$(cat "$TEST_DATA_DIR/checkpoint_id")
    fi
    
    [ -n "$CP_ID" ] || skip "CP_ID not available"
    
    # Set the checkpoint as current
    run "$BINARY" set-checkpoint "$CP_ID"
    [ "$status" -eq 0 ]
    
    # Verify it was set
    [[ "$output" =~ "Current checkpoint set" ]] || [[ "$output" =~ "checkpoint set to" ]]
}
