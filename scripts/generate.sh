#!/bin/bash
# generate.sh - Generate Go code from OpenAPI specification
# Usage: ./scripts/generate.sh [spec-file]

set -euo pipefail

# Configuration
SPEC_FILE="${1:-specs/api.yaml}"
OUTPUT_DIR="src/api"
PACKAGE_NAME="api"

# Colours for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Colour

# Check if spec file exists
if [ ! -f "$SPEC_FILE" ]; then
    echo -e "${RED}Error: OpenAPI spec not found at $SPEC_FILE${NC}"
    echo "Please place your OpenAPI spec file at: $SPEC_FILE"
    exit 1
fi

# Check if oapi-codegen is installed
if ! command -v oapi-codegen &> /dev/null; then
    echo -e "${YELLOW}oapi-codegen not found. Installing...${NC}"
    go install github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@latest
fi

# Create output directory
mkdir -p "$OUTPUT_DIR"

echo -e "${GREEN}Generating Go code from $SPEC_FILE...${NC}"

# Generate types
echo "Generating types..."
oapi-codegen \
    -package "$PACKAGE_NAME" \
    -generate types \
    -o "$OUTPUT_DIR/types.gen.go" \
    "$SPEC_FILE"

# Generate client
echo "Generating client..."
oapi-codegen \
    -package "$PACKAGE_NAME" \
    -generate client \
    -o "$OUTPUT_DIR/client.gen.go" \
    "$SPEC_FILE"

# Generate spec (embedded)
echo "Generating embedded spec..."
oapi-codegen \
    -package "$PACKAGE_NAME" \
    -generate spec \
    -o "$OUTPUT_DIR/spec.gen.go" \
    "$SPEC_FILE"

# Create a configuration file for future use
cat > "$OUTPUT_DIR/.oapi-codegen.yaml" << EOF
package: $PACKAGE_NAME
generate:
  - types
  - client
  - spec
output: $OUTPUT_DIR/generated.go
EOF

echo -e "${GREEN}✅ Code generation complete!${NC}"
echo "Generated files:"
echo "  - $OUTPUT_DIR/types.gen.go"
echo "  - $OUTPUT_DIR/client.gen.go"
echo "  - $OUTPUT_DIR/spec.gen.go"

# Run go mod tidy to ensure dependencies are correct
echo -e "\n${GREEN}Running go mod tidy...${NC}"
go mod tidy

echo -e "\n${GREEN}All done! You can now build the CLI with 'make build'${NC}"
