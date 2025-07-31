#!/bin/bash

# Script to create IAM user with proper CDK deployment permissions
# Run this from AWS CloudShell or with proper AWS CLI credentials

set -e

echo "Creating IAM user for CDK deployment..."

# Variables
USER_NAME="virtuoso-cdk-deploy"
POLICY_NAME="VirtuosoCDKDeployPolicy"

# Create the IAM user
aws iam create-user --user-name $USER_NAME

# Create access key for the user
echo "Creating access keys..."
CREDENTIALS=$(aws iam create-access-key --user-name $USER_NAME)
ACCESS_KEY=$(echo $CREDENTIALS | jq -r '.AccessKey.AccessKeyId')
SECRET_KEY=$(echo $CREDENTIALS | jq -r '.AccessKey.SecretAccessKey')

# Create the policy document
cat > cdk-deploy-policy.json << 'EOF'
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "cloudformation:*",
        "lambda:*",
        "apigatewayv2:*",
        "logs:*",
        "secretsmanager:*",
        "s3:*",
        "ssm:*",
        "ec2:DescribeAvailabilityZones",
        "ec2:DescribeVpcs",
        "ec2:DescribeSubnets",
        "ec2:DescribeSecurityGroups"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "iam:CreateRole",
        "iam:DeleteRole",
        "iam:AttachRolePolicy",
        "iam:DetachRolePolicy",
        "iam:PutRolePolicy",
        "iam:DeleteRolePolicy",
        "iam:GetRole",
        "iam:GetRolePolicy",
        "iam:PassRole",
        "iam:ListRolePolicies",
        "iam:ListAttachedRolePolicies",
        "iam:ListInstanceProfilesForRole",
        "iam:CreatePolicy",
        "iam:DeletePolicy",
        "iam:GetPolicy",
        "iam:GetPolicyVersion",
        "iam:ListPolicyVersions"
      ],
      "Resource": [
        "arn:aws:iam::*:role/VirtuosoApi*",
        "arn:aws:iam::*:role/cdk-*",
        "arn:aws:iam::*:policy/VirtuosoApi*"
      ]
    }
  ]
}
EOF

# Create the IAM policy
POLICY_ARN=$(aws iam create-policy \
    --policy-name $POLICY_NAME \
    --policy-document file://cdk-deploy-policy.json \
    --query 'Policy.Arn' \
    --output text)

# Attach the policy to the user
aws iam attach-user-policy \
    --user-name $USER_NAME \
    --policy-arn $POLICY_ARN

# Also attach PowerUserAccess for broader permissions if needed
# aws iam attach-user-policy \
#     --user-name $USER_NAME \
#     --policy-arn arn:aws:iam::aws:policy/PowerUserAccess

echo "IAM user created successfully!"
echo "================================="
echo "User Name: $USER_NAME"
echo "Access Key: $ACCESS_KEY"
echo "Secret Key: $SECRET_KEY"
echo "================================="
echo ""
echo "Configure AWS CLI with these credentials:"
echo "aws configure --profile virtuoso-cdk"
echo ""
echo "Then deploy with:"
echo "export AWS_PROFILE=virtuoso-cdk"
echo "cd cdk && npm run deploy"
echo ""
echo "IMPORTANT: Save these credentials securely and never commit them to version control!"

# Clean up
rm -f cdk-deploy-policy.json