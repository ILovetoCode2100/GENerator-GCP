# Virtuoso API Gateway Deployment Guide

This guide walks you through deploying the complete Virtuoso API Gateway infrastructure on AWS using CDK.

## Overview

The infrastructure includes:
- **HTTP API Gateway** with 19 endpoints
- **19 Lambda functions** (one per endpoint) using Node.js 20.x ARM64
- **Custom authorizer** for Bearer token validation
- **Secrets Manager** for secure API configuration
- **CloudWatch Logs** for monitoring
- **IAM roles** with least-privilege permissions

## Prerequisites

### Required Software
- **AWS CLI** v2.x configured with appropriate permissions
- **Node.js** 18+ and npm
- **AWS CDK CLI** v2.x (`npm install -g aws-cdk`)

### Required AWS Permissions
Your AWS credentials need permissions for:
- CloudFormation (create/update/delete stacks)
- Lambda (create/update functions)
- API Gateway (create/configure HTTP APIs)
- IAM (create roles and policies)
- Secrets Manager (create/update secrets)
- CloudWatch Logs (create log groups)

## Quick Start

### Step 1: Clone and Setup

```bash
# Navigate to the CDK directory
cd /path/to/api-lambdav2/cdk

# Install dependencies
npm install
cd lambda && npm install && cd ..
```

### Step 2: Configure AWS

```bash
# Configure AWS CLI (if not already done)
aws configure

# Verify configuration
aws sts get-caller-identity
```

### Step 3: Deploy with Script

```bash
# Use the automated deployment script
./scripts/deploy.sh

# Or deploy manually
npm run deploy
```

### Step 4: Configure API Key

```bash
# Update the secret with your Virtuoso API key
./scripts/update-secrets.sh --api-key YOUR_VIRTUOSO_API_KEY
```

### Step 5: Test the Deployment

```bash
# Test all endpoints
./scripts/test-endpoints.sh --token YOUR_BEARER_TOKEN
```

## Manual Deployment Steps

If you prefer manual deployment:

### 1. Install Dependencies

```bash
# CDK dependencies
npm install

# Lambda dependencies
cd lambda
npm install
cd ..
```

### 2. Build the Project

```bash
npm run build
```

### 3. Bootstrap CDK (First Time Only)

```bash
# Get your AWS account ID and region
export AWS_ACCOUNT=$(aws sts get-caller-identity --query Account --output text)
export AWS_REGION=$(aws configure get region)

# Bootstrap CDK
cdk bootstrap aws://$AWS_ACCOUNT/$AWS_REGION
```

### 4. Deploy the Stack

```bash
# Set environment variables
export CDK_DEFAULT_ACCOUNT=$AWS_ACCOUNT  
export CDK_DEFAULT_REGION=$AWS_REGION

# Deploy
cdk deploy
```

### 5. Update Secrets Manager

After deployment, update the secret `virtuoso-api-config`:

```bash
aws secretsmanager update-secret \
  --secret-id virtuoso-api-config \
  --secret-string '{
    "virtuosoApiBaseUrl": "https://api-app2.virtuoso.qa/api",
    "organizationId": "2242", 
    "apiKey": "YOUR_ACTUAL_API_KEY"
  }'
```

## Configuration Options

### Environment Variables

Set these before deployment to customize the stack:

```bash
export ENVIRONMENT=production          # deployment environment
export CDK_DEFAULT_ACCOUNT=123456789   # AWS account ID
export CDK_DEFAULT_REGION=us-west-2    # AWS region
```

### Stack Parameters

Modify `/lib/virtuoso-api-stack.ts` to customize:

- Lambda memory size (default: 256MB)
- Lambda timeout (default: 30 seconds)
- Log retention (default: 7 days)
- API throttling limits
- CORS configuration

### Secrets Configuration

The `virtuoso-api-config` secret should contain:

```json
{
  "virtuosoApiBaseUrl": "https://api-app2.virtuoso.qa/api",
  "organizationId": "2242",
  "apiKey": "your-virtuoso-api-key-here"
}
```

## Testing the Deployment

### 1. Get the API Gateway URL

```bash
# From CloudFormation outputs
aws cloudformation describe-stacks \
  --stack-name VirtuosoApiStack \
  --query 'Stacks[0].Outputs[?OutputKey==`ApiGatewayUrl`].OutputValue' \
  --output text
```

### 2. Test Basic Endpoints

```bash
export API_URL="https://your-api-id.execute-api.region.amazonaws.com"
export TOKEN="your-bearer-token"

# Test user endpoint
curl -H "Authorization: Bearer $TOKEN" "$API_URL/api/user"

# Test projects endpoint  
curl -H "Authorization: Bearer $TOKEN" "$API_URL/api/projects"

# Create a project
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Project","organizationId":"2242"}' \
  "$API_URL/api/projects"
```

### 3. Automated Testing

```bash
# Test all endpoints with the provided script
./scripts/test-endpoints.sh --token YOUR_BEARER_TOKEN
```

## Monitoring and Troubleshooting

### CloudWatch Logs

Each Lambda function has its own log group:

```bash
# View logs for a specific function
aws logs tail /aws/lambda/virtuoso-get-user --follow

# View logs for the authorizer
aws logs tail /aws/lambda/virtuoso-api-authorizer --follow
```

### Common Issues

#### 1. Authorization Failed (401/403)

**Causes:**
- Invalid or missing Bearer token
- Token not properly formatted
- Secrets Manager configuration incorrect

**Solutions:**
```bash
# Check token format (should start with 'Bearer ')
echo "Authorization: Bearer $TOKEN"

# Verify secrets configuration
aws secretsmanager get-secret-value --secret-id virtuoso-api-config

# Check authorizer logs
aws logs tail /aws/lambda/virtuoso-api-authorizer --follow
```

#### 2. Internal Server Error (500)

**Causes:**
- Lambda function timeout
- Network connectivity issues
- Virtuoso API errors

**Solutions:**
```bash
# Check Lambda function logs
aws logs tail /aws/lambda/virtuoso-get-user --follow

# Verify timeout settings in CDK stack
# Increase timeout if needed in lib/virtuoso-api-stack.ts
```

#### 3. CORS Errors

**Causes:**
- Frontend origin not allowed
- Missing preflight handling

**Solutions:**
- Update CORS configuration in CDK stack
- Ensure OPTIONS method is handled
- Verify allowed origins match your frontend domain

#### 4. Rate Limiting (429)

**Causes:**
- Exceeded API Gateway throttling limits
- Virtuoso API rate limits

**Solutions:**
```bash
# Check API Gateway throttling settings
# Increase limits in CDK stack if needed

# Monitor CloudWatch metrics for throttling
aws cloudwatch get-metric-statistics \
  --namespace AWS/ApiGateway \
  --metric-name ThrottledRequests \
  --start-time 2023-01-01T00:00:00Z \
  --end-time 2023-01-02T00:00:00Z \
  --period 3600 \
  --statistics Sum
```

### Performance Monitoring

Monitor Lambda functions through CloudWatch:

```bash
# View function metrics
aws cloudwatch get-metric-statistics \
  --namespace AWS/Lambda \
  --metric-name Duration \
  --dimensions Name=FunctionName,Value=virtuoso-get-user \
  --start-time 2023-01-01T00:00:00Z \
  --end-time 2023-01-02T00:00:00Z \
  --period 3600 \
  --statistics Average,Maximum
```

## Updating the Deployment

### Code Changes

```bash
# After making changes to Lambda functions
npm run build
cdk deploy
```

### Configuration Changes

```bash
# Update secrets
./scripts/update-secrets.sh --api-key NEW_API_KEY

# Update CDK stack
cdk deploy
```

### Rolling Back

```bash
# Rollback to previous version
cdk deploy --previous-parameters

# Or destroy and redeploy
cdk destroy
cdk deploy
```

## Cost Optimization

### Lambda Settings
- **ARM64 architecture**: Better price/performance ratio
- **256MB memory**: Optimized for most workloads
- **30s timeout**: Prevents runaway costs

### API Gateway
- **HTTP API**: 70% cheaper than REST API
- **Rate limiting**: Prevents abuse and unexpected costs

### Monitoring
- **7-day log retention**: Balances debugging needs with costs
- **CloudWatch metrics**: Monitor usage patterns

### Estimated Monthly Costs

For moderate usage (1M requests/month):
- **API Gateway**: ~$1.00
- **Lambda**: ~$3.50 (including authorizer)
- **Secrets Manager**: ~$0.40
- **CloudWatch Logs**: ~$0.50
- **Total**: ~$5.40/month

## Security Best Practices

### Authentication
- Bearer tokens validated by custom authorizer
- Tokens forwarded to Virtuoso API for verification
- No token storage in Lambda functions

### Network Security
- HTTPS-only communication
- CORS properly configured
- No VPC required (uses AWS backbone)

### IAM Permissions
- Least-privilege principle
- Lambda functions can only access required services
- Secrets Manager access limited to specific secret

### Data Protection
- API keys stored in Secrets Manager
- CloudWatch logs don't expose sensitive data
- Request/response logging excludes authorization headers

## Production Considerations

### Scaling
- Lambda auto-scales to handle load
- API Gateway handles up to 10,000 RPS by default
- Consider reserved concurrency for critical functions

### High Availability
- Multi-AZ deployment automatic
- API Gateway has 99.95% SLA
- Lambda has 99.95% SLA

### Disaster Recovery
- CDK code in version control
- Secrets Manager cross-region replication available
- CloudFormation enables infrastructure as code

### Compliance
- All services support SOC, PCI DSS, HIPAA
- CloudTrail logs all API calls
- VPC deployment option available if needed

## Clean Up

### Destroy the Stack

```bash
# Using the deployment script
./scripts/deploy.sh destroy

# Or manually
cdk destroy
```

### Manual Cleanup

If automated cleanup fails:

```bash
# Delete CloudFormation stack
aws cloudformation delete-stack --stack-name VirtuosoApiStack

# Delete any remaining log groups
aws logs describe-log-groups --log-group-name-prefix "/aws/lambda/virtuoso-" \
  --query 'logGroups[].logGroupName' --output text | \
  xargs -I {} aws logs delete-log-group --log-group-name {}
```

## Support

### Documentation
- [AWS CDK Documentation](https://docs.aws.amazon.com/cdk/)
- [AWS Lambda Documentation](https://docs.aws.amazon.com/lambda/)
- [AWS API Gateway Documentation](https://docs.aws.amazon.com/apigateway/)

### Troubleshooting Resources
- CloudWatch Logs for detailed error messages
- AWS X-Ray for request tracing (can be enabled)
- CDK diff command to see changes before deployment

### Getting Help
1. Check CloudWatch Logs for specific error messages
2. Use AWS CLI to inspect resource configurations
3. Review CDK documentation for configuration options
4. Test individual endpoints to isolate issues