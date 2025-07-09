#!/bin/bash
# validate-spec.sh - Validate OpenAPI specification
# Usage: ./scripts/validate-spec.sh [spec-file]

set -euo pipefail

SPEC_FILE="${1:-specs/api.yaml}"

# Colours
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}OpenAPI Specification Validator${NC}"
echo "================================="

# Check if file exists
if [ ! -f "$SPEC_FILE" ]; then
    echo -e "${RED}Error: Spec file not found at $SPEC_FILE${NC}"
    exit 1
fi

echo -e "Validating: ${YELLOW}$SPEC_FILE${NC}\n"

# Method 1: Using swagger-cli (if available)
if command -v swagger-cli &> /dev/null; then
    echo -e "${GREEN}Using swagger-cli...${NC}"
    swagger-cli validate "$SPEC_FILE"
elif command -v npx &> /dev/null; then
    echo -e "${GREEN}Using npx @apidevtools/swagger-cli...${NC}"
    npx @apidevtools/swagger-cli validate "$SPEC_FILE"
else
    echo -e "${YELLOW}No OpenAPI validator found. Installing minimal validator...${NC}"
    # Use Go-based validator as fallback
    go install github.com/getkin/kin-openapi/cmd/validate@latest
    validate "$SPEC_FILE"
fi

# Additional checks
echo -e "\n${BLUE}Running additional checks...${NC}"

# Check OpenAPI version
VERSION=$(grep -E "^openapi:" "$SPEC_FILE" | cut -d: -f2 | tr -d ' ')
echo -e "OpenAPI Version: ${GREEN}$VERSION${NC}"

# Count operations
PATHS=$(grep -E "^\s+/" "$SPEC_FILE" | wc -l)
echo -e "Number of paths: ${GREEN}$PATHS${NC}"

# Check for required fields
echo -e "\n${BLUE}Checking required fields...${NC}"
for field in "openapi" "info" "paths"; do
    if grep -q "^$field:" "$SPEC_FILE"; then
        echo -e "✅ $field: ${GREEN}Found${NC}"
    else
        echo -e "❌ $field: ${RED}Missing${NC}"
    fi
done

echo -e "\n${GREEN}✅ Validation complete!${NC}"
