#!/bin/bash

# Rocketshop E-commerce Test Steps
# Using hint text instead of selectors where possible

CHECKPOINT_ID=1682489
POS=1

echo "Adding test steps to checkpoint $CHECKPOINT_ID..."

# Navigate to URL
./bin/api-cli step-navigate to $CHECKPOINT_ID "https://rocketshop.virtuoso.qa" $POS
POS=$((POS + 1))

# Assert "Border Not Found" product is visible
./bin/api-cli step-assert exists $CHECKPOINT_ID "Border Not Found" $POS
POS=$((POS + 1))

# Wait for page to load
./bin/api-cli step-wait time $CHECKPOINT_ID 20000 $POS
POS=$((POS + 1))

# Click Add to Bag button (using hint text)
./bin/api-cli step-interact click $CHECKPOINT_ID "Add to Bag" $POS
POS=$((POS + 1))

# Click Shopping Bag link
./bin/api-cli step-interact click $CHECKPOINT_ID "Shopping Bag" $POS
POS=$((POS + 1))

# Assert Shopping Bag page loaded
./bin/api-cli step-assert exists $CHECKPOINT_ID "Shopping Bag" $POS
POS=$((POS + 1))

# Click Go to Checkout button
./bin/api-cli step-interact click $CHECKPOINT_ID "Go to Checkout" $POS
POS=$((POS + 1))

# Fill in checkout form
# Full name
./bin/api-cli step-interact write $CHECKPOINT_ID "Full name" "John Doe" $POS
POS=$((POS + 1))

# Email
./bin/api-cli step-interact write $CHECKPOINT_ID "Email" "johndoe@example.com" $POS
POS=$((POS + 1))

# Address
./bin/api-cli step-interact write $CHECKPOINT_ID "Address" "123 Elm Street" $POS
POS=$((POS + 1))

# Phone
./bin/api-cli step-interact write $CHECKPOINT_ID "Phone numbers" "555-1234" $POS
POS=$((POS + 1))

# ZIP code
./bin/api-cli step-interact write $CHECKPOINT_ID "ZIP code" "90210" $POS
POS=$((POS + 1))

# Card number
./bin/api-cli step-interact write $CHECKPOINT_ID "Card number" "4111 1111 1111 1111" $POS
POS=$((POS + 1))

# CVV/CVC (the field might have placeholder "xxx")
./bin/api-cli step-interact write $CHECKPOINT_ID "CVV" "234" $POS
POS=$((POS + 1))

# Click Confirm and Pay button
./bin/api-cli step-interact click $CHECKPOINT_ID "Confirm and Pay" $POS
POS=$((POS + 1))

# Wait for confirmation
./bin/api-cli step-wait time $CHECKPOINT_ID 20000 $POS
POS=$((POS + 1))

# Assert purchase confirmed
./bin/api-cli step-assert exists $CHECKPOINT_ID "Purchase Confirmed!" $POS
POS=$((POS + 1))

# Click Download Confirmation button
./bin/api-cli step-interact click $CHECKPOINT_ID "Download Confirmation" $POS

echo "All test steps added successfully!"
