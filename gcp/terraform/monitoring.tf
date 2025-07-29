# Monitoring and observability configuration for Virtuoso API CLI
# Includes Cloud Logging, Monitoring, Alerts, and Uptime Checks

# Log sinks for different purposes
resource "google_logging_project_sink" "security_logs" {
  name        = "virtuoso-security-logs"
  destination = "bigquery.googleapis.com/projects/${var.project_id}/datasets/${google_bigquery_dataset.logs.dataset_id}"

  filter = <<EOF
    severity >= WARNING
    AND (
      resource.type="cloud_run_revision"
      OR resource.type="cloud_function"
      OR protoPayload.authenticationInfo.principalEmail:*
    )
  EOF

  unique_writer_identity = true

  depends_on = [google_project_service.apis["logging.googleapis.com"]]
}

resource "google_logging_project_sink" "api_analytics" {
  name        = "virtuoso-api-analytics"
  destination = "storage.googleapis.com/${google_storage_bucket.app_data.name}/logs"

  filter = <<EOF
    resource.type="cloud_run_revision"
    AND jsonPayload.endpoint =~ "^/api/"
    AND severity="INFO"
  EOF

  unique_writer_identity = true
}

resource "google_logging_project_sink" "error_logs" {
  name        = "virtuoso-error-logs"
  destination = "pubsub.googleapis.com/projects/${var.project_id}/topics/${google_pubsub_topic.system_events.name}"

  filter = <<EOF
    severity >= ERROR
    AND (
      resource.type="cloud_run_revision"
      OR resource.type="cloud_function"
      OR resource.type="cloud_tasks_queue"
    )
  EOF

  unique_writer_identity = true
}

# BigQuery dataset for logs
resource "google_bigquery_dataset" "logs" {
  dataset_id                  = "virtuoso_logs"
  friendly_name               = "Virtuoso Logs"
  description                 = "Log storage for Virtuoso API CLI"
  location                    = var.bigquery_location
  default_table_expiration_ms = 2592000000 # 30 days

  labels = merge(var.common_labels, {
    component = "bigquery"
    purpose   = "logs"
  })

  depends_on = [google_project_service.apis["bigquery.googleapis.com"]]
}

# BigQuery dataset for analytics (optional)
resource "google_bigquery_dataset" "analytics" {
  count = var.enable_bigquery ? 1 : 0

  dataset_id                  = "virtuoso_analytics"
  friendly_name               = "Virtuoso Analytics"
  description                 = "Analytics data for Virtuoso API CLI"
  location                    = var.bigquery_location
  default_table_expiration_ms = 7776000000 # 90 days

  labels = merge(var.common_labels, {
    component = "bigquery"
    purpose   = "analytics"
  })
}

# Grant permissions to log sink service accounts
resource "google_bigquery_dataset_iam_member" "logs_writer" {
  dataset_id = google_bigquery_dataset.logs.dataset_id
  role       = "roles/bigquery.dataEditor"
  member     = google_logging_project_sink.security_logs.writer_identity
}

resource "google_storage_bucket_iam_member" "analytics_writer" {
  bucket = google_storage_bucket.app_data.name
  role   = "roles/storage.objectCreator"
  member = google_logging_project_sink.api_analytics.writer_identity
}

resource "google_pubsub_topic_iam_member" "error_publisher" {
  project = var.project_id
  topic   = google_pubsub_topic.system_events.name
  role    = "roles/pubsub.publisher"
  member  = google_logging_project_sink.error_logs.writer_identity
}

# Monitoring dashboard
resource "google_monitoring_dashboard" "main" {
  dashboard_json = jsonencode({
    displayName = "Virtuoso API CLI Dashboard"
    gridLayout = {
      columns = 12
      widgets = [
        {
          title = "API Request Rate"
          xyChart = {
            dataSets = [{
              timeSeriesQuery = {
                timeSeriesFilter = {
                  filter = "metric.type=\"run.googleapis.com/request_count\" resource.type=\"cloud_run_revision\""
                  aggregation = {
                    alignmentPeriod   = "60s"
                    perSeriesAligner  = "ALIGN_RATE"
                    crossSeriesReducer = "REDUCE_SUM"
                    groupByFields      = ["resource.label.service_name"]
                  }
                }
              }
            }]
          }
        },
        {
          title = "API Latency (p95)"
          xyChart = {
            dataSets = [{
              timeSeriesQuery = {
                timeSeriesFilter = {
                  filter = "metric.type=\"run.googleapis.com/request_latencies\" resource.type=\"cloud_run_revision\""
                  aggregation = {
                    alignmentPeriod    = "60s"
                    perSeriesAligner   = "ALIGN_PERCENTILE_95"
                    crossSeriesReducer = "REDUCE_MEAN"
                    groupByFields      = ["resource.label.service_name"]
                  }
                }
              }
            }]
          }
        },
        {
          title = "Error Rate"
          xyChart = {
            dataSets = [{
              timeSeriesQuery = {
                timeSeriesFilter = {
                  filter = "metric.type=\"logging.googleapis.com/user/error_count\" resource.type=\"cloud_run_revision\""
                  aggregation = {
                    alignmentPeriod   = "60s"
                    perSeriesAligner  = "ALIGN_RATE"
                    crossSeriesReducer = "REDUCE_SUM"
                  }
                }
              }
            }]
          }
        },
        {
          title = "Cloud Run Instances"
          xyChart = {
            dataSets = [{
              timeSeriesQuery = {
                timeSeriesFilter = {
                  filter = "metric.type=\"run.googleapis.com/container/instance_count\" resource.type=\"cloud_run_revision\""
                  aggregation = {
                    alignmentPeriod   = "60s"
                    perSeriesAligner  = "ALIGN_MEAN"
                    crossSeriesReducer = "REDUCE_SUM"
                    groupByFields      = ["resource.label.service_name"]
                  }
                }
              }
            }]
          }
        },
        {
          title = "Redis Memory Usage"
          xyChart = {
            dataSets = [{
              timeSeriesQuery = {
                timeSeriesFilter = {
                  filter = "metric.type=\"redis.googleapis.com/stats/memory/usage_ratio\" resource.type=\"redis_instance\""
                  aggregation = {
                    alignmentPeriod   = "60s"
                    perSeriesAligner  = "ALIGN_MEAN"
                  }
                }
              }
            }]
          }
        },
        {
          title = "Firestore Operations"
          xyChart = {
            dataSets = [{
              timeSeriesQuery = {
                timeSeriesFilter = {
                  filter = "metric.type=\"firestore.googleapis.com/document/read_count\" resource.type=\"firestore_database\""
                  aggregation = {
                    alignmentPeriod   = "60s"
                    perSeriesAligner  = "ALIGN_RATE"
                    crossSeriesReducer = "REDUCE_SUM"
                  }
                }
              }
            }]
          }
        }
      ]
    }
  })

  depends_on = [google_project_service.apis["monitoring.googleapis.com"]]
}

# Uptime checks
resource "google_monitoring_uptime_check_config" "api_health" {
  display_name = "Virtuoso API Health Check"
  timeout      = "10s"
  period       = "60s"

  http_check {
    path           = "/health"
    port           = "443"
    use_ssl        = true
    validate_ssl   = true
    request_method = "GET"

    accepted_response_status_codes {
      status_class = "STATUS_CLASS_2XX"
    }
  }

  monitored_resource {
    type = "uptime_url"
    labels = {
      host       = var.api_domains[0]
      project_id = var.project_id
    }
  }

  lifecycle {
    create_before_destroy = true
  }
}

# Alert policies
resource "google_monitoring_alert_policy" "high_error_rate" {
  display_name = "Virtuoso - High Error Rate"
  combiner     = "OR"

  conditions {
    display_name = "Error rate > 5%"

    condition_threshold {
      filter          = "metric.type=\"run.googleapis.com/request_count\" resource.type=\"cloud_run_revision\" metric.label.\"response_code_class\"=\"5xx\""
      duration        = "300s"
      comparison      = "COMPARISON_GT"
      threshold_value = 0.05

      aggregations {
        alignment_period     = "60s"
        per_series_aligner   = "ALIGN_RATE"
        cross_series_reducer = "REDUCE_SUM"
        group_by_fields      = ["resource.label.service_name"]
      }
    }
  }

  notification_channels = var.notification_channels

  alert_strategy {
    auto_close = "1800s"
  }

  documentation {
    content = "The error rate for Virtuoso API has exceeded 5%. Check the logs for details."
  }
}

resource "google_monitoring_alert_policy" "high_latency" {
  display_name = "Virtuoso - High Latency"
  combiner     = "OR"

  conditions {
    display_name = "95th percentile latency > 5s"

    condition_threshold {
      filter          = "metric.type=\"run.googleapis.com/request_latencies\" resource.type=\"cloud_run_revision\""
      duration        = "300s"
      comparison      = "COMPARISON_GT"
      threshold_value = 5000 # milliseconds

      aggregations {
        alignment_period     = "60s"
        per_series_aligner   = "ALIGN_PERCENTILE_95"
        cross_series_reducer = "REDUCE_MEAN"
        group_by_fields      = ["resource.label.service_name"]
      }
    }
  }

  notification_channels = var.notification_channels

  alert_strategy {
    auto_close = "1800s"
  }
}

resource "google_monitoring_alert_policy" "redis_memory" {
  display_name = "Virtuoso - Redis High Memory Usage"
  combiner     = "OR"

  conditions {
    display_name = "Redis memory usage > 90%"

    condition_threshold {
      filter          = "metric.type=\"redis.googleapis.com/stats/memory/usage_ratio\" resource.type=\"redis_instance\""
      duration        = "300s"
      comparison      = "COMPARISON_GT"
      threshold_value = 0.9

      aggregations {
        alignment_period   = "60s"
        per_series_aligner = "ALIGN_MEAN"
      }
    }
  }

  notification_channels = var.notification_channels
}

resource "google_monitoring_alert_policy" "uptime_check_failure" {
  display_name = "Virtuoso - Uptime Check Failure"
  combiner     = "OR"

  conditions {
    display_name = "Uptime check failed"

    condition_threshold {
      filter          = "metric.type=\"monitoring.googleapis.com/uptime_check/check_passed\" resource.type=\"uptime_url\" metric.label.\"check_id\"=\"${google_monitoring_uptime_check_config.api_health.uptime_check_id}\""
      duration        = "300s"
      comparison      = "COMPARISON_LT"
      threshold_value = 1

      aggregations {
        alignment_period     = "60s"
        per_series_aligner   = "ALIGN_FRACTION_TRUE"
        cross_series_reducer = "REDUCE_MEAN"
      }
    }
  }

  notification_channels = var.notification_channels
}

# Custom log-based metrics
resource "google_logging_metric" "command_execution_count" {
  name   = "virtuoso_command_execution_count"
  filter = <<EOF
    resource.type="cloud_run_revision"
    AND jsonPayload.event_type="command.executed"
  EOF

  metric_descriptor {
    metric_kind = "DELTA"
    value_type  = "INT64"
    unit        = "1"

    labels {
      key         = "command_type"
      value_type  = "STRING"
      description = "Type of command executed"
    }
  }

  label_extractors = {
    "command_type" = "EXTRACT(jsonPayload.command_type)"
  }
}

resource "google_logging_metric" "api_response_time" {
  name   = "virtuoso_api_response_time"
  filter = <<EOF
    resource.type="cloud_run_revision"
    AND jsonPayload.duration_ms=~".+"
  EOF

  metric_descriptor {
    metric_kind = "GAUGE"
    value_type  = "DISTRIBUTION"
    unit        = "ms"
  }

  value_extractor = "EXTRACT(jsonPayload.duration_ms)"
}

# Notification channels (placeholder - configure with actual values)
resource "google_monitoring_notification_channel" "email" {
  count = length(var.alert_email_addresses) > 0 ? 1 : 0

  display_name = "Virtuoso Email Alerts"
  type         = "email"

  labels = {
    email_address = var.alert_email_addresses[0]
  }
}

resource "google_monitoring_notification_channel" "slack" {
  count = var.slack_webhook_url != "" ? 1 : 0

  display_name = "Virtuoso Slack Alerts"
  type         = "slack"

  labels = {
    channel_name = var.slack_channel_name
    url          = var.slack_webhook_url
  }

  sensitive_labels {
    auth_token = var.slack_auth_token
  }
}

# SLO (Service Level Objective) configuration
resource "google_monitoring_slo" "api_availability" {
  count = var.enable_slos ? 1 : 0

  service      = google_monitoring_service.api[0].service_id
  display_name = "API Availability SLO"

  goal                = 0.999 # 99.9% availability
  rolling_period_days = 30

  request_based_sli {
    good_total_ratio {
      good_service_filter = <<EOF
        metric.type="run.googleapis.com/request_count"
        AND resource.type="cloud_run_revision"
        AND (metric.label.response_code_class="2xx" OR metric.label.response_code_class="3xx")
      EOF

      total_service_filter = <<EOF
        metric.type="run.googleapis.com/request_count"
        AND resource.type="cloud_run_revision"
      EOF
    }
  }
}

# Service for SLO
resource "google_monitoring_service" "api" {
  count = var.enable_slos ? 1 : 0

  service_id   = "virtuoso-api"
  display_name = "Virtuoso API Service"

  basic_service {
    service_type = "CLOUD_RUN"
    service_labels = {
      service_name = google_cloud_run_v2_service.api.name
      location     = var.region
    }
  }
}

# Outputs
output "dashboard_url" {
  value       = "https://console.cloud.google.com/monitoring/dashboards/custom/${google_monitoring_dashboard.main.id}?project=${var.project_id}"
  description = "URL to the monitoring dashboard"
}

output "log_sink_destinations" {
  value = {
    security_logs  = google_logging_project_sink.security_logs.destination
    api_analytics  = google_logging_project_sink.api_analytics.destination
    error_logs     = google_logging_project_sink.error_logs.destination
  }
  description = "Log sink destinations"
}

output "alert_policies" {
  value = {
    high_error_rate = google_monitoring_alert_policy.high_error_rate.name
    high_latency    = google_monitoring_alert_policy.high_latency.name
    redis_memory    = google_monitoring_alert_policy.redis_memory.name
    uptime_check    = google_monitoring_alert_policy.uptime_check_failure.name
  }
  description = "Alert policy names"
}
