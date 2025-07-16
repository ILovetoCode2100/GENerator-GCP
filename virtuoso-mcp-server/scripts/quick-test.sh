#!/bin/bash
# Quick test script to validate MCP server functionality

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}Virtuoso MCP Server Quick Test${NC}\n"

# Check if config exists
CONFIG_FILE="$HOME/.api-cli/virtuoso-config.yaml"
if [ ! -f "$CONFIG_FILE" ]; then
    echo -e "${YELLOW}Warning: Config file not found at $CONFIG_FILE${NC}"
    echo "Please create the config file with your Virtuoso API credentials."
    echo ""
    echo "Example config:"
    echo "api:"
    echo "  auth_token: your-api-key-here"
    echo "  base_url: https://api-app2.virtuoso.qa/api"
    echo "organization:"
    echo "  id: \"2242\""
    exit 1
fi

# Build if needed
if [ ! -d "dist" ]; then
    echo -e "${YELLOW}Building server...${NC}"
    npm run build
fi

# Run validation
echo -e "\n${YELLOW}Running tool validation...${NC}"
npm run validate

# Run server test
echo -e "\n${YELLOW}Running server integration test...${NC}"
npm run test:server

echo -e "\n${GREEN}✅ Quick test completed!${NC}"
