#!/bin/bash
# Test configuration script for Virtuoso credentials
# This script should be sourced by all test scripts to load credentials
# Source: . scripts/test/config.sh

# Exit on any error
set -e

# Function to safely get secrets from various sources
get_secret() {
    local secret_name=$1
    local default_env_var=$2
    
    # Priority order for secret retrieval:
    # 1. CI/CD secret store (GitHub Actions, GitLab CI, etc.)
    # 2. Local secret manager (AWS Secrets Manager, HashiCorp Vault, etc.)
    # 3. Environment variables
    # 4. Local .env file (for development only, not in version control)
    
    # Check if we're in GitHub Actions
    if [ -n "$GITHUB_ACTIONS" ]; then
        # In GitHub Actions, secrets are already in environment
        eval "echo \$$default_env_var"
        return
    fi
    
    # Check if we're in GitLab CI
    if [ -n "$GITLAB_CI" ]; then
        # In GitLab CI, secrets are already in environment
        eval "echo \$$default_env_var"
        return
    fi
    
    # Check if we're in Jenkins
    if [ -n "$JENKINS_HOME" ]; then
        # In Jenkins, secrets might be in credentials binding
        eval "echo \$$default_env_var"
        return
    fi
    
    # Check for AWS Secrets Manager (if AWS CLI is available)
    if command -v aws &> /dev/null && [ -n "$AWS_SECRET_NAME" ]; then
        aws secretsmanager get-secret-value \
            --secret-id "$AWS_SECRET_NAME" \
            --query "SecretString" \
            --output text 2>/dev/null | jq -r ".$secret_name" 2>/dev/null || true
    fi
    
    # Check for HashiCorp Vault (if vault CLI is available)
    if command -v vault &> /dev/null && [ -n "$VAULT_ADDR" ]; then
        vault kv get -field="$secret_name" secret/virtuoso 2>/dev/null || true
    fi
    
    # Fall back to environment variable
    if [ -n "${!default_env_var}" ]; then
        echo "${!default_env_var}"
        return
    fi
    
    # Check for local .env file (development only)
    if [ -f "scripts/test/.env" ]; then
        grep "^$default_env_var=" scripts/test/.env 2>/dev/null | cut -d'=' -f2- || true
    fi
}

# Function to validate that all required credentials are set
validate_credentials() {
    local missing=()
    
    [ -z "$VIRTUOSO_BASE_URL" ] && missing+=("VIRTUOSO_BASE_URL")
    [ -z "$VIRTUOSO_AUTH_TOKEN" ] && missing+=("VIRTUOSO_AUTH_TOKEN")
    [ -z "$VIRTUOSO_CLIENT_ID" ] && missing+=("VIRTUOSO_CLIENT_ID")
    [ -z "$VIRTUOSO_CLIENT_NAME" ] && missing+=("VIRTUOSO_CLIENT_NAME")
    
    if [ ${#missing[@]} -gt 0 ]; then
        echo "ERROR: Missing required credentials: ${missing[*]}" >&2
        echo "Please ensure the following environment variables are set:" >&2
        echo "  - VIRTUOSO_BASE_URL: The base URL for the Virtuoso API" >&2
        echo "  - VIRTUOSO_AUTH_TOKEN: Authentication token for Virtuoso" >&2
        echo "  - VIRTUOSO_CLIENT_ID: Your Virtuoso client ID" >&2
        echo "  - VIRTUOSO_CLIENT_NAME: Your Virtuoso client name" >&2
        echo "" >&2
        echo "For local development, you can create scripts/test/.env file with these values." >&2
        echo "For CI/CD, configure these as secret environment variables." >&2
        return 1
    fi
    
    return 0
}

# Export Virtuoso credentials
export VIRTUOSO_BASE_URL=$(get_secret "virtuoso_base_url" "VIRTUOSO_BASE_URL")
export VIRTUOSO_AUTH_TOKEN=$(get_secret "virtuoso_auth_token" "VIRTUOSO_AUTH_TOKEN")
export VIRTUOSO_CLIENT_ID=$(get_secret "virtuoso_client_id" "VIRTUOSO_CLIENT_ID")
export VIRTUOSO_CLIENT_NAME=$(get_secret "virtuoso_client_name" "VIRTUOSO_CLIENT_NAME")

# Validate that all credentials are present
if ! validate_credentials; then
    # If we're in a CI environment, fail hard
    if [ -n "$CI" ] || [ -n "$GITHUB_ACTIONS" ] || [ -n "$GITLAB_CI" ] || [ -n "$JENKINS_HOME" ]; then
        exit 1
    fi
    # In development, just warn
    echo "WARNING: Running in development mode with missing credentials" >&2
fi

# Optional: Set additional test configuration
export VIRTUOSO_TEST_TIMEOUT="${VIRTUOSO_TEST_TIMEOUT:-30}"
export VIRTUOSO_TEST_RETRY="${VIRTUOSO_TEST_RETRY:-3}"
export VIRTUOSO_TEST_DEBUG="${VIRTUOSO_TEST_DEBUG:-false}"

# Print configuration status (without revealing secrets)
if [ "$VIRTUOSO_TEST_DEBUG" = "true" ]; then
    echo "Virtuoso test configuration loaded:"
    echo "  VIRTUOSO_BASE_URL: ${VIRTUOSO_BASE_URL:+[SET]}"
    echo "  VIRTUOSO_AUTH_TOKEN: ${VIRTUOSO_AUTH_TOKEN:+[SET]}"
    echo "  VIRTUOSO_CLIENT_ID: ${VIRTUOSO_CLIENT_ID:+[SET]}"
    echo "  VIRTUOSO_CLIENT_NAME: ${VIRTUOSO_CLIENT_NAME:+[SET]}"
    echo "  VIRTUOSO_TEST_TIMEOUT: $VIRTUOSO_TEST_TIMEOUT"
    echo "  VIRTUOSO_TEST_RETRY: $VIRTUOSO_TEST_RETRY"
fi
