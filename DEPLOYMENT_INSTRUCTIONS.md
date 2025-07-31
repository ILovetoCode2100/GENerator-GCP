# Virtuoso API Gateway - Deployment Instructions

## Section 4: Deployment Instructions

### Prerequisites

1. **AWS CLI** configured with appropriate credentials
2. **Node.js 20.x** or later installed
3. **AWS CDK** installed globally: `npm install -g aws-cdk`
4. **Git** for version control

### Step 1: Install Dependencies

```bash
cd /Users/marklovelady/_dev/_projects/api-lambdav2/cdk
npm install
```

### Step 2: Bootstrap CDK (First Time Only)

If this is your first time using CDK in this AWS account/region:

```bash
cdk bootstrap aws://ACCOUNT-NUMBER/REGION
# Example: cdk bootstrap aws://123456789012/us-east-1
```

### Step 3: Configure Virtuoso API Key

After deployment, you'll need to set your Virtuoso API key in AWS Secrets Manager:

```bash
# Option 1: AWS CLI
aws secretsmanager put-secret-value \
  --secret-id virtuoso-api-key \
  --secret-string '{"apiKey":"YOUR_VIRTUOSO_API_KEY"}'

# Option 2: AWS Console
# Navigate to AWS Secrets Manager and update the 'virtuoso-api-key' secret
```

### Step 4: Deploy the Stack

```bash
# Build and deploy
npm run deploy

# Or manually:
npm run build
cdk deploy --require-approval never
```

### Step 5: Test the Deployment

After deployment, CDK will output the API endpoint URL. Test it:

```bash
# Get API endpoint from CDK output
API_ENDPOINT=$(aws cloudformation describe-stacks \
  --stack-name VirtuosoApiStack \
  --query 'Stacks[0].Outputs[?OutputKey==`ApiEndpoint`].OutputValue' \
  --output text)

# Test with curl (replace with your Bearer token)
curl -H "Authorization: Bearer vrt_your_token_here" \
  $API_ENDPOINT/api/user

# Test project listing
curl -H "Authorization: Bearer vrt_your_token_here" \
  $API_ENDPOINT/api/projects

# Test goal execution
curl -X POST \
  -H "Authorization: Bearer vrt_your_token_here" \
  -H "Content-Type: application/json" \
  -d '{"startingUrl":"https://example.com"}' \
  $API_ENDPOINT/api/goals/YOUR_GOAL_ID/execute
```

### Step 6: Monitor and Debug

View Lambda logs in CloudWatch:

```bash
# List all Lambda functions
aws lambda list-functions --query 'Functions[?starts_with(FunctionName, `virtuoso-`)].FunctionName'

# View logs for specific function
aws logs tail /aws/lambda/virtuoso-get-user --follow
```

### Advanced Configuration

#### Custom Domain

To add a custom domain:

```typescript
// Add to CDK stack
import * as certificatemanager from 'aws-cdk-lib/aws-certificatemanager';
import * as route53 from 'aws-cdk-lib/aws-route53';

const domainName = 'api.yourdomain.com';
const certificate = new certificatemanager.Certificate(this, 'Certificate', {
  domainName,
  validation: certificatemanager.CertificateValidation.fromDns(),
});

const customDomain = new apigatewayv2.DomainName(this, 'Domain', {
  domainName,
  certificate,
});

new apigatewayv2.HttpApiMapping(this, 'Mapping', {
  api: httpApi,
  domainName: customDomain,
});
```

#### Environment Variables

Set custom Virtuoso API base URL:

```bash
export VIRTUOSO_API_BASE_URL=https://custom-api.virtuoso.qa
cdk deploy
```

### Cleanup

To remove all resources:

```bash
cdk destroy
```

### Cost Optimization Tips

1. **Use ARM64 Lambda**: Already configured for better price/performance
2. **Adjust memory**: Lower to 256MB if response times are acceptable
3. **Enable caching**: Add API Gateway caching for read endpoints
4. **Set up alarms**: Monitor for excessive usage

### Security Best Practices

1. **Rotate API keys** regularly in Secrets Manager
2. **Implement proper token validation** in the authorizer
3. **Enable AWS WAF** for DDoS protection
4. **Use VPC endpoints** if calling from within AWS
5. **Enable API Gateway access logging**

### Troubleshooting

**Common Issues:**

1. **401 Unauthorized**: Check Bearer token format (must start with `vrt_`)
2. **500 Internal Server Error**: Check Lambda logs for details
3. **Timeout errors**: Increase Lambda timeout or optimize Virtuoso API calls
4. **CORS errors**: Verify origin is allowed in API Gateway configuration

**Debug Commands:**

```bash
# Check stack status
cdk diff

# View synthesized CloudFormation
cdk synth

# Check Lambda environment variables
aws lambda get-function-configuration --function-name virtuoso-get-user

# Test Lambda directly
aws lambda invoke \
  --function-name virtuoso-get-user \
  --payload '{"headers":{"authorization":"Bearer vrt_test"}}' \
  response.json
```