#!/bin/bash

echo "Deploying GCP API with real CLI execution support..."
echo ""

# First, ensure we have the CLI binary built
if [ ! -f "bin/api-cli" ]; then
    echo "Building CLI binary..."
    make build
fi

# Copy necessary files to api directory
echo "Preparing deployment files..."
cp bin/api-cli api/
cp -r config api/ 2>/dev/null || true

# Create updated Dockerfile
cat > api/Dockerfile <<'EOF'
FROM python:3.11-slim

# Install system dependencies
RUN apt-get update && apt-get install -y \
    curl \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy Python requirements and install
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY . .

# Copy CLI binary
COPY api-cli /usr/local/bin/api-cli
RUN chmod +x /usr/local/bin/api-cli

# Create config directory and copy config
RUN mkdir -p /root/.api-cli
COPY config/virtuoso-config.yaml /root/.api-cli/virtuoso-config.yaml || true

# Set environment variables
ENV CLI_PATH=/usr/local/bin/api-cli
ENV CLI_CONFIG_PATH=/root/.api-cli/virtuoso-config.yaml
ENV PYTHONUNBUFFERED=1
ENV PORT=8080

# Expose port
EXPOSE 8080

# Run the application
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8080"]
EOF

# Deploy to Cloud Run
echo "Deploying to Google Cloud Run..."
cd api

gcloud run deploy virtuoso-api \
  --source . \
  --region=us-central1 \
  --platform=managed \
  --allow-unauthenticated \
  --set-env-vars="AUTH_ENABLED=false,SKIP_AUTH=true,VIRTUOSO_API_KEY=f7a55516-5cc4-4529-b2ae-8e106a7d164e,CLI_PATH=/usr/local/bin/api-cli" \
  --timeout=600 \
  --memory=2Gi \
  --cpu=2 \
  --min-instances=1 \
  --max-instances=10

echo ""
echo "Deployment initiated!"
echo ""
echo "Once complete, test with:"
echo "curl -X POST https://virtuoso-api-5e22h3hywa-uc.a.run.app/api/v1/tests/run \\"
echo "  -H 'X-API-Key: 6a54e405ab1277b555f13ccfcd68f32343a21debcb2f7fe12ce845ca8dfd5e2d' \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -d '{\"definition\": {\"name\": \"Test\", \"steps\": [{\"action\": \"navigate\", \"url\": \"https://example.com\"}], \"config\": {\"project_name\": \"Real Project Name\"}}}'"
