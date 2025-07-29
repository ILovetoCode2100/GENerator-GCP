# Virtuoso API CLI - Render Deployment Guide

This guide provides comprehensive instructions for deploying the Virtuoso API CLI to [Render](https://render.com), a modern cloud platform that makes deployment simple and scalable.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Initial Setup](#initial-setup)
3. [Environment Variables](#environment-variables)
4. [Deployment](#deployment)
5. [Custom Domain](#custom-domain)
6. [Monitoring & Logs](#monitoring--logs)
7. [Scaling & Performance](#scaling--performance)
8. [Cost Optimization](#cost-optimization)
9. [Troubleshooting](#troubleshooting)
10. [Maintenance](#maintenance)

## Prerequisites

Before deploying to Render, ensure you have:

- A [Render account](https://render.com/sign-up)
- Your Virtuoso API credentials:
  - API Auth Token
  - Organization ID
  - Base URL (usually `https://api-app2.virtuoso.qa/api`)
- Git repository with the Virtuoso API CLI code
- (Optional) Render CLI installed: `brew install render`

## Initial Setup

### 1. Connect Your GitHub Repository

1. Log in to your [Render Dashboard](https://dashboard.render.com)
2. Click "New +" and select "Web Service"
3. Connect your GitHub account if not already connected
4. Select your `virtuoso-GENerator` repository
5. Configure the following settings:

   ```yaml
   Name: virtuoso-api-cli
   Region: Oregon (US West) # or closest to your location
   Branch: main
   Root Directory: . # leave empty for root
   Environment: Docker
   Build Command: # leave empty - using Dockerfile
   Start Command: # leave empty - using Dockerfile
   ```

### 2. Docker Configuration

Render will automatically detect and use the Dockerfile. Ensure your Dockerfile includes:

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o bin/api-cli cmd/api-cli/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/bin/api-cli .
COPY --from=builder /app/deployment/render/virtuoso-config.yaml.template /root/.api-cli/virtuoso-config.yaml
CMD ["./api-cli", "serve"]
```

### 3. Service Configuration

In the Render dashboard, configure:

- **Instance Type**: Standard (for production) or Free (for testing)
- **Auto-Deploy**: Yes (automatically deploy on git push)
- **Health Check Path**: `/health`

## Environment Variables

### Setting Environment Variables

1. In your Render service dashboard, go to "Environment"
2. Add the following environment variables:

```bash
# Required - Virtuoso API Configuration
VIRTUOSO_API_TOKEN=your-api-key-here
VIRTUOSO_ORG_ID=2242
VIRTUOSO_BASE_URL=https://api-app2.virtuoso.qa/api

# Optional - Service Configuration
PORT=8080
LOG_LEVEL=info
ENABLE_CORS=true
CORS_ORIGINS=https://yourdomain.com
REQUEST_TIMEOUT=30s
MAX_REQUEST_SIZE=10MB

# Optional - Performance
WORKER_POOL_SIZE=10
CACHE_TTL=300
ENABLE_RATE_LIMITING=true
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m

# Optional - Monitoring
ENABLE_METRICS=true
METRICS_PORT=9090
TRACE_ENABLED=false
TRACE_ENDPOINT=https://your-tracing-service.com
```

### Secret Management

For sensitive values:

1. Use Render's encrypted environment variables
2. Never commit secrets to your repository
3. Rotate API tokens regularly
4. Use Render's secret files for complex configurations

## Deployment

### Automatic Deployment (Recommended)

1. Push to your connected branch:

   ```bash
   git add .
   git commit -m "Deploy to Render"
   git push origin main
   ```

2. Render will automatically:
   - Detect the push
   - Build your Docker image
   - Deploy the new version
   - Run health checks
   - Switch traffic to the new version

### Manual Deployment

Using Render CLI:

```bash
# Install Render CLI if not already installed
brew install render

# Deploy
./deployment/render/deploy.sh

# Or manually
render deploy --service virtuoso-api-cli
```

### Deployment Verification

After deployment:

1. Check the deployment logs in Render dashboard
2. Verify the health endpoint:
   ```bash
   curl https://virtuoso-api-cli.onrender.com/health
   ```
3. Run the health check script:
   ```bash
   ./deployment/render/health-check.sh
   ```

## Custom Domain

### Setting Up a Custom Domain

1. In your Render service dashboard, go to "Settings"
2. Under "Custom Domains", click "Add Custom Domain"
3. Enter your domain: `api.yourdomain.com`
4. Add the provided DNS records to your domain provider:

   ```
   Type: CNAME
   Name: api
   Value: virtuoso-api-cli.onrender.com
   ```

5. Wait for DNS propagation (usually 10-30 minutes)
6. Render will automatically provision an SSL certificate

### SSL/TLS Configuration

- Render provides free SSL certificates via Let's Encrypt
- Automatic renewal every 90 days
- Force HTTPS redirect is enabled by default
- No additional configuration needed

## Monitoring & Logs

### Viewing Logs

1. **Dashboard**: Navigate to your service → "Logs" tab
2. **CLI**:

   ```bash
   render logs --service virtuoso-api-cli --tail
   ```

3. **Log Streaming**:
   ```bash
   # Stream logs in real-time
   render logs --service virtuoso-api-cli --tail --follow
   ```

### Setting Up Alerts

1. Go to service settings → "Notifications"
2. Configure alerts for:
   - Deploy failures
   - Service downtime
   - High error rates
   - Resource usage

### External Monitoring

Integrate with monitoring services:

```yaml
# DataDog
DATADOG_API_KEY=your-datadog-key
DATADOG_APP_KEY=your-app-key

# New Relic
NEW_RELIC_LICENSE_KEY=your-license-key
NEW_RELIC_APP_NAME=virtuoso-api-cli

# Sentry
SENTRY_DSN=https://your-sentry-dsn
```

### Metrics Endpoint

If metrics are enabled, access them at:

```
https://virtuoso-api-cli.onrender.com/metrics
```

## Scaling & Performance

### Horizontal Scaling

1. Go to service settings → "Scaling"
2. Configure:
   - **Min Instances**: 1 (minimum running instances)
   - **Max Instances**: 10 (maximum for auto-scaling)
   - **Target CPU**: 70% (scale up when CPU exceeds this)
   - **Target Memory**: 70% (scale up when memory exceeds this)

### Vertical Scaling

Upgrade instance types as needed:

- **Free**: 512 MB RAM, 0.1 CPU (development only)
- **Starter**: 512 MB RAM, 0.5 CPU ($7/month)
- **Standard**: 2 GB RAM, 1 CPU ($25/month)
- **Pro**: 4 GB RAM, 2 CPU ($85/month)
- **Pro Plus**: 8 GB RAM, 4 CPU ($175/month)

### Performance Optimization

1. **Enable Caching**:

   ```bash
   ENABLE_CACHE=true
   CACHE_TTL=300
   REDIS_URL=redis://your-redis-instance
   ```

2. **Configure CDN** (for static assets):

   - Enable Render's built-in CDN
   - Or integrate with Cloudflare

3. **Database Connection Pooling**:
   ```bash
   DB_POOL_SIZE=20
   DB_MAX_IDLE_CONNS=5
   DB_CONN_MAX_LIFETIME=1h
   ```

## Cost Optimization

### Free Tier Limitations

- Spins down after 15 minutes of inactivity
- Limited to 750 hours/month
- No custom domains
- Limited build minutes

### Cost-Saving Strategies

1. **Use Free Tier for Development**:

   ```yaml
   # render.yaml
   services:
     - type: web
       name: virtuoso-api-cli-dev
       env: docker
       plan: free
   ```

2. **Optimize Build Times**:

   - Use multi-stage Docker builds
   - Cache dependencies
   - Minimize image size

3. **Resource Monitoring**:

   - Set up usage alerts
   - Review metrics regularly
   - Scale down during off-hours

4. **Scheduled Scaling**:

   ```bash
   # Scale down at night
   0 20 * * * render scale --service virtuoso-api-cli --min-instances 1

   # Scale up in morning
   0 8 * * * render scale --service virtuoso-api-cli --min-instances 3
   ```

## Troubleshooting

### Common Issues and Solutions

#### 1. Build Failures

**Problem**: Docker build fails

```
Error: failed to build: exit status 1
```

**Solution**:

- Check Dockerfile syntax
- Ensure all dependencies are specified
- Review build logs for specific errors
- Test build locally: `docker build .`

#### 2. Environment Variable Issues

**Problem**: Missing configuration

```
Error: VIRTUOSO_API_TOKEN not set
```

**Solution**:

- Verify all required environment variables are set
- Check for typos in variable names
- Ensure no trailing spaces in values
- Restart service after adding variables

#### 3. Health Check Failures

**Problem**: Service marked as unhealthy

```
Health check failed: connection refused
```

**Solution**:

- Verify health endpoint is implemented
- Check PORT environment variable
- Ensure service starts on correct port
- Review application logs

#### 4. Memory Issues

**Problem**: Out of memory errors

```
Error: container killed due to memory limit
```

**Solution**:

- Upgrade to larger instance type
- Optimize memory usage in code
- Enable swap if available
- Implement memory profiling

#### 5. Timeout Errors

**Problem**: Request timeout

```
Error: request timeout after 30s
```

**Solution**:

- Increase timeout settings
- Optimize slow endpoints
- Implement caching
- Use background jobs for long operations

### Debug Mode

Enable debug logging:

```bash
# Environment variables
DEBUG=true
LOG_LEVEL=debug
VERBOSE_ERRORS=true
```

### Support Resources

- [Render Documentation](https://render.com/docs)
- [Render Community](https://community.render.com)
- [Status Page](https://status.render.com)
- Support ticket via dashboard

## Maintenance

### Regular Tasks

1. **Weekly**:

   - Review error logs
   - Check resource usage
   - Monitor costs

2. **Monthly**:

   - Update dependencies
   - Review and rotate API keys
   - Performance analysis
   - Cost optimization review

3. **Quarterly**:
   - Security audit
   - Disaster recovery test
   - Scale testing

### Backup and Recovery

1. **Configuration Backup**:

   ```bash
   # Export environment variables
   render env --service virtuoso-api-cli > env-backup.txt
   ```

2. **Rollback Procedures**:
   - Render keeps previous deployments
   - Roll back via dashboard or CLI:
   ```bash
   render rollback --service virtuoso-api-cli
   ```

### Zero-Downtime Updates

1. Enable "Zero Downtime Deploys" in service settings
2. Ensure proper health checks
3. Use rolling deployments
4. Test in staging first

### Monitoring Checklist

- [ ] Health endpoint responding
- [ ] Error rate < 1%
- [ ] Response time < 500ms (p95)
- [ ] CPU usage < 80%
- [ ] Memory usage < 80%
- [ ] No critical errors in logs

## Security Best Practices

1. **API Security**:

   - Use environment variables for secrets
   - Enable rate limiting
   - Implement request validation
   - Use HTTPS only

2. **Access Control**:

   - Restrict Render dashboard access
   - Use service accounts for CI/CD
   - Enable 2FA on Render account

3. **Network Security**:
   - Configure allowed IP ranges if needed
   - Use private services for internal APIs
   - Enable DDoS protection

## Conclusion

Deploying the Virtuoso API CLI on Render provides a robust, scalable solution with minimal operational overhead. Follow this guide for initial setup, then customize based on your specific needs. Regular monitoring and maintenance ensure optimal performance and reliability.

For additional help, consult the Render documentation or contact their support team.
