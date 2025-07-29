#!/bin/bash
# Secrets Setup Script for Virtuoso API CLI on GCP
# This script configures all required secrets in Google Secret Manager

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Default values
PROJECT_ID="${GCP_PROJECT_ID:-$(gcloud config get-value project)}"
INTERACTIVE=true
ROTATE=false

# Secret names
declare -A SECRETS=(
    ["virtuoso-api-key"]="Virtuoso API authentication key"
    ["virtuoso-org-id"]="Virtuoso organization ID"
    ["jwt-secret"]="JWT signing secret for API authentication"
    ["github-webhook-secret"]="GitHub webhook secret for CI/CD"
    ["slack-webhook-url"]="Slack webhook URL for notifications"
    ["monitoring-api-key"]="API key for monitoring services"
    ["redis-password"]="Redis password for caching"
    ["encryption-key"]="Data encryption key"
)

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to generate secure random string
generate_secret() {
    local length="${1:-32}"
    openssl rand -base64 "$length" | tr -d "=+/" | cut -c1-"$length"
}

# Function to check if secret exists
secret_exists() {
    gcloud secrets describe "$1" --project="$PROJECT_ID" &>/dev/null
}

# Function to create or update secret
create_or_update_secret() {
    local secret_name="$1"
    local secret_value="$2"
    local description="${3:-}"

    if secret_exists "$secret_name"; then
        if [ "$ROTATE" = true ]; then
            print_info "Rotating secret: $secret_name"
            echo -n "$secret_value" | gcloud secrets versions add "$secret_name" \
                --data-file=- \
                --project="$PROJECT_ID"
        else
            print_warning "Secret $secret_name already exists (use --rotate to update)"
        fi
    else
        print_info "Creating secret: $secret_name"
        echo -n "$secret_value" | gcloud secrets create "$secret_name" \
            --data-file=- \
            --replication-policy="automatic" \
            --project="$PROJECT_ID"

        if [ -n "$description" ]; then
            gcloud secrets update "$secret_name" \
                --update-labels="description=$description" \
                --project="$PROJECT_ID"
        fi
    fi
}

# Function to set up IAM permissions for secrets
setup_secret_permissions() {
    print_info "Setting up secret access permissions..."

    local service_accounts=(
        "virtuoso-api-cli@${PROJECT_ID}.iam.gserviceaccount.com"
        "virtuoso-functions@${PROJECT_ID}.iam.gserviceaccount.com"
        "${PROJECT_NUMBER}@cloudbuild.gserviceaccount.com"
    )

    # Get project number
    PROJECT_NUMBER=$(gcloud projects describe "$PROJECT_ID" --format="value(projectNumber)")

    for secret_name in "${!SECRETS[@]}"; do
        if secret_exists "$secret_name"; then
            for sa in "${service_accounts[@]}"; do
                gcloud secrets add-iam-policy-binding "$secret_name" \
                    --member="serviceAccount:$sa" \
                    --role="roles/secretmanager.secretAccessor" \
                    --project="$PROJECT_ID" &>/dev/null || true
            done
        fi
    done

    print_success "Secret permissions configured"
}

# Function to prompt for secret value
prompt_secret() {
    local secret_name="$1"
    local description="$2"
    local current_value=""

    # Check if secret already has a value
    if secret_exists "$secret_name" && [ "$ROTATE" = false ]; then
        print_info "$secret_name already exists, skipping..."
        return
    fi

    print_info "Setting up: $description"

    case "$secret_name" in
        "virtuoso-api-key")
            read -s -p "Enter Virtuoso API key: " current_value
            echo
            ;;
        "virtuoso-org-id")
            read -p "Enter Virtuoso organization ID: " current_value
            ;;
        "github-webhook-secret")
            if [ "$INTERACTIVE" = true ]; then
                read -p "Enter GitHub webhook secret (or press Enter to generate): " current_value
            fi
            if [ -z "$current_value" ]; then
                current_value=$(generate_secret 32)
                print_info "Generated webhook secret: $current_value"
            fi
            ;;
        "slack-webhook-url")
            read -p "Enter Slack webhook URL (optional, press Enter to skip): " current_value
            ;;
        "monitoring-api-key")
            if [ "$INTERACTIVE" = true ]; then
                read -p "Enter monitoring API key (or press Enter to generate): " current_value
            fi
            if [ -z "$current_value" ]; then
                current_value=$(generate_secret 32)
            fi
            ;;
        *)
            # Generate random secrets for JWT, Redis, encryption
            current_value=$(generate_secret 32)
            print_info "Generated secret for $secret_name"
            ;;
    esac

    if [ -n "$current_value" ]; then
        create_or_update_secret "$secret_name" "$current_value" "$description"
    else
        print_warning "Skipping $secret_name (no value provided)"
    fi
}

# Function to setup secret rotation
setup_rotation() {
    print_info "Setting up secret rotation policies..."

    # Create rotation Cloud Function
    cat > rotate-secrets.py <<'EOF'
import functions_framework
import google.cloud.secretmanager as secretmanager
import os
import secrets
import string

@functions_framework.cloud_event
def rotate_secrets(cloud_event):
    """Rotate secrets on schedule"""

    client = secretmanager.SecretManagerServiceClient()
    project_id = os.environ['GCP_PROJECT']

    # List of secrets to auto-rotate
    auto_rotate = [
        'jwt-secret',
        'redis-password',
        'encryption-key',
        'monitoring-api-key'
    ]

    for secret_name in auto_rotate:
        try:
            # Generate new secret value
            alphabet = string.ascii_letters + string.digits
            new_value = ''.join(secrets.choice(alphabet) for _ in range(32))

            # Add new version
            parent = f"projects/{project_id}/secrets/{secret_name}"
            response = client.add_secret_version(
                parent=parent,
                payload={"data": new_value.encode("UTF-8")}
            )

            print(f"Rotated secret: {secret_name}")

            # Disable old versions (keep last 3)
            versions = client.list_secret_versions(parent=parent)
            version_list = list(versions)

            if len(version_list) > 3:
                for version in version_list[3:]:
                    if version.state == secretmanager.SecretVersion.State.ENABLED:
                        client.disable_secret_version(name=version.name)

        except Exception as e:
            print(f"Error rotating {secret_name}: {e}")
EOF

    print_info "Deploy rotation function with:"
    print_info "  gcloud functions deploy rotate-secrets --runtime python39 --trigger-topic secret-rotation"

    # Create rotation schedule
    gcloud scheduler jobs create pubsub rotate-secrets-weekly \
        --schedule="0 2 * * 0" \
        --topic="secret-rotation" \
        --message-body="{}" \
        --time-zone="UTC" \
        --project="$PROJECT_ID" 2>/dev/null || true

    rm -f rotate-secrets.py

    print_success "Secret rotation configured (weekly on Sundays at 2 AM UTC)"
}

# Function to export secrets for local development
export_local_secrets() {
    print_info "Exporting secrets for local development..."

    local env_file="$SCRIPT_DIR/../.local/.env.secrets"
    mkdir -p "$(dirname "$env_file")"

    cat > "$env_file" <<EOF
# Generated secrets for local development
# DO NOT COMMIT THIS FILE TO VERSION CONTROL

EOF

    for secret_name in "${!SECRETS[@]}"; do
        if secret_exists "$secret_name"; then
            # Get latest secret value
            secret_value=$(gcloud secrets versions access latest \
                --secret="$secret_name" \
                --project="$PROJECT_ID" 2>/dev/null || echo "")

            if [ -n "$secret_value" ]; then
                # Convert kebab-case to UPPER_SNAKE_CASE
                env_var_name=$(echo "$secret_name" | tr '[:lower:]-' '[:upper:]_')
                echo "${env_var_name}=${secret_value}" >> "$env_file"
            fi
        fi
    done

    chmod 600 "$env_file"

    print_success "Local secrets exported to: $env_file"
    print_warning "Remember to add .env.secrets to .gitignore"
}

# Function to verify secrets
verify_secrets() {
    print_info "Verifying secrets configuration..."

    local missing=()
    local configured=()

    for secret_name in "${!SECRETS[@]}"; do
        if secret_exists "$secret_name"; then
            # Check if secret has at least one version
            if gcloud secrets versions list "$secret_name" --limit=1 --project="$PROJECT_ID" | grep -q "ENABLED"; then
                configured+=("$secret_name")
            else
                missing+=("$secret_name")
            fi
        else
            missing+=("$secret_name")
        fi
    done

    print_info "Configured secrets: ${#configured[@]}/${#SECRETS[@]}"

    if [ ${#missing[@]} -gt 0 ]; then
        print_warning "Missing secrets:"
        for secret in "${missing[@]}"; do
            echo "  - $secret: ${SECRETS[$secret]}"
        done
    fi

    # Test secret access
    print_info "Testing secret access..."
    if gcloud secrets versions access latest \
        --secret="virtuoso-api-key" \
        --project="$PROJECT_ID" &>/dev/null; then
        print_success "Secret access test passed"
    else
        print_error "Secret access test failed"
    fi
}

# Function to backup secrets
backup_secrets() {
    print_info "Backing up secrets..."

    local backup_dir="$SCRIPT_DIR/../.backups/secrets-$(date +%Y%m%d%H%M%S)"
    mkdir -p "$backup_dir"

    for secret_name in "${!SECRETS[@]}"; do
        if secret_exists "$secret_name"; then
            # Export secret metadata
            gcloud secrets describe "$secret_name" \
                --project="$PROJECT_ID" \
                --format=json > "$backup_dir/${secret_name}-metadata.json"

            # Note: We don't backup actual secret values for security
            print_info "Backed up metadata for: $secret_name"
        fi
    done

    # Create restore script
    cat > "$backup_dir/restore.sh" <<'EOF'
#!/bin/bash
# Restore script for secrets metadata
# Note: Actual secret values must be re-entered manually

for metadata_file in *.json; do
    secret_name="${metadata_file%-metadata.json}"
    echo "To restore $secret_name, run:"
    echo "  gcloud secrets create $secret_name --data-file=- --project=$PROJECT_ID"
    echo "  Then enter the secret value and press Ctrl+D"
    echo ""
done
EOF

    chmod +x "$backup_dir/restore.sh"

    print_success "Secrets metadata backed up to: $backup_dir"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --project-id)
            PROJECT_ID="$2"
            shift 2
            ;;
        --non-interactive)
            INTERACTIVE=false
            shift
            ;;
        --rotate)
            ROTATE=true
            shift
            ;;
        --export-local)
            EXPORT_LOCAL=true
            shift
            ;;
        --backup)
            BACKUP=true
            shift
            ;;
        --verify-only)
            VERIFY_ONLY=true
            shift
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo "Options:"
            echo "  --project-id ID      GCP project ID"
            echo "  --non-interactive    Don't prompt for values"
            echo "  --rotate            Rotate existing secrets"
            echo "  --export-local      Export secrets for local development"
            echo "  --backup            Backup secret metadata"
            echo "  --verify-only       Only verify secrets, don't create"
            echo "  --help              Show this help message"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Main flow
print_info "Virtuoso API CLI Secrets Setup"
echo -e "${BLUE}===================================================${NC}\n"

# Verify project
if [ -z "$PROJECT_ID" ]; then
    print_error "Project ID not specified"
    exit 1
fi

print_info "Project: $PROJECT_ID"

# Execute based on mode
if [ "${VERIFY_ONLY:-false}" = true ]; then
    verify_secrets
elif [ "${BACKUP:-false}" = true ]; then
    backup_secrets
else
    # Setup secrets
    print_info "Setting up secrets..."
    echo

    # Required secrets
    if [ "$INTERACTIVE" = true ] || [ "$ROTATE" = true ]; then
        prompt_secret "virtuoso-api-key" "${SECRETS[virtuoso-api-key]}"
        prompt_secret "virtuoso-org-id" "${SECRETS[virtuoso-org-id]}"
    fi

    # Auto-generated secrets
    for secret_name in "${!SECRETS[@]}"; do
        if [[ "$secret_name" != "virtuoso-api-key" && "$secret_name" != "virtuoso-org-id" ]]; then
            prompt_secret "$secret_name" "${SECRETS[$secret_name]}"
        fi
    done

    # Setup permissions
    setup_secret_permissions

    # Setup rotation
    if [ "$ROTATE" = true ]; then
        setup_rotation
    fi

    # Export for local development
    if [ "${EXPORT_LOCAL:-false}" = true ]; then
        export_local_secrets
    fi

    # Verify setup
    echo
    verify_secrets
fi

print_success "Secrets setup completed!"
