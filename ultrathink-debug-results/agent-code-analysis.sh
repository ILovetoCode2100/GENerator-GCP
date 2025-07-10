#!/bin/bash
source "$(dirname "$0")/../ultrathink-common.sh"

analyze_command_implementations() {
    log "INFO" "CodeAnalysis" "Analyzing command implementations..."
    
    # Find all create-step commands
    local cmd_files=$(find src/cmd -name "create-step-*.go" | sort)
    local modern_count=0
    local legacy_count=0
    
    echo "# Command Implementation Analysis" > "$RESULTS_DIR/code-analysis-report.md"
    echo "Generated: $(date)" >> "$RESULTS_DIR/code-analysis-report.md"
    echo "" >> "$RESULTS_DIR/code-analysis-report.md"
    
    for file in $cmd_files; do
        local cmd_name=$(basename "$file" .go)
        
        # Check for modern pattern indicators
        if grep -q "resolveStepContext" "$file"; then
            echo "- $cmd_name: MODERN (uses resolveStepContext)" >> "$RESULTS_DIR/code-analysis-report.md"
            ((modern_count++))
        else
            echo "- $cmd_name: LEGACY (no session context)" >> "$RESULTS_DIR/code-analysis-report.md"
            ((legacy_count++))
        fi
        
        # Check for checkpoint flag
        if grep -q "addCheckpointFlag" "$file"; then
            echo "  ✓ Has --checkpoint flag" >> "$RESULTS_DIR/code-analysis-report.md"
        fi
    done
    
    echo "" >> "$RESULTS_DIR/code-analysis-report.md"
    echo "## Summary" >> "$RESULTS_DIR/code-analysis-report.md"
    echo "- Modern implementations: $modern_count" >> "$RESULTS_DIR/code-analysis-report.md"
    echo "- Legacy implementations: $legacy_count" >> "$RESULTS_DIR/code-analysis-report.md"
    
    log "SUCCESS" "CodeAnalysis" "Analysis complete: $modern_count modern, $legacy_count legacy"
}

analyze_command_implementations
