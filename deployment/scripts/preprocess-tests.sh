#!/bin/bash
# preprocess-tests.sh - Update test files to use environment variables

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
TEST_DIR="$PROJECT_ROOT/d365-virtuoso-tests-final"
PROCESSED_DIR="$PROJECT_ROOT/deployment/processed-tests"
BACKUP_DIR="$PROJECT_ROOT/deployment/backups"

# Log function
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Create backup of original tests
create_backup() {
    log "Creating backup of original test files..."

    local backup_timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_path="$BACKUP_DIR/tests_backup_$backup_timestamp"

    mkdir -p "$backup_path"
    cp -r "$TEST_DIR"/* "$backup_path/"

    # Create backup manifest
    find "$backup_path" -name "*.yaml" -type f | sort > "$backup_path/manifest.txt"

    success "Backup created at: $backup_path"
    echo "$backup_path" > "$BACKUP_DIR/latest_backup.txt"
}

# Process a single YAML file
process_yaml_file() {
    local input_file="$1"
    local output_file="$2"

    # Create output directory if needed
    mkdir -p "$(dirname "$output_file")"

    # Replace [instance] with ${D365_INSTANCE}
    # This handles multiple patterns:
    # - https://[instance].crm.dynamics.com
    # - [instance].crm.dynamics.com
    # - Any other occurrence of [instance]

    sed 's/\[instance\]/${D365_INSTANCE}/g' "$input_file" > "$output_file"

    # Validate the processed file has valid YAML syntax
    if command -v python3 &> /dev/null; then
        python3 -c "import yaml; yaml.safe_load(open('$output_file'))" 2>/dev/null || {
            error "Invalid YAML syntax in processed file: $output_file"
            return 1
        }
    fi

    return 0
}

# Process all test files
process_all_tests() {
    log "Processing test files..."

    # Clear processed directory
    rm -rf "$PROCESSED_DIR"
    mkdir -p "$PROCESSED_DIR"

    local total_files=0
    local processed_files=0
    local updated_files=0
    local error_files=0

    # Process each module directory
    for module_dir in "$TEST_DIR"/*; do
        if [ -d "$module_dir" ]; then
            local module_name=$(basename "$module_dir")
            log "Processing module: $module_name"

            # Process each YAML file in the module
            for yaml_file in "$module_dir"/*.yaml; do
                if [ -f "$yaml_file" ]; then
                    ((total_files++))

                    local relative_path="${yaml_file#$TEST_DIR/}"
                    local output_file="$PROCESSED_DIR/$relative_path"

                    if process_yaml_file "$yaml_file" "$output_file"; then
                        ((processed_files++))

                        # Check if file was actually updated
                        if grep -q '\${D365_INSTANCE}' "$output_file"; then
                            ((updated_files++))
                        fi
                    else
                        ((error_files++))
                        error "Failed to process: $yaml_file"
                    fi
                fi
            done
        fi
    done

    # Generate processing report
    local report_file="$PROJECT_ROOT/deployment/reports/preprocessing-report-$(date +%Y%m%d_%H%M%S).json"
    mkdir -p "$(dirname "$report_file")"

    cat > "$report_file" <<EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "total_files": $total_files,
  "processed_files": $processed_files,
  "updated_files": $updated_files,
  "error_files": $error_files,
  "modules": {
$(find "$PROCESSED_DIR" -type d -mindepth 1 -maxdepth 1 | while read -r module_dir; do
    local module_name=$(basename "$module_dir")
    local module_count=$(find "$module_dir" -name "*.yaml" | wc -l | tr -d ' ')
    echo "    \"$module_name\": $module_count,"
done | sed '$ s/,$//')
  }
}
EOF

    log "Processing complete:"
    log "  Total files: $total_files"
    log "  Processed successfully: $processed_files"
    log "  Files updated with environment variable: $updated_files"
    log "  Errors: $error_files"

    if [ $error_files -gt 0 ]; then
        error "Some files failed to process. Check the report for details."
        return 1
    fi

    success "All test files processed successfully!"
    success "Report saved to: $report_file"
}

# Validate processed tests
validate_processed_tests() {
    log "Validating processed test files..."

    # Check that all files have required fields
    local validation_errors=0

    find "$PROCESSED_DIR" -name "*.yaml" -type f | while read -r yaml_file; do
        # Check for required fields using grep
        for field in "name:" "description:" "starting_url:" "steps:"; do
            if ! grep -q "^$field" "$yaml_file"; then
                error "Missing required field '$field' in: $yaml_file"
                ((validation_errors++))
            fi
        done

        # Check that ${D365_INSTANCE} is present in URLs
        if ! grep -q '\${D365_INSTANCE}' "$yaml_file"; then
            error "Missing environment variable substitution in: $yaml_file"
            ((validation_errors++))
        fi
    done

    if [ $validation_errors -gt 0 ]; then
        error "Validation found $validation_errors errors"
        return 1
    fi

    success "All processed tests validated successfully"
}

# Create test summary
create_test_summary() {
    log "Creating test summary..."

    local summary_file="$PROJECT_ROOT/deployment/reports/test-summary.md"

    cat > "$summary_file" <<EOF
# D365 Virtuoso Test Summary

Generated: $(date)

## Test Distribution by Module

| Module | Test Count | Description |
|--------|------------|-------------|
EOF

    # Add module summaries
    for module_dir in "$PROCESSED_DIR"/*; do
        if [ -d "$module_dir" ]; then
            local module_name=$(basename "$module_dir")
            local test_count=$(find "$module_dir" -name "*.yaml" | wc -l | tr -d ' ')
            local module_desc=""

            case "$module_name" in
                "commerce") module_desc="B2B, CX, POS, Pricing, Products, Store Management" ;;
                "customer-service") module_desc="Cases, Knowledge Base, Omnichannel, SLA, Satisfaction" ;;
                "field-service") module_desc="Work Orders, IoT, Mobile, Scheduling, Inventory" ;;
                "finance-operations") module_desc="AP, AR, GL, Budget, Fixed Assets, Financial Reporting" ;;
                "human-resources") module_desc="Benefits, Employees, Leave, Performance, Training" ;;
                "marketing") module_desc="Email, Events, Journeys, Landing Pages, Lead Scoring" ;;
                "project-operations") module_desc="Project Management, Resources, Time & Expense" ;;
                "sales") module_desc="Leads, Opportunities, Orders, Pipeline, Integration" ;;
                "supply-chain") module_desc="Inventory, Planning, Production, Quality, Sales Orders" ;;
            esac

            echo "| $module_name | $test_count | $module_desc |" >> "$summary_file"
        fi
    done

    local total_tests=$(find "$PROCESSED_DIR" -name "*.yaml" | wc -l | tr -d ' ')
    echo "| **TOTAL** | **$total_tests** | **All D365 modules** |" >> "$summary_file"

    # Add test listing
    echo -e "\n## Complete Test Listing\n" >> "$summary_file"

    for module_dir in "$PROCESSED_DIR"/*; do
        if [ -d "$module_dir" ]; then
            local module_name=$(basename "$module_dir")
            echo -e "\n### $module_name\n" >> "$summary_file"

            find "$module_dir" -name "*.yaml" -type f | sort | while read -r test_file; do
                local test_name=$(basename "$test_file" .yaml)
                echo "- $test_name" >> "$summary_file"
            done
        fi
    done

    success "Test summary created at: $summary_file"
}

# Main preprocessing function
main() {
    echo -e "${BLUE}=== D365 Virtuoso Test Preprocessing ===${NC}"
    echo

    # Check environment
    if [ -z "${D365_INSTANCE:-}" ]; then
        error "D365_INSTANCE environment variable is not set"
        echo "Please run: export D365_INSTANCE=your-instance-name"
        exit 1
    fi

    log "Using D365 instance: $D365_INSTANCE"

    # Create backup
    create_backup || exit 1

    # Process all tests
    process_all_tests || exit 1

    # Validate processed tests
    validate_processed_tests || exit 1

    # Create summary
    create_test_summary || exit 1

    echo
    success "Test preprocessing completed successfully!"
    echo
    echo "Processed tests are available at: $PROCESSED_DIR"
    echo "You can now run the deployment script: ./deployment/scripts/deploy-tests.sh"
}

# Run main function
main "$@"
