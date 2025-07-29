#!/bin/bash

# Set the checkpoint ID for session context
export VIRTUOSO_SESSION_ID=1683912

echo "Deploying NLP test steps to checkpoint 1683912..."

# Step 1: Navigate to Rocketshop
echo "Step 1: Navigate to Rocketshop..."
./bin/api-cli step-navigate to "https://rocketshop.virtuoso.qa"

# Step 2: Verify Border Not Found product is visible
echo "Step 2: Verify Border Not Found product..."
./bin/api-cli step-assert exists "text=Border Not Found"

# Step 3: Wait for page to fully load
echo "Step 3: Wait for page load..."
./bin/api-cli library-step-create 1683912 WAIT '{"duration": 20000}'

# Step 4: Click on Add to Bag button
echo "Step 4: Click on Add to Bag..."
./bin/api-cli step-interact click "text=Add to Bag"

# Step 5: Click on Shopping Bag link
echo "Step 5: Click on Shopping Bag..."
./bin/api-cli step-interact click "text=Shopping Bag"

# Step 6: Verify Shopping Bag page
echo "Step 6: Verify Shopping Bag page..."
./bin/api-cli step-assert exists "text=Shopping Bag"

# Step 7: Click Go to Checkout
echo "Step 7: Click Go to Checkout..."
./bin/api-cli step-interact click "text=Go to Checkout"

echo "Adding checkout steps to new checkpoint..."

# Create second checkpoint for checkout process
CHECKPOINT2=$(./bin/api-cli create-checkpoint 611011 14582 44337 "Checkout and Payment" --output json | jq -r '.checkpoint_id')
export VIRTUOSO_SESSION_ID=$CHECKPOINT2

echo "Created checkpoint $CHECKPOINT2 for checkout steps"

# Step 8: Enter Full name
echo "Step 8: Enter Full name..."
./bin/api-cli step-interact write "John Doe" "text=Full name"

# Step 9: Enter Email
echo "Step 9: Enter Email..."
./bin/api-cli step-interact write "johndoe@example.com" "text=Email"

# Step 10: Enter Address
echo "Step 10: Enter Address..."
./bin/api-cli step-interact write "123 Elm Street" "text=Address"

# Step 11: Enter Phone numbers
echo "Step 11: Enter Phone numbers..."
./bin/api-cli step-interact write "555-1234" "text=Phone numbers"

# Step 12: Enter ZIP code
echo "Step 12: Enter ZIP code..."
./bin/api-cli step-interact write "90210" "text=ZIP code"

# Step 13: Enter Card number
echo "Step 13: Enter Card number..."
./bin/api-cli step-interact write "4111 1111 1111 1111" "text=Card number"

# Step 14: Enter CVV
echo "Step 14: Enter CVV..."
./bin/api-cli step-interact write "234" "placeholder=xxx"

# Step 15: Click Confirm and Pay
echo "Step 15: Click Confirm and Pay..."
./bin/api-cli step-interact click "text=Confirm and Pay"

# Step 16: Wait for confirmation
echo "Step 16: Wait for confirmation..."
./bin/api-cli library-step-create $CHECKPOINT2 WAIT '{"duration": 20000}'

# Step 17: Verify Purchase Confirmed
echo "Step 17: Verify Purchase Confirmed..."
./bin/api-cli step-assert exists "text=Purchase Confirmed!"

# Step 18: Click Download Confirmation
echo "Step 18: Click Download Confirmation..."
./bin/api-cli step-interact click "text=Download Confirmation"

echo "All NLP test steps deployed successfully!"
echo "Project: 9410"
echo "Goal: 14582"
echo "Journey: 611011"
echo "Checkpoints: 1683912, $CHECKPOINT2"
