# AWS Deployment Guide - Virtuoso API Lambda Functions

## Prerequisites

Before deploying, ensure you have:

1. **AWS CLI** configured with credentials:
   ```bash
   aws configure
   ```

2. **AWS SAM CLI** installed:
   ```bash
   # macOS
   brew install aws-sam-cli
   
   # Windows/Linux
   # Follow: https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html
   ```

3. **Node.js 18.x** or later:
   ```bash
   node --version  # Should show v18.x or higher
   ```

4. **Virtuoso API Token** from your Virtuoso account

## Quick Deployment

```bash
# Clone and enter the project
git clone <repository-url>
cd lambda-api-gen

# Generate Lambda functions
node generate-lambdas.js

# Deploy to AWS
./deploy.sh YOUR_VIRTUOSO_API_TOKEN
```

## Step-by-Step Deployment

### 1. Generate Lambda Functions

The generator script creates Lambda function code and SAM template:

```bash
node generate-lambdas.js
```

This creates:
- `lambda-functions/` - Individual Lambda function directories
- `template.yaml` - AWS SAM template for deployment

### 2. Configure Deployment

Set deployment parameters via environment variables:

```bash
# Set AWS region (default: us-east-1)
export AWS_REGION=eu-west-1

# Set custom stack name (default: virtuoso-api-stack)
export STACK_NAME=my-virtuoso-stack

# Set S3 bucket for deployment artifacts
export S3_BUCKET=my-deployment-bucket
```

### 3. Deploy to AWS

Run the deployment script with your API token:

```bash
./deploy.sh YOUR_VIRTUOSO_API_TOKEN
```

The script will:
1. Create S3 bucket for deployment artifacts
2. Install Lambda layer dependencies
3. Package the application
4. Deploy via CloudFormation
5. Display the API Gateway URL

### 4. Verify Deployment

Test your deployment:

```bash
# Run test suite
node test-virtuoso-api.js

# Test specific endpoint
curl https://YOUR_API_GATEWAY_URL/virtuoso/projects
```

## Manual Deployment Steps

If you prefer manual deployment:

```bash
# 1. Install layer dependencies
cd lambda-layer/nodejs
npm install
cd ../..

# 2. Create S3 bucket
aws s3 mb s3://my-lambda-deployments

# 3. Package application
sam package \
  --template-file template.yaml \
  --s3-bucket my-lambda-deployments \
  --output-template-file packaged.yaml

# 4. Deploy stack
sam deploy \
  --template-file packaged.yaml \
  --stack-name virtuoso-api-stack \
  --capabilities CAPABILITY_IAM \
  --parameter-overrides ApiTokenValue=YOUR_TOKEN
```

## Post-Deployment Configuration

### 1. Update API Token

To update the Virtuoso API token:

```bash
aws ssm put-parameter \
  --name /virtuoso/api-token \
  --value "NEW_API_TOKEN" \
  --type SecureString \
  --overwrite
```

### 2. Enable CloudWatch Logs

View Lambda logs:

```bash
# List recent log streams
aws logs describe-log-streams \
  --log-group-name /aws/lambda/VirtuosoProjectHandler \
  --order-by LastEventTime \
  --descending

# View logs
aws logs tail /aws/lambda/VirtuosoProjectHandler --follow
```

### 3. Set Up Alarms

Create CloudWatch alarms for monitoring:

```bash
aws cloudwatch put-metric-alarm \
  --alarm-name "VirtuosoAPIErrors" \
  --alarm-description "Alert on Lambda errors" \
  --metric-name Errors \
  --namespace AWS/Lambda \
  --statistic Sum \
  --period 300 \
  --threshold 10 \
  --comparison-operator GreaterThanThreshold \
  --evaluation-periods 1
```

## Updating Functions

After code changes:

```bash
# Regenerate functions
node generate-lambdas.js

# Update all Lambda functions
node update-all-lambdas.js

# Or redeploy entire stack
./deploy.sh YOUR_VIRTUOSO_API_TOKEN
```

## Troubleshooting

### Lambda Function Not Found

```bash
# Check if stack deployed successfully
aws cloudformation describe-stacks --stack-name virtuoso-api-stack
```

### API Gateway 403 Errors

```bash
# Verify API Gateway deployment
aws apigateway get-rest-apis

# Check Lambda permissions
aws lambda get-policy --function-name VirtuosoProjectHandler
```

### SSM Parameter Access Denied

```bash
# Check Lambda execution role
aws iam get-role --role-name virtuoso-api-stack-VirtuosoProjectHandlerRole-XXX
```

## Cost Optimization

### 1. Configure Reserved Concurrency

Limit concurrent executions to control costs:

```bash
aws lambda put-function-concurrency \
  --function-name VirtuosoProjectHandler \
  --reserved-concurrent-executions 10
```

### 2. Set Up Cost Alerts

```bash
aws budgets create-budget \
  --account-id $(aws sts get-caller-identity --query Account --output text) \
  --budget file://budget.json
```

### 3. Enable X-Ray Sampling

Reduce X-Ray costs with sampling:

```bash
aws xray create-sampling-rule --cli-input-json file://sampling-rule.json
```

## Clean Up

To remove all resources:

```bash
# Delete the stack
aws cloudformation delete-stack --stack-name virtuoso-api-stack

# Remove S3 bucket
aws s3 rb s3://my-lambda-deployments --force

# Delete SSM parameter
aws ssm delete-parameter --name /virtuoso/api-token
```

## Next Steps

- Set up CI/CD pipeline with AWS CodePipeline
- Configure custom domain with Route 53
- Enable AWS WAF for additional security
- Implement caching with API Gateway
- Set up multi-region deployment

## Support

For issues or questions:
- Check CloudWatch Logs for error details
- Review Lambda function configuration
- Verify API token is valid
- Ensure IAM permissions are correct

## Security Best Practices

1. **Rotate API Tokens** regularly using AWS Secrets Manager rotation
2. **Enable VPC** for Lambda functions if accessing private resources
3. **Use IAM roles** with least privilege principle
4. **Enable API Gateway throttling** to prevent abuse
5. **Monitor with AWS GuardDuty** for security threats

---

*Last Updated: January 2025*