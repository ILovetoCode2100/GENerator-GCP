#!/bin/bash

# Virtuoso API CLI - Health Check Script
# This script verifies the deployment health and tests all endpoints

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERVICE_URL="${SERVICE_URL:-https://virtuoso-api-cli.onrender.com}"
CUSTOM_DOMAIN="${CUSTOM_DOMAIN:-}"
TIMEOUT="${TIMEOUT:-10}"
VERBOSE="${VERBOSE:-false}"

# Use custom domain if set
if [ -n "$CUSTOM_DOMAIN" ]; then
    SERVICE_URL="https://$CUSTOM_DOMAIN"
fi

# Script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

# Test results
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
WARNINGS=0

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[✓]${NC} $1"
    ((PASSED_TESTS++))
    ((TOTAL_TESTS++))
}

log_warning() {
    echo -e "${YELLOW}[!]${NC} $1"
    ((WARNINGS++))
    ((TOTAL_TESTS++))
}

log_error() {
    echo -e "${RED}[✗]${NC} $1"
    ((FAILED_TESTS++))
    ((TOTAL_TESTS++))
}

verbose_log() {
    if [ "$VERBOSE" = "true" ]; then
        echo "    $1"
    fi
}

# Health check functions
check_endpoint() {
    local endpoint="$1"
    local expected_status="$2"
    local description="$3"

    local url="${SERVICE_URL}${endpoint}"
    verbose_log "Testing: $url"

    local response=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout "$TIMEOUT" "$url" 2>/dev/null || echo "000")

    if [ "$response" = "$expected_status" ]; then
        log_success "$description (HTTP $response)"
        return 0
    elif [ "$response" = "000" ]; then
        log_error "$description - Connection failed"
        return 1
    else
        log_error "$description - Expected HTTP $expected_status, got $response"
        return 1
    fi
}

check_json_endpoint() {
    local endpoint="$1"
    local expected_field="$2"
    local description="$3"

    local url="${SERVICE_URL}${endpoint}"
    verbose_log "Testing JSON endpoint: $url"

    local response=$(curl -s --connect-timeout "$TIMEOUT" "$url" 2>/dev/null)
    local http_code=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout "$TIMEOUT" "$url" 2>/dev/null || echo "000")

    if [ "$http_code" != "200" ]; then
        log_error "$description - HTTP $http_code"
        return 1
    fi

    # Check if response is valid JSON
    if ! echo "$response" | jq . >/dev/null 2>&1; then
        log_error "$description - Invalid JSON response"
        verbose_log "Response: $response"
        return 1
    fi

    # Check for expected field if provided
    if [ -n "$expected_field" ]; then
        if echo "$response" | jq -e ".$expected_field" >/dev/null 2>&1; then
            log_success "$description"
            verbose_log "Response contains field: $expected_field"
        else
            log_error "$description - Missing field: $expected_field"
            return 1
        fi
    else
        log_success "$description"
    fi

    return 0
}

check_response_time() {
    local endpoint="$1"
    local max_time="$2"
    local description="$3"

    local url="${SERVICE_URL}${endpoint}"
    verbose_log "Testing response time: $url"

    local start_time=$(date +%s%N)
    local response=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout "$TIMEOUT" "$url" 2>/dev/null || echo "000")
    local end_time=$(date +%s%N)

    if [ "$response" = "000" ]; then
        log_error "$description - Connection failed"
        return 1
    fi

    local elapsed=$(( (end_time - start_time) / 1000000 )) # Convert to milliseconds

    if [ "$elapsed" -le "$max_time" ]; then
        log_success "$description - ${elapsed}ms (max: ${max_time}ms)"
    else
        log_warning "$description - ${elapsed}ms (exceeded max: ${max_time}ms)"
    fi
}

check_cli_command() {
    local command="$1"
    local description="$2"

    verbose_log "Testing CLI command: $command"

    # Create a test request
    local response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "{\"command\": \"$command\"}" \
        --connect-timeout "$TIMEOUT" \
        "${SERVICE_URL}/api/cli/execute" 2>/dev/null)

    local http_code=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
        -H "Content-Type: application/json" \
        -d "{\"command\": \"$command\"}" \
        --connect-timeout "$TIMEOUT" \
        "${SERVICE_URL}/api/cli/execute" 2>/dev/null || echo "000")

    if [ "$http_code" = "200" ]; then
        log_success "$description"
    elif [ "$http_code" = "401" ]; then
        log_warning "$description - Authentication required"
    elif [ "$http_code" = "000" ]; then
        log_error "$description - Connection failed"
    else
        log_error "$description - HTTP $http_code"
    fi
}

check_ssl_certificate() {
    local domain="$1"

    verbose_log "Checking SSL certificate for: $domain"

    # Extract domain from URL
    domain=$(echo "$SERVICE_URL" | sed -e 's|^[^/]*//||' -e 's|/.*$||')

    # Check certificate expiry
    local cert_info=$(echo | openssl s_client -servername "$domain" -connect "$domain:443" 2>/dev/null | openssl x509 -noout -dates 2>/dev/null)

    if [ -z "$cert_info" ]; then
        log_error "SSL Certificate - Unable to retrieve certificate"
        return 1
    fi

    local not_after=$(echo "$cert_info" | grep "notAfter" | cut -d= -f2)
    local expiry_epoch=$(date -j -f "%b %d %H:%M:%S %Y %Z" "$not_after" +%s 2>/dev/null || date -d "$not_after" +%s 2>/dev/null)
    local current_epoch=$(date +%s)
    local days_until_expiry=$(( (expiry_epoch - current_epoch) / 86400 ))

    if [ "$days_until_expiry" -gt 30 ]; then
        log_success "SSL Certificate - Valid for $days_until_expiry days"
    elif [ "$days_until_expiry" -gt 7 ]; then
        log_warning "SSL Certificate - Expires in $days_until_expiry days"
    else
        log_error "SSL Certificate - Expires in $days_until_expiry days!"
    fi
}

check_headers() {
    local endpoint="$1"
    local header="$2"
    local expected_value="$3"
    local description="$4"

    local url="${SERVICE_URL}${endpoint}"
    verbose_log "Checking headers for: $url"

    local headers=$(curl -s -I --connect-timeout "$TIMEOUT" "$url" 2>/dev/null)

    if [ -z "$headers" ]; then
        log_error "$description - Unable to retrieve headers"
        return 1
    fi

    local header_value=$(echo "$headers" | grep -i "^$header:" | cut -d' ' -f2- | tr -d '\r')

    if [ -n "$expected_value" ]; then
        if [[ "$header_value" == *"$expected_value"* ]]; then
            log_success "$description - $header: $header_value"
        else
            log_error "$description - Expected $header: $expected_value, got: $header_value"
        fi
    else
        if [ -n "$header_value" ]; then
            log_success "$description - $header present: $header_value"
        else
            log_error "$description - $header header missing"
        fi
    fi
}

# Main health checks
run_health_checks() {
    echo "======================================"
    echo "   Virtuoso API CLI Health Check"
    echo "======================================"
    echo "Service URL: $SERVICE_URL"
    echo "Timeout: ${TIMEOUT}s"
    echo ""

    # Basic connectivity
    log_info "Testing basic connectivity..."
    check_endpoint "/" "200" "Root endpoint"
    check_endpoint "/health" "200" "Health endpoint"

    # API endpoints
    log_info ""
    log_info "Testing API endpoints..."
    check_endpoint "/api/version" "200" "Version endpoint"
    check_endpoint "/api/status" "200" "Status endpoint"
    check_json_endpoint "/api/info" "version" "Info endpoint"

    # Response times
    log_info ""
    log_info "Testing response times..."
    check_response_time "/health" "500" "Health check response time"
    check_response_time "/api/version" "1000" "API response time"

    # Security headers
    log_info ""
    log_info "Testing security headers..."
    check_headers "/" "X-Content-Type-Options" "nosniff" "X-Content-Type-Options"
    check_headers "/" "X-Frame-Options" "" "X-Frame-Options"
    check_headers "/" "Strict-Transport-Security" "" "HSTS header"

    # SSL certificate
    log_info ""
    log_info "Testing SSL/TLS..."
    check_ssl_certificate

    # CLI commands (if API key is available)
    if [ -n "$VIRTUOSO_API_TOKEN" ]; then
        log_info ""
        log_info "Testing CLI commands..."
        check_cli_command "list-commands" "List commands"
        check_cli_command "validate-config" "Validate config"
    else
        log_info ""
        log_warning "Skipping CLI command tests (VIRTUOSO_API_TOKEN not set)"
    fi

    # Metrics endpoint (if enabled)
    log_info ""
    log_info "Testing monitoring endpoints..."
    check_endpoint "/metrics" "200" "Metrics endpoint" || log_warning "Metrics endpoint not available (may be disabled)"

    # Load test (optional)
    if [ "$RUN_LOAD_TEST" = "true" ]; then
        log_info ""
        log_info "Running load test..."
        run_load_test
    fi
}

run_load_test() {
    local concurrent_requests=10
    local total_requests=100

    log_info "Sending $total_requests requests with $concurrent_requests concurrent connections..."

    if command -v ab &> /dev/null; then
        ab -n "$total_requests" -c "$concurrent_requests" -s "$TIMEOUT" "${SERVICE_URL}/health" 2>&1 | \
            grep -E "(Requests per second:|Time per request:|Failed requests:)" | \
            while read -r line; do
                verbose_log "$line"
            done
        log_success "Load test completed"
    else
        log_warning "Apache Bench (ab) not installed, skipping load test"
    fi
}

show_summary() {
    echo ""
    echo "======================================"
    echo "   Health Check Summary"
    echo "======================================"
    echo "Total Tests: $TOTAL_TESTS"
    echo -e "${GREEN}Passed: $PASSED_TESTS${NC}"

    if [ "$WARNINGS" -gt 0 ]; then
        echo -e "${YELLOW}Warnings: $WARNINGS${NC}"
    fi

    if [ "$FAILED_TESTS" -gt 0 ]; then
        echo -e "${RED}Failed: $FAILED_TESTS${NC}"
    fi

    echo ""

    if [ "$FAILED_TESTS" -eq 0 ]; then
        if [ "$WARNINGS" -eq 0 ]; then
            echo -e "${GREEN}✓ All health checks passed!${NC}"
            return 0
        else
            echo -e "${YELLOW}⚠ Health checks passed with warnings${NC}"
            return 0
        fi
    else
        echo -e "${RED}✗ Some health checks failed${NC}"
        return 1
    fi
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --url)
            SERVICE_URL="$2"
            shift 2
            ;;
        --timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        --load-test)
            RUN_LOAD_TEST=true
            shift
            ;;
        --help)
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  --url URL          Service URL to test (default: https://virtuoso-api-cli.onrender.com)"
            echo "  --timeout SECONDS  Request timeout in seconds (default: 10)"
            echo "  --verbose, -v      Enable verbose output"
            echo "  --load-test        Run load test"
            echo "  --help             Show this help message"
            echo ""
            echo "Environment variables:"
            echo "  SERVICE_URL        Service URL to test"
            echo "  CUSTOM_DOMAIN      Custom domain to test"
            echo "  VIRTUOSO_API_TOKEN API token for authenticated tests"
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Check dependencies
if ! command -v curl &> /dev/null; then
    log_error "curl is required but not installed"
    exit 1
fi

if ! command -v jq &> /dev/null; then
    log_warning "jq is not installed - some JSON tests will be limited"
fi

# Run health checks
run_health_checks
show_summary

# Exit with appropriate code
if [ "$FAILED_TESTS" -gt 0 ]; then
    exit 1
else
    exit 0
fi
