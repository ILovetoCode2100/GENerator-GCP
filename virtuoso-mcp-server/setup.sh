#!/bin/bash

# Virtuoso MCP Server Setup Script

set -e

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}Virtuoso MCP Server Setup${NC}"
echo "=========================="

# Check Node.js version
echo -e "\n${YELLOW}Checking prerequisites...${NC}"
NODE_VERSION=$(node -v | cut -d'v' -f2 | cut -d'.' -f1)
if [ "$NODE_VERSION" -lt 18 ]; then
    echo -e "${RED}Error: Node.js 18 or higher required${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Node.js version OK${NC}"

# Install dependencies
echo -e "\n${YELLOW}Installing dependencies...${NC}"
npm install

# Check for CLI binary
echo -e "\n${YELLOW}Checking for Virtuoso CLI...${NC}"
CLI_PATH="../bin/api-cli"
if [ ! -f "$CLI_PATH" ]; then
    echo -e "${RED}Error: API CLI not found at $CLI_PATH${NC}"
    echo "Please build the Virtuoso API CLI first"
    exit 1
fi
echo -e "${GREEN}✓ CLI found at $CLI_PATH${NC}"

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo -e "\n${YELLOW}Creating .env file...${NC}"
    cp .env.example .env

    # Update CLI path in .env
    sed -i.bak "s|VIRTUOSO_CLI_PATH=.*|VIRTUOSO_CLI_PATH=$(pwd)/$CLI_PATH|" .env
    rm .env.bak

    echo -e "${GREEN}✓ Created .env file${NC}"
    echo -e "${YELLOW}Please update .env with your configuration paths${NC}"
fi

# Build the server
echo -e "\n${YELLOW}Building TypeScript...${NC}"
npm run build
echo -e "${GREEN}✓ Build complete${NC}"

# Generate Claude Desktop config
echo -e "\n${YELLOW}Generating Claude Desktop configuration...${NC}"
cat > claude-desktop-config.json << EOF
{
  "mcpServers": {
    "virtuoso": {
      "command": "node",
      "args": ["$(pwd)/dist/index.js"],
      "env": {
        "VIRTUOSO_CLI_PATH": "$(pwd)/$CLI_PATH",
        "VIRTUOSO_CONFIG_PATH": "$HOME/.api-cli/virtuoso-config.yaml"
      }
    }
  }
}
EOF

echo -e "${GREEN}✓ Generated claude-desktop-config.json${NC}"

# Instructions
echo -e "\n${BLUE}Setup Complete!${NC}"
echo -e "\n${YELLOW}Next Steps:${NC}"
echo "1. Update .env with your configuration"
echo "2. Ensure virtuoso-config.yaml exists with your API credentials"
echo "3. Add the configuration to Claude Desktop:"
echo "   - macOS: ~/Library/Application Support/Claude/claude_desktop_config.json"
echo "   - Windows: %APPDATA%/Claude/claude_desktop_config.json"
echo "   - Linux: ~/.config/Claude/claude_desktop_config.json"
echo -e "\n${YELLOW}Configuration to add:${NC}"
cat claude-desktop-config.json

echo -e "\n${YELLOW}To test the server:${NC}"
echo "npm test"
echo -e "\n${YELLOW}To run in development mode:${NC}"
echo "npm run dev"
