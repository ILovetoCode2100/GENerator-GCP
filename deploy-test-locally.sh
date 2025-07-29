#!/bin/bash

echo "Deploying Rocketshop test using local CLI..."
echo ""

# Create a comprehensive YAML test file
cat > rocketshop-complete-test.yaml << 'EOF'
test: "Rocketshop Purchase Flow - Complete Test"
description: "Full e-commerce purchase flow with all steps"
project_id: 9349
goal_name: "E-commerce Purchase Flow"
journey_name: "Complete Purchase Journey"
checkpoint_name: "Full Test Run"
nav: https://rocketshop.virtuoso.qa
data:
  customerName: John Doe
  customerEmail: johndoe@example.com
  streetAddress: 123 Elm Street
  phoneNumber: 555-1234
  postalCode: 90210
  creditCard: 4111 1111 1111 1111
  securityCode: 234
do:
  # Verify page loaded
  - ch: "Border Not Found"

  # Wait for page to fully load
  - wait: 20000

  # Add item to shopping bag
  - c: "Add to Bag"

  # Navigate to shopping bag
  - c: "Shopping Bag"

  # Verify on shopping bag page
  - ch: Shopping Bag in h1

  # Proceed to checkout
  - c: "Go to Checkout"

  # Fill checkout form
  - t: $customerName in "Full name"
  - t: $customerEmail in "Email"
  - t: $streetAddress in "Address"
  - t: $phoneNumber in "Phone numbers"
  - t: $postalCode in "ZIP code"

  # Enter payment information
  - t: $creditCard in "Card number"
  - t: $securityCode in "xxx"

  # Complete purchase
  - c: "Confirm and Pay"

  # Wait for confirmation
  - wait: 20000

  # Verify purchase confirmation
  - ch: Purchase Confirmed! in h3

  # Download confirmation
  - c: "Download Confirmation"
EOF

# Deploy using the CLI
echo "Running test deployment..."
./bin/api-cli run-test rocketshop-complete-test.yaml --execute --output json > test-result.json

# Check result
if [ $? -eq 0 ]; then
    echo ""
    echo "Test successfully deployed!"
    echo ""
    echo "Results:"
    cat test-result.json | jq '.'

    # Extract checkpoint URL
    CHECKPOINT_ID=$(cat test-result.json | jq -r '.checkpoint_id')
    echo ""
    echo "View your test at: https://app.virtuoso.qa/#/checkpoint/$CHECKPOINT_ID"
else
    echo ""
    echo "Test deployment failed. Check test-result.json for details."
fi

echo ""
echo "Note: The GCP Cloud Run API deployment can be completed by running:"
echo "  cd gcp && ./deploy.sh"
echo ""
echo "Once deployed, you can use the GCP API endpoint to run tests remotely."
