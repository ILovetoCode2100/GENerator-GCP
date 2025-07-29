# Cloud Functions configuration for Virtuoso API CLI
# Includes health check, webhook handler, and cleanup functions

# Cloud Functions (2nd gen) for lightweight operations

# Health Check Function
resource "google_cloudfunctions2_function" "health_check" {
  name        = "virtuoso-health-check"
  location    = var.region
  description = "Comprehensive health check for all services"

  build_config {
    runtime     = "python311"
    entry_point = "health_check"

    source {
      storage_source {
        bucket = google_storage_bucket.functions_source.name
        object = google_storage_bucket_object.health_check_source.name
      }
    }
  }

  service_config {
    max_instance_count = 10
    min_instance_count = 0
    available_memory   = "256M"
    available_cpu      = "0.166"
    timeout_seconds    = 30

    environment_variables = {
      PROJECT_ID        = var.project_id
      CLOUD_RUN_URL     = google_cloud_run_v2_service.api.uri
      FIRESTORE_PROJECT = var.project_id
      REDIS_HOST        = google_redis_instance.cache.host
      REDIS_PORT        = google_redis_instance.cache.port
    }

    secret_environment_variables {
      key        = "REDIS_AUTH"
      project_id = var.project_id
      secret     = google_secret_manager_secret.additional["redis-auth"].secret_id
      version    = "latest"
    }

    service_account_email = google_service_account.cloud_functions.email

    ingress_settings               = "ALLOW_ALL"
    all_traffic_on_latest_revision = true
  }

  labels = merge(var.common_labels, {
    component = "functions"
    purpose   = "health-check"
  })

  depends_on = [
    google_project_service.apis["cloudfunctions.googleapis.com"],
    google_storage_bucket_object.health_check_source,
  ]
}

# Webhook Handler Function
resource "google_cloudfunctions2_function" "webhook_handler" {
  name        = "virtuoso-webhook-handler"
  location    = var.region
  description = "Handle incoming webhooks from Virtuoso"

  build_config {
    runtime     = "python311"
    entry_point = "webhook_handler"

    source {
      storage_source {
        bucket = google_storage_bucket.functions_source.name
        object = google_storage_bucket_object.webhook_handler_source.name
      }
    }
  }

  service_config {
    max_instance_count = 100
    min_instance_count = 0
    available_memory   = "512M"
    available_cpu      = "1"
    timeout_seconds    = 60

    environment_variables = {
      PROJECT_ID          = var.project_id
      PUBSUB_TOPIC_EVENTS = google_pubsub_topic.command_events.name
    }

    secret_environment_variables {
      key        = "WEBHOOK_SECRET"
      project_id = var.project_id
      secret     = google_secret_manager_secret.additional["webhook-secret"].secret_id
      version    = "latest"
    }

    service_account_email = google_service_account.cloud_functions.email

    ingress_settings               = "ALLOW_ALL"
    all_traffic_on_latest_revision = true
  }

  event_trigger {
    trigger_region = var.region
    event_type     = "google.cloud.firestore.document.v1.written"

    event_filters {
      attribute = "database"
      value     = "(default)"
    }

    event_filters {
      attribute = "document"
      value     = "webhooks/{webhook}"
    }

    retry_policy = "RETRY_POLICY_RETRY"
  }

  labels = merge(var.common_labels, {
    component = "functions"
    purpose   = "webhook-handler"
  })
}

# Cleanup Function
resource "google_cloudfunctions2_function" "cleanup" {
  name        = "virtuoso-cleanup"
  location    = var.region
  description = "Clean up old data and expired sessions"

  build_config {
    runtime     = "go121"
    entry_point = "Cleanup"

    source {
      storage_source {
        bucket = google_storage_bucket.functions_source.name
        object = google_storage_bucket_object.cleanup_source.name
      }
    }
  }

  service_config {
    max_instance_count = 5
    min_instance_count = 0
    available_memory   = "512M"
    available_cpu      = "1"
    timeout_seconds    = 300

    environment_variables = {
      PROJECT_ID        = var.project_id
      FIRESTORE_PROJECT = var.project_id
      STORAGE_BUCKET    = google_storage_bucket.app_data.name
      RETENTION_DAYS    = var.data_retention_days
    }

    service_account_email = google_service_account.cloud_functions.email

    ingress_settings               = "ALLOW_INTERNAL_ONLY"
    all_traffic_on_latest_revision = true
  }

  labels = merge(var.common_labels, {
    component = "functions"
    purpose   = "cleanup"
  })
}

# Analytics Function
resource "google_cloudfunctions2_function" "analytics" {
  name        = "virtuoso-analytics"
  location    = var.region
  description = "Process analytics and generate reports"

  build_config {
    runtime     = "python311"
    entry_point = "process_analytics"

    source {
      storage_source {
        bucket = google_storage_bucket.functions_source.name
        object = google_storage_bucket_object.analytics_source.name
      }
    }
  }

  service_config {
    max_instance_count = 10
    min_instance_count = 0
    available_memory   = "1024M"
    available_cpu      = "2"
    timeout_seconds    = 540

    environment_variables = {
      PROJECT_ID        = var.project_id
      FIRESTORE_PROJECT = var.project_id
      STORAGE_BUCKET    = google_storage_bucket.app_data.name
      BIGQUERY_DATASET  = var.enable_bigquery ? google_bigquery_dataset.analytics[0].dataset_id : ""
    }

    service_account_email = google_service_account.cloud_functions.email

    ingress_settings               = "ALLOW_INTERNAL_ONLY"
    all_traffic_on_latest_revision = true
  }

  event_trigger {
    trigger_region = var.region
    event_type     = "google.cloud.pubsub.topic.v1.messagePublished"
    pubsub_topic   = google_pubsub_topic.system_events.id

    retry_policy = "RETRY_POLICY_RETRY"
  }

  labels = merge(var.common_labels, {
    component = "functions"
    purpose   = "analytics"
  })
}

# Storage bucket for function source code
resource "google_storage_bucket" "functions_source" {
  name          = "${var.project_id}-functions-source-${random_id.suffix.hex}"
  location      = var.region
  force_destroy = true

  uniform_bucket_level_access = true

  versioning {
    enabled = true
  }

  lifecycle_rule {
    condition {
      num_newer_versions = 3
    }
    action {
      type = "Delete"
    }
  }

  labels = merge(var.common_labels, {
    component = "functions"
    purpose   = "source-code"
  })
}

# Upload function source code (placeholder - replace with actual code)
resource "google_storage_bucket_object" "health_check_source" {
  name   = "health-check-${timestamp()}.zip"
  bucket = google_storage_bucket.functions_source.name

  content = base64decode(data.archive_file.health_check_source.output_base64sha256)

  lifecycle {
    ignore_changes = [name]
  }
}

resource "google_storage_bucket_object" "webhook_handler_source" {
  name   = "webhook-handler-${timestamp()}.zip"
  bucket = google_storage_bucket.functions_source.name

  content = base64decode(data.archive_file.webhook_handler_source.output_base64sha256)

  lifecycle {
    ignore_changes = [name]
  }
}

resource "google_storage_bucket_object" "cleanup_source" {
  name   = "cleanup-${timestamp()}.zip"
  bucket = google_storage_bucket.functions_source.name

  content = base64decode(data.archive_file.cleanup_source.output_base64sha256)

  lifecycle {
    ignore_changes = [name]
  }
}

resource "google_storage_bucket_object" "analytics_source" {
  name   = "analytics-${timestamp()}.zip"
  bucket = google_storage_bucket.functions_source.name

  content = base64decode(data.archive_file.analytics_source.output_base64sha256)

  lifecycle {
    ignore_changes = [name]
  }
}

# Archive files for function source (placeholder)
data "archive_file" "health_check_source" {
  type        = "zip"
  output_path = "/tmp/health-check.zip"

  source {
    content  = file("${path.module}/functions/health_check.py")
    filename = "main.py"
  }

  source {
    content  = file("${path.module}/functions/requirements.txt")
    filename = "requirements.txt"
  }
}

data "archive_file" "webhook_handler_source" {
  type        = "zip"
  output_path = "/tmp/webhook-handler.zip"

  source {
    content  = file("${path.module}/functions/webhook_handler.py")
    filename = "main.py"
  }

  source {
    content  = file("${path.module}/functions/requirements.txt")
    filename = "requirements.txt"
  }
}

data "archive_file" "cleanup_source" {
  type        = "zip"
  output_path = "/tmp/cleanup.zip"

  source {
    content  = file("${path.module}/functions/cleanup.go")
    filename = "cleanup.go"
  }

  source {
    content  = file("${path.module}/functions/go.mod")
    filename = "go.mod"
  }
}

data "archive_file" "analytics_source" {
  type        = "zip"
  output_path = "/tmp/analytics.zip"

  source {
    content  = file("${path.module}/functions/analytics.py")
    filename = "main.py"
  }

  source {
    content  = file("${path.module}/functions/requirements.txt")
    filename = "requirements.txt"
  }
}

# Backend service for health check function (for load balancer integration)
resource "google_compute_backend_service" "health_function_backend" {
  count = length(var.api_domains) > 0 ? 1 : 0

  name                  = "virtuoso-health-function-backend"
  protocol              = "HTTPS"
  load_balancing_scheme = "EXTERNAL_MANAGED"

  backend {
    group = google_compute_region_network_endpoint_group.health_function_neg[0].id
  }

  timeout_sec = 30
}

resource "google_compute_region_network_endpoint_group" "health_function_neg" {
  count = length(var.api_domains) > 0 ? 1 : 0

  name                  = "virtuoso-health-function-neg"
  network_endpoint_type = "SERVERLESS"
  region                = var.region

  cloud_function {
    function = google_cloudfunctions2_function.health_check.name
  }
}

# IAM bindings for function invocation
resource "google_cloudfunctions2_function_iam_member" "health_check_invoker" {
  project        = var.project_id
  location       = google_cloudfunctions2_function.health_check.location
  cloud_function = google_cloudfunctions2_function.health_check.name
  role           = "roles/cloudfunctions.invoker"
  member         = "allUsers" # Public access for health checks
}

resource "google_cloudfunctions2_function_iam_member" "webhook_invoker" {
  project        = var.project_id
  location       = google_cloudfunctions2_function.webhook_handler.location
  cloud_function = google_cloudfunctions2_function.webhook_handler.name
  role           = "roles/cloudfunctions.invoker"
  member         = "allUsers" # Public access for webhooks
}

resource "google_cloudfunctions2_function_iam_member" "cleanup_invoker" {
  project        = var.project_id
  location       = google_cloudfunctions2_function.cleanup.location
  cloud_function = google_cloudfunctions2_function.cleanup.name
  role           = "roles/cloudfunctions.invoker"
  member         = "serviceAccount:${google_service_account.cloud_scheduler.email}"
}

# Outputs
output "function_urls" {
  value = {
    health_check    = google_cloudfunctions2_function.health_check.service_config[0].uri
    webhook_handler = google_cloudfunctions2_function.webhook_handler.service_config[0].uri
    cleanup         = google_cloudfunctions2_function.cleanup.service_config[0].uri
    analytics       = google_cloudfunctions2_function.analytics.service_config[0].uri
  }
  description = "Cloud Function URLs"
  sensitive   = true
}
