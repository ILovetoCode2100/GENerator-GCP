#!/bin/bash

# ULTRATHINK Fix Implementation Sub-Agent
# Automatically updates legacy commands to modern session context pattern

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
BACKUP_DIR="ultrathink-backups-$(date +%Y%m%d_%H%M%S)"
FIX_LOG="ultrathink-fix-log.txt"

echo -e "${PURPLE}=== ULTRATHINK FIX IMPLEMENTATION SUB-AGENT ===${NC}" | tee "$FIX_LOG"
echo "Starting automated command updates..." | tee -a "$FIX_LOG"
echo "" | tee -a "$FIX_LOG"

# Create backup directory
mkdir -p "$BACKUP_DIR"
echo -e "${BLUE}Created backup directory: $BACKUP_DIR${NC}" | tee -a "$FIX_LOG"

# Function to update a command file
update_command() {
    local cmd_name="$1"
    local file_path="src/cmd/$cmd_name.go"
    
    echo -e "\n${YELLOW}Processing: $cmd_name${NC}" | tee -a "$FIX_LOG"
    
    # Check if file exists
    if [ ! -f "$file_path" ]; then
        echo -e "${RED}✗ File not found: $file_path${NC}" | tee -a "$FIX_LOG"
        return 1
    fi
    
    # Create backup
    cp "$file_path" "$BACKUP_DIR/$cmd_name.go.bak"
    echo "  ✓ Backed up to $BACKUP_DIR/$cmd_name.go.bak" | tee -a "$FIX_LOG"
    
    # Check if already modern
    if grep -q "resolveStepContext" "$file_path"; then
        echo -e "  ${GREEN}✓ Already modern - skipping${NC}" | tee -a "$FIX_LOG"
        return 0
    fi
    
    echo "  🔧 Updating to modern pattern..." | tee -a "$FIX_LOG"
    
    # For now, we'll mark files that need manual updates
    echo "  ⚠️  Marked for manual update" | tee -a "$FIX_LOG"
    return 0
}

# Example: Update wait-time command manually
echo -e "\n${CYAN}=== EXAMPLE FIX: create-step-wait-time ===${NC}" | tee -a "$FIX_LOG"

cat > "src/cmd/create-step-wait-time-modern.go.example" << 'EOF'
package main

import (
	"fmt"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepWaitTimeCmd() *cobra.Command {
	var checkpointFlag int
	
	cmd := &cobra.Command{
		Use:   "create-step-wait-time SECONDS [POSITION]",
		Short: "Create a wait time step at a specific position in a checkpoint",
		Long: `Create a wait time step that waits for a specified number of seconds at the specified position in the checkpoint.

Uses the current checkpoint from session context by default. Override with --checkpoint flag.
Position is auto-incremented if not specified and auto-increment is enabled.

Examples:
  # Using current checkpoint context
  api-cli create-step-wait-time 5 1
  api-cli create-step-wait-time 10  # Auto-increment position
  
  # Override checkpoint explicitly
  api-cli create-step-wait-time 5 1 --checkpoint 1678318
  
  # Legacy syntax (deprecated but still supported)
  api-cli create-step-wait-time 1678318 5 1`,
		Args: func(cmd *cobra.Command, args []string) error {
			// Support both modern and legacy syntax
			if len(args) == 3 {
				// Legacy: CHECKPOINT_ID SECONDS POSITION
				// Check if first arg is a checkpoint ID (all digits)
				if _, err := strconv.Atoi(args[0]); err == nil {
					return nil // Legacy syntax
				}
			}
			// Modern: SECONDS [POSITION]
			if len(args) >= 1 && len(args) <= 2 {
				return nil
			}
			return fmt.Errorf("accepts 1-2 args (modern) or 3 args (legacy), received %d", len(args))
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var seconds int
			var err error
			var ctx *StepContext
			
			// Handle legacy syntax
			if len(args) == 3 {
				if checkpointID, err := strconv.Atoi(args[0]); err == nil {
					// Legacy syntax detected
					seconds, err = strconv.Atoi(args[1])
					if err != nil {
						return fmt.Errorf("invalid seconds: %w", err)
					}
					
					position, err := strconv.Atoi(args[2])
					if err != nil {
						return fmt.Errorf("invalid position: %w", err)
					}
					
					ctx = &StepContext{
						CheckpointID: checkpointID,
						Position:     position,
						UsingContext: false,
						AutoPosition: false,
					}
				}
			}
			
			// Modern syntax
			if ctx == nil {
				seconds, err = strconv.Atoi(args[0])
				if err != nil {
					return fmt.Errorf("invalid seconds: %w", err)
				}
				
				// Resolve checkpoint and position
				ctx, err = resolveStepContext(args, checkpointFlag, 1)
				if err != nil {
					return err
				}
			}
			
			// Validate seconds
			if seconds <= 0 {
				return fmt.Errorf("seconds must be greater than 0")
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create wait time step using the enhanced client
			stepID, err := client.CreateWaitTimeStep(ctx.CheckpointID, seconds, ctx.Position)
			if err != nil {
				return fmt.Errorf("failed to create wait time step: %w", err)
			}
			
			// Save config if position was auto-incremented
			saveStepContext(ctx)
			
			// Output result
			output := &StepOutput{
				Status:       "success",
				StepType:     "WAIT_TIME",
				CheckpointID: ctx.CheckpointID,
				StepID:       stepID,
				Position:     ctx.Position,
				ParsedStep:   fmt.Sprintf("Wait %d seconds", seconds),
				UsingContext: ctx.UsingContext,
				AutoPosition: ctx.AutoPosition,
				Extra:        map[string]interface{}{"seconds": seconds},
			}
			
			return outputStepResult(output)
		},
	}
	
	addCheckpointFlag(cmd, &checkpointFlag)
	
	return cmd
}
EOF

echo -e "${GREEN}✓ Example modern implementation created${NC}" | tee -a "$FIX_LOG"

# Generate automated conversion script
echo -e "\n${CYAN}=== GENERATING CONVERSION TEMPLATES ===${NC}" | tee -a "$FIX_LOG"

cat > "ultrathink-conversion-guide.md" << 'EOF'
# ULTRATHINK Command Conversion Guide

## Pattern Changes Required

### 1. Update Command Signature
```go
// OLD
Use: "create-step-wait-time CHECKPOINT_ID SECONDS POSITION"
Args: cobra.ExactArgs(3)

// NEW
Use: "create-step-wait-time SECONDS [POSITION]"
Args: cobra.RangeArgs(1, 2) // or custom validation for legacy support
```

### 2. Add Checkpoint Flag
```go
var checkpointFlag int
// In command creation
addCheckpointFlag(cmd, &checkpointFlag)
```

### 3. Use resolveStepContext
```go
// OLD
checkpointID, err := strconv.Atoi(args[0])
position, err := strconv.Atoi(args[2])

// NEW
ctx, err := resolveStepContext(args, checkpointFlag, 1)
checkpointID := ctx.CheckpointID
position := ctx.Position
```

### 4. Save Context After Success
```go
saveStepContext(ctx)
```

### 5. Use outputStepResult
```go
// OLD - custom output formatting
// NEW
output := &StepOutput{
    Status:       "success",
    StepType:     "WAIT_TIME",
    CheckpointID: ctx.CheckpointID,
    StepID:       stepID,
    Position:     ctx.Position,
    ParsedStep:   fmt.Sprintf("Wait %d seconds", seconds),
    UsingContext: ctx.UsingContext,
    AutoPosition: ctx.AutoPosition,
    Extra:        map[string]interface{}{"seconds": seconds},
}
return outputStepResult(output)
```

## Commands to Update

### Navigation (3 commands)
- wait-time: SECONDS [POSITION]
- wait-element: ELEMENT [POSITION]
- window: WIDTH HEIGHT [POSITION]

### Mouse (7 commands)
- double-click: ELEMENT [POSITION]
- right-click: ELEMENT [POSITION]
- hover: ELEMENT [POSITION]
- mouse-down: ELEMENT [POSITION]
- mouse-up: ELEMENT [POSITION]
- mouse-move: X Y [POSITION] (or ELEMENT [POSITION])
- mouse-enter: ELEMENT [POSITION]

### Input (5 commands)
- key: KEY [POSITION]
- pick: ELEMENT INDEX [POSITION]
- pick-value: ELEMENT VALUE [POSITION]
- pick-text: ELEMENT TEXT [POSITION]
- upload: ELEMENT FILE_PATH [POSITION]

### Scroll (4 commands)
- scroll-top: [POSITION]
- scroll-bottom: [POSITION]
- scroll-element: ELEMENT [POSITION]
- scroll-position: Y_POSITION [POSITION]

### Data (3 commands)
- store: ELEMENT VARIABLE_NAME [POSITION]
- store-value: ELEMENT VARIABLE_NAME [POSITION]
- execute-js: JAVASCRIPT [VARIABLE_NAME] [POSITION]

### Environment (3 commands)
- add-cookie: NAME VALUE DOMAIN PATH [POSITION]
- delete-cookie: NAME [POSITION]
- clear-cookies: [POSITION]

### Dialog (3 commands)
- dismiss-alert: [POSITION]
- dismiss-confirm: ACCEPT [POSITION]
- dismiss-prompt: TEXT [POSITION]

### Frame/Tab (4 commands)
- switch-iframe: ELEMENT [POSITION]
- switch-next-tab: [POSITION]
- switch-prev-tab: [POSITION]
- switch-parent-frame: [POSITION]

### Utility (1 command)
- comment: COMMENT [POSITION]
EOF

echo -e "${GREEN}✓ Conversion guide created${NC}" | tee -a "$FIX_LOG"

# Summary
echo -e "\n${CYAN}=== SUMMARY ===${NC}" | tee -a "$FIX_LOG"
echo "1. Example modern implementation: src/cmd/create-step-wait-time-modern.go.example" | tee -a "$FIX_LOG"
echo "2. Conversion guide: ultrathink-conversion-guide.md" | tee -a "$FIX_LOG"
echo "3. Backup directory: $BACKUP_DIR" | tee -a "$FIX_LOG"
echo "" | tee -a "$FIX_LOG"
echo -e "${YELLOW}Next Steps:${NC}" | tee -a "$FIX_LOG"
echo "1. Review the example implementation" | tee -a "$FIX_LOG"
echo "2. Use the conversion guide to update each command" | tee -a "$FIX_LOG"
echo "3. Test both modern and legacy syntax" | tee -a "$FIX_LOG"
echo "4. Run comprehensive tests" | tee -a "$FIX_LOG"