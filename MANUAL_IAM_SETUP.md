# Manual IAM Setup Guide for Virtuoso CDK Deployment

## Current Situation
The AWS CLI is configured but the credentials appear to be invalid. You need to:
1. Log into AWS Console with valid credentials
2. Create an IAM user manually
3. Configure AWS CLI with the new user credentials

## Step-by-Step Manual Setup

### 1. Log into AWS Console
1. Go to https://console.aws.amazon.com/
2. Sign in with your AWS account (NOT using root credentials for production)

### 2. Create IAM User
1. Navigate to IAM → Users → Create User
2. User name: `virtuoso-cdk-deploy`
3. Check "Provide user access to the AWS Management Console" (optional)
4. Select "I want to create an IAM user"
5. Click Next

### 3. Set Permissions
1. Select "Attach policies directly"
2. Click "Create policy" to open a new tab
3. Switch to JSON editor and paste the policy from `iam-policy-virtuoso-cdk.json`
4. Name the policy: `VirtuosoCDKDeploymentPolicy`
5. Go back to user creation tab and refresh policies
6. Select the newly created policy
7. Click Next and Create User

### 4. Create Access Keys
1. Click on the newly created user
2. Go to "Security credentials" tab
3. Under "Access keys", click "Create access key"
4. Select "Command Line Interface (CLI)"
5. Click "Create access key"
6. **SAVE THE CREDENTIALS SECURELY** - You won't see the secret key again

### 5. Configure AWS CLI
```bash
# Option 1: Configure a named profile (recommended)
aws configure --profile virtuoso-cdk-deploy

# Enter when prompted:
# AWS Access Key ID: [Your Access Key]
# AWS Secret Access Key: [Your Secret Key]
# Default region name: us-east-1
# Default output format: json

# Option 2: Configure default profile
aws configure

# Test the configuration
export AWS_PROFILE=virtuoso-cdk-deploy
aws sts get-caller-identity
```

### 6. Alternative: Use AWS SSO (Recommended for Teams)
If your organization uses AWS SSO:
```bash
aws configure sso
# Follow the prompts to set up SSO access
```

## Quick Verification Commands

```bash
# Check current user
aws sts get-caller-identity --profile virtuoso-cdk-deploy

# Verify permissions
aws iam list-attached-user-policies --user-name virtuoso-cdk-deploy --profile virtuoso-cdk-deploy

# Test basic operations
aws s3 ls --profile virtuoso-cdk-deploy
aws lambda list-functions --profile virtuoso-cdk-deploy
```

## Environment Setup for Deployment

```bash
# Set the profile for your session
export AWS_PROFILE=virtuoso-cdk-deploy

# Verify it's working
aws sts get-caller-identity

# Should show:
# {
#     "UserId": "AID...",
#     "Account": "824988897938",
#     "Arn": "arn:aws:iam::824988897938:user/virtuoso-cdk-deploy"
# }
```

## Next Steps

Once you have valid AWS credentials configured:

1. Navigate to CDK directory:
   ```bash
   cd /Users/marklovelady/_dev/_projects/api-lambdav2/cdk
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Bootstrap CDK (first time only):
   ```bash
   npx cdk bootstrap aws://824988897938/us-east-1
   ```

4. Deploy the stack:
   ```bash
   npm run deploy
   ```

## Security Best Practices

1. **Never use root credentials** for application deployment
2. **Rotate access keys** every 90 days
3. **Use MFA** on your IAM user when possible
4. **Monitor CloudTrail** for unusual activity
5. **Consider using temporary credentials** via STS when possible

## Troubleshooting

### Invalid Credentials Error
- Check if access keys are active in IAM console
- Ensure no typos in credentials
- Verify the keys haven't been rotated

### Permission Denied Errors
- Ensure the IAM policy is attached to the user
- Check CloudTrail logs for specific permission failures
- Verify the policy JSON is correctly formatted

### Region Issues
- Ensure AWS_DEFAULT_REGION is set correctly
- Some services might not be available in all regions