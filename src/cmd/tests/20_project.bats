#!/usr/bin/env bats
# 20_project.bats - Project listing and creation tests

load 00_env

# Test data
TEST_PROJECT_NAME="test-project-$(date +%s)"
TEST_PROJECT_DESC="Test project created by BATS"

setup() {
    # Inherit parent setup
    load 00_env
    setup
}

@test "project list command exists" {
    run "$BINARY" project list --help
    [ "$status" -eq 0 ]
}

@test "can list projects in JSON format" {
    run "$BINARY" project list --format json
    [ "$status" -eq 0 ]

    # Check JSON is valid and has at least one object
    echo "$output" | jq 'length > 0' > /dev/null
    [ $? -eq 0 ]
}

@test "project list supports pagination" {
    run "$BINARY" project list --limit 5 --offset 0
    [ "$status" -eq 0 ]
}

@test "project create command exists" {
    run "$BINARY" project create --help
    [ "$status" -eq 0 ]
}

@test "can create a new project and capture id" {
    run "$BINARY" project create "$TEST_PROJECT_NAME" --output json
    [ "$status" -eq 0 ]

    # Capture id and save it to shared test data for use in subsequent test files
    PROJECT_ID=$(echo "$output" | jq -r '.id')
    echo "$PROJECT_ID" > "$TEST_DATA_DIR/project_id"
    
    # Verify we got a valid ID
    [ -n "$PROJECT_ID" ]
}

@test "cannot create project with duplicate name" {
    # First create
    run "$BINARY" project create --name "$TEST_PROJECT_NAME-dup" --description "First"
    [ "$status" -eq 0 ]
    
    # Duplicate attempt
    run "$BINARY" project create --name "$TEST_PROJECT_NAME-dup" --description "Second"
    [ "$status" -ne 0 ]
    [[ "$output" =~ "exists" ]] || [[ "$output" =~ "duplicate" ]]
}

@test "new project id appears in list" {
    # Read project id from shared test data
    PROJECT_ID=$(cat "$TEST_DATA_DIR/project_id")

    run "$BINARY" project list --format json
    [ "$status" -eq 0 ]

    # Assert that the new id is in the list
    echo "$output" | jq -e ".[] | select(.id == \"$PROJECT_ID\")" > /dev/null
    [ $? -eq 0 ]
}

@test "can update project" {
    NEW_DESC="Updated description"
    run "$BINARY" project update --name "$TEST_PROJECT_NAME" --description "$NEW_DESC"
    [ "$status" -eq 0 ]
    
    # Verify update
    run "$BINARY" project get --name "$TEST_PROJECT_NAME"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "$NEW_DESC" ]]
}

@test "can delete project" {
    skip "Implement if your API supports project deletion"
    # run "$BINARY" project delete --name "$TEST_PROJECT_NAME" --force
    # [ "$status" -eq 0 ]
}

@test "project list filters work" {
    skip "Implement based on available filters"
    # run "$BINARY" project list --filter "name=$TEST_PROJECT_NAME"
    # [ "$status" -eq 0 ]
    # [[ "$output" =~ "$TEST_PROJECT_NAME" ]]
}
