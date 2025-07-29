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

# Create config directory
RUN mkdir -p /root/.api-cli
# Config will be provided via environment variables

# Set environment variables
ENV CLI_PATH=/usr/local/bin/api-cli
ENV CLI_CONFIG_PATH=/root/.api-cli/virtuoso-config.yaml
ENV PYTHONUNBUFFERED=1
ENV PORT=8080

# Expose port
EXPOSE 8080

# Run the application
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8080"]
