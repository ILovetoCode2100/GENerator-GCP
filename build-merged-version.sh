#!/bin/bash

# Build script for merged Version A with Version B enhancements

set -e  # Exit on error

VERSION_A_DIR="/Users/marklovelady/_dev/virtuoso-api-cli-generator"

echo "========================================="
echo "Building Merged Version A"
echo "========================================="

# Change to Version A directory
cd "$VERSION_A_DIR"

# Step 1: Clean previous builds
echo "📧 Cleaning previous builds..."
rm -f bin/api-cli
mkdir -p bin

# Step 2: Update dependencies
echo "📦 Updating Go dependencies..."
go mod tidy 2>&1 | tee go-mod-tidy.log

# Step 3: Download dependencies
echo "⬇️  Downloading dependencies..."
go mod download

# Step 4: Run go vet
echo "🔍 Running go vet..."
if go vet ./... 2>&1 | tee go-vet.log; then
    echo "✅ Go vet passed"
else
    echo "⚠️  Go vet found issues (see go-vet.log)"
fi

# Step 5: Build the binary
echo "🔨 Building api-cli binary..."
if go build -v -o bin/api-cli ./src/cmd 2>&1 | tee build.log; then
    echo "✅ Build successful!"
    echo "📍 Binary location: $VERSION_A_DIR/bin/api-cli"
    
    # Check binary size
    SIZE=$(ls -lh bin/api-cli | awk '{print $5}')
    echo "📊 Binary size: $SIZE"
    
    # Make it executable
    chmod +x bin/api-cli
else
    echo "❌ Build failed! Check build.log for details"
    exit 1
fi

# Step 6: Quick sanity check
echo ""
echo "🧪 Running sanity check..."
if ./bin/api-cli --version 2>&1 | grep -q "api-cli version"; then
    echo "✅ Binary executes successfully"
else
    echo "⚠️  Binary may have issues with --version flag"
fi

# Step 7: List available commands
echo ""
echo "📋 Checking available commands..."
./bin/api-cli --help 2>&1 | grep "create-step-" | wc -l | xargs -I {} echo "Found {} create-step commands"

# Step 8: Check for Version B commands
echo ""
echo "🔍 Verifying Version B commands are present..."
VERSION_B_COMMANDS=(
    "create-step-cookie-create"
    "create-step-cookie-wipe-all"
    "create-step-execute-script"
    "create-step-mouse-move-to"
    "create-step-mouse-move-by"
    "create-step-pick-index"
    "create-step-store-element-text"
    "create-step-window-resize"
)

MISSING_COMMANDS=0
for cmd in "${VERSION_B_COMMANDS[@]}"; do
    if ./bin/api-cli --help 2>&1 | grep -q "$cmd"; then
        echo "✅ Found: $cmd"
    else
        echo "❌ Missing: $cmd"
        MISSING_COMMANDS=$((MISSING_COMMANDS + 1))
    fi
done

if [ $MISSING_COMMANDS -eq 0 ]; then
    echo ""
    echo "✅ All Version B commands are available!"
else
    echo ""
    echo "⚠️  $MISSING_COMMANDS Version B commands are missing"
fi

echo ""
echo "========================================="
echo "Build Summary"
echo "========================================="
echo "✅ Build completed successfully"
echo "📍 Binary: $VERSION_A_DIR/bin/api-cli"
echo "📝 Logs saved in: $VERSION_A_DIR/"
echo ""
echo "Next step: Run test-merged-version.sh to test functionality"
echo "========================================="