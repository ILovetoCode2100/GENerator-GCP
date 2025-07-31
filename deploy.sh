#!/bin/bash

# Virtuoso API Lambda Deployment Script

set -e

echo "🚀 Deploying Virtuoso API Lambda Functions"

# Check if API token is provided
if [ -z "$1" ]; then
  echo "❌ Error: API token required"
  echo "Usage: ./deploy.sh <API_TOKEN>"
  exit 1
fi

API_TOKEN=$1
STACK_NAME=${STACK_NAME:-virtuoso-api-stack}
REGION=${AWS_REGION:-us-east-1}
S3_BUCKET=${S3_BUCKET:-virtuoso-lambda-deployments-$RANDOM}

# Create S3 bucket for deployments if it doesn't exist
echo "📦 Creating deployment bucket..."
aws s3 mb s3://$S3_BUCKET --region $REGION 2>/dev/null || true

# Install layer dependencies
echo "📥 Installing layer dependencies..."
cd lambda-layer/nodejs && npm install && cd ../..

# Package the application
echo "📦 Packaging application..."
sam package \
  --template-file template.yaml \
  --s3-bucket $S3_BUCKET \
  --output-template-file packaged.yaml \
  --region $REGION

# Deploy the application
echo "🚀 Deploying application..."
sam deploy \
  --template-file packaged.yaml \
  --stack-name $STACK_NAME \
  --capabilities CAPABILITY_IAM \
  --parameter-overrides ApiTokenValue=$API_TOKEN \
  --region $REGION \
  --no-fail-on-empty-changeset

# Get outputs
echo "✅ Deployment complete!"
echo "📋 Stack outputs:"
aws cloudformation describe-stacks \
  --stack-name $STACK_NAME \
  --region $REGION \
  --query 'Stacks[0].Outputs' \
  --output table

# Cleanup
rm -f packaged.yaml
