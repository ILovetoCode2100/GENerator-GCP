#!/bin/bash
source "$(dirname "$0")/../ultrathink-common.sh"

analyze_signatures() {
    log "INFO" "SignaturePattern" "Analyzing command signatures..."
    
    echo "# Command Signature Patterns" > "$RESULTS_DIR/signature-patterns.md"
    echo "" >> "$RESULTS_DIR/signature-patterns.md"
    
    # Pattern categories
    echo "## Identified Patterns" >> "$RESULTS_DIR/signature-patterns.md"
    echo "" >> "$RESULTS_DIR/signature-patterns.md"
    
    echo "### Modern Pattern (Session Context)" >> "$RESULTS_DIR/signature-patterns.md"
    echo '```' >> "$RESULTS_DIR/signature-patterns.md"
    echo "Args: cobra.MinimumNArgs(1) or similar" >> "$RESULTS_DIR/signature-patterns.md"
    echo "Usage: ELEMENT [POSITION] with optional --checkpoint flag" >> "$RESULTS_DIR/signature-patterns.md"
    echo "Key: Uses resolveStepContext() from step_helpers.go" >> "$RESULTS_DIR/signature-patterns.md"
    echo '```' >> "$RESULTS_DIR/signature-patterns.md"
    echo "" >> "$RESULTS_DIR/signature-patterns.md"
    
    echo "### Legacy Pattern (Checkpoint Required)" >> "$RESULTS_DIR/signature-patterns.md"
    echo '```' >> "$RESULTS_DIR/signature-patterns.md"
    echo "Args: cobra.ExactArgs(3) or similar" >> "$RESULTS_DIR/signature-patterns.md"
    echo "Usage: CHECKPOINT_ID ELEMENT POSITION" >> "$RESULTS_DIR/signature-patterns.md"
    echo "Key: Direct checkpoint ID parsing" >> "$RESULTS_DIR/signature-patterns.md"
    echo '```' >> "$RESULTS_DIR/signature-patterns.md"
    
    # Find specific patterns
    echo "" >> "$RESULTS_DIR/signature-patterns.md"
    echo "## Commands by Pattern" >> "$RESULTS_DIR/signature-patterns.md"
    
    # Modern pattern search
    echo "### Modern Commands:" >> "$RESULTS_DIR/signature-patterns.md"
    grep -l "resolveStepContext" src/cmd/create-step-*.go | while read file; do
        echo "- $(basename "$file" .go)" >> "$RESULTS_DIR/signature-patterns.md"
    done
    
    log "SUCCESS" "SignaturePattern" "Signature pattern analysis complete"
}

analyze_signatures
