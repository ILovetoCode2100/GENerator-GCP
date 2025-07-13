#!/usr/bin/env bats
# 50_steps.bats - All step creation commands with modern session pattern

load 00_env

# Test data
TEST_PROJECT="steps-test-project-$(date +%s)"
TEST_JOURNEY="steps-test-journey-$(date +%s)"
TEST_GOAL="steps-test-goal-$(date +%s)"
TEST_CHECKPOINT="steps-test-checkpoint-$(date +%s)"
TEST_CHECKPOINT_ID=""

setup() {
    load 00_env
    setup
    
    # Create test hierarchy
    run "$BINARY" project create --name "$TEST_PROJECT" --description "Steps test project"
    run "$BINARY" journey create --project "$TEST_PROJECT" --name "$TEST_JOURNEY" --description "Test journey"
    run "$BINARY" goal create --project "$TEST_PROJECT" --journey "$TEST_JOURNEY" --name "$TEST_GOAL" --description "Test goal"
    run "$BINARY" checkpoint create --project "$TEST_PROJECT" --journey "$TEST_JOURNEY" --goal "$TEST_GOAL" --name "$TEST_CHECKPOINT" --description "Test checkpoint"
    
    # Extract checkpoint ID from the output
    TEST_CHECKPOINT_ID=$(echo "$output" | grep -oE "checkpoint.*created.*ID:? ?([0-9]+)" | grep -oE "[0-9]+$" || echo "1")
    
    # Set up session context
    run "$BINARY" set-checkpoint "$TEST_CHECKPOINT_ID"
}

teardown() {
    # Clean up session context
    rm -f ~/.api-cli/virtuoso-config.yaml
    :
}

@test "step create command exists" {
    run "$BINARY" step create --help
    [ "$status" -eq 0 ]
}

@test "can create a basic step" {
    run "$BINARY" step create \
        --project "$TEST_PROJECT" \
        --journey "$TEST_JOURNEY" \
        --goal "$TEST_GOAL" \
        --checkpoint "$TEST_CHECKPOINT" \
        --name "basic-step" \
        --description "Basic step test"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "created" ]] || [[ "$output" =~ "basic-step" ]]
}

@test "can create step with code block" {
    run "$BINARY" step create \
        --project "$TEST_PROJECT" \
        --journey "$TEST_JOURNEY" \
        --goal "$TEST_GOAL" \
        --checkpoint "$TEST_CHECKPOINT" \
        --name "code-step" \
        --description "Step with code" \
        --code 'echo "Hello World"' \
        --language "bash"
    [ "$status" -eq 0 ]
}

@test "can create step with file reference" {
    # Create a test file
    echo "test content" > "$TEST_DATA_DIR/test-file.txt"
    
    run "$BINARY" step create \
        --project "$TEST_PROJECT" \
        --journey "$TEST_JOURNEY" \
        --goal "$TEST_GOAL" \
        --checkpoint "$TEST_CHECKPOINT" \
        --name "file-step" \
        --description "Step with file" \
        --file "$TEST_DATA_DIR/test-file.txt"
    [ "$status" -eq 0 ]
}

@test "can create step with multiple tags" {
    run "$BINARY" step create \
        --project "$TEST_PROJECT" \
        --journey "$TEST_JOURNEY" \
        --goal "$TEST_GOAL" \
        --checkpoint "$TEST_CHECKPOINT" \
        --name "tagged-step" \
        --description "Step with tags" \
        --tag "backend" \
        --tag "api" \
        --tag "test"
    [ "$status" -eq 0 ]
}

@test "can create step with metadata" {
    run "$BINARY" step create \
        --project "$TEST_PROJECT" \
        --journey "$TEST_JOURNEY" \
        --goal "$TEST_GOAL" \
        --checkpoint "$TEST_CHECKPOINT" \
        --name "metadata-step" \
        --description "Step with metadata" \
        --metadata "priority=high" \
        --metadata "type=implementation"
    [ "$status" -eq 0 ]
}

@test "can list steps in checkpoint" {
    # Create a few steps first
    for i in {1..3}; do
        run "$BINARY" step create \
            --project "$TEST_PROJECT" \
            --journey "$TEST_JOURNEY" \
            --goal "$TEST_GOAL" \
            --checkpoint "$TEST_CHECKPOINT" \
            --name "step-$i" \
            --description "Step $i"
    done
    
    run "$BINARY" step list \
        --project "$TEST_PROJECT" \
        --journey "$TEST_JOURNEY" \
        --goal "$TEST_GOAL" \
        --checkpoint "$TEST_CHECKPOINT"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "step-1" ]]
    [[ "$output" =~ "step-2" ]]
    [[ "$output" =~ "step-3" ]]
}

@test "can get step details" {
    run "$BINARY" step create \
        --project "$TEST_PROJECT" \
        --journey "$TEST_JOURNEY" \
        --goal "$TEST_GOAL" \
        --checkpoint "$TEST_CHECKPOINT" \
        --name "detail-step" \
        --description "Step for detail test"
    
    run "$BINARY" step get \
        --project "$TEST_PROJECT" \
        --journey "$TEST_JOURNEY" \
        --goal "$TEST_GOAL" \
        --checkpoint "$TEST_CHECKPOINT" \
        --step "detail-step"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "detail-step" ]]
    [[ "$output" =~ "Step for detail test" ]]
}

@test "can update step" {
    run "$BINARY" step create \
        --project "$TEST_PROJECT" \
        --journey "$TEST_JOURNEY" \
        --goal "$TEST_GOAL" \
        --checkpoint "$TEST_CHECKPOINT" \
        --name "update-step" \
        --description "Original description"
    
    run "$BINARY" step update \
        --project "$TEST_PROJECT" \
        --journey "$TEST_JOURNEY" \
        --goal "$TEST_GOAL" \
        --checkpoint "$TEST_CHECKPOINT" \
        --step "update-step" \
        --description "Updated description"
    [ "$status" -eq 0 ]
}

@test "can mark step as complete" {
    skip "Implement step completion"
    # run "$BINARY" step complete \
    #     --project "$TEST_PROJECT" \
    #     --journey "$TEST_JOURNEY" \
    #     --goal "$TEST_GOAL" \
    #     --checkpoint "$TEST_CHECKPOINT" \
    #     --step "basic-step"
    # [ "$status" -eq 0 ]
}

@test "step creation validates required fields" {
    # Missing checkpoint
    run "$BINARY" step create \
        --project "$TEST_PROJECT" \
        --journey "$TEST_JOURNEY" \
        --goal "$TEST_GOAL" \
        --name "invalid-step"
    [ "$status" -ne 0 ]
}

# Modern session pattern tests
@test "navigation step with session context" {
    run "$BINARY" step create --session --type navigation --selector "#nav" --value "https://example.com"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "created" ]] || [[ "$output" =~ "navigation" ]]
    
    # Check position auto-incremented
    run bash -c "cat ~/.api-cli/virtuoso-config.yaml | yq '.session.next_position'"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "2" ]]
}

@test "interaction step (click) with session context" {
    run "$BINARY" step create --session --type click --selector "#button" --value "foo"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "created" ]] || [[ "$output" =~ "click" ]]
    
    # Check position auto-incremented
    run bash -c "cat ~/.api-cli/virtuoso-config.yaml | yq '.session.next_position'"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "3" ]]
}

@test "interaction step (write) with session context" {
    run "$BINARY" step create --session --type write --selector "#input" --value "test text"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "created" ]] || [[ "$output" =~ "write" ]]
    
    # Check position auto-incremented
    run bash -c "cat ~/.api-cli/virtuoso-config.yaml | yq '.session.next_position'"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "4" ]]
}

@test "assertion step (assert-exists) with session context" {
    run "$BINARY" step create --session --type assert-exists --selector "#element" --value "foo"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "created" ]] || [[ "$output" =~ "assert" ]]
    
    # Check position auto-incremented
    run bash -c "cat ~/.api-cli/virtuoso-config.yaml | yq '.session.next_position'"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "5" ]]
}

@test "assertion step (assert-equals) with session context" {
    run "$BINARY" step create --session --type assert-equals --selector "#element" --value "expected value"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "created" ]] || [[ "$output" =~ "assert" ]]
    
    # Check position auto-incremented
    run bash -c "cat ~/.api-cli/virtuoso-config.yaml | yq '.session.next_position'"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "6" ]]
}

@test "wait step with session context" {
    run "$BINARY" step create --session --type wait-element --selector "#element" --value "foo"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "created" ]] || [[ "$output" =~ "wait" ]]
    
    # Check position auto-incremented
    run bash -c "cat ~/.api-cli/virtuoso-config.yaml | yq '.session.next_position'"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "7" ]]
}

@test "store step with session context" {
    run "$BINARY" step create --session --type store --selector "#element" --value "variable_name"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "created" ]] || [[ "$output" =~ "store" ]]
    
    # Check position auto-incremented
    run bash -c "cat ~/.api-cli/virtuoso-config.yaml | yq '.session.next_position'"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "8" ]]
}

@test "hover step with session context" {
    run "$BINARY" step create --session --type hover --selector "#element" --value "foo"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "created" ]] || [[ "$output" =~ "hover" ]]
    
    # Check position auto-incremented
    run bash -c "cat ~/.api-cli/virtuoso-config.yaml | yq '.session.next_position'"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "9" ]]
}

@test "scroll step with session context" {
    run "$BINARY" step create --session --type scroll --selector "#element" --value "bottom"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "created" ]] || [[ "$output" =~ "scroll" ]]
    
    # Check position auto-incremented
    run bash -c "cat ~/.api-cli/virtuoso-config.yaml | yq '.session.next_position'"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "10" ]]
}

@test "execute-js step with session context" {
    run "$BINARY" step create --session --type execute-js --selector "body" --value "console.log('test')"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "created" ]] || [[ "$output" =~ "execute" ]]
    
    # Check position auto-incremented
    run bash -c "cat ~/.api-cli/virtuoso-config.yaml | yq '.session.next_position'"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "11" ]]
}

@test "comment step with session context" {
    run "$BINARY" step create --session --type comment --selector "" --value "This is a comment"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "created" ]] || [[ "$output" =~ "comment" ]]
    
    # Check position auto-incremented
    run bash -c "cat ~/.api-cli/virtuoso-config.yaml | yq '.session.next_position'"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "12" ]]
}

@test "key step with session context" {
    run "$BINARY" step create --session --type key --selector "#input" --value "Enter"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "created" ]] || [[ "$output" =~ "key" ]]
    
    # Check position auto-incremented
    run bash -c "cat ~/.api-cli/virtuoso-config.yaml | yq '.session.next_position'"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "13" ]]
}

@test "can delete step" {
    skip "Implement if API supports step deletion"
    # run "$BINARY" step delete \
    #     --project "$TEST_PROJECT" \
    #     --journey "$TEST_JOURNEY" \
    #     --goal "$TEST_GOAL" \
    #     --checkpoint "$TEST_CHECKPOINT" \
    #     --step "basic-step" \
    #     --force
    # [ "$status" -eq 0 ]
}
