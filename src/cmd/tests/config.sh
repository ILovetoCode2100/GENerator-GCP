#!/usr/bin/env bash
# config.sh - Test configuration

# API CLI binary location
export API_CLI_BIN="${API_CLI_BIN:-./bin/api-cli}"

# Test prefix for resource tagging
export TEST_TAG_PREFIX="${TEST_TAG_PREFIX:-test-$$}"

# Timeout for long operations
export TEST_TIMEOUT="${TEST_TIMEOUT:-120}"

# Enable debug output if requested
if [ "${DEBUG}" = "true" ]; then
    set -x
fi