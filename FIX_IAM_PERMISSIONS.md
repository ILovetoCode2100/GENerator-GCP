# How to Fix IAM Permissions for CDK Deployment

## ⚠️ Security Warning
**NEVER use AWS root credentials for deployments.** Root credentials have unrestricted access and should only be used for initial account setup and creating IAM users.

## Recommended Approach

### Option 1: Create a New IAM User with Proper Permissions

1. **Log into AWS Console** (using root account if necessary)
2. **Navigate to IAM** → Users → Add User
3. **Create user** `virtuoso-cdk-deploy` with programmatic access
4. **Attach this policy** (create as custom policy):

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "cloudformation:*",
        "lambda:*",
        "apigatewayv2:*",
        "iam:*",
        "logs:*",
        "secretsmanager:*",
        "s3:*",
        "ssm:*"
      ],
      "Resource": "*"
    }
  ]
}
```

5. **Configure AWS CLI** with the new credentials:
```bash
aws configure --profile virtuoso-cdk
# Enter the new access key and secret key
```

6. **Deploy using the profile**:
```bash
cd cdk
export AWS_PROFILE=virtuoso-cdk
npm run deploy
```

### Option 2: Fix Existing virtuoso-dev User

1. **In AWS Console**, go to IAM → Users → virtuoso-dev
2. **Add these managed policies**:
   - PowerUserAccess (or AdministratorAccess temporarily)
   - Or attach the custom policy above

3. **Verify permissions**:
```bash
aws sts get-caller-identity --profile virtuoso-dev
aws iam simulate-principal-policy \
  --policy-source-arn $(aws sts get-caller-identity --query Arn --output text) \
  --action-names cloudformation:CreateStack \
  --profile virtuoso-dev
```

### Option 3: Use Temporary Admin Access

1. **Create temporary credentials** with admin access:
```bash
# In AWS Console, create a new IAM user with AdministratorAccess
# Use it only for CDK deployment
# Delete or disable after deployment
```

## Alternative: Switch to GCP

Given the IAM complexity, consider using the already-working GCP deployment:

```bash
cd gcp
./one-click-deploy.sh
```

GCP has simpler permission models and the deployment is already configured and tested.

## Security Best Practices

1. **Never commit credentials** to version control
2. **Use IAM roles** for production deployments
3. **Enable MFA** on all privileged accounts
4. **Rotate credentials** regularly
5. **Use least privilege** principle

## If You Must Use Provided Credentials

⚠️ **NOT RECOMMENDED** - Only if you understand the risks:

1. The credentials you provided appear invalid (same value for both)
2. If they were valid root credentials, you should:
   - Immediately create proper IAM users
   - Enable MFA on the root account
   - Never use root credentials again
   - Rotate the root access keys

## Recommended Next Steps

1. **Create proper IAM user** with limited CDK deployment permissions
2. **Or switch to GCP** which is already working
3. **Enable CloudTrail** to audit all actions
4. **Set up AWS Organizations** for better account management

Remember: Security is not optional. Proper IAM setup may take time but prevents catastrophic breaches.