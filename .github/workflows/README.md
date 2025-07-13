# GitHub Actions Workflow Documentation

## Overview

This directory contains GitHub Actions workflows for the Virtuoso API CLI Generator project.

## Workflows

### test.yml

The main test workflow that runs on push and pull requests to `main` and `develop` branches.

#### Jobs

1. **test** - Runs the main test suite
   - Checks out code
   - Sets up Go 1.21
   - Caches Go modules
   - Installs dependencies (bats, jq, oapi-codegen)
   - Builds the CLI for testing
   - Runs BATS integration tests
   - Uploads test artifacts (report.md and api-cli.log)

2. **cleanup** - Runs after tests to clean up test resources
   - Always runs, even if tests fail
   - Builds the CLI
   - Deletes all test resources created with the test run tag

#### Required Secrets

The following secrets must be configured in your GitHub repository:

- `VIRTUOSO_API_KEY` - API key for accessing Virtuoso services
- `VIRTUOSO_API_URL` - Base URL for the Virtuoso API

To set these secrets:
1. Go to your repository on GitHub
2. Navigate to Settings > Secrets and variables > Actions
3. Click "New repository secret"
4. Add each secret with the appropriate value

#### Environment Variables

- `GITHUB_RUN_ID` - Automatically set by GitHub Actions
- `TEST_TAG_PREFIX` - Set to `test-${GITHUB_RUN_ID}` for tagging test resources

#### Artifacts

The workflow uploads the following artifacts:
- `test-report` - Contains the test execution report (src/cmd/tests/report.md)
- `api-cli-log` - Contains the CLI log file (~/.api-cli/api-cli.log)

#### Cleanup Strategy

The cleanup job deletes resources in the following order to handle dependencies:
1. Executions
2. Goals
3. Jobs
4. Snapshots
5. Tests
6. Checkpoints
7. Journeys
8. Projects

Each resource type is filtered by the test tag prefix `test-${GITHUB_RUN_ID}` to ensure only test resources are deleted.

## Local Testing

To test the workflow locally, you can use [act](https://github.com/nektos/act):

```bash
# Install act
brew install act

# Run the workflow
act -s VIRTUOSO_API_KEY=your-key -s VIRTUOSO_API_URL=your-url
```

## Troubleshooting

### Tests fail due to missing credentials
Ensure the `VIRTUOSO_API_KEY` and `VIRTUOSO_API_URL` secrets are properly configured in your repository settings.

### Cleanup job fails
The cleanup job uses `|| true` to continue even if individual delete commands fail. Check the logs to see which resources couldn't be deleted.

### Missing dependencies
The workflow installs all required dependencies, but if you encounter issues:
- Ensure the Go version matches your project requirements
- Check that oapi-codegen installation succeeds
- Verify bats is properly installed

## Future Improvements

Consider implementing:
- Matrix testing across multiple Go versions
- Parallel test execution for faster CI
- Test result visualization
- Coverage reporting integration
- Performance benchmarking
