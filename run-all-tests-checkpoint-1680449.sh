#!/bin/bash

# Master test runner for checkpoint 1680449
# Executes both basic and ULTRATHINK test frameworks

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
PURPLE='\033[0;35m'
NC='\033[0m'

# Configuration
CHECKPOINT_ID="1680449"
export VIRTUOSO_API_TOKEN="f7a55516-5cc4-4529-b2ae-8e106a7d164e"

echo -e "${CYAN}╔══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║          COMPREHENSIVE STEP TESTING SUITE                    ║${NC}"
echo -e "${CYAN}║                 Checkpoint: $CHECKPOINT_ID                      ║${NC}"
echo -e "${CYAN}╚══════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Ensure CLI is built
echo -e "${BLUE}[1/4] Building CLI...${NC}"
if make build; then
    echo -e "${GREEN}✓ CLI built successfully${NC}"
else
    echo -e "${YELLOW}⚠ Build failed, attempting to use existing binary${NC}"
fi

# Validate configuration
echo -e "\n${BLUE}[2/4] Validating configuration...${NC}"
if ./bin/api-cli validate-config; then
    echo -e "${GREEN}✓ Configuration valid${NC}"
else
    echo -e "${YELLOW}⚠ Configuration validation failed, continuing anyway${NC}"
fi

# Run basic comprehensive test
echo -e "\n${BLUE}[3/4] Running comprehensive step tests...${NC}"
echo -e "${PURPLE}This will test all 47 step commands with basic scenarios${NC}"
echo ""

if ./test-all-steps-checkpoint-1680449.sh; then
    echo -e "\n${GREEN}✓ Basic comprehensive tests completed${NC}"
    BASIC_RESULT="PASSED"
else
    echo -e "\n${YELLOW}⚠ Some basic tests failed${NC}"
    BASIC_RESULT="FAILED"
fi

# Run ULTRATHINK framework
echo -e "\n${BLUE}[4/4] Running ULTRATHINK test framework...${NC}"
echo -e "${PURPLE}This uses sub-agents for advanced testing scenarios${NC}"
echo ""

if ./ultrathink-test-framework.sh; then
    echo -e "\n${GREEN}✓ ULTRATHINK tests completed${NC}"
    ULTRA_RESULT="PASSED"
else
    echo -e "\n${YELLOW}⚠ Some ULTRATHINK tests failed${NC}"
    ULTRA_RESULT="FAILED"
fi

# Final summary
echo -e "\n${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}                        FINAL SUMMARY                          ${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo ""
echo -e "Checkpoint ID: ${PURPLE}$CHECKPOINT_ID${NC}"
echo -e "Basic Tests: ${BASIC_RESULT}"
echo -e "ULTRATHINK Tests: ${ULTRA_RESULT}"
echo ""
echo -e "Test Logs:"
echo -e "  - Basic: ${BLUE}test-results-checkpoint-${CHECKPOINT_ID}.log${NC}"
echo -e "  - ULTRATHINK: ${BLUE}ultrathink-results/${NC}"
echo ""

if [ "$BASIC_RESULT" = "PASSED" ] && [ "$ULTRA_RESULT" = "PASSED" ]; then
    echo -e "${GREEN}✓ ALL TESTS PASSED!${NC}"
    echo -e "${GREEN}Checkpoint $CHECKPOINT_ID is fully validated for all 47 step commands${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠ Some tests failed - review logs for details${NC}"
    exit 1
fi