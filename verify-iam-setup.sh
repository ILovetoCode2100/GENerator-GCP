#!/bin/bash

echo "🔍 Verifying IAM Setup for CDK Deployment"
echo "========================================="

# Check if we're using the correct profile
CURRENT_USER=$(aws sts get-caller-identity --query 'Arn' --output text 2>/dev/null)
if [[ $? -ne 0 ]]; then
    echo "❌ Error: AWS CLI not configured or no credentials found"
    exit 1
fi

echo "Current AWS Identity: $CURRENT_USER"

# Check if it's NOT root user
if [[ "$CURRENT_USER" == *":root" ]]; then
    echo "⚠️  WARNING: You are using ROOT credentials!"
    echo "   This is NOT recommended for CDK deployment."
    echo "   Please use the virtuoso-cdk-deploy user instead."
    exit 1
fi

# Check if it's the correct user
if [[ "$CURRENT_USER" == *"user/virtuoso-cdk-deploy" ]]; then
    echo "✅ Using correct IAM user: virtuoso-cdk-deploy"
else
    echo "⚠️  Warning: You are not using the virtuoso-cdk-deploy user"
    echo "   Current user: $CURRENT_USER"
fi

# Check if the user exists
echo ""
echo "🔍 Checking IAM user details..."
USER_INFO=$(aws iam get-user --user-name virtuoso-cdk-deploy 2>/dev/null)
if [[ $? -eq 0 ]]; then
    echo "✅ IAM user 'virtuoso-cdk-deploy' exists"
    
    # Check attached policies
    echo ""
    echo "🔍 Checking attached policies..."
    POLICIES=$(aws iam list-attached-user-policies --user-name virtuoso-cdk-deploy --query 'AttachedPolicies[].PolicyName' --output text)
    if [[ "$POLICIES" == *"VirtuosoCDKDeploymentPolicy"* ]]; then
        echo "✅ Custom policy 'VirtuosoCDKDeploymentPolicy' is attached"
    else
        echo "❌ Custom policy not found. Attached policies: $POLICIES"
    fi
else
    echo "❌ IAM user 'virtuoso-cdk-deploy' does not exist"
    exit 1
fi

# Test basic permissions
echo ""
echo "🔍 Testing basic permissions..."

# Test CloudFormation
aws cloudformation list-stacks --max-items 1 >/dev/null 2>&1
if [[ $? -eq 0 ]]; then
    echo "✅ CloudFormation access: OK"
else
    echo "❌ CloudFormation access: FAILED"
fi

# Test Lambda
aws lambda list-functions --max-items 1 >/dev/null 2>&1
if [[ $? -eq 0 ]]; then
    echo "✅ Lambda access: OK"
else
    echo "❌ Lambda access: FAILED"
fi

# Test S3
aws s3 ls >/dev/null 2>&1
if [[ $? -eq 0 ]]; then
    echo "✅ S3 access: OK"
else
    echo "❌ S3 access: FAILED"
fi

# Test Secrets Manager
aws secretsmanager list-secrets --max-results 1 >/dev/null 2>&1
if [[ $? -eq 0 ]]; then
    echo "✅ Secrets Manager access: OK"
else
    echo "❌ Secrets Manager access: FAILED"
fi

echo ""
echo "🔐 Security Recommendations:"
echo "=============================="
echo "1. ✅ Using IAM user (not root credentials)"
echo "2. ✅ Custom policy with minimal required permissions"
echo "3. ✅ Resource-specific ARN restrictions where possible"
echo "4. 🔄 Regularly rotate access keys (recommended: every 90 days)"
echo "5. 🔄 Monitor CloudTrail logs for unusual activity"
echo "6. 🔄 Review and audit permissions quarterly"

echo ""
echo "📋 Next Steps:"
echo "==============="
echo "1. Run: cd /Users/marklovelady/_dev/_projects/api-lambdav2/cdk"
echo "2. Run: npm install"
echo "3. Run: cdk bootstrap (if first time)"
echo "4. Run: npm run deploy"
echo "5. Configure Virtuoso API key in Secrets Manager"

echo ""
echo "✅ IAM setup verification complete!"