#!/bin/bash
# Simple test for batch structure creation

echo "🧪 Simple Batch Structure Test"
echo "=============================="

export API_CLI_CONFIG="/Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator/config/virtuoso-config.yaml"
cd /Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator

# Test with a simple structure file
cat > /tmp/simple-test.yaml << EOF
project:
  name: "Simple Test Project $(date +%Y%m%d-%H%M%S)"
  
goals:
  - name: "Simple Goal"
    url: "https://simple.example.com"
    journeys:
      - name: "Simple Journey"
        checkpoints:
          - name: "Homepage"
            navigation_url: "https://simple.example.com"
            steps:
              - type: wait
                selector: "body"
                timeout: 1000
EOF

echo "📄 Structure file created at /tmp/simple-test.yaml"
echo ""
echo "🔍 Running dry-run first..."
./bin/api-cli create-structure --file /tmp/simple-test.yaml --dry-run

echo ""
echo "🏗️  Creating actual structure..."
./bin/api-cli create-structure --file /tmp/simple-test.yaml --verbose

echo ""
echo "✅ Test complete!"