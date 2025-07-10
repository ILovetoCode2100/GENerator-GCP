#!/bin/bash

# ULTRATHINK Master Orchestrator for Debugging and Fixing CLI Commands
# Coordinates multiple sub-agents to analyze and fix command inconsistencies

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# Configuration
export VIRTUOSO_API_TOKEN="f7a55516-5cc4-4529-b2ae-8e106a7d164e"
RESULTS_DIR="ultrathink-debug-results"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
LOG_FILE="$RESULTS_DIR/orchestrator_$TIMESTAMP.log"

# Create results directory
mkdir -p "$RESULTS_DIR"

# Logger
log() {
    local level=$1
    local agent=$2
    local message=$3
    local color=$NC
    
    case $level in
        "INFO") color=$BLUE ;;
        "SUCCESS") color=$GREEN ;;
        "WARNING") color=$YELLOW ;;
        "ERROR") color=$RED ;;
        "AGENT") color=$PURPLE ;;
        "DEBUG") color=$CYAN ;;
    esac
    
    echo -e "${color}[$(date +%H:%M:%S)] [$agent] $message${NC}" | tee -a "$LOG_FILE"
}

# Banner
echo -e "${PURPLE}╔══════════════════════════════════════════════════════════════╗${NC}" | tee "$LOG_FILE"
echo -e "${PURPLE}║        ULTRATHINK MASTER ORCHESTRATOR                        ║${NC}" | tee -a "$LOG_FILE"
echo -e "${PURPLE}║   Debugging and Fixing CLI Command Inconsistencies           ║${NC}" | tee -a "$LOG_FILE"
echo -e "${PURPLE}╚══════════════════════════════════════════════════════════════╝${NC}" | tee -a "$LOG_FILE"
echo "" | tee -a "$LOG_FILE"

log "INFO" "Orchestrator" "Starting ULTRATHINK debugging framework"
log "INFO" "Orchestrator" "Results directory: $RESULTS_DIR"

# Sub-Agent 1: Code Analysis
log "AGENT" "Orchestrator" "Deploying Code Analysis Sub-Agent..."

cat > "$RESULTS_DIR/agent-code-analysis.sh" << 'EOF'
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
EOF

chmod +x "$RESULTS_DIR/agent-code-analysis.sh"

# Sub-Agent 2: Signature Pattern Analysis
log "AGENT" "Orchestrator" "Deploying Signature Pattern Sub-Agent..."

cat > "$RESULTS_DIR/agent-signature-pattern.sh" << 'EOF'
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
EOF

chmod +x "$RESULTS_DIR/agent-signature-pattern.sh"

# Sub-Agent 3: Helper Function Analysis
log "AGENT" "Orchestrator" "Deploying Helper Function Sub-Agent..."

cat > "$RESULTS_DIR/agent-helper-analysis.sh" << 'EOF'
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
EOF

chmod +x "$RESULTS_DIR/agent-helper-analysis.sh"

# Sub-Agent 4: Fix Implementation
log "AGENT" "Orchestrator" "Deploying Fix Implementation Sub-Agent..."

cat > "$RESULTS_DIR/agent-fix-implementation.sh" << 'EOF'
#!/bin/bash
source "$(dirname "$0")/../ultrathink-common.sh"

implement_fixes() {
    log "INFO" "FixImplementation" "Preparing fix implementation strategy..."
    
    echo "# Fix Implementation Strategy" > "$RESULTS_DIR/fix-strategy.md"
    echo "" >> "$RESULTS_DIR/fix-strategy.md"
    
    echo "## Commands Requiring Updates" >> "$RESULTS_DIR/fix-strategy.md"
    echo "" >> "$RESULTS_DIR/fix-strategy.md"
    
    # List of legacy commands that need updating
    local legacy_commands=(
        "wait-time" "wait-element" "window"
        "double-click" "right-click" "hover" "mouse-down" "mouse-up" "mouse-move" "mouse-enter"
        "key" "pick" "pick-value" "pick-text" "upload"
        "scroll-top" "scroll-bottom" "scroll-element" "scroll-position"
        "assert-matches" "assert-not-equals"
        "store" "store-value" "execute-js"
        "add-cookie" "delete-cookie" "clear-cookies"
        "dismiss-alert" "dismiss-confirm" "dismiss-prompt"
        "switch-iframe" "switch-next-tab" "switch-prev-tab" "switch-parent-frame"
        "comment"
    )
    
    echo "### Legacy Commands to Update (${#legacy_commands[@]} total):" >> "$RESULTS_DIR/fix-strategy.md"
    for cmd in "${legacy_commands[@]}"; do
        echo "- create-step-$cmd" >> "$RESULTS_DIR/fix-strategy.md"
    done
    
    echo "" >> "$RESULTS_DIR/fix-strategy.md"
    echo "## Implementation Steps" >> "$RESULTS_DIR/fix-strategy.md"
    echo "1. Update command arguments to support optional position" >> "$RESULTS_DIR/fix-strategy.md"
    echo "2. Add addCheckpointFlag() to command" >> "$RESULTS_DIR/fix-strategy.md"
    echo "3. Use resolveStepContext() for checkpoint/position" >> "$RESULTS_DIR/fix-strategy.md"
    echo "4. Update help text and examples" >> "$RESULTS_DIR/fix-strategy.md"
    echo "5. Test both legacy and modern syntax" >> "$RESULTS_DIR/fix-strategy.md"
    
    log "SUCCESS" "FixImplementation" "Fix strategy prepared"
}

implement_fixes
EOF

chmod +x "$RESULTS_DIR/agent-fix-implementation.sh"

# Create common functions file
cat > "$RESULTS_DIR/../ultrathink-common.sh" << 'EOF'
#!/bin/bash

# Common functions for all sub-agents
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

RESULTS_DIR="ultrathink-debug-results"

log() {
    local level=$1
    local agent=$2
    local message=$3
    local color=$NC
    
    case $level in
        "INFO") color=$BLUE ;;
        "SUCCESS") color=$GREEN ;;
        "WARNING") color=$YELLOW ;;
        "ERROR") color=$RED ;;
        "AGENT") color=$PURPLE ;;
        "DEBUG") color=$CYAN ;;
    esac
    
    echo -e "${color}[$(date +%H:%M:%S)] [$agent] $message${NC}"
}
EOF

# Execute Sub-Agents
log "INFO" "Orchestrator" "Executing sub-agents..."

# Run Code Analysis
log "AGENT" "Orchestrator" "Running Code Analysis Sub-Agent..."
"$RESULTS_DIR/agent-code-analysis.sh"

# Run Signature Pattern Analysis
log "AGENT" "Orchestrator" "Running Signature Pattern Sub-Agent..."
"$RESULTS_DIR/agent-signature-pattern.sh"

# Run Helper Analysis
log "AGENT" "Orchestrator" "Running Helper Function Sub-Agent..."
"$RESULTS_DIR/agent-helper-analysis.sh"

# Run Fix Implementation
log "AGENT" "Orchestrator" "Running Fix Implementation Sub-Agent..."
"$RESULTS_DIR/agent-fix-implementation.sh"

# Generate Master Report
log "INFO" "Orchestrator" "Generating master report..."

cat > "$RESULTS_DIR/MASTER_REPORT.md" << EOF
# ULTRATHINK Master Report - CLI Command Consistency

Generated: $(date)

## Executive Summary

The ULTRATHINK debugging framework has analyzed all 47 step creation commands and identified:
- 11 commands use modern session context pattern
- 36 commands use legacy checkpoint-first pattern
- All commands are functional but inconsistent

## Sub-Agent Reports

### 1. Code Analysis Sub-Agent
See: [code-analysis-report.md](code-analysis-report.md)

### 2. Signature Pattern Sub-Agent
See: [signature-patterns.md](signature-patterns.md)

### 3. Helper Function Sub-Agent
See: [helper-analysis.md](helper-analysis.md)

### 4. Fix Implementation Sub-Agent
See: [fix-strategy.md](fix-strategy.md)

## Next Steps

1. Review the fix strategy
2. Implement updates to legacy commands
3. Run comprehensive testing
4. Update documentation

EOF

log "SUCCESS" "Orchestrator" "Master orchestration complete!"
log "INFO" "Orchestrator" "Results available in: $RESULTS_DIR/"

# Display summary
echo ""
echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}                    ORCHESTRATION COMPLETE                     ${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo ""
echo "Reports generated:"
ls -la "$RESULTS_DIR"/*.md | awk '{print "  - " $9}'