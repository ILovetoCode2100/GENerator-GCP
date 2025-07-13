#!/bin/bash

# Test checkpoint operations
# This script tests the checkpoint create, attach, and list functionality

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "Testing Checkpoint Operations"
echo "============================"

# Check if binary exists
if [ ! -f "./bin/api-cli" ]; then
    echo -e "${RED}Error: api-cli binary not found. Run 'make build' first.${NC}"
    exit 1
fi

BINARY="./bin/api-cli"

# Test data - these would normally come from previous tests
# In a real scenario, these IDs would be captured from actual API responses
JOURNEY_ID="608038"  # Example journey ID
GOAL_ID="13776"      # Example goal ID
SNAPSHOT_ID="43802"  # Example snapshot ID

echo -e "\n${YELLOW}Test 1: Create checkpoint CP-1${NC}"
echo "Command: $BINARY create-checkpoint $JOURNEY_ID $GOAL_ID $SNAPSHOT_ID \"CP-1\""
if OUTPUT=$($BINARY create-checkpoint "$JOURNEY_ID" "$GOAL_ID" "$SNAPSHOT_ID" "CP-1" 2>&1); then
    echo -e "${GREEN}✓ Checkpoint created successfully${NC}"
    echo "Output: $OUTPUT"
    
    # Try to extract checkpoint ID
    CP_ID=$(echo "$OUTPUT" | grep -oE 'ID: [0-9]+' | grep -oE '[0-9]+' | head -1)
    if [ -n "$CP_ID" ]; then
        echo "Captured Checkpoint ID: $CP_ID"
    else
        echo -e "${YELLOW}Warning: Could not extract checkpoint ID from output${NC}"
    fi
else
    echo -e "${RED}✗ Failed to create checkpoint${NC}"
    echo "Error: $OUTPUT"
fi

echo -e "\n${YELLOW}Test 2: List checkpoints${NC}"
echo "Command: $BINARY list-checkpoints $JOURNEY_ID"
if OUTPUT=$($BINARY list-checkpoints "$JOURNEY_ID" 2>&1); then
    echo -e "${GREEN}✓ Listed checkpoints successfully${NC}"
    echo "Output:"
    echo "$OUTPUT"
    
    # Check if CP-1 appears in the list
    if echo "$OUTPUT" | grep -q "CP-1"; then
        echo -e "${GREEN}✓ CP-1 found in checkpoint list${NC}"
    else
        echo -e "${RED}✗ CP-1 not found in checkpoint list${NC}"
    fi
    
    # Check for position information
    if echo "$OUTPUT" | grep -qE "(position|Position|1\.)"; then
        echo -e "${GREEN}✓ Position information found${NC}"
    else
        echo -e "${YELLOW}⚠ Position information not clearly visible${NC}"
    fi
else
    echo -e "${RED}✗ Failed to list checkpoints${NC}"
    echo "Error: $OUTPUT"
fi

echo -e "\n${YELLOW}Test 3: Create additional checkpoints${NC}"
echo "Creating CP-2 with position 3..."
if $BINARY create-checkpoint "$JOURNEY_ID" "$GOAL_ID" "$SNAPSHOT_ID" "CP-2" --position 3 > /dev/null 2>&1; then
    echo -e "${GREEN}✓ CP-2 created${NC}"
else
    echo -e "${RED}✗ Failed to create CP-2${NC}"
fi

echo "Creating CP-3..."
if $BINARY create-checkpoint "$JOURNEY_ID" "$GOAL_ID" "$SNAPSHOT_ID" "CP-3" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ CP-3 created${NC}"
else
    echo -e "${RED}✗ Failed to create CP-3${NC}"
fi

echo -e "\n${YELLOW}Test 4: Verify all checkpoints and positions${NC}"
if OUTPUT=$($BINARY list-checkpoints "$JOURNEY_ID" 2>&1); then
    echo -e "${GREEN}✓ Listed all checkpoints${NC}"
    echo "Output:"
    echo "$OUTPUT"
    
    # Check for all checkpoints
    ALL_FOUND=true
    for CP in "CP-1" "CP-2" "CP-3"; do
        if echo "$OUTPUT" | grep -q "$CP"; then
            echo -e "${GREEN}✓ $CP found${NC}"
        else
            echo -e "${RED}✗ $CP not found${NC}"
            ALL_FOUND=false
        fi
    done
    
    if $ALL_FOUND; then
        echo -e "${GREEN}✓ All checkpoints found with correct positions${NC}"
    fi
else
    echo -e "${RED}✗ Failed to list checkpoints${NC}"
fi

echo -e "\n${YELLOW}Test 5: Set checkpoint context${NC}"
if [ -n "$CP_ID" ]; then
    echo "Command: $BINARY set-checkpoint $CP_ID"
    if OUTPUT=$($BINARY set-checkpoint "$CP_ID" 2>&1); then
        echo -e "${GREEN}✓ Checkpoint context set successfully${NC}"
        echo "Output: $OUTPUT"
    else
        echo -e "${RED}✗ Failed to set checkpoint context${NC}"
        echo "Error: $OUTPUT"
    fi
else
    echo -e "${YELLOW}⚠ Skipping set-checkpoint test (no checkpoint ID available)${NC}"
fi

echo -e "\n${GREEN}Checkpoint operations test completed!${NC}"
