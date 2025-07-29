#!/bin/bash

# Set the checkpoint ID for session context
export VIRTUOSO_SESSION_ID=1683907

# Add wait step using library command
echo "Adding wait steps..."
./bin/api-cli library-step create 1683907 WAIT '{"duration": 20000}'

# Fill ZIP code
echo "Filling ZIP code..."
./bin/api-cli step-interact write "90210" "label=ZIP code"

# Add another wait before payment
./bin/api-cli library-step create 1683907 WAIT '{"duration": 20000}'

echo "Remaining steps added successfully!"
