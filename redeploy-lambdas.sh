#!/bin/bash

# Redeploy Lambda Functions Script

set -e

echo "🚀 Redeploying Virtuoso API Lambda Functions"
echo "=========================================="
echo ""

# Configuration
STACK_NAME="virtuoso-api-stack"
REGION="us-east-1"
S3_BUCKET="virtuoso-deployment-986639207129-1753922127"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "📦 Installing layer dependencies..."
cd lambda-layer/nodejs
npm install --production
cd ../..

echo ""
echo "📦 Packaging Lambda functions..."

# Package each Lambda function
for func_dir in lambda-functions/*/; do
    if [ -d "$func_dir" ]; then
        func_name=$(basename "$func_dir")
        echo "  - Packaging $func_name..."
        
        # Create zip file
        cd "$func_dir"
        zip -r "../../${func_name}.zip" . -x "*.git*" > /dev/null 2>&1
        cd ../..
        
        # Upload to S3
        aws s3 cp "${func_name}.zip" "s3://${S3_BUCKET}/lambdas/${func_name}.zip" --region $REGION > /dev/null 2>&1
        rm "${func_name}.zip"
    fi
done

echo ""
echo "📦 Packaging Lambda layer..."
cd lambda-layer
zip -r ../lambda-layer.zip . -x "*.git*" > /dev/null 2>&1
cd ..
aws s3 cp lambda-layer.zip "s3://${S3_BUCKET}/layers/lambda-layer.zip" --region $REGION > /dev/null 2>&1
rm lambda-layer.zip

echo ""
echo "🔄 Updating Lambda functions..."

# Update each Lambda function
LAMBDA_FUNCTIONS=(
    "VirtuosoProjectHandler:project"
    "VirtuosoGoalHandler:goal"
    "VirtuosoJourneyHandler:journey"
    "VirtuosoCheckpointHandler:checkpoint"
    "VirtuosoStepHandler:step"
    "VirtuosoExecutionHandler:execution"
    "VirtuosoLibraryHandler:library"
    "VirtuosoDataHandler:data"
    "VirtuosoEnvironmentHandler:environment"
)

for func_info in "${LAMBDA_FUNCTIONS[@]}"; do
    IFS=':' read -r func_name func_key <<< "$func_info"
    
    echo -e "  ${YELLOW}Updating ${func_name}...${NC}"
    
    # Update function code
    aws lambda update-function-code \
        --function-name $func_name \
        --s3-bucket $S3_BUCKET \
        --s3-key "lambdas/${func_key}.zip" \
        --region $REGION > /dev/null 2>&1
    
    # Wait for update to complete
    aws lambda wait function-updated \
        --function-name $func_name \
        --region $REGION
    
    echo -e "  ${GREEN}✓ ${func_name} updated${NC}"
done

echo ""
echo "🔄 Updating Lambda layer..."

# Get current layer version
LAYER_ARN=$(aws lambda publish-layer-version \
    --layer-name virtuoso-lambda-layer \
    --content S3Bucket=$S3_BUCKET,S3Key=layers/lambda-layer.zip \
    --compatible-runtimes nodejs18.x \
    --region $REGION \
    --query 'LayerVersionArn' \
    --output text)

echo "  New layer version: $LAYER_ARN"

# Update all functions to use new layer
for func_info in "${LAMBDA_FUNCTIONS[@]}"; do
    IFS=':' read -r func_name func_key <<< "$func_info"
    
    aws lambda update-function-configuration \
        --function-name $func_name \
        --layers $LAYER_ARN \
        --region $REGION > /dev/null 2>&1
done

echo ""
echo "✅ Deployment complete!"
echo ""
echo "🌐 API Gateway Endpoint:"
echo "   https://4sswk1wyv9.execute-api.us-east-1.amazonaws.com/Prod/virtuoso"
echo ""
echo "📝 Next step: Run the test suite"
echo "   node test-virtuoso-api.js"