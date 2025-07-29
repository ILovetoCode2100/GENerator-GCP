#!/bin/bash

# Run the RocketShop shopping cart test
echo "Running RocketShop shopping cart test..."
./bin/api-cli run-test rocketshop-shopping-test.yaml

# Check if you want to see the results
echo ""
echo "Test execution initiated. You can check the results with:"
echo "./bin/api-cli execution list"
