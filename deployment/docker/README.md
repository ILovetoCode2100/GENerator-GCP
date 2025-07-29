# Virtuoso API CLI Docker Deployment Guide

This directory contains production-ready Docker configurations for the Virtuoso API CLI system.

## Architecture

The deployment consists of the following services:

- **API**: FastAPI service handling HTTP requests
- **CLI**: Go-based command-line interface
- **Redis**: Caching and rate limiting
- **PostgreSQL**: Database for persistent storage (optional)
- **Nginx**: Reverse proxy and load balancer

## Prerequisites

- Docker Engine 20.10+
- Docker Compose 2.0+
- 2GB+ available RAM
- Valid Virtuoso API credentials

## Quick Start

### 1. Clone the repository

```bash
git clone https://github.com/your-org/virtuoso-generator.git
cd virtuoso-generator/deployment/docker
```

### 2. Configure environment

```bash
cp .env.example .env
# Edit .env with your actual values
vim .env
```

### 3. Build images

```bash
# Build all images
docker-compose -f docker-compose.prod.yml build

# Or build specific service
docker-compose -f docker-compose.prod.yml build api
```

### 4. Start services

```bash
# Start all services
docker-compose -f docker-compose.prod.yml up -d

# Check status
docker-compose -f docker-compose.prod.yml ps

# View logs
docker-compose -f docker-compose.prod.yml logs -f api
```

## Development Setup

For local development with hot reload:

```bash
# From project root
docker-compose up -d

# This starts:
# - CLI builder with file watching
# - API with volume mounts
# - Redis
# - PostgreSQL (with --profile with-db)
```

## Production Deployment

### 1. Pre-deployment checklist

- [ ] Update all passwords in `.env`
- [ ] Configure SSL certificates
- [ ] Set appropriate resource limits
- [ ] Configure backup strategy
- [ ] Set up monitoring

### 2. SSL/TLS Configuration

Create SSL certificates directory:

```bash
mkdir -p nginx/ssl
# Copy your certificates
cp /path/to/cert.pem nginx/ssl/
cp /path/to/key.pem nginx/ssl/
```

### 3. Nginx Configuration

Create `nginx/conf.d/default.conf`:

```nginx
upstream api_backend {
    server api:8000;
}

server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;

    ssl_certificate /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    location / {
        proxy_pass http://api_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /health {
        access_log off;
        proxy_pass http://api_backend/health;
    }
}
```

### 4. Deploy to production

```bash
# Deploy with Docker Compose
docker-compose -f docker-compose.prod.yml up -d

# Or use Docker Swarm
docker stack deploy -c docker-compose.prod.yml virtuoso

# Or use Kubernetes (convert with Kompose)
kompose convert -f docker-compose.prod.yml
```

## Scaling

### Horizontal scaling

```bash
# Scale API service
docker-compose -f docker-compose.prod.yml up -d --scale api=3

# With Docker Swarm
docker service scale virtuoso_api=3
```

### Resource limits

Resource limits are pre-configured in `docker-compose.prod.yml`. Adjust based on your needs:

```yaml
deploy:
  resources:
    limits:
      cpus: "2"
      memory: 1G
```

## Maintenance

### Backup

```bash
# Backup PostgreSQL
docker-compose -f docker-compose.prod.yml exec postgres \
  pg_dump -U $POSTGRES_USER $POSTGRES_DB > backup.sql

# Backup Redis
docker-compose -f docker-compose.prod.yml exec redis \
  redis-cli --rdb /data/dump.rdb
```

### Update

```bash
# Pull latest changes
git pull origin main

# Rebuild images
docker-compose -f docker-compose.prod.yml build

# Rolling update
docker-compose -f docker-compose.prod.yml up -d --no-deps api
```

### Monitoring

```bash
# Check health endpoints
curl http://localhost:8000/health

# View metrics (if enabled)
curl http://localhost:9090/metrics

# Check logs
docker-compose -f docker-compose.prod.yml logs --tail=100 -f
```

## Troubleshooting

### Common issues

1. **Port conflicts**

   ```bash
   # Check what's using the port
   lsof -i :8000
   # Change port in .env file
   ```

2. **Permission errors**

   ```bash
   # Fix volume permissions
   docker-compose -f docker-compose.prod.yml exec api chown -R virtuoso:virtuoso /app
   ```

3. **Memory issues**
   ```bash
   # Check memory usage
   docker stats
   # Adjust limits in docker-compose.prod.yml
   ```

### Debug mode

```bash
# Enable debug mode
DEBUG=true docker-compose -f docker-compose.prod.yml up

# Or update .env and restart
```

## Security Best Practices

1. **Never commit `.env` files**
2. **Use secrets management** (Docker Swarm secrets, Kubernetes secrets)
3. **Regular updates** of base images
4. **Network isolation** between services
5. **Rate limiting** configured in API
6. **SSL/TLS** for all external traffic
7. **Non-root users** in containers

## Environment Variables Reference

See `.env.example` for complete list. Key variables:

- `VIRTUOSO_API_KEY`: Your Virtuoso API key (required)
- `VIRTUOSO_ORG_ID`: Your organization ID (required)
- `JWT_SECRET_KEY`: Strong secret for JWT tokens
- `POSTGRES_PASSWORD`: Database password
- `REDIS_PASSWORD`: Redis password

## Support

For issues or questions:

1. Check logs: `docker-compose logs`
2. Review this documentation
3. Check GitHub issues
4. Contact support team

## License

See LICENSE file in project root.
