# Deployment Options for Virtuoso API Gateway

## Current Situation
The `virtuoso-dev` AWS profile lacks sufficient permissions to deploy the CDK stack. Here are your options:

## Option 1: Use AWS Admin/PowerUser Credentials (Recommended)

### Steps:
1. Configure AWS CLI with admin credentials:
```bash
aws configure --profile virtuoso-admin
# Enter credentials with AdministratorAccess or PowerUserAccess policy
```

2. Deploy with the admin profile:
```bash
cd /Users/marklovelady/_dev/_projects/api-lambdav2/cdk
export AWS_PROFILE=virtuoso-admin
npx cdk bootstrap
npm run deploy
```

## Option 2: Add Required Permissions to virtuoso-dev User

### Required IAM Policies:
Add these AWS managed policies to the `virtuoso-dev` user:
- `IAMFullAccess` (for creating Lambda execution roles)
- `AWSCloudFormationFullAccess` (for stack operations)
- `AmazonSSMFullAccess` (for bootstrap parameters)
- `AmazonS3FullAccess` (for CDK assets)

Or use the custom policy from `iam-policy-virtuoso-cdk.json`

## Option 3: Manual Deployment Without CDK

### Convert to CloudFormation:
```bash
# Synthesize the template
npx cdk synth > virtuoso-api-template.json

# Deploy manually via AWS Console
# 1. Go to CloudFormation console
# 2. Create stack → Upload template
# 3. Use virtuoso-api-template.json
```

## Option 4: Use Alternative Deployment Methods

### A. Deploy to AWS Lambda + API Gateway manually:
1. Create Lambda functions manually in AWS Console
2. Create HTTP API Gateway
3. Configure routes and integrations

### B. Use Serverless Framework:
```bash
# Convert CDK to Serverless Framework
npm install -g serverless
serverless create --template aws-nodejs --path virtuoso-api
# Configure serverless.yml with the same functions
```

### C. Use AWS SAM:
```bash
# Convert to SAM template
sam init
# Configure template.yaml with Lambda functions
sam deploy --guided
```

## Option 5: Deploy to Alternative Platforms

### A. Deploy to Google Cloud Platform:
- Already have GCP deployment ready in `/gcp` directory
- Use Cloud Functions + API Gateway

### B. Deploy to Render.com:
- Containerized deployment option available
- See `RENDER_DEPLOYMENT.md`

### C. Deploy to Vercel/Netlify:
- Convert Lambda functions to serverless functions
- Use their native API routing

## Quick Fix for Current Situation

If you have access to AWS Console with admin privileges:

1. **Fix Bootstrap Stack**:
   - Go to CloudFormation console
   - Find `CDKToolkit` stack
   - Delete the stack (may need to skip resources)
   - Re-run bootstrap with admin credentials

2. **Grant Temporary Permissions**:
   - Attach `AdministratorAccess` policy to `virtuoso-dev` user
   - Deploy the stack
   - Remove admin access after deployment

## Recommended Next Steps

1. **For Production**: Use Option 1 with proper admin credentials
2. **For Testing**: Use Option 4A (GCP deployment) which is already configured
3. **For Quick Demo**: Use Option 3 (manual CloudFormation upload)

## Security Notes

- Never use root AWS credentials
- Create deployment-specific IAM users
- Use temporary credentials when possible
- Rotate access keys regularly
- Enable MFA on privileged accounts