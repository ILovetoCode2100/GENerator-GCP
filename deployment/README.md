# D365 Virtuoso Test Deployment Guide

This directory contains all scripts and configuration needed to deploy your D365 Virtuoso tests.

## Prerequisites

1. **D365 Instance**: You need access to a D365 instance
2. **Virtuoso API Token**: Get this from your Virtuoso platform account
3. **API CLI Built**: Run `make build` in the project root to build the CLI

## Quick Start

For the fastest deployment, use the quick deploy script:

```bash
# Set your environment variables
export D365_INSTANCE=your-instance-name  # e.g., contoso-dev
export VIRTUOSO_API_TOKEN=your-api-token

# Run quick deployment
./deployment/quick-deploy.sh
```

This will automatically:

1. Setup the environment
2. Process all test files
3. Deploy to Virtuoso
4. Validate the deployment

## Manual Deployment Steps

### 1. Environment Setup

```bash
# Set required environment variables
export D365_INSTANCE=your-instance-name
export VIRTUOSO_API_TOKEN=your-api-token

# Run environment setup
./deployment/scripts/setup-environment.sh
```

This script will:

- Create necessary directories
- Validate environment variables
- Check API CLI availability
- Test Virtuoso API connection
- Create environment template

### 2. Preprocess Test Files

```bash
./deployment/scripts/preprocess-tests.sh
```

This script will:

- Create backup of original tests
- Replace `[instance]` with `${D365_INSTANCE}` in all YAML files
- Validate YAML syntax
- Generate preprocessing report
- Create test summary

### 3. Deploy Tests

```bash
# Normal deployment
./deployment/scripts/deploy-tests.sh

# Dry run (no actual deployment)
./deployment/scripts/deploy-tests.sh --dry-run

# Continue from checkpoint (if deployment was interrupted)
./deployment/scripts/deploy-tests.sh --continue

# Use existing project
./deployment/scripts/deploy-tests.sh --project-id YOUR_PROJECT_ID
```

The deployment script will:

- Create a new project in Virtuoso (or use existing)
- Create goals for each module (9 total)
- Deploy all 169 tests
- Track progress and handle errors
- Generate deployment report

### 4. Validate Deployment

```bash
./deployment/scripts/validate-deployment.sh
```

This will:

- Check deployment state
- Verify project and goals exist
- Count deployed tests
- Test API connectivity
- Generate health report

## Directory Structure

```
deployment/
├── config/                 # Configuration files
│   └── deployment.config.yaml
├── scripts/               # Deployment scripts
│   ├── setup-environment.sh
│   ├── preprocess-tests.sh
│   ├── deploy-tests.sh
│   ├── rollback-deployment.sh
│   └── validate-deployment.sh
├── state/                 # Deployment state tracking
├── logs/                  # Deployment logs
├── backups/              # Test file backups
├── reports/              # Deployment reports
├── processed-tests/      # Processed YAML files
└── quick-deploy.sh       # Quick deployment script

Test structure:
d365-virtuoso-tests-final/
├── commerce/             # 21 tests
├── customer-service/     # 19 tests
├── field-service/        # 20 tests
├── finance-operations/   # 17 tests
├── human-resources/      # 19 tests
├── marketing/            # 16 tests
├── project-operations/   # 18 tests
├── sales/                # 16 tests
└── supply-chain/         # 23 tests
Total: 169 tests
```

## Environment Variables

### Required

- `D365_INSTANCE`: Your D365 instance name (without .crm.dynamics.com)
- `VIRTUOSO_API_TOKEN`: Your Virtuoso API authentication token

### Optional

- `DEPLOYMENT_PARALLEL_UPLOADS`: Number of parallel uploads (default: 5)
- `DEPLOYMENT_BATCH_SIZE`: Batch size for processing (default: 10)
- `DEPLOYMENT_RETRY_ATTEMPTS`: Retry attempts for failures (default: 3)
- `DEBUG`: Enable debug logging (default: false)

## Rollback Procedures

If you need to rollback a deployment:

```bash
./deployment/scripts/rollback-deployment.sh
```

Options:

1. Delete deployed project (removes from Virtuoso)
2. Restore original test files
3. Clean deployment artifacts
4. Full rollback (all of the above)

## Reports and Monitoring

After deployment, check these reports:

- **Preprocessing Report**: `deployment/reports/preprocessing-report-*.json`
- **Test Summary**: `deployment/reports/test-summary.md`
- **Deployment Report**: `deployment/reports/deployment-report-*.md`
- **Health Report**: `deployment/reports/health-report-*.md`

## Troubleshooting

### Common Issues

1. **Authentication Failed**

   - Verify your API token is correct
   - Check token hasn't expired
   - Ensure token has necessary permissions

2. **Tests Not Found**

   - Verify test directory exists: `d365-virtuoso-tests-final/`
   - Check preprocessing completed successfully

3. **Deployment Interrupted**

   - Use `--continue` flag to resume: `./deploy-tests.sh --continue`
   - Check state file: `deployment/state/deployment-state.json`

4. **Environment Variable Issues**
   - Ensure variables are exported, not just set
   - Check for typos in instance name
   - Verify instance URL format

### Getting Help

1. Check deployment logs in `deployment/logs/`
2. Review error messages in reports
3. Validate environment with `setup-environment.sh`
4. Run health check with `validate-deployment.sh`

## Manual Steps (Cannot be Automated)

The following steps require manual intervention:

1. **Virtuoso Platform Configuration**

   - Set up test execution schedules
   - Configure test environments
   - Set up notifications and alerts
   - Assign team members and permissions

2. **D365 Configuration**

   - Ensure test users are created
   - Configure necessary permissions
   - Set up test data if needed
   - Configure any custom entities used in tests

3. **Post-Deployment Verification**
   - Manually run a sample test to verify setup
   - Check test results and logs
   - Adjust timeouts if needed
   - Configure parallel execution settings

## Best Practices

1. **Before Deployment**

   - Always run in dry-run mode first
   - Backup your test files
   - Verify environment variables
   - Check API connectivity

2. **During Deployment**

   - Monitor logs for errors
   - Don't interrupt unless necessary
   - Use checkpoint recovery if needed

3. **After Deployment**
   - Validate deployment health
   - Run sample tests
   - Review all reports
   - Document any customizations

## Support

For issues with:

- **Scripts**: Check script comments and error messages
- **API CLI**: Refer to CLAUDE.md in project root
- **Virtuoso Platform**: Contact Virtuoso support
- **D365 Access**: Contact your D365 administrator
