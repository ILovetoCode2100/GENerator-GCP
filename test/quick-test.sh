#!/bin/bash
# Quick test of the enhanced CLI features

cd /Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator

echo "🧪 Quick Test of Enhanced Virtuoso CLI"
echo "======================================"

echo -e "\n1️⃣ Testing Configuration Validation..."
./bin/api-cli validate-config --config ./config/virtuoso-config.yaml

echo -e "\n2️⃣ Testing Batch Creation (Dry Run)..."
./bin/api-cli create-structure --file examples/test-small.yaml --config ./config/virtuoso-config.yaml --dry-run

echo -e "\n3️⃣ Creating a Test Structure..."
# Create unique test file
TIMESTAMP=$(date +%s)
cat > /tmp/quick-test-$TIMESTAMP.yaml << EOF
project:
  name: "Quick Test $TIMESTAMP"
  description: "Quick test at $(date)"
goals:
  - name: "Quick Goal"
    url: "https://example.com"
    journeys:
      - name: "Quick Journey"
        checkpoints:
          - name: "Quick Check"
            steps:
              - type: navigate
                url: "https://example.com"
              - type: wait
                selector: "body"
                timeout: 1000
EOF

./bin/api-cli create-structure --file /tmp/quick-test-$TIMESTAMP.yaml --config ./config/virtuoso-config.yaml

echo -e "\n4️⃣ Testing Output Formats..."
echo "JSON output:"
./bin/api-cli validate-config --config ./config/virtuoso-config.yaml -o json | jq .status

echo -e "\n✅ Quick test complete!"
echo "Check the Virtuoso UI at: https://app2.virtuoso.qa"

# Cleanup
rm -f /tmp/quick-test-$TIMESTAMP.yaml