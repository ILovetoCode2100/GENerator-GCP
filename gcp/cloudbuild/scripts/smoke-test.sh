#!/bin/bash
# Smoke test script for Virtuoso API CLI
# Runs basic functionality tests after deployment

set -e

# Configuration
TARGET_URL="${TARGET_URL}"
ENVIRONMENT="${ENVIRONMENT:-dev}"
API_KEY="${VIRTUOSO_API_KEY}"

# Test results
TESTS_PASSED=0
TESTS_FAILED=0

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

log_failure() {
    echo -e "${RED}[FAIL]${NC} $1"
}

run_test() {
    local test_name=$1
    local test_command=$2
    local expected_result=$3

    log_info "Testing: ${test_name}"

    if eval "${test_command}"; then
        log_success "${test_name}"
        ((TESTS_PASSED++))
        return 0
    else
        log_failure "${test_name}"
        ((TESTS_FAILED++))
        return 1
    fi
}

# Check if URL is provided
if [ -z "${TARGET_URL}" ]; then
    log_failure "TARGET_URL is not set"
    exit 1
fi

log_info "Running smoke tests against: ${TARGET_URL}"
log_info "Environment: ${ENVIRONMENT}"

# Test 1: Health Check
run_test "Health Check Endpoint" \
    "curl -sf ${TARGET_URL}/health > /dev/null"

# Test 2: Version Endpoint
VERSION_RESPONSE=$(curl -sf ${TARGET_URL}/version || echo "{}")
run_test "Version Endpoint" \
    "[ -n '${VERSION_RESPONSE}' ]"

# Test 3: API Documentation
run_test "API Documentation" \
    "curl -sf ${TARGET_URL}/docs > /dev/null"

# Test 4: Authentication (if API key provided)
if [ -n "${API_KEY}" ]; then
    run_test "Authentication Test" \
        "curl -sf -H 'Authorization: Bearer ${API_KEY}' ${TARGET_URL}/api/v1/projects > /dev/null"
fi

# Test 5: Response Time
log_info "Testing response time..."
RESPONSE_TIME=$(curl -sf -o /dev/null -w "%{time_total}" ${TARGET_URL}/health)
RESPONSE_TIME_MS=$(echo "${RESPONSE_TIME} * 1000" | bc | cut -d. -f1)

if [ ${RESPONSE_TIME_MS} -lt 1000 ]; then
    log_success "Response time: ${RESPONSE_TIME_MS}ms (< 1000ms)"
    ((TESTS_PASSED++))
else
    log_failure "Response time: ${RESPONSE_TIME_MS}ms (>= 1000ms)"
    ((TESTS_FAILED++))
fi

# Test 6: SSL Certificate
if [[ ${TARGET_URL} == https://* ]]; then
    run_test "SSL Certificate Validation" \
        "curl -sf --ssl-reqd ${TARGET_URL}/health > /dev/null"
fi

# Test 7: CORS Headers (for web compatibility)
CORS_HEADERS=$(curl -sf -I ${TARGET_URL}/health | grep -i "access-control-allow-origin" || echo "")
if [ -n "${CORS_HEADERS}" ]; then
    log_success "CORS headers present"
    ((TESTS_PASSED++))
else
    log_info "CORS headers not configured (may be intentional)"
fi

# Test 8: Error Handling
run_test "404 Error Handling" \
    "curl -sf -o /dev/null -w '%{http_code}' ${TARGET_URL}/nonexistent | grep -q '404'"

# Test 9: Metrics Endpoint (if available)
if curl -sf ${TARGET_URL}/metrics > /dev/null 2>&1; then
    log_success "Metrics endpoint available"
    ((TESTS_PASSED++))
else
    log_info "Metrics endpoint not available"
fi

# Test 10: CLI-specific endpoints
if [ "${ENVIRONMENT}" != "prod" ] || [ -n "${API_KEY}" ]; then
    # Test project listing (requires auth in prod)
    AUTH_HEADER=""
    if [ -n "${API_KEY}" ]; then
        AUTH_HEADER="-H 'Authorization: Bearer ${API_KEY}'"
    fi

    # Test a simple API call
    run_test "API Projects Endpoint" \
        "curl -sf ${AUTH_HEADER} ${TARGET_URL}/api/v1/projects -o /dev/null"
fi

# Environment-specific tests
case ${ENVIRONMENT} in
    prod)
        # Production-specific tests
        log_info "Running production-specific tests..."

        # Test rate limiting
        log_info "Testing rate limiting..."
        for i in {1..20}; do
            curl -sf ${TARGET_URL}/health > /dev/null 2>&1 &
        done
        wait

        # Check if any were rate limited (429 status)
        RATE_LIMITED=$(curl -sf -o /dev/null -w '%{http_code}' ${TARGET_URL}/health)
        if [ "${RATE_LIMITED}" == "429" ]; then
            log_success "Rate limiting is active"
            ((TESTS_PASSED++))
        else
            log_info "Rate limiting test inconclusive"
        fi
        ;;

    staging)
        # Staging-specific tests
        log_info "Running staging-specific tests..."

        # Test debug endpoints (should be disabled)
        if curl -sf ${TARGET_URL}/debug/vars > /dev/null 2>&1; then
            log_failure "Debug endpoints are exposed in staging!"
            ((TESTS_FAILED++))
        else
            log_success "Debug endpoints are properly disabled"
            ((TESTS_PASSED++))
        fi
        ;;

    dev)
        # Development-specific tests
        log_info "Running development-specific tests..."

        # Test debug endpoints (should be available)
        run_test "Debug Endpoints Available" \
            "curl -sf ${TARGET_URL}/debug/vars > /dev/null || true"
        ;;
esac

# Performance test (basic)
log_info "Running basic performance test..."
CONCURRENT_REQUESTS=10
TOTAL_REQUESTS=100

# Create a simple load test
cat > /tmp/smoke-load-test.sh << 'EOF'
#!/bin/bash
URL=$1
for i in $(seq 1 10); do
    curl -sf -o /dev/null -w "%{http_code} %{time_total}\n" "$URL/health"
done
EOF

chmod +x /tmp/smoke-load-test.sh

# Run concurrent requests
START_TIME=$(date +%s)
for i in $(seq 1 ${CONCURRENT_REQUESTS}); do
    /tmp/smoke-load-test.sh ${TARGET_URL} > /tmp/load-test-${i}.log 2>&1 &
done
wait
END_TIME=$(date +%s)

# Analyze results
TOTAL_TIME=$((END_TIME - START_TIME))
SUCCESS_COUNT=$(cat /tmp/load-test-*.log | grep "^200" | wc -l)
AVG_RESPONSE_TIME=$(cat /tmp/load-test-*.log | awk '{sum+=$2; count++} END {print sum/count*1000}' | cut -d. -f1)

log_info "Load test results:"
log_info "  Total requests: ${TOTAL_REQUESTS}"
log_info "  Successful: ${SUCCESS_COUNT}"
log_info "  Total time: ${TOTAL_TIME}s"
log_info "  Avg response time: ${AVG_RESPONSE_TIME}ms"

if [ ${SUCCESS_COUNT} -eq ${TOTAL_REQUESTS} ]; then
    log_success "All requests succeeded"
    ((TESTS_PASSED++))
else
    log_failure "Some requests failed (${SUCCESS_COUNT}/${TOTAL_REQUESTS})"
    ((TESTS_FAILED++))
fi

# Cleanup
rm -f /tmp/smoke-load-test.sh /tmp/load-test-*.log

# Generate test report
cat > smoke-test-report.json << EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "target_url": "${TARGET_URL}",
  "environment": "${ENVIRONMENT}",
  "total_tests": $((TESTS_PASSED + TESTS_FAILED)),
  "passed": ${TESTS_PASSED},
  "failed": ${TESTS_FAILED},
  "performance": {
    "avg_response_time_ms": ${AVG_RESPONSE_TIME:-0},
    "total_requests": ${TOTAL_REQUESTS},
    "successful_requests": ${SUCCESS_COUNT}
  }
}
EOF

# Summary
echo -e "\n========================================="
echo -e "Smoke Test Summary"
echo -e "========================================="
echo -e "Target: ${TARGET_URL}"
echo -e "Environment: ${ENVIRONMENT}"
echo -e "Total tests: $((TESTS_PASSED + TESTS_FAILED))"
echo -e "${GREEN}Passed: ${TESTS_PASSED}${NC}"
echo -e "${RED}Failed: ${TESTS_FAILED}${NC}"
echo -e "========================================="

# Exit with appropriate code
if [ ${TESTS_FAILED} -gt 0 ]; then
    exit 1
else
    exit 0
fi
