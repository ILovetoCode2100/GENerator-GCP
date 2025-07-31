#!/bin/bash

set -e  # Exit on any error

echo "🚀 Setting up IAM User for CDK Deployment"
echo "=========================================="

USER_NAME="virtuoso-cdk-deploy"
POLICY_NAME="VirtuosoCDKDeploymentPolicy"
POLICY_FILE="/Users/marklovelady/_dev/_projects/api-lambdav2/iam-policy-virtuoso-cdk.json"

# Check if AWS CLI is installed and configured
if ! command -v aws &> /dev/null; then
    echo "❌ AWS CLI is not installed. Please install it first."
    exit 1
fi

# Check if we have AWS credentials
aws sts get-caller-identity >/dev/null 2>&1
if [[ $? -ne 0 ]]; then
    echo "❌ AWS CLI is not configured. Please run 'aws configure' first."
    exit 1
fi

# Get AWS account ID
AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
echo "📋 AWS Account ID: $AWS_ACCOUNT_ID"

# Check if user already exists
if aws iam get-user --user-name "$USER_NAME" >/dev/null 2>&1; then
    echo "⚠️  User '$USER_NAME' already exists. Skipping user creation."
else
    echo "👤 Creating IAM user: $USER_NAME"
    aws iam create-user --user-name "$USER_NAME"
    
    # Tag the user
    aws iam tag-user --user-name "$USER_NAME" --tags \
        Key=Purpose,Value=CDKDeployment \
        Key=Project,Value=VirtuosoAPI \
        Key=CreatedBy,Value=SetupScript
    
    echo "✅ User created successfully"
fi

# Check if policy already exists
POLICY_ARN="arn:aws:iam::${AWS_ACCOUNT_ID}:policy/${POLICY_NAME}"
if aws iam get-policy --policy-arn "$POLICY_ARN" >/dev/null 2>&1; then
    echo "⚠️  Policy '$POLICY_NAME' already exists. Skipping policy creation."
else
    echo "📜 Creating custom IAM policy: $POLICY_NAME"
    aws iam create-policy \
        --policy-name "$POLICY_NAME" \
        --policy-document "file://$POLICY_FILE" \
        --description "Minimal permissions for Virtuoso API CDK deployment"
    
    echo "✅ Policy created successfully"
fi

# Check if policy is already attached
if aws iam list-attached-user-policies --user-name "$USER_NAME" | grep -q "$POLICY_NAME"; then
    echo "⚠️  Policy already attached to user. Skipping attachment."
else
    echo "🔗 Attaching policy to user"
    aws iam attach-user-policy \
        --user-name "$USER_NAME" \
        --policy-arn "$POLICY_ARN"
    
    echo "✅ Policy attached successfully"
fi

# Check if access keys already exist
EXISTING_KEYS=$(aws iam list-access-keys --user-name "$USER_NAME" --query 'AccessKeyMetadata[].AccessKeyId' --output text)
if [[ -n "$EXISTING_KEYS" ]]; then
    echo "⚠️  Access keys already exist for user:"
    echo "   $EXISTING_KEYS"
    echo "   You can use existing keys or create new ones (max 2 per user)."
    read -p "Create new access keys? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "ℹ️  Skipping access key creation."
        SKIP_KEYS=true
    fi
fi

if [[ "$SKIP_KEYS" != "true" ]]; then
    echo "🔑 Creating access keys"
    KEY_OUTPUT=$(aws iam create-access-key --user-name "$USER_NAME")
    
    ACCESS_KEY=$(echo "$KEY_OUTPUT" | jq -r '.AccessKey.AccessKeyId')
    SECRET_KEY=$(echo "$KEY_OUTPUT" | jq -r '.AccessKey.SecretAccessKey')
    
    echo ""
    echo "🎉 Access Keys Created Successfully!"
    echo "==================================="
    echo "Access Key ID: $ACCESS_KEY"
    echo "Secret Access Key: $SECRET_KEY"
    echo ""
    echo "⚠️  IMPORTANT: Save these credentials securely!"
    echo "   The secret key will not be shown again."
    echo ""
    
    # Configure AWS CLI profile
    read -p "Configure AWS CLI profile 'virtuoso-cdk-deploy'? (Y/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Nn]$ ]]; then
        echo "🔧 Configuring AWS CLI profile..."
        
        # Get current region
        CURRENT_REGION=$(aws configure get region 2>/dev/null || echo "us-east-1")
        
        # Configure the profile
        aws configure set aws_access_key_id "$ACCESS_KEY" --profile virtuoso-cdk-deploy
        aws configure set aws_secret_access_key "$SECRET_KEY" --profile virtuoso-cdk-deploy
        aws configure set region "$CURRENT_REGION" --profile virtuoso-cdk-deploy
        aws configure set output json --profile virtuoso-cdk-deploy
        
        echo "✅ AWS CLI profile 'virtuoso-cdk-deploy' configured"
        echo ""
        echo "To use this profile, run:"
        echo "  export AWS_PROFILE=virtuoso-cdk-deploy"
        echo "  # or add --profile virtuoso-cdk-deploy to AWS commands"
    fi
fi

echo ""
echo "🎯 Setup Complete! Summary:"
echo "=========================="
echo "✅ IAM User: $USER_NAME"
echo "✅ Custom Policy: $POLICY_NAME"
echo "✅ Policy Attached: Yes"
if [[ "$SKIP_KEYS" != "true" ]]; then
    echo "✅ Access Keys: Created"
    echo "✅ AWS CLI Profile: virtuoso-cdk-deploy"
fi

echo ""
echo "🔍 Next Steps:"
echo "=============="
echo "1. Run the verification script:"
echo "   ./verify-iam-setup.sh"
echo ""
echo "2. Test with the new profile:"
echo "   export AWS_PROFILE=virtuoso-cdk-deploy"
echo "   aws sts get-caller-identity"
echo ""
echo "3. Deploy your CDK stack:"
echo "   cd /Users/marklovelady/_dev/_projects/api-lambdav2/cdk"
echo "   npm install"
echo "   cdk bootstrap"
echo "   npm run deploy"
echo ""
echo "4. Configure Virtuoso API key:"
echo "   aws secretsmanager put-secret-value \\"
echo "     --secret-id virtuoso-api-key \\"
echo "     --secret-string '{\"apiKey\":\"YOUR_VIRTUOSO_API_KEY\"}'"

echo ""
echo "🔐 Security Reminders:"
echo "====================="
echo "• Never use root AWS credentials for deployment"
echo "• Rotate access keys every 90 days"
echo "• Monitor CloudTrail for unusual activity"
echo "• Review IAM permissions quarterly"
echo "• Store credentials securely (consider AWS SSO for teams)"

echo ""
echo "✅ IAM user setup completed successfully!"