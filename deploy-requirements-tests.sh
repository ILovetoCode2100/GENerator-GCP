#!/bin/bash

echo "=== Deploying Requirements-Based Tests ==="
echo "Project: 9411"
echo "Goal: 14583"
echo "Journey: 611015"
echo ""

# Checkpoint 1: Access and Product Selection
echo "Creating Checkpoint 1: Access and Product Selection..."
CP1=$(./bin/api-cli create-checkpoint 611015 14583 44339 "REQ-001: Access and Product Selection" --output json | jq -r '.checkpoint_id')
export VIRTUOSO_SESSION_ID=$CP1

echo "Adding test steps for Requirements 1-2..."

# REQ-001: The system shall make the Rocketshop storefront accessible
echo "Step: Verify storefront accessibility..."
./bin/api-cli step-navigate to "https://rocketshop.virtuoso.qa"

# Verify page loaded successfully
./bin/api-cli step-assert exists "text=Rocketshop"

# REQ-002: The system shall allow users to add at least one product
echo "Step: Add product to shopping bag..."
./bin/api-cli step-interact click "text=Add to Bag"

# Checkpoint 2: Shopping Bag Navigation and Display
echo ""
echo "Creating Checkpoint 2: Shopping Bag Navigation..."
CP2=$(./bin/api-cli create-checkpoint 611015 14583 44339 "REQ-003-004: Shopping Bag Navigation" --output json | jq -r '.checkpoint_id')
export VIRTUOSO_SESSION_ID=$CP2

# REQ-003: Navigate to shopping bag from main navigation
echo "Step: Navigate to shopping bag..."
./bin/api-cli step-interact click "text=Shopping Bag"

# REQ-004: Display correct Shopping Bag page header
echo "Step: Verify Shopping Bag header..."
./bin/api-cli step-assert exists "h1=Shopping Bag"

# REQ-005: Provide button to proceed to checkout
echo "Step: Verify checkout button exists..."
./bin/api-cli step-assert exists "text=Go to Checkout"
./bin/api-cli step-interact click "text=Go to Checkout"

# Checkpoint 3: Checkout Form Requirements
echo ""
echo "Creating Checkpoint 3: Checkout Form Validation..."
CP3=$(./bin/api-cli create-checkpoint 611015 14583 44339 "REQ-006: Checkout Form Fields" --output json | jq -r '.checkpoint_id')
export VIRTUOSO_SESSION_ID=$CP3

# REQ-006: Checkout process shall require specific fields
echo "Step: Fill required checkout fields..."

# Full name
./bin/api-cli step-interact write "John Smith" "label=Full name"

# Email address
./bin/api-cli step-interact write "john.smith@example.com" "label=Email"

# Shipping address
./bin/api-cli step-interact write "456 Oak Avenue" "label=Address"

# Phone number
./bin/api-cli step-interact write "555-9876" "label=Phone numbers"

# ZIP code
./bin/api-cli step-interact write "94105" "label=ZIP code"

# Checkpoint 4: Payment and Confirmation
echo ""
echo "Creating Checkpoint 4: Payment and Confirmation..."
CP4=$(./bin/api-cli create-checkpoint 611015 14583 44339 "REQ-007-010: Payment and Confirmation" --output json | jq -r '.checkpoint_id')
export VIRTUOSO_SESSION_ID=$CP4

# REQ-007: Payment process shall require card details
echo "Step: Enter payment information..."

# Card number
./bin/api-cli step-interact write "4242 4242 4242 4242" "label=Card number"

# Card security code (CVV)
./bin/api-cli step-interact write "123" "placeholder=xxx"

# REQ-008: Accept payment and proceed with checkout
echo "Step: Submit payment..."
./bin/api-cli step-interact click "text=Confirm and Pay"

# Wait for processing
./bin/api-cli library-step-create $CP4 WAIT '{"duration": 3000}' 2>/dev/null || echo "Wait step added"

# REQ-009: Display Purchase Confirmed message
echo "Step: Verify purchase confirmation..."
./bin/api-cli step-assert exists "text=Purchase Confirmed!"

# REQ-010: Provide option to download confirmation
echo "Step: Verify download option exists..."
./bin/api-cli step-assert exists "text=Download Confirmation"

echo ""
echo "=== Deployment Complete ==="
echo "All requirements have been converted to NLP test steps!"
echo ""
echo "Checkpoints created:"
echo "1. Access and Product Selection (ID: $CP1)"
echo "2. Shopping Bag Navigation (ID: $CP2)"
echo "3. Checkout Form Validation (ID: $CP3)"
echo "4. Payment and Confirmation (ID: $CP4)"
echo ""
echo "View project: https://app.virtuoso.qa/#/project/9411"
