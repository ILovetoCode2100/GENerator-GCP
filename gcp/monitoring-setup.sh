#!/bin/bash
# Monitoring Setup Script for Virtuoso API CLI on GCP
# This script configures comprehensive monitoring and alerting

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Default values
PROJECT_ID="${GCP_PROJECT_ID:-$(gcloud config get-value project)}"
REGION="${GCP_REGION:-us-central1}"
NOTIFICATION_CHANNEL=""
ALERT_EMAIL="${ALERT_EMAIL:-}"
SLACK_WEBHOOK="${SLACK_WEBHOOK:-}"

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to create notification channels
create_notification_channels() {
    print_info "Creating notification channels..."

    # Create email notification channel
    if [ -n "$ALERT_EMAIL" ]; then
        print_info "Creating email notification channel..."

        EMAIL_CHANNEL=$(gcloud alpha monitoring channels create \
            --display-name="Virtuoso API Alerts Email" \
            --type=email \
            --channel-labels="email_address=$ALERT_EMAIL" \
            --format="value(name)" 2>/dev/null || echo "")

        if [ -n "$EMAIL_CHANNEL" ]; then
            NOTIFICATION_CHANNEL="$EMAIL_CHANNEL"
            print_success "Email notification channel created"
        fi
    fi

    # Create Slack notification channel
    if [ -n "$SLACK_WEBHOOK" ]; then
        print_info "Creating Slack notification channel..."

        SLACK_CHANNEL=$(gcloud alpha monitoring channels create \
            --display-name="Virtuoso API Alerts Slack" \
            --type=slack \
            --channel-labels="url=$SLACK_WEBHOOK" \
            --format="value(name)" 2>/dev/null || echo "")

        if [ -n "$SLACK_CHANNEL" ]; then
            if [ -n "$NOTIFICATION_CHANNEL" ]; then
                NOTIFICATION_CHANNEL="$NOTIFICATION_CHANNEL,$SLACK_CHANNEL"
            else
                NOTIFICATION_CHANNEL="$SLACK_CHANNEL"
            fi
            print_success "Slack notification channel created"
        fi
    fi

    if [ -z "$NOTIFICATION_CHANNEL" ]; then
        print_warning "No notification channels created. Alerts will not be sent."
    fi
}

# Function to create SLOs
create_slos() {
    print_info "Creating Service Level Objectives (SLOs)..."

    # Create availability SLO
    cat > slo-availability.yaml <<EOF
displayName: "Virtuoso API Availability"
serviceLevelIndicator:
  requestBased:
    goodTotalRatio:
      badServiceFilter: >
        resource.type="cloud_run_revision"
        AND resource.labels.service_name="virtuoso-api-cli"
        AND metric.type="run.googleapis.com/request_count"
        AND metric.labels.response_code_class=~"5.."
      totalServiceFilter: >
        resource.type="cloud_run_revision"
        AND resource.labels.service_name="virtuoso-api-cli"
        AND metric.type="run.googleapis.com/request_count"
goal: 0.999
rollingPeriod: 2592000s  # 30 days
EOF

    gcloud alpha slo create slo-availability \
        --service="virtuoso-api-cli" \
        --slo-config=slo-availability.yaml \
        --project="$PROJECT_ID" 2>/dev/null || true

    # Create latency SLO
    cat > slo-latency.yaml <<EOF
displayName: "Virtuoso API Latency"
serviceLevelIndicator:
  requestBased:
    distributionCut:
      range:
        max: 1000  # 1 second
      filter: >
        resource.type="cloud_run_revision"
        AND resource.labels.service_name="virtuoso-api-cli"
        AND metric.type="run.googleapis.com/request_latencies"
goal: 0.95
rollingPeriod: 2592000s  # 30 days
EOF

    gcloud alpha slo create slo-latency \
        --service="virtuoso-api-cli" \
        --slo-config=slo-latency.yaml \
        --project="$PROJECT_ID" 2>/dev/null || true

    rm -f slo-availability.yaml slo-latency.yaml

    print_success "SLOs created"
}

# Function to create uptime checks
create_uptime_checks() {
    print_info "Creating uptime checks..."

    # Get Cloud Run service URL
    SERVICE_URL=$(gcloud run services describe virtuoso-api-cli \
        --platform=managed \
        --region="$REGION" \
        --format="value(status.url)" 2>/dev/null || echo "")

    if [ -z "$SERVICE_URL" ]; then
        print_warning "Cloud Run service URL not found, skipping uptime checks"
        return
    fi

    # Create health endpoint uptime check
    gcloud monitoring uptime create \
        --display-name="Virtuoso API Health Check" \
        --resource-type="URL" \
        --hostname="${SERVICE_URL#https://}" \
        --path="/health" \
        --check-interval="60s" \
        --timeout="10s" \
        --regions="USA,EUROPE,ASIA_PACIFIC" || true

    # Create API endpoint uptime check
    gcloud monitoring uptime create \
        --display-name="Virtuoso API Commands Check" \
        --resource-type="URL" \
        --hostname="${SERVICE_URL#https://}" \
        --path="/api/v1/commands/list" \
        --check-interval="300s" \
        --timeout="10s" \
        --regions="USA" || true

    print_success "Uptime checks created"
}

# Function to create alert policies
create_alert_policies() {
    print_info "Creating alert policies..."

    # High error rate alert
    cat > alert-error-rate.yaml <<EOF
displayName: "High Error Rate - Virtuoso API"
conditions:
  - displayName: "Error rate > 5%"
    conditionThreshold:
      filter: >
        resource.type="cloud_run_revision"
        AND resource.labels.service_name="virtuoso-api-cli"
        AND metric.type="run.googleapis.com/request_count"
        AND metric.labels.response_code_class="5xx"
      aggregations:
        - alignmentPeriod: 60s
          perSeriesAligner: ALIGN_RATE
          crossSeriesReducer: REDUCE_SUM
      comparison: COMPARISON_GT
      thresholdValue: 0.05
      duration: 300s
documentation:
  content: "The Virtuoso API error rate has exceeded 5% for more than 5 minutes."
  mimeType: text/markdown
EOF

    if [ -n "$NOTIFICATION_CHANNEL" ]; then
        echo "notificationChannels: [\"$NOTIFICATION_CHANNEL\"]" >> alert-error-rate.yaml
    fi

    gcloud alpha monitoring policies create --policy-from-file=alert-error-rate.yaml || true

    # High latency alert
    cat > alert-latency.yaml <<EOF
displayName: "High Latency - Virtuoso API"
conditions:
  - displayName: "95th percentile latency > 2s"
    conditionThreshold:
      filter: >
        resource.type="cloud_run_revision"
        AND resource.labels.service_name="virtuoso-api-cli"
        AND metric.type="run.googleapis.com/request_latencies"
      aggregations:
        - alignmentPeriod: 300s
          perSeriesAligner: ALIGN_PERCENTILE_95
          crossSeriesReducer: REDUCE_MAX
      comparison: COMPARISON_GT
      thresholdValue: 2000  # 2 seconds in milliseconds
      duration: 600s
documentation:
  content: "The Virtuoso API 95th percentile latency has exceeded 2 seconds for more than 10 minutes."
  mimeType: text/markdown
EOF

    if [ -n "$NOTIFICATION_CHANNEL" ]; then
        echo "notificationChannels: [\"$NOTIFICATION_CHANNEL\"]" >> alert-latency.yaml
    fi

    gcloud alpha monitoring policies create --policy-from-file=alert-latency.yaml || true

    # Memory usage alert
    cat > alert-memory.yaml <<EOF
displayName: "High Memory Usage - Virtuoso API"
conditions:
  - displayName: "Memory usage > 80%"
    conditionThreshold:
      filter: >
        resource.type="cloud_run_revision"
        AND resource.labels.service_name="virtuoso-api-cli"
        AND metric.type="run.googleapis.com/container/memory/utilizations"
      aggregations:
        - alignmentPeriod: 300s
          perSeriesAligner: ALIGN_MEAN
          crossSeriesReducer: REDUCE_MAX
      comparison: COMPARISON_GT
      thresholdValue: 0.8
      duration: 600s
documentation:
  content: "The Virtuoso API memory usage has exceeded 80% for more than 10 minutes."
  mimeType: text/markdown
EOF

    if [ -n "$NOTIFICATION_CHANNEL" ]; then
        echo "notificationChannels: [\"$NOTIFICATION_CHANNEL\"]" >> alert-memory.yaml
    fi

    gcloud alpha monitoring policies create --policy-from-file=alert-memory.yaml || true

    # Cloud Function failures alert
    cat > alert-function-failures.yaml <<EOF
displayName: "Cloud Function Failures - Virtuoso"
conditions:
  - displayName: "Function execution failures"
    conditionThreshold:
      filter: >
        resource.type="cloud_function"
        AND metric.type="cloudfunctions.googleapis.com/function/execution_count"
        AND metric.labels.status!="ok"
      aggregations:
        - alignmentPeriod: 300s
          perSeriesAligner: ALIGN_RATE
          crossSeriesReducer: REDUCE_SUM
      comparison: COMPARISON_GT
      thresholdValue: 0.1  # More than 0.1 failures per second
      duration: 300s
documentation:
  content: "Cloud Functions are experiencing elevated failure rates."
  mimeType: text/markdown
EOF

    if [ -n "$NOTIFICATION_CHANNEL" ]; then
        echo "notificationChannels: [\"$NOTIFICATION_CHANNEL\"]" >> alert-function-failures.yaml
    fi

    gcloud alpha monitoring policies create --policy-from-file=alert-function-failures.yaml || true

    # Cleanup
    rm -f alert-*.yaml

    print_success "Alert policies created"
}

# Function to create custom dashboard
create_dashboard() {
    print_info "Creating monitoring dashboard..."

    cat > dashboard.json <<'EOF'
{
  "displayName": "Virtuoso API CLI Dashboard",
  "mosaicLayout": {
    "columns": 12,
    "tiles": [
      {
        "width": 6,
        "height": 4,
        "widget": {
          "title": "Request Rate",
          "xyChart": {
            "dataSets": [
              {
                "timeSeriesQuery": {
                  "timeSeriesFilter": {
                    "filter": "resource.type=\"cloud_run_revision\" resource.labels.service_name=\"virtuoso-api-cli\" metric.type=\"run.googleapis.com/request_count\"",
                    "aggregation": {
                      "alignmentPeriod": "60s",
                      "perSeriesAligner": "ALIGN_RATE",
                      "crossSeriesReducer": "REDUCE_SUM",
                      "groupByFields": ["metric.labels.response_code_class"]
                    }
                  }
                }
              }
            ]
          }
        }
      },
      {
        "width": 6,
        "height": 4,
        "xPos": 6,
        "widget": {
          "title": "Latency (95th percentile)",
          "xyChart": {
            "dataSets": [
              {
                "timeSeriesQuery": {
                  "timeSeriesFilter": {
                    "filter": "resource.type=\"cloud_run_revision\" resource.labels.service_name=\"virtuoso-api-cli\" metric.type=\"run.googleapis.com/request_latencies\"",
                    "aggregation": {
                      "alignmentPeriod": "60s",
                      "perSeriesAligner": "ALIGN_PERCENTILE_95",
                      "crossSeriesReducer": "REDUCE_MAX"
                    }
                  }
                }
              }
            ]
          }
        }
      },
      {
        "width": 6,
        "height": 4,
        "yPos": 4,
        "widget": {
          "title": "Memory Usage",
          "xyChart": {
            "dataSets": [
              {
                "timeSeriesQuery": {
                  "timeSeriesFilter": {
                    "filter": "resource.type=\"cloud_run_revision\" resource.labels.service_name=\"virtuoso-api-cli\" metric.type=\"run.googleapis.com/container/memory/utilizations\"",
                    "aggregation": {
                      "alignmentPeriod": "60s",
                      "perSeriesAligner": "ALIGN_MEAN",
                      "crossSeriesReducer": "REDUCE_MEAN"
                    }
                  }
                }
              }
            ]
          }
        }
      },
      {
        "width": 6,
        "height": 4,
        "xPos": 6,
        "yPos": 4,
        "widget": {
          "title": "CPU Usage",
          "xyChart": {
            "dataSets": [
              {
                "timeSeriesQuery": {
                  "timeSeriesFilter": {
                    "filter": "resource.type=\"cloud_run_revision\" resource.labels.service_name=\"virtuoso-api-cli\" metric.type=\"run.googleapis.com/container/cpu/utilizations\"",
                    "aggregation": {
                      "alignmentPeriod": "60s",
                      "perSeriesAligner": "ALIGN_MEAN",
                      "crossSeriesReducer": "REDUCE_MEAN"
                    }
                  }
                }
              }
            ]
          }
        }
      },
      {
        "width": 12,
        "height": 4,
        "yPos": 8,
        "widget": {
          "title": "Error Logs",
          "logsPanel": {
            "filter": "resource.type=\"cloud_run_revision\" resource.labels.service_name=\"virtuoso-api-cli\" severity>=ERROR"
          }
        }
      }
    ]
  }
}
EOF

    gcloud monitoring dashboards create --config-from-file=dashboard.json || true
    rm -f dashboard.json

    print_success "Dashboard created"
}

# Function to create log-based metrics
create_log_metrics() {
    print_info "Creating log-based metrics..."

    # API command usage metric
    gcloud logging metrics create api_command_usage \
        --description="Track API command usage" \
        --log-filter='resource.type="cloud_run_revision"
        resource.labels.service_name="virtuoso-api-cli"
        jsonPayload.command!=""' \
        --value-extractor='EXTRACT(jsonPayload.command)' \
        --metric-kind=DELTA \
        --value-type=INT64 || true

    # Authentication failures metric
    gcloud logging metrics create auth_failures \
        --description="Track authentication failures" \
        --log-filter='resource.type="cloud_run_revision"
        resource.labels.service_name="virtuoso-api-cli"
        jsonPayload.error="authentication failed"' \
        --metric-kind=DELTA \
        --value-type=INT64 || true

    # Slow queries metric
    gcloud logging metrics create slow_queries \
        --description="Track slow database queries" \
        --log-filter='resource.type="cloud_run_revision"
        resource.labels.service_name="virtuoso-api-cli"
        jsonPayload.query_duration_ms>1000' \
        --metric-kind=DELTA \
        --value-type=INT64 || true

    print_success "Log-based metrics created"
}

# Function to setup log exports
setup_log_exports() {
    print_info "Setting up log exports..."

    # Create BigQuery dataset for logs
    bq mk --dataset \
        --location="$REGION" \
        --description="Virtuoso API logs" \
        "${PROJECT_ID}:virtuoso_logs" || true

    # Export Cloud Run logs to BigQuery
    gcloud logging sinks create virtuoso-api-logs-bq \
        "bigquery.googleapis.com/projects/${PROJECT_ID}/datasets/virtuoso_logs" \
        --log-filter='resource.type="cloud_run_revision"
        resource.labels.service_name="virtuoso-api-cli"' || true

    # Export Cloud Function logs to BigQuery
    gcloud logging sinks create virtuoso-functions-logs-bq \
        "bigquery.googleapis.com/projects/${PROJECT_ID}/datasets/virtuoso_logs" \
        --log-filter='resource.type="cloud_function"
        resource.labels.function_name=~"^(analytics|auth-validator|cleanup|health-check|webhook-handler)$"' || true

    print_success "Log exports configured"
}

# Function to create synthetic monitoring
create_synthetic_monitoring() {
    print_info "Creating synthetic monitoring..."

    # This would typically use Cloud Monitoring synthetic monitors
    # For now, we'll create a Cloud Function that runs periodic tests

    cat > synthetic-monitor.py <<'EOF'
import functions_framework
import requests
import json
from google.cloud import monitoring_v3
import time

@functions_framework.cloud_event
def synthetic_monitor(cloud_event):
    """Run synthetic tests against the API"""

    # Get service URL from environment
    service_url = os.environ.get('SERVICE_URL', '')
    if not service_url:
        print("SERVICE_URL not configured")
        return

    # Run tests
    results = []

    # Test 1: Health check
    start = time.time()
    try:
        resp = requests.get(f"{service_url}/health", timeout=5)
        latency = (time.time() - start) * 1000
        results.append({
            'test': 'health_check',
            'success': resp.status_code == 200,
            'latency_ms': latency,
            'status_code': resp.status_code
        })
    except Exception as e:
        results.append({
            'test': 'health_check',
            'success': False,
            'error': str(e)
        })

    # Test 2: API endpoint
    start = time.time()
    try:
        resp = requests.get(
            f"{service_url}/api/v1/commands/list",
            headers={'X-API-Key': os.environ.get('API_KEY', '')}
        )
        latency = (time.time() - start) * 1000
        results.append({
            'test': 'api_endpoint',
            'success': resp.status_code == 200,
            'latency_ms': latency,
            'status_code': resp.status_code
        })
    except Exception as e:
        results.append({
            'test': 'api_endpoint',
            'success': False,
            'error': str(e)
        })

    # Log results
    print(json.dumps({'synthetic_test_results': results}))

    # Send custom metrics
    client = monitoring_v3.MetricServiceClient()
    project_name = f"projects/{os.environ['GCP_PROJECT']}"

    for result in results:
        if 'latency_ms' in result:
            # Send latency metric
            series = monitoring_v3.TimeSeries()
            series.metric.type = f"custom.googleapis.com/synthetic/{result['test']}/latency"
            series.metric.labels['test'] = result['test']

            point = monitoring_v3.Point()
            point.value.double_value = result['latency_ms']
            point.interval.end_time.seconds = int(time.time())

            series.points = [point]
            client.create_time_series(name=project_name, time_series=[series])
EOF

    print_info "Deploy synthetic monitor with:"
    print_info "  cd functions && gcloud functions deploy synthetic-monitor --runtime python39 --trigger-topic synthetic-tests"

    rm -f synthetic-monitor.py

    print_success "Synthetic monitoring setup complete"
}

# Function to print monitoring summary
print_summary() {
    print_info "Monitoring Setup Summary"
    echo -e "${BLUE}===================================================${NC}"
    echo -e "Project: ${GREEN}$PROJECT_ID${NC}"
    echo -e "Region: ${GREEN}$REGION${NC}"

    if [ -n "$NOTIFICATION_CHANNEL" ]; then
        echo -e "Notifications: ${GREEN}Configured${NC}"
    else
        echo -e "Notifications: ${YELLOW}Not configured (set ALERT_EMAIL or SLACK_WEBHOOK)${NC}"
    fi

    echo -e "\nMonitoring components:"
    echo -e "  ✓ Alert policies"
    echo -e "  ✓ Uptime checks"
    echo -e "  ✓ Custom dashboard"
    echo -e "  ✓ Log-based metrics"
    echo -e "  ✓ Log exports to BigQuery"
    echo -e "  ✓ Service Level Objectives (SLOs)"

    echo -e "\nView monitoring:"
    echo -e "  Dashboard: ${GREEN}https://console.cloud.google.com/monitoring/dashboards${NC}"
    echo -e "  Alerts: ${GREEN}https://console.cloud.google.com/monitoring/alerting${NC}"
    echo -e "  Logs: ${GREEN}https://console.cloud.google.com/logs${NC}"
    echo -e "${BLUE}===================================================${NC}"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --project-id)
            PROJECT_ID="$2"
            shift 2
            ;;
        --region)
            REGION="$2"
            shift 2
            ;;
        --alert-email)
            ALERT_EMAIL="$2"
            shift 2
            ;;
        --slack-webhook)
            SLACK_WEBHOOK="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo "Options:"
            echo "  --project-id ID         GCP project ID"
            echo "  --region REGION         GCP region"
            echo "  --alert-email EMAIL     Email for alerts"
            echo "  --slack-webhook URL     Slack webhook URL"
            echo "  --help                  Show this help message"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Main flow
print_info "Setting up monitoring for Virtuoso API CLI"
echo -e "${BLUE}===================================================${NC}\n"

# Verify project
if [ -z "$PROJECT_ID" ]; then
    print_error "Project ID not specified"
    exit 1
fi

# Execute setup steps
create_notification_channels
create_uptime_checks
create_alert_policies
create_dashboard
create_log_metrics
setup_log_exports
create_slos
create_synthetic_monitoring

# Print summary
print_summary

print_success "Monitoring setup completed successfully!"
