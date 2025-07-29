# 🔐 Render Credentials Setup Guide

## What Credentials Do You Need?

### 1. **Render Account** (Required)

- Sign up at [render.com](https://render.com)
- Free account is sufficient to start
- No credit card required for free tier

### 2. **Virtuoso API Credentials** (Required)

- Your Virtuoso API token
- Organization ID (default: 2242)
- Get from your Virtuoso account

### 3. **API Keys for Your Service** (Required)

- You generate these yourself
- Used to authenticate requests to your API
- Can create multiple keys for different clients

## 📝 Step-by-Step Setup

### Step 1: Create Render Account

1. **Sign Up**

   ```
   1. Go to https://render.com
   2. Click "Get Started for Free"
   3. Sign up with GitHub, GitLab, or email
   4. Verify your email
   ```

2. **Complete Profile**
   - No credit card needed for free tier
   - Optional: Add payment method for paid features

### Step 2: Get Your Virtuoso API Token

1. **Log into Virtuoso**

   ```
   1. Go to https://app.virtuoso.qa
   2. Navigate to Settings > API
   3. Click "Generate API Token"
   4. Copy the token (you won't see it again!)
   ```

2. **Find Organization ID**
   ```
   Default: 2242
   Or check: Settings > Organization > ID
   ```

### Step 3: Generate API Keys for Your Service

Generate secure API keys for clients to access your service:

```bash
# Generate a secure API key
openssl rand -hex 32

# Example output:
# a7f3d8e9c2b5a1f7e3d9c5b1a7f3d8e9c2b5a1f7e3d9c5b1a7f3d8e9c2b5a1f7

# Generate multiple keys for different clients
echo "Client 1: $(openssl rand -hex 32)"
echo "Client 2: $(openssl rand -hex 32)"
echo "Client 3: $(openssl rand -hex 32)"
```

### Step 4: Connect GitHub Repository

1. **In Render Dashboard**
   ```
   1. Click "New +"
   2. Select "Blueprint"
   3. Connect your GitHub account
   4. Select your repository
   5. Render detects render.yaml automatically
   ```

### Step 5: Set Environment Variables

In Render Dashboard > Your Service > Environment:

```bash
# Required Secrets
VIRTUOSO_API_TOKEN=<your-virtuoso-api-token>
API_KEYS=["key1_from_step3", "key2_from_step3"]

# Optional Configuration
RATE_LIMIT_PER_MINUTE=60
LOG_LEVEL=INFO
CORS_ALLOWED_ORIGINS=https://yourdomain.com
```

## 🔑 Environment Variables Reference

### Required Variables

| Variable             | Description                             | How to Get                           |
| -------------------- | --------------------------------------- | ------------------------------------ |
| `VIRTUOSO_API_TOKEN` | Your Virtuoso API authentication token  | Virtuoso Dashboard > Settings > API  |
| `API_KEYS`           | JSON array of API keys for your clients | Generate with `openssl rand -hex 32` |

### Optional Variables

| Variable                | Default        | Description                                 |
| ----------------------- | -------------- | ------------------------------------------- |
| `VIRTUOSO_ORG_ID`       | 2242           | Your Virtuoso organization ID               |
| `RATE_LIMIT_PER_MINUTE` | 60             | API rate limit per client                   |
| `LOG_LEVEL`             | INFO           | Logging level (DEBUG, INFO, WARNING, ERROR) |
| `CORS_ALLOWED_ORIGINS`  | \*             | Allowed CORS origins                        |
| `JWT_SECRET`            | auto-generated | Secret for JWT tokens                       |
| `ENCRYPTION_KEY`        | auto-generated | Key for encrypting sensitive data           |

## 🚀 Quick Setup Script

```bash
#!/bin/bash
# save as setup-render-env.sh

echo "🔐 Render Environment Setup"
echo "=========================="

# Generate API keys
echo ""
echo "📝 Generating API Keys..."
API_KEY_1=$(openssl rand -hex 32)
API_KEY_2=$(openssl rand -hex 32)
API_KEY_3=$(openssl rand -hex 32)

echo ""
echo "Your API Keys (save these!):"
echo "Client 1: $API_KEY_1"
echo "Client 2: $API_KEY_2"
echo "Client 3: $API_KEY_3"

echo ""
echo "📋 Environment Variables for Render:"
echo "===================================="
echo ""
echo "VIRTUOSO_API_TOKEN=<paste-your-virtuoso-token-here>"
echo "API_KEYS=[\"$API_KEY_1\", \"$API_KEY_2\", \"$API_KEY_3\"]"
echo ""
echo "Copy and paste these into Render Dashboard > Environment"
```

## 🔒 Security Best Practices

### 1. **API Token Security**

- Never commit tokens to Git
- Use Render's secret management
- Rotate tokens periodically
- Use different tokens for different environments

### 2. **API Key Management**

```python
# Example: Store API keys with metadata
API_KEYS = [
    {
        "key": "a7f3d8e9...",
        "name": "Frontend App",
        "created": "2024-01-01",
        "permissions": ["read", "write"]
    },
    {
        "key": "b8e4c9f0...",
        "name": "Mobile App",
        "created": "2024-01-01",
        "permissions": ["read"]
    }
]
```

### 3. **Environment Isolation**

```bash
# Development
VIRTUOSO_API_TOKEN_DEV=dev-token
API_KEYS_DEV=["dev-key-1", "dev-key-2"]

# Production
VIRTUOSO_API_TOKEN=prod-token
API_KEYS=["prod-key-1", "prod-key-2"]
```

## 🆘 Troubleshooting

### "Invalid API Token"

```bash
# Verify token in Virtuoso
curl -H "Authorization: Bearer YOUR_TOKEN" \
  https://api-app2.virtuoso.qa/api/projects
```

### "API Key Not Recognized"

```bash
# Check format - must be JSON array
# ✅ Correct
API_KEYS=["key1", "key2"]

# ❌ Wrong
API_KEYS=key1,key2
API_KEYS="key1, key2"
```

### "Environment Variable Not Set"

1. Go to Render Dashboard
2. Select your service
3. Click "Environment"
4. Add missing variable
5. Save (triggers redeploy)

## 📊 Verify Your Setup

After deployment, test your credentials:

```bash
# 1. Check health (no auth required)
curl https://your-app.onrender.com/health

# 2. Test API key
curl -H "X-API-Key: your-generated-key" \
  https://your-app.onrender.com/api/v1/commands

# 3. Check Virtuoso connection
curl -X POST https://your-app.onrender.com/api/v1/commands/list/projects \
  -H "X-API-Key: your-generated-key" \
  -H "Content-Type: application/json"
```

## 🎯 Next Steps

1. **Set up monitoring alerts** in Render Dashboard
2. **Configure custom domain** if needed
3. **Enable auto-deploy** from GitHub
4. **Set up different environments** (dev, staging, prod)

## 📚 Additional Resources

- [Render Environment Variables](https://render.com/docs/environment-variables)
- [Render Secrets Management](https://render.com/docs/configure-environment-variables#secret-files)
- [Virtuoso API Documentation](https://docs.virtuoso.qa/api)

That's it! You now have all the credentials needed to deploy to Render. 🚀
