#!/usr/bin/env bats
# 70_session.bats - Session context persistence and auto-increment tests

load 00_env

# Test data
TEST_PROJECT="session-test-project-$(date +%s)"
TEST_JOURNEY="session-test-journey-$(date +%s)"
TEST_GOAL="session-test-goal-$(date +%s)"
TEST_CHECKPOINT="session-test-checkpoint-$(date +%s)"
CONFIG_FILE="$HOME/.api-cli/virtuoso-config.yaml"

setup() {
    load 00_env
    setup
    
    # Clean up any existing config to ensure fresh start
    rm -f "$CONFIG_FILE"
    
    # Create test hierarchy
    run "$BINARY" project create --name "$TEST_PROJECT" --description "Session test project"
    run "$BINARY" journey create --project "$TEST_PROJECT" --name "$TEST_JOURNEY" --description "Test journey"
    run "$BINARY" goal create --project "$TEST_PROJECT" --journey "$TEST_JOURNEY" --name "$TEST_GOAL" --description "Test goal"
    run "$BINARY" checkpoint create --project "$TEST_PROJECT" --journey "$TEST_JOURNEY" --goal "$TEST_GOAL" --name "$TEST_CHECKPOINT" --description "Test checkpoint"
    
    # Set checkpoint in session
    run "$BINARY" set-checkpoint "$TEST_CHECKPOINT"
}

teardown() {
    # Clean up config file
    rm -f "$CONFIG_FILE"
}

@test "session persistence between invocations - auto-increment test" {
    # First, let's verify the config file exists after setup
    [[ -f "$CONFIG_FILE" ]]
    
    # Open a fresh shell, source config, and create first step
    run bash -c "
        # Source the config if needed
        if [[ -f '$PROJECT_ROOT/config.sh' ]]; then
            source '$PROJECT_ROOT/config.sh'
        fi
        
        # Create first step with session context
        '$BINARY' step create --session --type navigation --selector '#nav' --value 'https://example.com'
    "
    [ "$status" -eq 0 ]
    [[ "$output" =~ "created" ]] || [[ "$output" =~ "navigation" ]]
    
    # Get the position after first step
    run bash -c "cat '$CONFIG_FILE' | yq '.session.next_position'"
    [ "$status" -eq 0 ]
    first_position="$output"
    
    # Open another fresh shell, source config, and create second step
    run bash -c "
        # Source the config if needed
        if [[ -f '$PROJECT_ROOT/config.sh' ]]; then
            source '$PROJECT_ROOT/config.sh'
        fi
        
        # Create second step with session context
        '$BINARY' step create --session --type click --selector '#button' --value 'Submit'
    "
    [ "$status" -eq 0 ]
    [[ "$output" =~ "created" ]] || [[ "$output" =~ "click" ]]
    
    # Get the position after second step
    run bash -c "cat '$CONFIG_FILE' | yq '.session.next_position'"
    [ "$status" -eq 0 ]
    second_position="$output"
    
    # Assert that second position equals first position + 1
    # The first step should have position 1, next_position should be 2
    # The second step should have position 2, next_position should be 3
    expected_second_position=$((first_position + 1))
    [ "$second_position" -eq "$expected_second_position" ]
}

@test "session context persists across multiple fresh shell invocations" {
    # Create multiple steps in separate shell invocations
    local expected_position=1
    
    for step_type in "navigation" "click" "write" "assert-exists"; do
        run bash -c "
            if [[ -f '$PROJECT_ROOT/config.sh' ]]; then
                source '$PROJECT_ROOT/config.sh'
            fi
            
            '$BINARY' step create --session --type $step_type --selector '.test' --value 'test-value'
        "
        [ "$status" -eq 0 ]
        
        # Verify position incremented correctly
        run bash -c "cat '$CONFIG_FILE' | yq '.session.next_position'"
        [ "$status" -eq 0 ]
        expected_position=$((expected_position + 1))
        [ "$output" -eq "$expected_position" ]
    done
}

@test "session maintains checkpoint context between invocations" {
    # Verify checkpoint is persisted in config
    run bash -c "cat '$CONFIG_FILE' | yq '.session.checkpoint_id'"
    [ "$status" -eq 0 ]
    [[ -n "$output" ]]
    checkpoint_id="$output"
    
    # Create a step in a fresh shell to verify checkpoint is used
    run bash -c "
        if [[ -f '$PROJECT_ROOT/config.sh' ]]; then
            source '$PROJECT_ROOT/config.sh'
        fi
        
        '$BINARY' step create --session --type comment --selector '' --value 'Testing checkpoint persistence'
    "
    [ "$status" -eq 0 ]
    
    # Verify checkpoint ID is still the same
    run bash -c "cat '$CONFIG_FILE' | yq '.session.checkpoint_id'"
    [ "$status" -eq 0 ]
    [ "$output" = "$checkpoint_id" ]
}

@test "session config file is created if not exists" {
    # Remove config file
    rm -f "$CONFIG_FILE"
    
    # Create a step which should create the config file
    run bash -c "
        if [[ -f '$PROJECT_ROOT/config.sh' ]]; then
            source '$PROJECT_ROOT/config.sh'
        fi
        
        # Set checkpoint first
        '$BINARY' set-checkpoint '$TEST_CHECKPOINT'
        
        # Then create step
        '$BINARY' step create --session --type navigation --selector '#home' --value 'https://home.com'
    "
    [ "$status" -eq 0 ]
    
    # Verify config file was created
    [[ -f "$CONFIG_FILE" ]]
    
    # Verify it has the expected structure
    run bash -c "cat '$CONFIG_FILE' | yq '.session.next_position'"
    [ "$status" -eq 0 ]
    [[ "$output" =~ ^[0-9]+$ ]]
}

@test "position counter continues from last value after restart" {
    # Create several steps to establish a position
    for i in {1..3}; do
        run "$BINARY" step create --session --type comment --selector "" --value "Step $i"
        [ "$status" -eq 0 ]
    done
    
    # Get current position
    run bash -c "cat '$CONFIG_FILE' | yq '.session.next_position'"
    [ "$status" -eq 0 ]
    last_position="$output"
    
    # Simulate a restart by creating step in fresh shell
    run bash -c "
        if [[ -f '$PROJECT_ROOT/config.sh' ]]; then
            source '$PROJECT_ROOT/config.sh'
        fi
        
        '$BINARY' step create --session --type comment --selector '' --value 'After restart'
    "
    [ "$status" -eq 0 ]
    
    # Verify position continued from where it left off
    run bash -c "cat '$CONFIG_FILE' | yq '.session.next_position'"
    [ "$status" -eq 0 ]
    expected_position=$((last_position + 1))
    [ "$output" -eq "$expected_position" ]
}
