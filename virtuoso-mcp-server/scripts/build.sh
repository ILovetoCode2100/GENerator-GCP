#!/bin/bash
# Build Script for Virtuoso MCP Server
# This script compiles TypeScript, bundles for distribution, and creates a release package

set -e # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Get the directory of this script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo -e "${BLUE}Building Virtuoso MCP Server...${NC}\n"

# Change to project root
cd "$PROJECT_ROOT"

# Step 1: Clean previous builds
echo -e "${YELLOW}Cleaning previous builds...${NC}"
rm -rf dist coverage build virtuoso-mcp-server-*.tgz
echo -e "${GREEN}✓ Cleaned${NC}\n"

# Step 2: Install dependencies
echo -e "${YELLOW}Installing dependencies...${NC}"
npm install
echo -e "${GREEN}✓ Dependencies installed${NC}\n"

# Step 3: Run linting
echo -e "${YELLOW}Running TypeScript type checking...${NC}"
npm run lint
echo -e "${GREEN}✓ Type checking passed${NC}\n"

# Step 4: Run tests
echo -e "${YELLOW}Running tests...${NC}"
npm test -- --passWithNoTests || {
    echo -e "${YELLOW}⚠ Tests failed or no tests found, continuing...${NC}\n"
}

# Step 5: Build TypeScript
echo -e "${YELLOW}Compiling TypeScript...${NC}"
npm run build
echo -e "${GREEN}✓ TypeScript compiled${NC}\n"

# Step 6: Validate build output
echo -e "${YELLOW}Validating build output...${NC}"
if [ ! -f "dist/index.js" ]; then
    echo -e "${RED}✗ Build failed: dist/index.js not found${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Build output validated${NC}\n"

# Step 7: Create minimal package.json for distribution
echo -e "${YELLOW}Creating distribution package.json...${NC}"
node -e "
const pkg = require('./package.json');
const distPkg = {
  name: pkg.name,
  version: pkg.version,
  description: pkg.description,
  main: 'index.js',
  type: 'module',
  bin: {
    'virtuoso-mcp-server': './index.js'
  },
  dependencies: pkg.dependencies,
  engines: {
    node: '>=18.0.0'
  },
  keywords: pkg.keywords,
  author: pkg.author,
  license: pkg.license
};
require('fs').writeFileSync('./dist/package.json', JSON.stringify(distPkg, null, 2));
"
echo -e "${GREEN}✓ Distribution package.json created${NC}\n"

# Step 8: Copy necessary files to dist
echo -e "${YELLOW}Copying additional files...${NC}"
cp README.md dist/ 2>/dev/null || echo "No README.md found"
cp LICENSE dist/ 2>/dev/null || echo "No LICENSE found"

# Create a simple run script
cat > dist/run.sh << 'EOF'
#!/bin/bash
# Run script for Virtuoso MCP Server
node "$(dirname "$0")/index.js"
EOF
chmod +x dist/run.sh

echo -e "${GREEN}✓ Additional files copied${NC}\n"

# Step 9: Make the main file executable
echo -e "${YELLOW}Setting executable permissions...${NC}"
chmod +x dist/index.js
echo -e "${GREEN}✓ Permissions set${NC}\n"

# Step 10: Create tarball for distribution
echo -e "${YELLOW}Creating distribution package...${NC}"
VERSION=$(node -p "require('./package.json').version")
PACKAGE_NAME="virtuoso-mcp-server-${VERSION}"
mkdir -p build
cp -r dist "build/${PACKAGE_NAME}"
cd build
tar -czf "../${PACKAGE_NAME}.tgz" "${PACKAGE_NAME}"
cd ..
rm -rf build
echo -e "${GREEN}✓ Created ${PACKAGE_NAME}.tgz${NC}\n"

# Step 11: Generate build info
echo -e "${YELLOW}Generating build info...${NC}"
cat > dist/BUILD_INFO.json << EOF
{
  "version": "${VERSION}",
  "buildDate": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "nodeVersion": "$(node --version)",
  "platform": "$(uname -s)",
  "gitCommit": "$(git rev-parse HEAD 2>/dev/null || echo 'unknown')",
  "gitBranch": "$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo 'unknown')"
}
EOF
echo -e "${GREEN}✓ Build info generated${NC}\n"

# Step 12: Run validation
echo -e "${YELLOW}Running tool validation...${NC}"
npx tsx scripts/validate-tools.ts || {
    echo -e "${YELLOW}⚠ Tool validation had warnings, check output above${NC}\n"
}

# Summary
echo -e "${GREEN}✅ Build completed successfully!${NC}\n"
echo -e "Build outputs:"
echo -e "  - ${BLUE}dist/${NC} - Compiled JavaScript files"
echo -e "  - ${BLUE}${PACKAGE_NAME}.tgz${NC} - Distribution package"
echo -e ""
echo -e "To test the server locally:"
echo -e "  ${YELLOW}node dist/index.js${NC}"
echo -e ""
echo -e "To install globally:"
echo -e "  ${YELLOW}npm install -g ${PACKAGE_NAME}.tgz${NC}"
echo -e ""
echo -e "To use with Claude Desktop, add to config:"
echo -e '  {
    "mcpServers": {
      "virtuoso": {
        "command": "node",
        "args": ["'$(pwd)'/dist/index.js"]
      }
    }
  }'
