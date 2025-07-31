# Secure Deployment Guide for Virtuoso API

## 🔒 Critical Security Notice

**NEVER use root AWS credentials for deployment or share them with anyone.**

## Option 1: Use AWS Console (Recommended for Security)

1. **Log into AWS Console** using your root account
2. **Go to IAM** → Users → Create User
3. **Create user**: `virtuoso-cdk-deploy`
4. **Attach policy**: PowerUserAccess (or create custom policy)
5. **Generate access keys** and save them securely
6. **Configure locally**:
   ```bash
   aws configure --profile virtuoso-cdk
   # Enter the new credentials
   
   # Deploy
   export AWS_PROFILE=virtuoso-cdk
   cd cdk && npm run deploy
   ```

## Option 2: Use the Script (Requires Valid AWS Credentials)

If you have valid AWS credentials with admin permissions:

```bash
# Make the script executable
chmod +x create-iam-user-script.sh

# Run it (requires jq to be installed)
./create-iam-user-script.sh
```

## Option 3: Switch to GCP (Simplest)

Given the AWS complexity, use the working GCP deployment:

```bash
cd gcp
./one-click-deploy.sh
```

## Why Your Approach Won't Work

1. **Invalid Credentials**: The access key and secret key cannot be the same value
2. **Security Risk**: Root credentials should never be used for deployments
3. **Best Practice**: Always use IAM users with minimal required permissions

## Manual Steps to Fix Your Situation

### If You Have Access to AWS Console:

1. **Enable MFA on root account** immediately
2. **Create IAM admin user**:
   ```
   - Username: admin-yourname
   - Access: Programmatic + Console
   - Permissions: AdministratorAccess
   - Enable MFA
   ```
3. **Create deployment user**:
   ```
   - Username: virtuoso-cdk-deploy  
   - Access: Programmatic only
   - Permissions: Custom policy (see script)
   ```
4. **Delete any exposed root access keys**

### If You Only Have CLI Access:

You need valid credentials first. The ones provided won't work because:
- Access Key: AKIA4AFJLSKJCNF6VVH4
- Secret Key: AKIA4AFJLSKJCNF6VVH4 (same as access key - invalid)

## Alternative Deployment Options

### 1. Deploy to GCP Instead
```bash
# GCP has simpler permissions
cd gcp
gcloud auth login
./deploy-wizard.sh
```

### 2. Use Docker Locally
```bash
# No cloud permissions needed
docker-compose up -d
```

### 3. Manual Lambda Creation
- Create Lambda functions manually in AWS Console
- Upload code zips
- Configure API Gateway manually

## Getting Proper AWS Credentials

1. **From AWS Console**:
   - IAM → Users → Your User → Security Credentials
   - Create Access Key → CLI → Create
   
2. **From AWS CLI** (if you have working credentials):
   ```bash
   aws iam create-access-key --user-name your-username
   ```

## Next Steps

1. **Get valid AWS credentials** (not root)
2. **Run the IAM creation script** with those credentials
3. **Or switch to GCP** which is already configured

Remember: Security isn't optional. Taking shortcuts with root credentials can lead to:
- Complete account compromise
- Unexpected AWS bills
- Data breaches
- Loss of all resources

Please follow proper security practices. I'm here to help you do this the right way.