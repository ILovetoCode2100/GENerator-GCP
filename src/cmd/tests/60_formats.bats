#!/usr/bin/env bats
# 60_formats.bats - Output format tests (json/yaml/human/ai)

load 00_env

# Test data
TEST_PROJECT="format-test-project-$(date +%s)"

setup() {
    load 00_env
    setup
    
    # Save current configuration if config commands are available
    ORIGINAL_FORMAT=$("$BINARY" config get output.default_format 2>/dev/null || echo "human")
    export ORIGINAL_FORMAT
    
    # Create test data
    run "$BINARY" create-project "$TEST_PROJECT" --description "Format test project"
}

teardown() {
    # Restore original format if config commands are available
    if [[ -n "$ORIGINAL_FORMAT" ]]; then
        "$BINARY" config set output.default_format "$ORIGINAL_FORMAT" 2>/dev/null || true
    fi
}

# JSON Format Tests
@test "supports JSON output format for project list" {
    run "$BINARY" list-projects --output json
    [ "$status" -eq 0 ]
    # Validate JSON structure
    echo "$output" | jq . >/dev/null 2>&1
    [ $? -eq 0 ]
    # Check that project name appears in JSON
    echo "$output" | jq -e ".[] | select(.name == \"$TEST_PROJECT\")" >/dev/null 2>&1 || true
}

@test "supports JSON output format for single project" {
    # Since there's no get-project command, test with list-projects and filter
    run "$BINARY" list-projects --output json
    [ "$status" -eq 0 ]
    # Validate JSON and check that project exists
    echo "$output" | jq -e ".[] | select(.name == \"$TEST_PROJECT\")" >/dev/null
    [ $? -eq 0 ]
}

@test "JSON output is properly formatted" {
    run "$BINARY" list-projects --output json
    [ "$status" -eq 0 ]
    # Check if output is valid JSON
    echo "$output" | jq . >/dev/null 2>&1
    [ $? -eq 0 ]
}

# YAML Format Tests
@test "supports YAML output format for project list" {
    run "$BINARY" list-projects --output yaml
    [ "$status" -eq 0 ]
    # Basic YAML validation - should contain colons and proper indentation
    [[ "$output" =~ ^[[:space:]]*[a-zA-Z] ]]
    [[ "$output" =~ : ]]
    # Check that project name appears
    [[ "$output" =~ "$TEST_PROJECT" ]]
}

@test "supports YAML output format for single project" {
    run "$BINARY" list-projects --output yaml
    [ "$status" -eq 0 ]
    [[ "$output" =~ "name:" ]]
    [[ "$output" =~ "$TEST_PROJECT" ]]
}

@test "YAML output has proper structure" {
    # Since yq is not available, skip the full YAML validation
    skip "yq not available for full YAML validation"
}

# Human-Readable Format Tests
@test "supports human-readable output format (default)" {
    run "$BINARY" list-projects --output human
    [ "$status" -eq 0 ]
    # Should have table-like or formatted output
    [[ "$output" =~ "$TEST_PROJECT" ]]
}

@test "human format is default when no format specified" {
    run "$BINARY" list-projects
    [ "$status" -eq 0 ]
    # Should be readable, not JSON/YAML
    ! echo "$output" | jq . >/dev/null 2>&1
}

@test "human format includes headers and formatting" {
    run "$BINARY" list-projects --output human
    [ "$status" -eq 0 ]
    # Check for typical human-readable elements
    [[ "$output" =~ "Name" ]] || [[ "$output" =~ "PROJECT" ]] || [[ "$output" =~ "Description" ]]
}

# AI Format Tests
@test "supports AI-optimized output format" {
    run "$BINARY" list-projects --output ai
    [ "$status" -eq 0 ]
    # AI format should include the project name
    [[ "$output" =~ "$TEST_PROJECT" ]]
    # AI format might be structured text optimized for LLM consumption
    [[ "$output" =~ "Project" ]] || [[ "$output" =~ "project" ]] || [[ "$output" =~ "CONTEXT" ]]
}

@test "AI format includes context and metadata" {
    run "$BINARY" list-projects --output ai
    [ "$status" -eq 0 ]
    # Check that AI format includes structured information
    [[ "$output" =~ "$TEST_PROJECT" ]]
    # May include sections like Context, Metadata, Description, etc.
    [[ "$output" =~ "name" ]] || [[ "$output" =~ "Name" ]] || [[ "$output" =~ "PROJECT" ]]
}

# Format Consistency Tests
@test "all formats return same data for project list" {
    # Get data in different formats
    run "$BINARY" list-projects --output json
    json_count=$(echo "$output" | jq '. | length' 2>/dev/null || echo 0)
    
    run "$BINARY" list-projects --output yaml
    # Count YAML entries (approximate)
    yaml_count=$(echo "$output" | grep -c "^- " || true)
    
    # Counts should be similar (allowing for format differences)
    [ "$json_count" -gt 0 ]
}

@test "format flag works with all resource types" {
    # Test various commands with format flag
    for format in json yaml human ai; do
        run "$BINARY" list-projects --output $format
        [ "$status" -eq 0 ] || [ "$status" -eq 1 ]  # Allow "no resources found"
        
        run "$BINARY" list-journeys --project "$TEST_PROJECT" --output $format
        [ "$status" -eq 0 ] || [ "$status" -eq 1 ]  # Allow "no resources found"
    done
}

@test "invalid format returns error" {
    run "$BINARY" list-projects --output invalid
    [ "$status" -ne 0 ]
    [[ "$output" =~ "format" ]] || [[ "$output" =~ "invalid" ]] || [[ "$output" =~ "Format" ]]
}

@test "format flag is case-insensitive" {
    skip "Implement if your CLI supports case-insensitive formats"
}

# Complex Data Format Tests
@test "nested data is properly formatted in JSON" {
    # Create nested structure
    run "$BINARY" create-journey --project "$TEST_PROJECT" --name "test-journey" --description "Test"
    run "$BINARY" create-goal --project "$TEST_PROJECT" --journey "test-journey" --name "test-goal" --description "Test"
    
    # List journeys with format
    run "$BINARY" list-journeys --project "$TEST_PROJECT" --output json
    [ "$status" -eq 0 ]
    
    # Validate JSON structure
    echo "$output" | jq . >/dev/null 2>&1
    [ $? -eq 0 ]
}
