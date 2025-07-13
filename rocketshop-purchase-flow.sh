#!/bin/bash

# Script to create RocketShop purchase flow test in checkpoint 1680569

CHECKPOINT_ID=1680569
BINARY="/Users/marklovelady/_dev/virtuoso-api-cli-generator/bin/api-cli"

# Check if environment variables are set
if [ -z "$VIRTUOSO_API_TOKEN" ]; then
    echo "Error: VIRTUOSO_API_TOKEN environment variable is not set"
    exit 1
fi

if [ -z "$VIRTUOSO_API_BASE_URL" ]; then
    export VIRTUOSO_API_BASE_URL="https://api-app2.virtuoso.qa/api"
fi

echo "========================================="
echo "Creating RocketShop Purchase Flow Test"
echo "Checkpoint: $CHECKPOINT_ID"
echo "========================================="
echo ""

# Test Data
CUSTOMER_NAME="John Doe"
EMAIL="johndoe@example.com"
ADDRESS="123 Elm Street"
PHONE="555-1234"
ZIP="90210"
CARD_NUMBER="4111 1111 1111 1111"
CVV="234"

# Counter for position
POSITION=1

# Function to run a command and report status
run_step() {
    local description="$1"
    shift
    
    echo "[$POSITION] $description"
    
    if output=$("$@" 2>&1); then
        if echo "$output" | grep -q "successfully"; then
            echo "✅ Success"
            if echo "$output" | grep -q "Step ID:"; then
                step_id=$(echo "$output" | grep "Step ID:" | awk '{print $3}')
                echo "   Step ID: $step_id"
            fi
        else
            echo "⚠️  Command executed"
        fi
    else
        echo "❌ Failed: $output" | head -2
    fi
    echo ""
    
    POSITION=$((POSITION + 1))
}

echo "📋 Test Setup: Complete Purchase Flow on RocketShop"
echo "=================================================="
echo "Test Data:"
echo "- Customer: $CUSTOMER_NAME"
echo "- Email: $EMAIL"
echo "- Card: 4111 1111 1111 1111"
echo ""

echo "📍 Checkpoint 1: Navigate to Site"
echo "================================="

# Step 1: Navigate to RocketShop
run_step "Navigate to https://rocketshop.virtuoso.qa" \
    "$BINARY" create-step-navigate "$CHECKPOINT_ID" "https://rocketshop.virtuoso.qa" "$POSITION"

# Step 2: Wait for page to load
run_step "Wait for page to fully load" \
    "$BINARY" create-step-wait-for-element-default "$CHECKPOINT_ID" "body" "$POSITION"

# Step 3: Verify home page
run_step "Verify the home page is displayed" \
    "$BINARY" create-step-assert-exists "$CHECKPOINT_ID" "home" "$POSITION"

echo "📍 Checkpoint 2: Add Product to Bag"
echo "===================================="

# Step 4: Verify "Border Not Found" element
run_step "Verify 'Border Not Found' element is present" \
    "$BINARY" create-step-assert-exists "$CHECKPOINT_ID" "Border Not Found" "$POSITION"

# Step 5: Wait for product image button
run_step "Wait 20s for product image button to be clickable" \
    "$BINARY" create-step-wait-for-element-timeout "$CHECKPOINT_ID" "/html/body/div/div/div[1]/div[2]/div[1]/div/div[1]/button/img" 20000 "$POSITION"

# Step 6: Click "Add to Bag"
run_step "Click on 'Add to Bag' button" \
    "$BINARY" create-step-click "$CHECKPOINT_ID" "Add to Bag" "$POSITION"

# Step 7: Click "Shopping Bag"
run_step "Click on 'Shopping Bag' link/button" \
    "$BINARY" create-step-click "$CHECKPOINT_ID" "Shopping Bag" "$POSITION"

# Step 8: Verify Shopping Bag page
run_step "Verify 'Shopping Bag' page/section is displayed" \
    "$BINARY" create-step-assert-exists "$CHECKPOINT_ID" "Shopping Bag" "$POSITION"

# Step 9: Click "Go to Checkout"
run_step "Click on 'Go to Checkout' button" \
    "$BINARY" create-step-click "$CHECKPOINT_ID" "Go to Checkout" "$POSITION"

echo "📍 Checkpoint 3: Complete Checkout Process"
echo "========================================="

# Step 10a: Fill customer name
run_step "Enter '$CUSTOMER_NAME' in 'Full name' field" \
    "$BINARY" create-step-write "$CHECKPOINT_ID" "Full name" "$CUSTOMER_NAME" "$POSITION"

# Step 10b: Fill email
run_step "Enter '$EMAIL' in 'Email' field" \
    "$BINARY" create-step-write "$CHECKPOINT_ID" "Email" "$EMAIL" "$POSITION"

# Step 10c: Fill address
run_step "Enter '$ADDRESS' in 'Address' field" \
    "$BINARY" create-step-write "$CHECKPOINT_ID" "Address" "$ADDRESS" "$POSITION"

# Step 10d: Fill phone
run_step "Enter '$PHONE' in 'Phone numbers' field" \
    "$BINARY" create-step-write "$CHECKPOINT_ID" "Phone numbers" "$PHONE" "$POSITION"

# Step 10e: Fill ZIP code
run_step "Enter '$ZIP' in 'ZIP code' field" \
    "$BINARY" create-step-write "$CHECKPOINT_ID" "ZIP code" "$ZIP" "$POSITION"

# Step 11a: Fill card number
run_step "Enter card number in 'Card number' field" \
    "$BINARY" create-step-write "$CHECKPOINT_ID" "Card number" "$CARD_NUMBER" "$POSITION"

# Step 11b: Fill CVV
run_step "Enter '$CVV' in CVV field (labeled 'xxx')" \
    "$BINARY" create-step-write "$CHECKPOINT_ID" "xxx" "$CVV" "$POSITION"

# Step 12: Click Confirm and Pay
run_step "Click on 'Confirm and Pay' button" \
    "$BINARY" create-step-click "$CHECKPOINT_ID" "Confirm and Pay" "$POSITION"

# Step 13: Wait for confirmation
run_step "Wait up to 20s for confirmation message" \
    "$BINARY" create-step-wait-for-element-timeout "$CHECKPOINT_ID" "Purchase Confirmed!" 20000 "$POSITION"

# Step 14: Verify confirmation message
run_step "Verify 'Purchase Confirmed!' message is displayed" \
    "$BINARY" create-step-assert-exists "$CHECKPOINT_ID" "Purchase Confirmed!" "$POSITION"

# Step 15: Download confirmation
run_step "Click on 'Download Confirmation' button" \
    "$BINARY" create-step-click "$CHECKPOINT_ID" "Download Confirmation" "$POSITION"

echo "========================================="
echo "Test Setup Complete!"
echo "========================================="
echo "Total steps added: $((POSITION - 1))"
echo ""
echo "Test Case: Complete Purchase Flow on RocketShop"
echo "Checkpoint ID: $CHECKPOINT_ID"
echo ""
echo "Expected Results:"
echo "✓ Product added to shopping bag"
echo "✓ Checkout form accepts all information"
echo "✓ Payment processed successfully"
echo "✓ Purchase confirmation displayed"
echo "✓ Confirmation downloadable"
echo ""
echo "Check the Virtuoso UI to review and execute the test."
echo "========================================="