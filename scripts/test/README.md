# Test Configuration

This directory contains the test configuration script that manages Virtuoso API credentials securely.

## Setup

### For Local Development

1. Copy the template file:
   ```bash
   cp scripts/test/.env.template scripts/test/.env
   ```

2. Edit `scripts/test/.env` and fill in your actual Virtuoso credentials:
   - `VIRTUOSO_BASE_URL`: The base URL for the Virtuoso API
   - `VIRTUOSO_AUTH_TOKEN`: Your authentication token
   - `VIRTUOSO_CLIENT_ID`: Your client ID
   - `VIRTUOSO_CLIENT_NAME`: Your client name

3. **Important**: Never commit the `.env` file to version control!

### For CI/CD Environments

Configure the following environment variables as secrets in your CI/CD system:
- `VIRTUOSO_BASE_URL`
- `VIRTUOSO_AUTH_TOKEN`
- `VIRTUOSO_CLIENT_ID`
- `VIRTUOSO_CLIENT_NAME`

The config script automatically detects and uses secrets from:
- GitHub Actions
- GitLab CI
- Jenkins
- AWS Secrets Manager (if configured with `AWS_SECRET_NAME`)
- HashiCorp Vault (if configured with `VAULT_ADDR`)

## Usage in Tests

Source the configuration script at the beginning of your test scripts:

```bash
#!/bin/bash
# Source the test configuration
. scripts/test/config.sh

# Your test code here
# All VIRTUOSO_* variables are now available
```

## Security Features

- Credentials are never hard-coded in test scripts
- Multiple secret source support with priority ordering
- Validation ensures all required credentials are present
- Debug mode shows configuration status without revealing secrets
- CI/CD environments fail fast on missing credentials
- Development environments show warnings but continue

## Optional Configuration

The following optional variables can be set:
- `VIRTUOSO_TEST_TIMEOUT`: Request timeout in seconds (default: 30)
- `VIRTUOSO_TEST_RETRY`: Number of retry attempts (default: 3)
- `VIRTUOSO_TEST_DEBUG`: Enable debug output (default: false)

## Troubleshooting

To debug configuration loading, set `VIRTUOSO_TEST_DEBUG=true`:

```bash
VIRTUOSO_TEST_DEBUG=true . scripts/test/config.sh
```

This will show which credentials are set (without revealing their values).
