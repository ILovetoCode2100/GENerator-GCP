#!/bin/bash
# Test journey name updates

echo "🧪 Testing Journey Name Updates"
echo "=============================="

export API_CLI_CONFIG="/Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator/config/virtuoso-config.yaml"
cd /Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator

# Create a test structure with specific journey names
cat > /tmp/journey-name-test.yaml << EOF
project:
  name: "Journey Name Test $(date +%Y%m%d-%H%M%S)"
  
goals:
  - name: "Test Goal with Custom Journey"
    url: "https://example.com"
    journeys:
      - name: "My Custom Journey Name"
        checkpoints:
          - name: "Homepage"
            navigation_url: "https://example.com"
            steps:
              - type: wait
                selector: "body"
                timeout: 1000
                
      - name: "Second Custom Journey"
        checkpoints:
          - name: "About Page"
            navigation_url: "https://example.com/about"
EOF

echo "📄 Testing with custom journey names..."
echo ""

# Run with verbose to see the renaming process
./bin/api-cli create-structure --file /tmp/journey-name-test.yaml --verbose

echo ""
echo "✅ Test complete! Check the Virtuoso UI to verify journey names."