#!/bin/bash
source "$(dirname "$0")/../ultrathink-common.sh"

analyze_helpers() {
    log "INFO" "HelperAnalysis" "Analyzing step_helpers.go..."
    
    echo "# Helper Function Analysis" > "$RESULTS_DIR/helper-analysis.md"
    echo "" >> "$RESULTS_DIR/helper-analysis.md"
    
    # Key functions to analyze
    echo "## Key Helper Functions" >> "$RESULTS_DIR/helper-analysis.md"
    echo "" >> "$RESULTS_DIR/helper-analysis.md"
    
    # Extract function signatures
    echo "### resolveStepContext" >> "$RESULTS_DIR/helper-analysis.md"
    echo "Purpose: Resolves checkpoint ID and position from session or args" >> "$RESULTS_DIR/helper-analysis.md"
    echo "" >> "$RESULTS_DIR/helper-analysis.md"
    
    echo "### addCheckpointFlag" >> "$RESULTS_DIR/helper-analysis.md"
    echo "Purpose: Adds --checkpoint flag to commands" >> "$RESULTS_DIR/helper-analysis.md"
    echo "" >> "$RESULTS_DIR/helper-analysis.md"
    
    echo "### outputStepResult" >> "$RESULTS_DIR/helper-analysis.md"
    echo "Purpose: Consistent output formatting across formats" >> "$RESULTS_DIR/helper-analysis.md"
    echo "" >> "$RESULTS_DIR/helper-analysis.md"
    
    # Check which functions exist
    if [ -f "src/cmd/step_helpers.go" ]; then
        echo "## Available Functions in step_helpers.go:" >> "$RESULTS_DIR/helper-analysis.md"
        grep "^func " src/cmd/step_helpers.go | sed 's/func /- /' >> "$RESULTS_DIR/helper-analysis.md"
    fi
    
    log "SUCCESS" "HelperAnalysis" "Helper function analysis complete"
}

analyze_helpers
