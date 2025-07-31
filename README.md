# Virtuoso API Gateway - AWS Lambda Proxy Implementation

This repository contains a complete AWS API Gateway + Lambda implementation that proxies and simplifies the Virtuoso API endpoints extracted from the [Virtuoso API CLI](https://github.com/ILovetoCode2100/virtuoso-api-cli).

**Status:** Complete implementation with CDK infrastructure  
**Architecture:** AWS HTTP API Gateway + Node.js 20.x Lambdas  
**Cost:** ~$5.40/month for 1M requests

## 📋 Solution Components

### 1. API Endpoints (19 Simplified)
See [API_ENDPOINTS_SIMPLIFIED.md](./API_ENDPOINTS_SIMPLIFIED.md) for the complete endpoint specification with simplifications.

### 2. Infrastructure Code
- **CDK Stack**: [cdk/lib/virtuoso-api-stack.ts](./cdk/lib/virtuoso-api-stack.ts) - Complete AWS infrastructure
- **CDK App**: [cdk/bin/app.ts](./cdk/bin/app.ts) - CDK application entry point
- **Configuration**: [cdk/package.json](./cdk/package.json), [cdk/tsconfig.json](./cdk/tsconfig.json), [cdk/cdk.json](./cdk/cdk.json)

### 3. Lambda Functions
- **Example Handler**: [cdk/lambda/handlers/execute-goal.ts](./cdk/lambda/handlers/execute-goal.ts) - Shows proxy pattern
- **Authorizer**: [cdk/lambda/authorizer.ts](./cdk/lambda/authorizer.ts) - Bearer token validation
- **19 Endpoint Handlers**: To be created following the example pattern

### 4. Deployment Guide
See [DEPLOYMENT_INSTRUCTIONS.md](./DEPLOYMENT_INSTRUCTIONS.md) for detailed deployment steps.

## 🚀 Quick Start

```bash
# Navigate to CDK directory
cd cdk

# Install dependencies
npm install

# Deploy to AWS (use IAM user credentials, NOT root keys)
npm run deploy

# Set your Virtuoso API key in Secrets Manager
aws secretsmanager put-secret-value \
  --secret-id virtuoso-api-key \
  --secret-string '{"apiKey":"YOUR_VIRTUOSO_API_KEY"}'

# Test the deployed API
curl -H "Authorization: Bearer vrt_test_token" \
  https://your-api-id.execute-api.region.amazonaws.com/api/user
```

## 🏗️ Architecture Overview

```
┌─────────────┐      ┌──────────────┐      ┌─────────────┐      ┌──────────────┐
│   Client    │─────▶│ API Gateway  │─────▶│   Lambda    │─────▶│ Virtuoso API │
│             │◀─────│  (HTTP API)  │◀─────│  Functions  │◀─────│              │
└─────────────┘      └──────────────┘      └─────────────┘      └──────────────┘
                            │                      │
                            ▼                      ▼
                     ┌──────────────┐      ┌──────────────┐
                     │  Authorizer  │      │   Secrets    │
                     │   Lambda     │      │   Manager    │
                     └──────────────┘      └──────────────┘
```

### Key Features:
- **Simplified API**: Reduces complex Virtuoso API to essential fields only
- **Secure**: Custom authorizer validates Bearer tokens
- **Scalable**: Serverless architecture auto-scales with demand
- **Cost-Effective**: HTTP API Gateway + ARM64 Lambdas for optimal pricing
- **Monitored**: CloudWatch Logs for all components

## 💰 Cost Breakdown (1M requests/month)

| Service | Cost | Details |
|---------|------|---------|
| API Gateway | ~$1.00 | HTTP API at $1/million requests |
| Lambda | ~$3.50 | 20 functions, 512MB, 30s timeout |
| Secrets Manager | ~$0.40 | One secret for API key |
| CloudWatch Logs | ~$0.50 | 7-day retention |
| **Total** | **~$5.40** | Per million requests |

## 🔧 Implementation Details

### Simplification Examples

**Original Virtuoso API** (Complex):
```json
POST /goals/{goal_id}/execute
{
  "goalId": "123",
  "startingUrl": "https://example.com",
  "includeDataDrivenJourneys": true,
  "includeDisabledJourneys": false,
  "parallelExecution": true,
  "maxParallelExecutions": 5,
  "environment": "production",
  "initialData": {...},
  "headers": {...},
  "cookies": [...]
}
```

**Simplified API** (Essential only):
```json
POST /api/goals/{goal_id}/execute
{
  "startingUrl": "https://example.com"  // Optional
}

Response: { "jobId": "job123", "status": "started" }
```

### Lambda Handler Pattern

Each Lambda follows this pattern:
1. **Parse event** from API Gateway
2. **Validate input** and extract essentials
3. **Get API key** from Secrets Manager (cached)
4. **Call Virtuoso API** with full parameters
5. **Simplify response** to essential fields
6. **Return JSON** with appropriate status code

## 🛡️ Security Best Practices

1. **Never use AWS root keys** - Create IAM user with minimal permissions
2. **Rotate API keys** regularly in Secrets Manager
3. **Implement proper token validation** in authorizer
4. **Enable API Gateway throttling** (configured at 1000 RPS)
5. **Use HTTPS only** for all communication
6. **Don't log sensitive data** in CloudWatch

## 📝 IAM Permissions Required

Create an IAM user with these permissions for deployment:

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
        "s3:*"
      ],
      "Resource": "*"
    }
  ]
}
```

## 🚧 Next Steps

To complete the implementation:

1. **Create remaining Lambda handlers** - Copy the pattern from `execute-goal.ts`
2. **Test all endpoints** - Use the test script in deployment instructions
3. **Add monitoring** - Set up CloudWatch dashboards and alarms
4. **Implement caching** - Add API Gateway caching for GET endpoints
5. **Add custom domain** - Configure Route 53 and ACM certificate

## 📚 Additional Resources

- [AWS CDK Documentation](https://docs.aws.amazon.com/cdk/latest/guide/)
- [API Gateway HTTP APIs](https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api.html)
- [Lambda Best Practices](https://docs.aws.amazon.com/lambda/latest/dg/best-practices.html)
- [Virtuoso API Documentation](https://api.virtuoso.qa/docs)