# Getting Started with Virtuoso API Deployment

This guide will help you test the FastAPI wrapper locally and prepare for production deployment.

## Prerequisites

- Docker and Docker Compose installed
- Go 1.21+ (for building the CLI)
- Python 3.11+ (for local API development)
- Redis (via Docker)
- Your Virtuoso API credentials

## Step 1: Build the CLI Binary

```bash
# Build the CLI binary first
make build

# Verify it works
./bin/api-cli --help
```

## Step 2: Set Up Configuration

1. Create your Virtuoso configuration:

```bash
mkdir -p ~/.api-cli
cat > ~/.api-cli/virtuoso-config.yaml << EOF
api:
  auth_token: your-virtuoso-api-token-here
  base_url: https://api-app2.virtuoso.qa/api
organization:
  id: "2242"
EOF
```

2. Create API environment file:

```bash
cp api/.env.example api/.env
# Edit api/.env with your settings
```

## Step 3: Start Services Locally

```bash
# Start all services with Docker Compose
docker-compose up -d

# Check services are running
docker-compose ps

# View logs
docker-compose logs -f api
```

## Step 4: Test the API

1. **Check Health**:

```bash
curl http://localhost:8000/health
```

2. **View API Documentation**:
   Open http://localhost:8000/docs in your browser

3. **Create an API Key** (first time):

```bash
# Generate a secure API key
export API_KEY=$(openssl rand -hex 32)
echo "Your API Key: $API_KEY"

# Note: In production, store this in your auth service
```

4. **Test Command Execution**:

```bash
# List all commands
curl -X GET http://localhost:8000/api/v1/commands \
  -H "X-API-Key: $API_KEY"

# Execute a simple command
curl -X POST http://localhost:8000/api/v1/commands/step-navigate/to \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "checkpoint_id": "12345",
    "url": "https://example.com",
    "position": 1
  }'
```

5. **Run a Test from YAML**:

```bash
# Create a test file
cat > test.yaml << 'EOF'
name: "Quick Test"
steps:
  - navigate: "https://example.com"
  - assert: "Example Domain"
  - click: "More information..."
EOF

# Run the test
curl -X POST http://localhost:8000/api/v1/tests/run \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d "{\"content\": \"$(cat test.yaml | sed 's/"/\\"/g' | tr '\n' '\\n')\"}"
```

## Step 5: Development Workflow

### For API Development:

```bash
# Install Python dependencies
cd api
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
pip install -r requirements.txt

# Run FastAPI directly (with hot reload)
uvicorn app.main:app --reload --host 0.0.0.0 --port 8000
```

### For CLI Development:

```bash
# Make changes to Go code
# Rebuild
make build

# Test changes
./bin/api-cli step-assert exists "test" --dry-run
```

## Step 6: Running Tests

```bash
# API unit tests
cd api
pytest tests/

# CLI tests
make test

# Integration tests
./test-scripts/test-all-69-commands.sh
```

## Step 7: Production Deployment

### Using Kubernetes:

```bash
# Build and push images
docker build -t your-registry/virtuoso-cli:latest .
docker build -t your-registry/virtuoso-api:latest -f api/Dockerfile.api .
docker push your-registry/virtuoso-cli:latest
docker push your-registry/virtuoso-api:latest

# Deploy with kubectl
cd deployment/kubernetes
./scripts/deploy.sh production
```

### Using Helm:

```bash
cd deployment/helm
./install.sh --environment production \
  --api-token "your-token" \
  --org-id "2242" \
  --image-tag "latest" \
  install
```

## Troubleshooting

### API Won't Start

- Check Redis is running: `docker-compose ps redis`
- Verify CLI binary exists: `ls -la bin/api-cli`
- Check logs: `docker-compose logs api`

### Authentication Errors

- Ensure X-API-Key header is included
- Check the API key is valid
- Verify rate limits haven't been exceeded

### Command Execution Fails

- Verify Virtuoso credentials in config
- Check the checkpoint ID exists
- Review API logs for detailed errors

### Performance Issues

- Monitor Redis memory: `docker-compose exec redis redis-cli info memory`
- Check API metrics: `curl http://localhost:8000/health/metrics`
- Review resource limits in docker-compose.yml

## Next Steps

1. **Secure Your Deployment**:

   - Set up proper API key management
   - Configure SSL/TLS
   - Review security policies

2. **Monitor Your Services**:

   - Set up Prometheus/Grafana
   - Configure alerts
   - Enable audit logging

3. **Scale Your Deployment**:

   - Configure HPA in Kubernetes
   - Set up Redis clustering
   - Implement caching strategies

4. **Integrate with CI/CD**:
   - Set up automated testing
   - Configure deployment pipelines
   - Implement rollback procedures

## Useful Commands

```bash
# View all running containers
docker-compose ps

# Stop all services
docker-compose down

# Clean up everything (including volumes)
docker-compose down -v

# View API logs
docker-compose logs -f api

# Access Redis CLI
docker-compose exec redis redis-cli

# Rebuild after changes
docker-compose build api

# Scale API instances (local testing)
docker-compose up -d --scale api=3
```

## Support Resources

- API Documentation: http://localhost:8000/docs
- Redoc Documentation: http://localhost:8000/redoc
- Health Status: http://localhost:8000/health?detailed=true
- Metrics: http://localhost:8000/health/metrics

Ready to generate tests from requirements? Your Virtuoso API is now running!
