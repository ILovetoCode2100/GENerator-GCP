#!/bin/bash

DEPLOYMENT_BUCKET="virtuoso-deployment-986639207129-1753922127"
FUNCTIONS=("goal" "journey" "checkpoint" "step" "execution" "library" "data" "environment")

for func in "${FUNCTIONS[@]}"; do
    echo "Packaging function: $func"
    
    cd "/Users/marklovelady/_dev/_projects/lambda-api-gen/lambda-functions/$func"
    zip -r "/Users/marklovelady/_dev/_projects/lambda-api-gen/function-${func}.zip" . -x "*.git*" "*.DS_Store*"
    
    echo "Uploading function-${func}.zip to S3..."
    aws s3 cp "/Users/marklovelady/_dev/_projects/lambda-api-gen/function-${func}.zip" "s3://${DEPLOYMENT_BUCKET}/functions/function-${func}.zip"
    
    echo "Completed: $func"
    echo "---"
done

echo "All functions packaged and uploaded!"