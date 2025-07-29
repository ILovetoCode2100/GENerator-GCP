#!/bin/sh
# Simple health check script for containers

set -e

# Function to check service health
check_service() {
    service_name=$1
    check_command=$2

    echo -n "Checking $service_name... "
    if eval "$check_command" > /dev/null 2>&1; then
        echo "✓ Healthy"
        return 0
    else
        echo "✗ Unhealthy"
        return 1
    fi
}

# Check each service
failed=0

# API Service
check_service "API" "curl -f http://api:8000/health" || failed=$((failed + 1))

# Redis
check_service "Redis" "redis-cli -h redis ping" || failed=$((failed + 1))

# PostgreSQL (if enabled)
if [ -n "$POSTGRES_ENABLED" ]; then
    check_service "PostgreSQL" "pg_isready -h postgres -U $POSTGRES_USER" || failed=$((failed + 1))
fi

# CLI (check if binary exists and is executable)
check_service "CLI" "test -x /app/bin/api-cli" || failed=$((failed + 1))

# Summary
echo ""
if [ $failed -eq 0 ]; then
    echo "All services are healthy!"
    exit 0
else
    echo "$failed service(s) failed health check"
    exit 1
fi
