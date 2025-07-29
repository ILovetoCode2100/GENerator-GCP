#!/bin/bash

# Set checkpoint ID for session
export VIRTUOSO_SESSION_ID=1682631

# Deploy all steps
echo "Deploying Rocketshop test steps to checkpoint 1682631..."

./bin/api-cli step-assert exists "Border Not Found"
./bin/api-cli step-wait time 20000
./bin/api-cli step-interact click "Add to Bag"
./bin/api-cli step-interact click "Shopping Bag"
./bin/api-cli step-assert exists "Shopping Bag"
./bin/api-cli step-interact click "Go to Checkout"
./bin/api-cli step-interact write "Full name" "John Doe"
./bin/api-cli step-interact write "Email" "johndoe@example.com"
./bin/api-cli step-interact write "Address" "123 Elm Street"
./bin/api-cli step-interact write "Phone numbers" "555-1234"
./bin/api-cli step-interact write "ZIP code" "90210"
./bin/api-cli step-interact write "Card number" "4111 1111 1111 1111"
./bin/api-cli step-interact write "xxx" "234"
./bin/api-cli step-interact click "Confirm and Pay"
./bin/api-cli step-wait time 20000
./bin/api-cli step-assert exists "Purchase Confirmed!"
./bin/api-cli step-interact click "Download Confirmation"

echo "Deployment complete!"
echo "View test at: https://app.virtuoso.qa/#/checkpoint/1682631"
