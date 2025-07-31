# IAM User Setup Guide for CDK Deployment

This guide provides secure setup instructions for creating an IAM user dedicated to CDK deployment of the Virtuoso API Gateway project.

## 🚨 Security Warning

**NEVER use AWS root credentials for CDK deployment!** This guide creates a dedicated IAM user with minimal required permissions following AWS security best practices.

## 📋 Quick Setup

### Option 1: Automated Setup (Recommended)

```bash
# Run the automated setup script
./setup-iam-user.sh

# Verify the setup
./verify-iam-setup.sh

# Optional: Run security audit
./audit-iam-security.sh
```

### Option 2: Manual Setup

Follow the step-by-step instructions below.

## 🔧 Manual Setup Instructions

### Step 1: Create IAM User

```bash
# Create the user
aws iam create-user --user-name virtuoso-cdk-deploy

# Tag for organization
aws iam tag-user --user-name virtuoso-cdk-deploy --tags \
  Key=Purpose,Value=CDKDeployment \
  Key=Project,Value=VirtuosoAPI
```

### Step 2: Create Custom Policy

The custom policy (`iam-policy-virtuoso-cdk.json`) includes:

- **CloudFormation**: Stack management for CDK
- **Lambda**: Function deployment and management
- **API Gateway v2**: HTTP API creation and configuration
- **IAM**: Role creation for Lambda execution
- **CloudWatch Logs**: Log group management
- **Secrets Manager**: API key storage
- **S3**: CDK asset storage
- **STS/SSM**: Supporting services

```bash
# Get your AWS account ID
AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)

# Create the policy
aws iam create-policy \
  --policy-name VirtuosoCDKDeploymentPolicy \
  --policy-document file://iam-policy-virtuoso-cdk.json \
  --description "Minimal permissions for Virtuoso API CDK deployment"

# Attach to user
aws iam attach-user-policy \
  --user-name virtuoso-cdk-deploy \
  --policy-arn arn:aws:iam::${AWS_ACCOUNT_ID}:policy/VirtuosoCDKDeploymentPolicy
```

### Step 3: Create Access Keys

```bash
# Create access keys
aws iam create-access-key --user-name virtuoso-cdk-deploy
```

**⚠️ Important:** Save the Access Key ID and Secret Access Key securely. The secret will only be shown once.

### Step 4: Configure AWS CLI Profile

```bash
# Configure dedicated profile
aws configure --profile virtuoso-cdk-deploy

# Test the configuration
aws sts get-caller-identity --profile virtuoso-cdk-deploy
```

### Step 5: Deploy with New User

```bash
# Set the profile
export AWS_PROFILE=virtuoso-cdk-deploy

# Navigate to CDK directory
cd cdk

# Install dependencies
npm install

# Bootstrap CDK (first time only)
cdk bootstrap

# Deploy
npm run deploy
```

## 🔐 Security Features

### Principle of Least Privilege

The custom policy follows the principle of least privilege:

1. **Resource-specific ARNs** where possible (Lambda functions with `virtuoso-*` prefix)
2. **Service-specific permissions** (no wildcard `*:*` actions)
3. **Stack-scoped CloudFormation** access
4. **Limited S3 access** to CDK asset buckets only

### Security Boundaries

- **No console access** (API keys only)
- **No MFA requirement** (service account)
- **Resource tagging** for identification
- **No inline policies** (managed policies only)

## 📊 Monitoring and Auditing

### Regular Security Checks

```bash
# Run security audit
./audit-iam-security.sh

# Check recent activity
aws iam get-access-key-last-used --access-key-id YOUR_ACCESS_KEY_ID

# View CloudTrail events
aws logs filter-log-events --log-group-name CloudTrail/virtuoso-api
```

### Key Rotation Schedule

- **Review permissions**: Quarterly
- **Rotate access keys**: Every 90 days
- **Audit usage**: Monthly
- **Update policies**: As needed for new features

## 🚨 Troubleshooting

### Common Issues

1. **403 Forbidden Errors**
   - Check policy attachment: `aws iam list-attached-user-policies --user-name virtuoso-cdk-deploy`
   - Verify resource ARNs match your account/region

2. **CDK Bootstrap Fails**
   - Ensure S3 permissions for CDK buckets
   - Check CloudFormation stack permissions

3. **Lambda Deployment Fails**
   - Verify IAM role creation permissions
   - Check Lambda resource ARN patterns

### Debugging Commands

```bash
# Check current identity
aws sts get-caller-identity

# List user policies
aws iam list-attached-user-policies --user-name virtuoso-cdk-deploy

# Test specific permissions
aws lambda list-functions --max-items 1
aws cloudformation list-stacks --max-items 1
```

## 📁 File Structure

```
/Users/marklovelady/_dev/_projects/api-lambdav2/
├── iam-policy-virtuoso-cdk.json    # Custom IAM policy
├── setup-iam-user.sh               # Automated setup script
├── verify-iam-setup.sh             # Verification script
├── audit-iam-security.sh           # Security audit script
└── IAM_SETUP_GUIDE.md              # This guide
```

## 🔄 Maintenance Tasks

### Monthly
- [ ] Run security audit script
- [ ] Check CloudTrail logs for unusual activity
- [ ] Review access key usage

### Quarterly
- [ ] Review and update permissions
- [ ] Rotate access keys
- [ ] Audit policy effectiveness
- [ ] Update documentation

### As Needed
- [ ] Add new service permissions for features
- [ ] Remove unused permissions
- [ ] Update resource ARN patterns

## 📞 Support

If you encounter issues:

1. Run the verification script: `./verify-iam-setup.sh`
2. Check the troubleshooting section above
3. Review AWS CloudTrail logs for detailed error information
4. Ensure you're not using root credentials

## 🔗 Additional Resources

- [AWS IAM Best Practices](https://docs.aws.amazon.com/IAM/latest/UserGuide/best-practices.html)
- [CDK Security Best Practices](https://docs.aws.amazon.com/cdk/latest/guide/security.html)
- [AWS CloudTrail User Guide](https://docs.aws.amazon.com/awscloudtrail/latest/userguide/)
- [Principle of Least Privilege](https://docs.aws.amazon.com/IAM/latest/UserGuide/best-practices.html#grant-least-privilege)

---

**✅ Following this guide ensures secure, auditable, and maintainable CDK deployments without using root credentials.**