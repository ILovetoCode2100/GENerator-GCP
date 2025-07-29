#!/bin/bash

# Set the checkpoint ID for session context
export VIRTUOSO_SESSION_ID=1683907

# Add wait step
echo "Adding wait step..."
./bin/api-cli step-interact wait 20000

# Click Add to Bag button
echo "Clicking Add to Bag..."
./bin/api-cli step-interact click "text=Add to Bag"

# Click Shopping Bag link
echo "Navigating to Shopping Bag..."
./bin/api-cli step-interact click "text=Shopping Bag"

# Assert Shopping Bag heading
echo "Verifying Shopping Bag page..."
./bin/api-cli step-assert exists "h1=Shopping Bag"

# Click Go to Checkout
echo "Proceeding to checkout..."
./bin/api-cli step-interact click "text=Go to Checkout"

# Fill checkout form
echo "Filling checkout form..."
./bin/api-cli step-interact write "John Doe" "label=Full name"
./bin/api-cli step-interact write "johndoe@example.com" "label=Email"
./bin/api-cli step-interact write "123 Elm Street" "label=Address"
./bin/api-cli step-interact write "555-1234" "label=Phone numbers"
./bin/api-cli step-interact write "90210" "label=ZIP code"

# Fill payment info
echo "Filling payment information..."
./bin/api-cli step-interact write "4111 1111 1111 1111" "label=Card number"
./bin/api-cli step-interact write "234" "css=.w-auto > .focus\\:border-rocket-orange"

# Complete purchase
echo "Completing purchase..."
./bin/api-cli step-interact click "text=Confirm and Pay"

# Wait for confirmation
echo "Waiting for confirmation..."
./bin/api-cli step-interact wait 20000

# Verify purchase confirmation
echo "Verifying purchase confirmation..."
./bin/api-cli step-assert exists "h3=Purchase Confirmed!"

# Download confirmation
echo "Downloading confirmation..."
./bin/api-cli step-interact click "text=Download Confirmation"

echo "All steps added successfully!"
echo "Project ID: 9408"
echo "Goal ID: 14580"
echo "Journey ID: 611006"
echo "Checkpoint ID: 1683907"
