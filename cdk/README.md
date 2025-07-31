# Virtuoso API Gateway CDK Infrastructure

This CDK project creates a complete AWS infrastructure for proxying the Virtuoso API through AWS API Gateway and Lambda functions.

## Architecture

- **HTTP API Gateway**: Cost-efficient API Gateway with CORS support
- **19 Lambda Functions**: One per endpoint, using Node.js 20.x ARM64 runtime
- **Custom Authorizer**: Lambda-based authorization for Bearer token validation
- **Secrets Manager**: Secure storage for API configuration
- **CloudWatch Logs**: Comprehensive logging for all functions

## Prerequisites

- AWS CLI configured with appropriate permissions
- Node.js 18+ and npm/yarn
- AWS CDK CLI installed (`npm install -g aws-cdk`)

## Quick Start

1. **Install dependencies**:
   ```bash
   npm install
   ```

2. **Configure AWS credentials**:
   ```bash
   aws configure
   ```

3. **Bootstrap CDK** (first time only):
   ```bash
   cdk bootstrap
   ```

4. **Deploy the stack**:
   ```bash
   npm run deploy
   ```

5. **Update API configuration in Secrets Manager**:
   After deployment, update the secret `virtuoso-api-config` with your actual Virtuoso API key.

## API Endpoints

The deployed API Gateway provides the following endpoints:

### User Management
- `GET /api/user` - Get current user details

### Project Management
- `GET /api/projects` - List projects
- `POST /api/projects` - Create project
- `GET /api/projects/{project_id}/goals` - List goals in project

### Goal Management
- `POST /api/goals` - Create goal
- `GET /api/goals/{goal_id}/versions` - Get goal versions
- `POST /api/goals/{goal_id}/execute` - Execute goal
- `POST /api/goals/{goal_id}/snapshots/{snapshot_id}/execute` - Execute snapshot

### Journey & Checkpoint Management
- `POST /api/journeys` - Create journey
- `POST /api/checkpoints` - Create checkpoint
- `GET /api/checkpoints/{checkpoint_id}/steps` - Get checkpoint steps
- `POST /api/steps` - Create step

### Execution Management
- `POST /api/executions` - Start execution
- `GET /api/executions/{execution_id}` - Get execution status
- `GET /api/executions/{execution_id}/analysis` - Get execution analysis

### Library Management
- `POST /api/library/checkpoints` - Create library checkpoint
- `GET /api/library/checkpoints` - List library checkpoints

### Test Data & Environment Management
- `POST /api/testdata/tables` - Create test data table
- `POST /api/environments` - Create environment

## Configuration

### Environment Variables

Each Lambda function receives these environment variables:

- `VIRTUOSO_SECRET_ARN`: ARN of the Secrets Manager secret
- `NODE_ENV`: Set to 'production'
- `LOG_LEVEL`: Logging level (default: 'info')
- `TIMEOUT_MS`: Request timeout in milliseconds (default: 30000)
- `RETRY_ATTEMPTS`: Number of retry attempts (default: 3)

### Secrets Manager Configuration

The secret `virtuoso-api-config` should contain:

```json
{
  "virtuosoApiBaseUrl": "https://api-app2.virtuoso.qa/api",
  "organizationId": "2242",
  "apiKey": "your-api-key-here"
}
```

## Authentication

All endpoints require a Bearer token in the Authorization header:

```
Authorization: Bearer your-token-here
```

The custom authorizer validates the token format and forwards it to the Virtuoso API for actual authentication.

## CORS Configuration

The API Gateway is configured to allow:
- **Origins**: `*` (configure as needed)
- **Methods**: GET, POST, PUT, DELETE, OPTIONS
- **Headers**: Content-Type, Authorization, X-Api-Key, etc.
- **Credentials**: Enabled

## Monitoring & Logging

- All Lambda functions log to CloudWatch Logs
- Log retention is set to 7 days
- Function names follow the pattern: `virtuoso-{endpoint-name}`
- Log groups: `/aws/lambda/virtuoso-{endpoint-name}`

## Throttling

API Gateway throttling is configured with:
- **Rate limit**: 1000 requests per second
- **Burst limit**: 2000 concurrent requests

## Cost Optimization

- Uses HTTP API Gateway (cheaper than REST API)
- ARM64 Lambda architecture (better price/performance)
- Optimized memory allocation (256MB)
- Short log retention (7 days)
- Lambda bundling with minification

## Development

### Local Development

1. Install dependencies in lambda directory:
   ```bash
   cd lambda && npm install
   ```

2. Run TypeScript compilation:
   ```bash
   cd lambda && npx tsc
   ```

### Deployment Commands

- `npm run build` - Compile TypeScript
- `npm run deploy` - Deploy the stack
- `npm run destroy` - Destroy the stack
- `npm run synth` - Generate CloudFormation template
- `npm run diff` - Show differences before deployment

### Testing

After deployment, test endpoints with curl:

```bash
# Get API Gateway URL from CDK output
export API_URL="https://your-api-id.execute-api.region.amazonaws.com"

# Test user endpoint
curl -H "Authorization: Bearer your-token" "$API_URL/api/user"

# Test projects endpoint
curl -H "Authorization: Bearer your-token" "$API_URL/api/projects"
```

## Security Considerations

1. **API Keys**: Store in Secrets Manager, never in code
2. **CORS**: Configure allowed origins appropriately for production
3. **Rate Limiting**: Adjust throttling based on expected load
4. **Authorization**: Implement proper token validation in the authorizer
5. **VPC**: Consider VPC configuration for enhanced security

## Troubleshooting

### Common Issues

1. **Authorization Failed**: Check token format and Secrets Manager configuration
2. **502 Bad Gateway**: Check Lambda function logs for errors
3. **Timeout Errors**: Increase Lambda timeout or adjust TIMEOUT_MS environment variable
4. **CORS Errors**: Verify CORS configuration matches client requirements

### Checking Logs

```bash
# View logs for a specific function
aws logs tail /aws/lambda/virtuoso-get-user --follow

# View API Gateway access logs (if enabled)
aws logs tail /aws/apigateway/virtuoso-api-proxy --follow
```

## Clean Up

To avoid charges, destroy the stack when no longer needed:

```bash
npm run destroy
```

## Support

For issues or questions:
1. Check CloudWatch Logs for detailed error messages
2. Review the CDK documentation
3. Consult the Virtuoso API documentation