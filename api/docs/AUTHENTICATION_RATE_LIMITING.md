# Authentication and Rate Limiting

This document describes the authentication and rate limiting implementation for the Virtuoso API CLI.

## Authentication

### Overview

The API supports multiple authentication methods:

1. **API Key Authentication** (Primary)

   - Header: `X-API-Key: your-api-key`
   - Configured via environment variable `API_KEYS`

2. **JWT Bearer Token** (Future)
   - Header: `Authorization: Bearer <token>`
   - For OAuth2/SSO integration

### Configuration

API keys are configured in the environment:

```bash
# Single key
API_KEYS=my-secure-api-key

# Multiple keys (comma-separated)
API_KEYS=key1,key2,key3
```

### User Roles and Permissions

#### Roles

- **VIEWER**: Read-only access
- **DEVELOPER**: Read/write access to tests and projects
- **ADMIN**: Full access including user management
- **SERVICE_ACCOUNT**: Customizable permissions for automation

#### Permissions

- `read:projects` - View projects, goals, journeys, checkpoints
- `read:tests` - View test steps and configurations
- `read:executions` - View execution results
- `read:library` - View library items
- `read:templates` - View test templates
- `write:projects` - Create/modify projects
- `write:tests` - Create/modify test steps
- `write:executions` - Start/stop executions
- `write:library` - Manage library items
- `admin:*` - All administrative permissions

### Protected Endpoints

All `/api/v1/*` endpoints require authentication. Public endpoints include:

- `/health` - Health check
- `/docs` - API documentation
- `/openapi.json` - OpenAPI schema
- `/` - Root endpoint

### Authentication Flow

1. Client sends request with API key header
2. Middleware validates the API key
3. User context is loaded with permissions
4. Request is processed if authorized
5. Audit log records the action

### Error Responses

#### Authentication Failed (401)

```json
{
  "status": "error",
  "error_type": "authentication_error",
  "message": "Authentication required. Please provide a valid API key or JWT token.",
  "timestamp": "2024-01-20T10:30:00Z",
  "request_id": "req_123456"
}
```

#### Authorization Failed (403)

```json
{
  "status": "error",
  "error_type": "authorization_error",
  "message": "Insufficient permissions. Required: write:tests",
  "timestamp": "2024-01-20T10:30:00Z",
  "request_id": "req_123456"
}
```

## Rate Limiting

### Overview

Rate limiting is implemented using Redis with a sliding window algorithm. It supports multiple strategies and configurable limits per endpoint.

### Configuration

```bash
# Enable/disable rate limiting
RATE_LIMIT_ENABLED=true

# Default limits
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_PERIOD=60  # seconds

# Redis configuration
REDIS_URL=redis://localhost:6379
REDIS_DB=0
REDIS_TIMEOUT=5
```

### Rate Limit Strategies

1. **PER_USER** - Limits per authenticated user
2. **PER_IP** - Limits per client IP address
3. **PER_API_KEY** - Limits per API key
4. **PER_TENANT** - Limits per organization/tenant
5. **GLOBAL** - Global rate limit

### Default Limits by Endpoint

| Endpoint Pattern           | Requests | Window | Strategy   |
| -------------------------- | -------- | ------ | ---------- |
| `/health`                  | 1000     | 60s    | PER_IP     |
| `/api/v1/projects` (list)  | 100      | 60s    | PER_USER   |
| `/api/v1/projects/create`  | 20       | 60s    | PER_USER   |
| `/api/v1/commands/step/*`  | 200      | 60s    | PER_USER   |
| `/api/v1/commands/batch`   | 10       | 60s    | PER_USER   |
| `/api/v1/executions/start` | 5        | 300s   | PER_TENANT |
| Default                    | 100      | 60s    | PER_USER   |

### Rate Limit Headers

Successful responses include rate limit information:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1705750200
X-RateLimit-Strategy: per_user
```

### Rate Limit Exceeded Response (429)

```json
{
  "status": "error",
  "error_type": "rate_limit_exceeded",
  "message": "Rate limit exceeded. Please retry after 45 seconds.",
  "details": [
    {
      "message": "Rate limit exceeded",
      "context": {
        "limit": 100,
        "window_seconds": 60,
        "retry_after": 45
      }
    }
  ],
  "timestamp": "2024-01-20T10:30:00Z",
  "request_id": "req_123456"
}
```

Headers:

```
Retry-After: 45
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1705750245
```

### Burst Allowance

Some endpoints support burst traffic with a multiplier:

```python
# Step commands allow 50% burst
"step_commands": RateLimitConfig(
    requests=200,
    window_seconds=60,
    burst_multiplier=1.5  # Allows 300 requests
)
```

### Redis Fallback

If Redis is unavailable, the system falls back to in-memory rate limiting with the following limitations:

- Rate limits are per-process (not shared across instances)
- Limits reset on process restart
- Less accurate for distributed deployments

## Audit Logging

All authenticated requests are logged for security and compliance:

```json
{
  "timestamp": "2024-01-20T10:30:00Z",
  "user_id": "user_123",
  "tenant_id": "org_456",
  "username": "john.doe",
  "auth_method": "api_key",
  "api_key_id": "key_789",
  "action": "execute_command",
  "resource": "step-interact click",
  "details": {
    "success": true,
    "method": "POST",
    "client_host": "192.168.1.100",
    "user_agent": "Mozilla/5.0..."
  }
}
```

## Security Best Practices

1. **API Key Management**

   - Use strong, randomly generated API keys
   - Rotate keys regularly
   - Never commit keys to version control
   - Use environment variables or secrets management

2. **HTTPS Only**

   - Always use HTTPS in production
   - Disable HTTP endpoints
   - Use proper SSL certificates

3. **Rate Limiting**

   - Enable rate limiting in production
   - Monitor for abuse patterns
   - Adjust limits based on usage patterns
   - Consider IP-based limits for public endpoints

4. **Monitoring**
   - Monitor authentication failures
   - Track rate limit violations
   - Alert on suspicious patterns
   - Regular security audits

## Testing Authentication

### With cURL

```bash
# Successful request
curl -H "X-API-Key: your-api-key" https://api.example.com/api/v1/commands/list

# Failed authentication
curl https://api.example.com/api/v1/commands/list
# Returns 401 Unauthorized

# Rate limit test
for i in {1..110}; do
  curl -H "X-API-Key: your-api-key" https://api.example.com/api/v1/commands/list
done
# After 100 requests, returns 429 Too Many Requests
```

### With Python

```python
import requests

# Configure session with API key
session = requests.Session()
session.headers.update({"X-API-Key": "your-api-key"})

# Make authenticated request
response = session.get("https://api.example.com/api/v1/commands/list")

# Check rate limit headers
print(f"Remaining: {response.headers.get('X-RateLimit-Remaining')}")
print(f"Reset at: {response.headers.get('X-RateLimit-Reset')}")
```

## Troubleshooting

### Common Issues

1. **401 Unauthorized**

   - Check API key is correct
   - Ensure header name is `X-API-Key`
   - Verify key is active in configuration

2. **403 Forbidden**

   - User lacks required permissions
   - Check role assignments
   - Verify endpoint permissions

3. **429 Too Many Requests**

   - Rate limit exceeded
   - Wait for `Retry-After` seconds
   - Consider upgrading limits

4. **Redis Connection Failed**
   - Check Redis is running
   - Verify connection URL
   - Check network/firewall rules
   - System falls back to in-memory limiting

### Debug Mode

Enable debug logging:

```bash
DEBUG=true
LOG_LEVEL=DEBUG
```

This provides detailed authentication and rate limiting information in logs.
