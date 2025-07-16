#!/bin/bash

# Virtuoso MCP Server Installation Verification Script
# This script checks that the MCP server is properly installed and configured

set -e

echo "🔍 Verifying Virtuoso MCP Server Installation..."
echo "============================================="

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check Node.js
echo -n "Checking Node.js installation... "
if command -v node &> /dev/null; then
    NODE_VERSION=$(node --version)
    echo -e "${GREEN}✓${NC} Found Node.js $NODE_VERSION"
else
    echo -e "${RED}✗${NC} Node.js not found. Please install Node.js 18 or later."
    exit 1
fi

# Check npm
echo -n "Checking npm installation... "
if command -v npm &> /dev/null; then
    NPM_VERSION=$(npm --version)
    echo -e "${GREEN}✓${NC} Found npm $NPM_VERSION"
else
    echo -e "${RED}✗${NC} npm not found. Please install npm."
    exit 1
fi

# Check TypeScript
echo -n "Checking TypeScript installation... "
if npm list typescript &> /dev/null; then
    TS_VERSION=$(npm list typescript --depth=0 | grep typescript | awk '{print $2}')
    echo -e "${GREEN}✓${NC} Found TypeScript $TS_VERSION"
else
    echo -e "${YELLOW}⚠${NC} TypeScript not installed. Run 'npm install' to install dependencies."
fi

# Check required files
echo ""
echo "Checking required files..."
REQUIRED_FILES=(
    "package.json"
    "tsconfig.json"
    "src/index.ts"
    "src/server.ts"
    "src/cli-wrapper.ts"
    "claude_desktop_config.json"
)

for file in "${REQUIRED_FILES[@]}"; do
    echo -n "  Checking $file... "
    if [ -f "$file" ]; then
        echo -e "${GREEN}✓${NC}"
    else
        echo -e "${RED}✗${NC} Missing"
        exit 1
    fi
done

# Check tool implementations
echo ""
echo "Checking tool implementations..."
TOOL_FILES=(
    "assert" "interact" "navigate" "data" "wait" "dialog"
    "window" "mouse" "select" "file" "misc" "library"
)

for tool in "${TOOL_FILES[@]}"; do
    echo -n "  Checking src/tools/$tool.ts... "
    if [ -f "src/tools/$tool.ts" ]; then
        echo -e "${GREEN}✓${NC}"
    else
        echo -e "${RED}✗${NC} Missing"
        exit 1
    fi
done

# Check if node_modules exists
echo ""
echo -n "Checking dependencies installation... "
if [ -d "node_modules" ]; then
    echo -e "${GREEN}✓${NC} Dependencies installed"
else
    echo -e "${YELLOW}⚠${NC} Dependencies not installed. Run 'npm install'"
fi

# Check if built
echo -n "Checking build output... "
if [ -d "dist" ] && [ -f "dist/index.js" ]; then
    echo -e "${GREEN}✓${NC} Project is built"
else
    echo -e "${YELLOW}⚠${NC} Project not built. Run 'npm run build'"
fi

# Check Virtuoso CLI
echo ""
echo -n "Checking Virtuoso CLI path... "
# Read CLI path from package.json or a config file
CLI_PATH="../bin/api-cli"
if [ -f "$CLI_PATH" ]; then
    echo -e "${GREEN}✓${NC} Found at $CLI_PATH"
else
    echo -e "${YELLOW}⚠${NC} Virtuoso CLI not found at expected path: $CLI_PATH"
    echo "    Update the path in claude_desktop_config.json"
fi

# Check Virtuoso config
echo -n "Checking Virtuoso config... "
VIRTUOSO_CONFIG_PATHS=(
    "$HOME/.api-cli/virtuoso-config.yaml"
    "./virtuoso-config.yaml"
    "../virtuoso-config.yaml"
)

CONFIG_FOUND=false
for config_path in "${VIRTUOSO_CONFIG_PATHS[@]}"; do
    if [ -f "$config_path" ]; then
        echo -e "${GREEN}✓${NC} Found at $config_path"
        CONFIG_FOUND=true
        break
    fi
done

if [ "$CONFIG_FOUND" = false ]; then
    echo -e "${YELLOW}⚠${NC} Virtuoso config not found. Create one at ~/.api-cli/virtuoso-config.yaml"
fi

# Summary
echo ""
echo "============================================="
echo "📊 Verification Summary"
echo "============================================="

# Count tools
TOOL_COUNT=$(ls src/tools/*.ts 2>/dev/null | wc -l | tr -d ' ')
echo "✅ Found $TOOL_COUNT tool implementations"

# Test compilation
echo ""
echo -n "Testing TypeScript compilation... "
if npx tsc --noEmit 2>/dev/null; then
    echo -e "${GREEN}✓${NC} TypeScript compilation successful"
else
    echo -e "${RED}✗${NC} TypeScript compilation failed"
    echo "   Run 'npx tsc' to see errors"
fi

echo ""
echo "============================================="
echo ""

# Final status
if [ -d "node_modules" ] && [ -d "dist" ] && [ "$CONFIG_FOUND" = true ]; then
    echo -e "${GREEN}✅ Virtuoso MCP Server is ready to use!${NC}"
    echo ""
    echo "To use with Claude Desktop:"
    echo "1. Copy claude_desktop_config.json to your Claude Desktop config directory"
    echo "2. Update the paths in the config file"
    echo "3. Restart Claude Desktop"
else
    echo -e "${YELLOW}⚠ Additional setup required:${NC}"
    [ ! -d "node_modules" ] && echo "  - Run 'npm install' to install dependencies"
    [ ! -d "dist" ] && echo "  - Run 'npm run build' to build the project"
    [ "$CONFIG_FOUND" = false ] && echo "  - Create Virtuoso config file"
fi

echo ""
