#!/bin/bash

# Add complete Selenium test conversion to checkpoint 1682637
export VIRTUOSO_SESSION_ID=1682637

echo "Adding complete Rocketshop test to checkpoint 1682637..."
echo ""

# Navigate to URL
echo "Step 1: Navigate to URL"
./bin/api-cli step-navigate to "https://rocketshop.virtuoso.qa"

# Verify Border Not Found element exists
echo "Step 2: Verify Border Not Found element"
./bin/api-cli step-assert exists "Border Not Found"

# Wait for page load
echo "Step 3: Wait 20 seconds"
./bin/api-cli step-wait time 20000

# Click Add to Bag
echo "Step 4: Click Add to Bag"
./bin/api-cli step-interact click "Add to Bag"

# Click Shopping Bag link
echo "Step 5: Navigate to Shopping Bag"
./bin/api-cli step-interact click "Shopping Bag"

# Verify Shopping Bag page heading
echo "Step 6: Verify Shopping Bag heading"
./bin/api-cli step-assert exists "Shopping Bag"

# Click Go to Checkout
echo "Step 7: Go to Checkout"
./bin/api-cli step-interact click "Go to Checkout"

# Fill checkout form
echo "Step 8-12: Fill checkout form"
./bin/api-cli step-interact write "Full name" "John Doe"
./bin/api-cli step-interact write "Email" "johndoe@example.com"
./bin/api-cli step-interact write "Address" "123 Elm Street"
./bin/api-cli step-interact write "Phone numbers" "555-1234"
./bin/api-cli step-interact write "ZIP code" "90210"

# Enter payment information
echo "Step 13-14: Enter payment details"
./bin/api-cli step-interact write "Card number" "4111 1111 1111 1111"
./bin/api-cli step-interact write "xxx" "234"

# Click Confirm and Pay
echo "Step 15: Confirm and Pay"
./bin/api-cli step-interact click "Confirm and Pay"

# Wait for confirmation
echo "Step 16: Wait for confirmation"
./bin/api-cli step-wait time 20000

# Verify Purchase Confirmed
echo "Step 17: Verify purchase confirmation"
./bin/api-cli step-assert exists "Purchase Confirmed!"

# Download confirmation
echo "Step 18: Download confirmation"
./bin/api-cli step-interact click "Download Confirmation"

echo ""
echo "Complete test added to checkpoint!"
echo "View at: https://app.virtuoso.qa/#/checkpoint/1682637"
