#!/bin/bash
# Final demonstration of auto-journey handling

echo "🎯 Final Demo: Auto-Journey Handling"
echo "===================================="

export API_CLI_CONFIG="/Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator/config/virtuoso-config.yaml"
cd /Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator

# Create test structure
cat > /tmp/final-demo.yaml << EOF
project:
  name: "Auto Journey Demo $(date +%Y%m%d-%H%M%S)"
  
goals:
  - name: "First Goal (Has Auto Journey)"
    url: "https://demo1.example.com"
    journeys:
      - name: "Renamed Auto Journey"
        checkpoints:
          - name: "Homepage Check"
            navigation_url: "https://demo1.example.com"
            steps:
              - type: wait
                selector: ".content"
                timeout: 2000
                
      - name: "Second Journey (Created New)"
        checkpoints:
          - name: "About Page"
            navigation_url: "https://demo1.example.com/about"
            
  - name: "Second Goal (Also Has Auto Journey)"
    url: "https://demo2.example.com"
    journeys:
      - name: "Another Renamed Journey"
        checkpoints:
          - name: "Login Page"
            navigation_url: "https://demo2.example.com/login"
EOF

echo "📄 Test structure created"
echo ""

# Run with verbose to see the process
echo "🚀 Creating structure with verbose output..."
./bin/api-cli create-structure --file /tmp/final-demo.yaml --verbose

echo ""
echo "✅ Demo complete!"
echo ""
echo "Key observations:"
echo "1. First goal's first journey uses the auto-created journey ID"
echo "2. First goal's second journey creates a new journey"
echo "3. Second goal also gets an auto-created journey that we rename"
echo "4. All journeys end up with the correct names"