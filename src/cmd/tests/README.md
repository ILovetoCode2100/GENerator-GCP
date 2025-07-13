# API CLI Test Suite

This directory contains the BATS (Bash Automated Testing System) test suite for the API CLI tool.

## Test Structure

The tests are organized into the following files:

- `00_env.bats` - Environment setup and binary verification
- `10_auth.bats` - Authentication and health check tests
- `20_project.bats` - Project CRUD operations
- `30_journey_goal.bats` - Journey and goal management
- `40_checkpoint.bats` - Checkpoint functionality
- `50_steps.bats` - Step management features
- `60_formats.bats` - Output format testing
- `70_session.bats` - Session handling
- `80_errors.bats` - Error scenarios and handling
- `99_report.bats` - Summary report generation

## Running Tests

### Prerequisites

Install BATS:
```bash
npm install -g bats
```

### Running Individual Test Files

```bash
bats 00_env.bats
```

### Running All Tests with Report Generation

Use the provided script to run all tests and generate a comprehensive report:

```bash
./generate_report.sh
```

This will:
- Run all BATS tests with TAP output format
- Parse the results
- Generate `report.md` with pass/fail counts and details
- Display a summary in the terminal

### CI/CD Integration

For CI pipelines, use the wrapper script:

```bash
./run_tests_with_report.sh
```

This script:
1. Runs all tests via `generate_report.sh`
2. Executes `99_report.bats` to display aggregated results in CI logs
3. Exits with appropriate status code (0 for success, 1 for failures)

## Test Report

After running `generate_report.sh`, a markdown report (`report.md`) is generated containing:

- Total test counts (passed, failed, skipped)
- Pass rate percentage
- Status of each test file
- Details of any failed tests
- Test category descriptions
- Test environment information

## TAP Output Format

The tests use TAP (Test Anything Protocol) format when run with `bats --tap`. Example output:

```
1..5
ok 1 binary exists and is executable
ok 2 binary shows version
ok 3 binary shows help
ok 4 config.sh loaded successfully
ok 5 required environment variables are set # skip Add environment variable checks as needed
```

## Writing New Tests

Follow the BATS syntax:

```bash
@test "description of test" {
    run "$BINARY" command args
    [ "$status" -eq 0 ]
    [[ "$output" =~ "expected output" ]]
}
```

## Troubleshooting

If BATS is not found:
- Install with: `npm install -g bats`
- Or use local installation: `npm install --save-dev bats`

If tests fail to find the binary:
- Ensure the binary is built: `make build` from project root
- Check the binary path in `00_env.bats`
