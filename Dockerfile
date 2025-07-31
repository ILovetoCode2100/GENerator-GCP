# Multi-stage optimized Dockerfile for complete API deployment
# Stage 1: Go builder for CLI binary
FROM golang:1.23-alpine as go-builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Copy Go source code
COPY go.mod go.sum ./
COPY cmd/ ./cmd/
COPY pkg/ ./pkg/

# Download dependencies
RUN go mod download

# Build the CLI binary for Linux
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o api-cli ./cmd/api-cli

# Stage 2: Python dependencies builder
FROM python:3.11-slim as python-builder

# Install build dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    g++ \
    && rm -rf /var/lib/apt/lists/*

# Create virtual environment
RUN python -m venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"

# Copy and install Python dependencies
COPY api/requirements.txt /tmp/requirements.txt
RUN pip install --no-cache-dir --upgrade pip && \
    pip install --no-cache-dir -r /tmp/requirements.txt

# Stage 3: Final runtime image
FROM python:3.11-slim

# Install runtime dependencies only
RUN apt-get update && apt-get install -y \
    ca-certificates \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Copy virtual environment from builder
COPY --from=python-builder /opt/venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"

# Create app directory and set as working directory
WORKDIR /app

# Copy the CLI binary from Go builder and make it executable
COPY --from=go-builder /build/api-cli /bin/api-cli
RUN chmod +x /bin/api-cli

# Copy API code maintaining module structure
COPY api/app /app/app
COPY api/requirements.txt /app/requirements.txt
COPY api/debug_startup.py /app/debug_startup.py

# Create necessary directories
RUN mkdir -p /root/.api-cli /var/log/api

# Set environment variables
ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    PYTHONPATH=/app:$PYTHONPATH \
    HOST=0.0.0.0 \
    PORT=8080 \
    CLI_PATH=/bin/api-cli \
    API_PATH=/api \
    LOG_DIR=/var/log/api \
    CONFIG_DIR=/root/.api-cli

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Expose port
EXPOSE 8080

# Run the FastAPI app
CMD ["python", "-m", "uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8080"]
