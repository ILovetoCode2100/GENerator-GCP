#!/bin/bash
# rollback-deployment.sh - Rollback a failed or unwanted deployment

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname $(dirname "$SCRIPT_DIR"))"
API_CLI="$PROJECT_ROOT/bin/api-cli"
STATE_DIR="$PROJECT_ROOT/deployment/state"
BACKUP_DIR="$PROJECT_ROOT/deployment/backups"

# State file
STATE_FILE="$STATE_DIR/deployment-state.json"

# Log functions
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Delete a project and all its contents
delete_project() {
    local project_id="$1"

    log "Deleting project: $project_id"

    # First, get project details
    local project_info=$("$API_CLI" get project "$project_id" --output json 2>&1) || {
        error "Failed to get project information"
        return 1
    }

    local project_name=$(echo "$project_info" | jq -r '.name // "Unknown"')

    warning "This will delete project: $project_name (ID: $project_id)"
    warning "This action cannot be undone!"

    read -p "Are you sure you want to delete this project? (yes/no): " confirmation

    if [ "$confirmation" != "yes" ]; then
        log "Rollback cancelled"
        return 1
    fi

    # Delete the project
    if "$API_CLI" delete project "$project_id" &>/dev/null; then
        success "Project deleted successfully"
        return 0
    else
        error "Failed to delete project"
        return 1
    fi
}

# Restore original test files
restore_test_files() {
    log "Restoring original test files..."

    # Find latest backup
    local latest_backup=""
    if [ -f "$BACKUP_DIR/latest_backup.txt" ]; then
        latest_backup=$(cat "$BACKUP_DIR/latest_backup.txt")
    fi

    if [ -z "$latest_backup" ] || [ ! -d "$latest_backup" ]; then
        error "No backup found to restore"
        return 1
    fi

    log "Restoring from backup: $latest_backup"

    # Restore files
    local test_dir="$PROJECT_ROOT/d365-virtuoso-tests-final"

    # Create backup of current state before restoring
    local current_backup="$BACKUP_DIR/pre_restore_$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$current_backup"
    cp -r "$test_dir"/* "$current_backup/" 2>/dev/null || true

    # Restore from backup
    cp -r "$latest_backup"/* "$test_dir/"

    success "Test files restored from backup"
}

# Clean deployment artifacts
clean_artifacts() {
    log "Cleaning deployment artifacts..."

    # Clean processed tests
    if [ -d "$PROJECT_ROOT/deployment/processed-tests" ]; then
        rm -rf "$PROJECT_ROOT/deployment/processed-tests"
        log "Cleaned processed tests"
    fi

    # Archive state files
    if [ -f "$STATE_FILE" ]; then
        local archive_name="$STATE_DIR/archived_state_$(date +%Y%m%d_%H%M%S).json"
        mv "$STATE_FILE" "$archive_name"
        log "Archived state file to: $archive_name"
    fi

    success "Deployment artifacts cleaned"
}

# Show rollback options
show_rollback_menu() {
    echo -e "${BLUE}=== D365 Virtuoso Deployment Rollback ===${NC}"
    echo
    echo "What would you like to rollback?"
    echo
    echo "1. Delete deployed project (removes all tests from Virtuoso)"
    echo "2. Restore original test files (undo preprocessing)"
    echo "3. Clean deployment artifacts only"
    echo "4. Full rollback (all of the above)"
    echo "5. Exit"
    echo
    read -p "Select option (1-5): " choice

    case $choice in
        1)
            rollback_project_only
            ;;
        2)
            restore_test_files
            ;;
        3)
            clean_artifacts
            ;;
        4)
            full_rollback
            ;;
        5)
            log "Exiting rollback"
            exit 0
            ;;
        *)
            error "Invalid option"
            exit 1
            ;;
    esac
}

# Rollback project only
rollback_project_only() {
    log "Rolling back deployed project..."

    # Get project ID from state
    local project_id=""
    if [ -f "$STATE_FILE" ]; then
        project_id=$(jq -r '.project_id // empty' "$STATE_FILE")
    fi

    if [ -z "$project_id" ]; then
        error "No project ID found in deployment state"
        return 1
    fi

    delete_project "$project_id"
}

# Full rollback
full_rollback() {
    log "Performing full rollback..."

    # Delete project if exists
    if [ -f "$STATE_FILE" ]; then
        local project_id=$(jq -r '.project_id // empty' "$STATE_FILE")
        if [ -n "$project_id" ]; then
            delete_project "$project_id" || warning "Failed to delete project"
        fi
    fi

    # Restore test files
    restore_test_files || warning "Failed to restore test files"

    # Clean artifacts
    clean_artifacts || warning "Failed to clean artifacts"

    success "Full rollback completed"
}

# Generate rollback report
generate_rollback_report() {
    local report_file="$PROJECT_ROOT/deployment/reports/rollback-report-$(date +%Y%m%d_%H%M%S).md"
    mkdir -p "$(dirname "$report_file")"

    cat > "$report_file" <<EOF
# Rollback Report

Generated: $(date)

## Actions Performed

- Project deletion: ${PROJECT_DELETED:-No}
- Test files restored: ${FILES_RESTORED:-No}
- Artifacts cleaned: ${ARTIFACTS_CLEANED:-No}

## Details

$(cat "$STATE_DIR/rollback.log" 2>/dev/null || echo "No additional details available")

EOF

    log "Rollback report generated: $report_file"
}

# Main function
main() {
    # Check if there's a deployment to rollback
    if [ ! -f "$STATE_FILE" ] && [ ! -d "$PROJECT_ROOT/deployment/processed-tests" ]; then
        warning "No deployment found to rollback"
        exit 0
    fi

    # Show current deployment status
    if [ -f "$STATE_FILE" ]; then
        local project_id=$(jq -r '.project_id // "Not found"' "$STATE_FILE")
        local status=$(jq -r '.status // "Unknown"' "$STATE_FILE")
        local deployed_count=$(jq '.tests_deployed | length' "$STATE_FILE" 2>/dev/null || echo "0")

        echo "Current deployment status:"
        echo "  Project ID: $project_id"
        echo "  Status: $status"
        echo "  Tests deployed: $deployed_count"
        echo
    fi

    # Show rollback menu
    show_rollback_menu

    # Generate report
    generate_rollback_report
}

# Run main function
main "$@"
