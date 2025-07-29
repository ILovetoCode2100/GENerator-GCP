# Core services configuration for Virtuoso API CLI
# Includes Cloud Run, Firestore, Memorystore, and Cloud Storage

# Cloud Run service for the main API
resource "google_cloud_run_v2_service" "api" {
  name     = "virtuoso-api-cli"
  location = var.region

  template {
    service_account = google_service_account.cloud_run.email

    # VPC connector for private resources access
    vpc_access {
      connector = google_vpc_access_connector.connector.id
      egress    = "PRIVATE_RANGES_ONLY"
    }

    scaling {
      min_instance_count = var.cloud_run_min_instances
      max_instance_count = var.cloud_run_max_instances
    }

    containers {
      image = var.api_image_url != "" ? var.api_image_url : "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.containers.repository_id}/virtuoso-api-cli:latest"

      resources {
        limits = {
          cpu    = var.cloud_run_cpu
          memory = var.cloud_run_memory
        }
        cpu_idle = true
        startup_cpu_boost = true
      }

      # Environment variables
      env {
        name  = "ENVIRONMENT"
        value = var.environment
      }

      env {
        name  = "PROJECT_ID"
        value = var.project_id
      }

      env {
        name  = "FIRESTORE_PROJECT"
        value = var.project_id
      }

      env {
        name = "REDIS_URL"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.redis_url.secret_id
            version = "latest"
          }
        }
      }

      env {
        name = "VIRTUOSO_API_KEY"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.virtuoso_api_key.secret_id
            version = "latest"
          }
        }
      }

      env {
        name  = "STORAGE_BUCKET"
        value = google_storage_bucket.app_data.name
      }

      env {
        name  = "PUBSUB_TOPIC_COMMANDS"
        value = google_pubsub_topic.command_events.name
      }

      env {
        name  = "PUBSUB_TOPIC_TESTS"
        value = google_pubsub_topic.test_events.name
      }

      # Liveness probe
      liveness_probe {
        initial_delay_seconds = 10
        period_seconds        = 10
        timeout_seconds       = 3
        failure_threshold     = 3

        http_get {
          path = "/health"
          port = 8080
        }
      }

      # Startup probe for slower initial startup
      startup_probe {
        initial_delay_seconds = 0
        period_seconds        = 5
        timeout_seconds       = 3
        failure_threshold     = 30

        http_get {
          path = "/health"
          port = 8080
        }
      }
    }

    # Execution environment
    execution_environment = "EXECUTION_ENVIRONMENT_GEN2"
    max_instance_request_concurrency = 100
    timeout = "300s"
  }

  traffic {
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
    percent = 100
  }

  labels = merge(var.common_labels, {
    component = "api"
    service   = "cloud-run"
  })

  depends_on = [
    google_project_service.apis["run.googleapis.com"],
    google_vpc_access_connector.connector,
    google_secret_manager_secret_version.redis_url,
    google_secret_manager_secret_version.virtuoso_api_key,
  ]
}

# Firestore database instance
resource "google_firestore_database" "main" {
  project     = var.project_id
  name        = "(default)"
  location_id = var.firestore_location
  type        = "FIRESTORE_NATIVE"

  # Concurrency mode for better performance
  concurrency_mode = "OPTIMISTIC"

  # Enable point-in-time recovery
  point_in_time_recovery_enablement = var.environment == "production" ? "POINT_IN_TIME_RECOVERY_ENABLED" : "POINT_IN_TIME_RECOVERY_DISABLED"

  depends_on = [google_project_service.apis["firestore.googleapis.com"]]
}

# Firestore indexes for common queries
resource "google_firestore_index" "session_user_timestamp" {
  project    = var.project_id
  database   = google_firestore_database.main.name
  collection = "sessions"

  fields {
    field_path = "user_id"
    order      = "ASCENDING"
  }

  fields {
    field_path = "updated_at"
    order      = "DESCENDING"
  }
}

resource "google_firestore_index" "test_runs_status_timestamp" {
  project    = var.project_id
  database   = google_firestore_database.main.name
  collection = "test_runs"

  fields {
    field_path = "status"
    order      = "ASCENDING"
  }

  fields {
    field_path = "created_at"
    order      = "DESCENDING"
  }
}

# Memorystore Redis instance
resource "google_redis_instance" "cache" {
  name           = "virtuoso-cache-${var.environment}"
  display_name   = "Virtuoso Cache ${title(var.environment)}"
  region         = var.region
  tier           = var.redis_tier
  memory_size_gb = var.redis_memory_gb

  redis_version = "REDIS_7_0"

  # Auth enabled for security
  auth_enabled = true

  # Transit encryption for security
  transit_encryption_mode = "SERVER_AUTHENTICATION"

  # Use the VPC network
  authorized_network = google_compute_network.vpc.id
  connect_mode      = "PRIVATE_SERVICE_ACCESS"

  # Maintenance window (Sunday 2-3 AM)
  maintenance_policy {
    weekly_maintenance_window {
      day = "SUNDAY"
      start_time {
        hours   = 2
        minutes = 0
        seconds = 0
        nanos   = 0
      }
    }
  }

  # Redis configuration
  redis_configs = {
    "maxmemory-policy"  = "allkeys-lru"
    "notify-keyspace-events" = "Ex"
    "timeout" = "300"
  }

  labels = merge(var.common_labels, {
    component = "cache"
    service   = "memorystore"
  })

  depends_on = [
    google_project_service.apis["redis.googleapis.com"],
    google_service_networking_connection.private_vpc_connection,
  ]
}

# Cloud Storage buckets for different purposes
# (Main bucket created in main.tf)

# Backup bucket with lifecycle rules
resource "google_storage_bucket" "backups" {
  name          = "${var.project_id}-virtuoso-backups-${random_id.suffix.hex}"
  location      = var.backup_location
  force_destroy = false

  uniform_bucket_level_access = true

  versioning {
    enabled = true
  }

  # Cross-region replication for disaster recovery
  dynamic "lifecycle_rule" {
    for_each = var.environment == "production" ? [1] : []
    content {
      condition {
        age = 365
      }
      action {
        type = "Delete"
      }
    }
  }

  # Retention policy for compliance
  dynamic "retention_policy" {
    for_each = var.environment == "production" ? [1] : []
    content {
      retention_period = 2592000 # 30 days in seconds
      is_locked        = false
    }
  }

  labels = merge(var.common_labels, {
    component = "storage"
    purpose   = "backups"
  })

  depends_on = [google_project_service.apis["storage.googleapis.com"]]
}

# Cloud Storage bucket for static assets (if needed)
resource "google_storage_bucket" "static_assets" {
  name          = "${var.project_id}-virtuoso-static-${random_id.suffix.hex}"
  location      = var.region
  force_destroy = var.environment != "production"

  uniform_bucket_level_access = false # Allow public access for static files

  website {
    main_page_suffix = "index.html"
    not_found_page   = "404.html"
  }

  cors {
    origin          = var.cors_origins
    method          = ["GET", "HEAD", "OPTIONS"]
    response_header = ["*"]
    max_age_seconds = 3600
  }

  labels = merge(var.common_labels, {
    component = "storage"
    purpose   = "static-assets"
  })

  depends_on = [google_project_service.apis["storage.googleapis.com"]]
}

# Make static assets bucket public (if needed)
resource "google_storage_bucket_iam_member" "static_public" {
  bucket = google_storage_bucket.static_assets.name
  role   = "roles/storage.objectViewer"
  member = "allUsers"
}

# Service account for Cloud Run
resource "google_service_account" "cloud_run" {
  account_id   = "virtuoso-cloud-run-sa"
  display_name = "Virtuoso Cloud Run Service Account"
  description  = "Service account for Virtuoso API CLI Cloud Run service"
}

# IAM roles for Cloud Run service account
resource "google_project_iam_member" "cloud_run_roles" {
  for_each = toset([
    "roles/datastore.user",
    "roles/redis.editor",
    "roles/cloudtasks.enqueuer",
    "roles/pubsub.publisher",
    "roles/storage.objectUser",
    "roles/secretmanager.secretAccessor",
    "roles/logging.logWriter",
    "roles/cloudtrace.agent",
  ])

  project = var.project_id
  role    = each.value
  member  = "serviceAccount:${google_service_account.cloud_run.email}"
}

# Cloud Run service IAM to allow public access (optional)
resource "google_cloud_run_v2_service_iam_member" "public_access" {
  count    = var.allow_public_access ? 1 : 0
  project  = google_cloud_run_v2_service.api.project
  location = google_cloud_run_v2_service.api.location
  name     = google_cloud_run_v2_service.api.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

# Outputs
output "cloud_run_url" {
  value       = google_cloud_run_v2_service.api.uri
  description = "URL of the Cloud Run service"
}

output "firestore_database" {
  value       = google_firestore_database.main.name
  description = "Firestore database name"
}

output "redis_host" {
  value       = google_redis_instance.cache.host
  description = "Redis instance host"
  sensitive   = true
}

output "redis_port" {
  value       = google_redis_instance.cache.port
  description = "Redis instance port"
}

output "static_assets_url" {
  value       = "https://storage.googleapis.com/${google_storage_bucket.static_assets.name}"
  description = "URL for static assets"
}
