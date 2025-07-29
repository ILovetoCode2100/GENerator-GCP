---
id: authentication
title: Authentication
sidebar_position: 2
---

# API Authentication

The Virtuoso Test Converter API uses API keys to authenticate requests. This guide covers how to obtain, use, and manage your API keys securely.

## Obtaining API Keys

### Via Dashboard

1. Log in to your [Virtuoso Dashboard](https://app.virtuoso.qa)
2. Navigate to **Settings** → **API Keys**
3. Click **Generate New Key**
4. Give your key a descriptive name (e.g., "Production CI/CD")
5. Select the appropriate permissions
6. Click **Create**

### Via Support

For enterprise customers or special access:

- Email: api-support@virtuoso.qa
- Include your organization ID and use case

## Using API Keys

### Header Authentication (Recommended)

Include your API key in the `X-API-Key` header:

```bash
curl -X GET https://api.virtuoso.qa/v1/patterns \
  -H "X-API-Key: vrt_live_abc123xyz789"
```

### Query Parameter Authentication

For tools that don't support custom headers:

```bash
curl "https://api.virtuoso.qa/v1/patterns?api_key=vrt_live_abc123xyz789"
```

⚠️ **Warning**: Query parameter authentication exposes your API key in logs and browser history. Use only when header authentication is not possible.

## API Key Format

API keys follow a consistent format:

```
vrt_[environment]_[unique_identifier]
```

- `vrt` - Virtuoso prefix
- `environment` - `live` or `test`
- `unique_identifier` - Random alphanumeric string

Example: `vrt_live_sk_1234567890abcdef`

## Key Permissions

API keys can have different permission levels:

### Read-Only

- View patterns
- Check job status
- List conversions
- Access documentation

### Write

- All read permissions
- Convert tests
- Submit feedback
- Create sessions

### Admin

- All write permissions
- Manage webhooks
- Access usage statistics
- Configure team settings

## Security Best Practices

### 1. Environment Variables

Store API keys in environment variables:

```bash
export VIRTUOSO_API_KEY="vrt_live_abc123xyz789"
```

Use in your application:

```javascript
const apiKey = process.env.VIRTUOSO_API_KEY;
```

### 2. Key Rotation

Rotate your API keys regularly:

```bash
# Revoke old key
curl -X DELETE https://api.virtuoso.qa/v1/auth/keys/vrt_live_old123 \
  -H "X-API-Key: vrt_live_admin_key"

# Use new key
export VIRTUOSO_API_KEY="vrt_live_new456"
```

### 3. Restrict Key Scope

Create separate keys for different environments:

- `vrt_live_prod_*` - Production environment
- `vrt_live_staging_*` - Staging environment
- `vrt_test_dev_*` - Development environment

### 4. IP Whitelisting

For production keys, enable IP whitelisting:

```json
{
  "key_id": "vrt_live_prod_123",
  "allowed_ips": ["192.168.1.100", "10.0.0.0/24"]
}
```

### 5. Monitor Usage

Track API key usage:

```bash
curl -X GET https://api.virtuoso.qa/v1/auth/keys/vrt_live_123/usage \
  -H "X-API-Key: vrt_live_admin_key"
```

## OAuth 2.0 (Enterprise)

Enterprise customers can use OAuth 2.0 for enhanced security:

### Authorization Code Flow

1. **Redirect to authorize**:

```
https://auth.virtuoso.qa/oauth/authorize?
  client_id=YOUR_CLIENT_ID&
  redirect_uri=YOUR_REDIRECT_URI&
  response_type=code&
  scope=read write
```

2. **Exchange code for token**:

```bash
curl -X POST https://auth.virtuoso.qa/oauth/token \
  -d "grant_type=authorization_code" \
  -d "code=AUTH_CODE" \
  -d "client_id=YOUR_CLIENT_ID" \
  -d "client_secret=YOUR_CLIENT_SECRET"
```

3. **Use access token**:

```bash
curl -X GET https://api.virtuoso.qa/v1/patterns \
  -H "Authorization: Bearer ACCESS_TOKEN"
```

### Client Credentials Flow

For machine-to-machine authentication:

```bash
curl -X POST https://auth.virtuoso.qa/oauth/token \
  -d "grant_type=client_credentials" \
  -d "client_id=YOUR_CLIENT_ID" \
  -d "client_secret=YOUR_CLIENT_SECRET" \
  -d "scope=read write"
```

## Service Accounts

For CI/CD and automation, use service accounts:

```json
{
  "name": "ci-service-account",
  "type": "service",
  "permissions": ["convert", "execute", "read"],
  "expires_at": null
}
```

## Multi-Factor Authentication

Enable MFA for accounts with API access:

1. Enable MFA in account settings
2. Generate app-specific API keys
3. Use time-based tokens for sensitive operations

## Error Responses

### Invalid API Key

```json
{
  "error": {
    "code": "INVALID_API_KEY",
    "message": "The provided API key is invalid or has been revoked"
  }
}
```

### Expired API Key

```json
{
  "error": {
    "code": "EXPIRED_API_KEY",
    "message": "The API key has expired",
    "expired_at": "2024-01-01T00:00:00Z"
  }
}
```

### Insufficient Permissions

```json
{
  "error": {
    "code": "INSUFFICIENT_PERMISSIONS",
    "message": "This API key does not have permission to perform this action",
    "required_permission": "write"
  }
}
```

## Testing Authentication

Verify your API key is working:

```bash
curl -X GET https://api.virtuoso.qa/v1/auth/verify \
  -H "X-API-Key: YOUR_API_KEY"
```

Success response:

```json
{
  "valid": true,
  "key_id": "vrt_live_abc123",
  "permissions": ["read", "write"],
  "organization_id": "org_123",
  "expires_at": null
}
```

## Troubleshooting

### Common Issues

1. **"Invalid API Key" error**

   - Verify key is copied correctly
   - Check for extra spaces or characters
   - Ensure key hasn't been revoked

2. **"Forbidden" error**

   - Check key has required permissions
   - Verify IP is whitelisted (if enabled)
   - Ensure organization is active

3. **"Rate limited" error**
   - Check current usage limits
   - Implement exponential backoff
   - Consider upgrading plan

### Debug Headers

Include debug headers for troubleshooting:

```bash
curl -X GET https://api.virtuoso.qa/v1/patterns \
  -H "X-API-Key: YOUR_API_KEY" \
  -H "X-Debug: true"
```

## Next Steps

- **[Rate Limiting](./rate-limiting)** - Understand API limits
- **[Error Handling](./error-handling)** - Handle auth errors
- **[Security Best Practices](../architecture/security)** - Secure your integration
