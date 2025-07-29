# Async processing configuration for Virtuoso API CLI
# Includes Pub/Sub, Cloud Tasks, and Cloud Scheduler

# Pub/Sub Topics
resource "google_pubsub_topic" "command_events" {
  name = "virtuoso-command-events"

  message_retention_duration = "604800s" # 7 days

  schema_settings {
    schema   = google_pubsub_schema.event_schema.id
    encoding = "JSON"
  }

  labels = merge(var.common_labels, {
    component = "pubsub"
    purpose   = "command-events"
  })

  depends_on = [google_project_service.apis["pubsub.googleapis.com"]]
}

resource "google_pubsub_topic" "test_events" {
  name = "virtuoso-test-events"

  message_retention_duration = "604800s" # 7 days

  schema_settings {
    schema   = google_pubsub_schema.event_schema.id
    encoding = "JSON"
  }

  labels = merge(var.common_labels, {
    component = "pubsub"
    purpose   = "test-events"
  })
}

resource "google_pubsub_topic" "system_events" {
  name = "virtuoso-system-events"

  message_retention_duration = "604800s" # 7 days

  labels = merge(var.common_labels, {
    component = "pubsub"
    purpose   = "system-events"
  })
}

# Dead letter topic for failed messages
resource "google_pubsub_topic" "dead_letter" {
  name = "virtuoso-dead-letter"

  message_retention_duration = "2592000s" # 30 days

  labels = merge(var.common_labels, {
    component = "pubsub"
    purpose   = "dead-letter"
  })
}

# Pub/Sub Schema for event validation
resource "google_pubsub_schema" "event_schema" {
  name       = "virtuoso-event-schema"
  type       = "AVRO"
  definition = <<EOF
{
  "type": "record",
  "name": "VirtuosoEvent",
  "fields": [
    {
      "name": "eventId",
      "type": "string",
      "doc": "Unique event identifier"
    },
    {
      "name": "eventType",
      "type": {
        "type": "enum",
        "name": "EventType",
        "symbols": [
          "command.created",
          "command.executed",
          "command.failed",
          "test.started",
          "test.completed",
          "test.failed",
          "system.health",
          "system.error"
        ]
      }
    },
    {
      "name": "timestamp",
      "type": "string",
      "doc": "ISO 8601 timestamp"
    },
    {
      "name": "data",
      "type": {
        "type": "map",
        "values": "string"
      }
    },
    {
      "name": "metadata",
      "type": {
        "type": "record",
        "name": "Metadata",
        "fields": [
          {"name": "source", "type": "string"},
          {"name": "version", "type": "string"},
          {"name": "userId", "type": ["null", "string"], "default": null},
          {"name": "sessionId", "type": ["null", "string"], "default": null}
        ]
      }
    }
  ]
}
EOF
}

# Pub/Sub Subscriptions
resource "google_pubsub_subscription" "command_processor" {
  name  = "command-processor-sub"
  topic = google_pubsub_topic.command_events.id

  push_config {
    push_endpoint = "${google_cloud_run_v2_service.api.uri}/webhooks/pubsub/command-processor"

    oidc_token {
      service_account_email = google_service_account.pubsub_invoker.email
    }

    attributes = {
      x-goog-version = "v1"
    }
  }

  ack_deadline_seconds = 60

  retry_policy {
    minimum_backoff = "10s"
    maximum_backoff = "600s"
  }

  dead_letter_policy {
    dead_letter_topic     = google_pubsub_topic.dead_letter.id
    max_delivery_attempts = 5
  }

  expiration_policy {
    ttl = "" # Never expire
  }

  labels = merge(var.common_labels, {
    component = "pubsub"
    purpose   = "command-processor"
  })
}

resource "google_pubsub_subscription" "test_executor" {
  name  = "test-executor-sub"
  topic = google_pubsub_topic.test_events.id

  push_config {
    push_endpoint = "${google_cloud_run_v2_service.api.uri}/webhooks/pubsub/test-executor"

    oidc_token {
      service_account_email = google_service_account.pubsub_invoker.email
    }
  }

  ack_deadline_seconds = 300 # 5 minutes for longer tests

  retry_policy {
    minimum_backoff = "30s"
    maximum_backoff = "600s"
  }

  dead_letter_policy {
    dead_letter_topic     = google_pubsub_topic.dead_letter.id
    max_delivery_attempts = 3
  }

  labels = merge(var.common_labels, {
    component = "pubsub"
    purpose   = "test-executor"
  })
}

# Pull subscription for monitoring/analytics
resource "google_pubsub_subscription" "analytics" {
  name  = "analytics-sub"
  topic = google_pubsub_topic.system_events.id

  # Pull subscription (no push config)
  ack_deadline_seconds = 60

  message_retention_duration = "604800s" # 7 days
  retain_acked_messages      = false

  enable_message_ordering = true

  labels = merge(var.common_labels, {
    component = "pubsub"
    purpose   = "analytics"
  })
}

# Service account for Pub/Sub to invoke services
resource "google_service_account" "pubsub_invoker" {
  account_id   = "virtuoso-pubsub-invoker"
  display_name = "Virtuoso Pub/Sub Invoker"
  description  = "Service account for Pub/Sub to invoke Cloud Run services"
}

resource "google_cloud_run_v2_service_iam_member" "pubsub_invoker" {
  project  = google_cloud_run_v2_service.api.project
  location = google_cloud_run_v2_service.api.location
  name     = google_cloud_run_v2_service.api.name
  role     = "roles/run.invoker"
  member   = "serviceAccount:${google_service_account.pubsub_invoker.email}"
}

# Cloud Tasks Queues
resource "google_cloud_tasks_queue" "command_execution" {
  name     = "command-execution"
  location = var.region

  rate_limits {
    max_dispatches_per_second = 100
    max_concurrent_dispatches = 1000
  }

  retry_config {
    max_attempts = 5
    max_retry_duration = "3600s"
    min_backoff = "10s"
    max_backoff = "3600s"
    max_doublings = 5
  }

  stackdriver_logging_config {
    sampling_ratio = 1.0
  }

  depends_on = [google_project_service.apis["cloudtasks.googleapis.com"]]
}

resource "google_cloud_tasks_queue" "test_execution" {
  name     = "test-execution"
  location = var.region

  rate_limits {
    max_dispatches_per_second = 50
    max_concurrent_dispatches = 500
  }

  retry_config {
    max_attempts = 3
    max_retry_duration = "1800s"
    min_backoff = "30s"
    max_backoff = "1800s"
    max_doublings = 3
  }

  stackdriver_logging_config {
    sampling_ratio = 1.0
  }
}

resource "google_cloud_tasks_queue" "batch_processing" {
  name     = "batch-processing"
  location = var.region

  rate_limits {
    max_dispatches_per_second = 10
    max_concurrent_dispatches = 50
  }

  retry_config {
    max_attempts = 3
    max_retry_duration = "3600s"
    min_backoff = "60s"
    max_backoff = "3600s"
    max_doublings = 2
  }
}

# Cloud Scheduler Jobs
resource "google_cloud_scheduler_job" "health_check" {
  name        = "virtuoso-health-check"
  description = "Regular health check of all services"
  region      = var.region
  schedule    = "*/5 * * * *" # Every 5 minutes
  time_zone   = "UTC"

  http_target {
    uri         = google_cloudfunctions2_function.health_check.service_config[0].uri
    http_method = "GET"

    oidc_token {
      service_account_email = google_service_account.cloud_scheduler.email
    }
  }

  retry_config {
    retry_count          = 1
    max_retry_duration   = "30s"
    min_backoff_duration = "5s"
    max_backoff_duration = "10s"
  }

  depends_on = [
    google_project_service.apis["cloudscheduler.googleapis.com"],
    google_cloudfunctions2_function.health_check
  ]
}

resource "google_cloud_scheduler_job" "cleanup" {
  name        = "virtuoso-cleanup"
  description = "Clean up old data and expired sessions"
  region      = var.region
  schedule    = "0 2 * * *" # Daily at 2 AM UTC
  time_zone   = "UTC"

  http_target {
    uri         = google_cloudfunctions2_function.cleanup.service_config[0].uri
    http_method = "POST"

    oidc_token {
      service_account_email = google_service_account.cloud_scheduler.email
    }

    headers = {
      "Content-Type" = "application/json"
    }

    body = base64encode(jsonencode({
      cleanup_types = ["sessions", "temp_files", "old_logs"]
      retention_days = 30
    }))
  }

  retry_config {
    retry_count = 3
  }
}

resource "google_cloud_scheduler_job" "backup" {
  name        = "virtuoso-backup"
  description = "Backup Firestore data"
  region      = var.region
  schedule    = "0 3 * * *" # Daily at 3 AM UTC
  time_zone   = "UTC"

  http_target {
    uri         = "${google_cloud_run_v2_service.api.uri}/api/v1/admin/backup"
    http_method = "POST"

    oidc_token {
      service_account_email = google_service_account.cloud_scheduler.email
    }

    headers = {
      "Content-Type" = "application/json"
    }

    body = base64encode(jsonencode({
      backup_type = "firestore"
      destination = "gs://${google_storage_bucket.backups.name}/firestore/${timestamp()}"
    }))
  }

  retry_config {
    retry_count = 3
  }
}

resource "google_cloud_scheduler_job" "metrics_aggregation" {
  name        = "virtuoso-metrics-aggregation"
  description = "Aggregate metrics for reporting"
  region      = var.region
  schedule    = "0 * * * *" # Hourly
  time_zone   = "UTC"

  pubsub_target {
    topic_name = google_pubsub_topic.system_events.id

    data = base64encode(jsonencode({
      eventType = "system.metrics.aggregate"
      timestamp = timestamp()
      data = {
        aggregation_period = "hourly"
      }
    }))

    attributes = {
      event_type = "metrics_aggregation"
    }
  }

  retry_config {
    retry_count = 1
  }
}

# Workflow for complex orchestration (optional)
resource "google_workflows_workflow" "test_orchestration" {
  count = var.enable_workflows ? 1 : 0

  name        = "virtuoso-test-orchestration"
  region      = var.region
  description = "Orchestrate complex test scenarios"

  service_account = google_service_account.workflows[0].id

  source_contents = <<EOF
main:
  params: [args]
  steps:
    - init:
        assign:
          - test_id: $${args.test_id}
          - steps: $${args.steps}
          - results: []

    - execute_steps:
        for:
          value: step
          in: $${steps}
          steps:
            - execute_step:
                call: http.post
                args:
                  url: $${sys.get_env("API_URL") + "/api/v1/commands/execute"}
                  auth:
                    type: OIDC
                  body:
                    command: $${step.command}
                    args: $${step.args}
                result: step_result

            - append_result:
                assign:
                  - results: $${list.concat(results, [step_result.body])}

    - return_results:
        return:
          test_id: $${test_id}
          results: $${results}
          status: "completed"
EOF

  labels = merge(var.common_labels, {
    component = "workflows"
    purpose   = "test-orchestration"
  })

  depends_on = [google_project_service.apis["workflows.googleapis.com"]]
}

# Service account for Workflows
resource "google_service_account" "workflows" {
  count = var.enable_workflows ? 1 : 0

  account_id   = "virtuoso-workflows-sa"
  display_name = "Virtuoso Workflows Service Account"
  description  = "Service account for Google Workflows"
}

resource "google_project_iam_member" "workflows_roles" {
  for_each = var.enable_workflows ? toset([
    "roles/run.invoker",
    "roles/logging.logWriter",
  ]) : toset([])

  project = var.project_id
  role    = each.value
  member  = "serviceAccount:${google_service_account.workflows[0].email}"
}

# Outputs
output "pubsub_topics" {
  value = {
    command_events = google_pubsub_topic.command_events.id
    test_events    = google_pubsub_topic.test_events.id
    system_events  = google_pubsub_topic.system_events.id
    dead_letter    = google_pubsub_topic.dead_letter.id
  }
  description = "Pub/Sub topic IDs"
}

output "cloud_tasks_queues" {
  value = {
    command_execution = google_cloud_tasks_queue.command_execution.id
    test_execution    = google_cloud_tasks_queue.test_execution.id
    batch_processing  = google_cloud_tasks_queue.batch_processing.id
  }
  description = "Cloud Tasks queue IDs"
}

output "scheduler_jobs" {
  value = {
    health_check       = google_cloud_scheduler_job.health_check.id
    cleanup            = google_cloud_scheduler_job.cleanup.id
    backup             = google_cloud_scheduler_job.backup.id
    metrics_aggregation = google_cloud_scheduler_job.metrics_aggregation.id
  }
  description = "Cloud Scheduler job IDs"
}
