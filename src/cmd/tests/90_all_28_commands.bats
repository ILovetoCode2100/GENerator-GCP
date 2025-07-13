#!/usr/bin/env bats

# Test file for all 28 CLI commands to ensure comprehensive coverage

load test_helper

@test "create-step-cookie-create command exists and runs" {
  run $API_CLI create-step-cookie-create --help
  assert_success
  assert_output --partial "Create a cookie"
}

@test "create-step-cookie-wipe-all command exists and runs" {
  run $API_CLI create-step-cookie-wipe-all --help
  assert_success
  assert_output --partial "Wipe all cookies"
}

@test "create-step-upload-url command exists and runs" {
  run $API_CLI create-step-upload-url --help
  assert_success
  assert_output --partial "Upload file from URL"
}

@test "create-step-mouse-move-to command exists and runs" {
  run $API_CLI create-step-mouse-move-to --help
  assert_success
  assert_output --partial "Move mouse to coordinates"
}

@test "create-step-mouse-move-by command exists and runs" {
  run $API_CLI create-step-mouse-move-by --help
  assert_success
  assert_output --partial "Move mouse by offset"
}

@test "create-step-switch-next-tab command exists and runs" {
  run $API_CLI create-step-switch-next-tab --help
  assert_success
  assert_output --partial "Switch to next tab"
}

@test "create-step-switch-prev-tab command exists and runs" {
  run $API_CLI create-step-switch-prev-tab --help
  assert_success
  assert_output --partial "Switch to previous tab"
}

@test "create-step-switch-parent-frame command exists and runs" {
  run $API_CLI create-step-switch-parent-frame --help
  assert_success
  assert_output --partial "Switch to parent frame"
}

@test "create-step-switch-iframe command exists and runs" {
  run $API_CLI create-step-switch-iframe --help
  assert_success
  assert_output --partial "Switch to iframe"
}

@test "create-step-execute-script command exists and runs" {
  run $API_CLI create-step-execute-script --help
  assert_success
  assert_output --partial "Execute script"
}

@test "create-step-pick-index command exists and runs" {
  run $API_CLI create-step-pick-index --help
  assert_success
  assert_output --partial "Pick option by index"
}

@test "create-step-pick-last command exists and runs" {
  run $API_CLI create-step-pick-last --help
  assert_success
  assert_output --partial "Pick last option"
}

@test "create-step-wait-for-element-timeout command exists and runs" {
  run $API_CLI create-step-wait-for-element-timeout --help
  assert_success
  assert_output --partial "Wait for element with timeout"
}

@test "create-step-wait-for-element-default command exists and runs" {
  run $API_CLI create-step-wait-for-element-default --help
  assert_success
  assert_output --partial "Wait for element"
}

@test "create-step-store-element-text command exists and runs" {
  run $API_CLI create-step-store-element-text --help
  assert_success
  assert_output --partial "Store element text"
}

@test "create-step-store-literal-value command exists and runs" {
  run $API_CLI create-step-store-literal-value --help
  assert_success
  assert_output --partial "Store literal value"
}

@test "create-step-assert-not-equals command exists and runs" {
  run $API_CLI create-step-assert-not-equals --help
  assert_success
  assert_output --partial "Assert not equals"
}

@test "create-step-assert-greater-than command exists and runs" {
  run $API_CLI create-step-assert-greater-than --help
  assert_success
  assert_output --partial "Assert greater than"
}

@test "create-step-assert-greater-than-or-equal command exists and runs" {
  run $API_CLI create-step-assert-greater-than-or-equal --help
  assert_success
  assert_output --partial "Assert greater than or equal"
}

@test "create-step-assert-matches command exists and runs" {
  run $API_CLI create-step-assert-matches --help
  assert_success
  assert_output --partial "Assert matches"
}

@test "create-step-dismiss-prompt-with-text command exists and runs" {
  run $API_CLI create-step-dismiss-prompt-with-text --help
  assert_success
  assert_output --partial "Dismiss prompt"
}

@test "create-step-navigate command exists and runs" {
  run $API_CLI create-step-navigate --help
  assert_success
  assert_output --partial "Navigate to URL"
}

@test "create-step-click command exists and runs" {
  run $API_CLI create-step-click --help
  assert_success
  assert_output --partial "Click element"
}

@test "create-step-write command exists and runs" {
  run $API_CLI create-step-write --help
  assert_success
  assert_output --partial "Write text"
}

@test "create-step-scroll-to-position command exists and runs" {
  run $API_CLI create-step-scroll-to-position --help
  assert_success
  assert_output --partial "Scroll to position"
}

@test "create-step-scroll-by-offset command exists and runs" {
  run $API_CLI create-step-scroll-by-offset --help
  assert_success
  assert_output --partial "Scroll by offset"
}

@test "create-step-scroll-to-top command exists and runs" {
  run $API_CLI create-step-scroll-to-top --help
  assert_success
  assert_output --partial "Scroll to top"
}

@test "create-step-window-resize command exists and runs" {
  run $API_CLI create-step-window-resize --help
  assert_success
  assert_output --partial "Resize window"
}

@test "create-step-key command exists and runs" {
  run $API_CLI create-step-key --help
  assert_success
  assert_output --partial "Press key"
}

@test "create-step-comment command exists and runs" {
  run $API_CLI create-step-comment --help
  assert_success
  assert_output --partial "Add comment"
}

# Integration test with mock checkpoint ID
@test "all commands handle missing checkpoint ID appropriately" {
  # Test a few representative commands without checkpoint ID
  run $API_CLI create-step-navigate
  assert_failure
  assert_output --partial "checkpoint"

  run $API_CLI create-step-click
  assert_failure
  assert_output --partial "checkpoint"

  run $API_CLI create-step-write
  assert_failure
  assert_output --partial "checkpoint"
}

# Test output formats are supported
@test "commands support multiple output formats" {
  run $API_CLI create-step-navigate --help
  assert_success
  assert_output --partial "--output"
  assert_output --partial "json"
  assert_output --partial "yaml"
  assert_output --partial "human"
}
