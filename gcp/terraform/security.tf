# Security configuration for Virtuoso API CLI
# Includes IAM, Secret Manager, Identity Platform, and Cloud Armor

# Secret Manager secrets
resource "google_secret_manager_secret" "virtuoso_api_key" {
  secret_id = "virtuoso-api-key"

  replication {
    auto {}
  }

  labels = merge(var.common_labels, {
    component = "secrets"
    purpose   = "api-key"
  })

  depends_on = [google_project_service.apis["secretmanager.googleapis.com"]]
}

resource "google_secret_manager_secret_version" "virtuoso_api_key" {
  secret      = google_secret_manager_secret.virtuoso_api_key.id
  secret_data = var.virtuoso_api_key != "" ? var.virtuoso_api_key : "PLACEHOLDER_API_KEY"

  lifecycle {
    ignore_changes = [secret_data]
  }
}

resource "google_secret_manager_secret" "redis_url" {
  secret_id = "redis-url"

  replication {
    auto {}
  }

  labels = merge(var.common_labels, {
    component = "secrets"
    purpose   = "redis-connection"
  })
}

resource "google_secret_manager_secret_version" "redis_url" {
  secret = google_secret_manager_secret.redis_url.id
  secret_data = format(
    "redis://:%s@%s:%s",
    google_redis_instance.cache.auth_string,
    google_redis_instance.cache.host,
    google_redis_instance.cache.port
  )
}

# Additional secrets for various services
locals {
  additional_secrets = {
    "jwt-secret"     = random_password.jwt_secret.result
    "webhook-secret" = random_password.webhook_secret.result
    "encryption-key" = random_password.encryption_key.result
  }
}

resource "google_secret_manager_secret" "additional" {
  for_each = local.additional_secrets

  secret_id = each.key

  replication {
    auto {}
  }

  labels = merge(var.common_labels, {
    component = "secrets"
    purpose   = each.key
  })
}

resource "google_secret_manager_secret_version" "additional" {
  for_each = local.additional_secrets

  secret      = google_secret_manager_secret.additional[each.key].id
  secret_data = each.value
}

# Random passwords for secrets
resource "random_password" "jwt_secret" {
  length  = 32
  special = true
}

resource "random_password" "webhook_secret" {
  length  = 32
  special = true
}

resource "random_password" "encryption_key" {
  length  = 32
  special = false # Base64 safe
}

# Service Account for Cloud Functions
resource "google_service_account" "cloud_functions" {
  account_id   = "virtuoso-functions-sa"
  display_name = "Virtuoso Cloud Functions Service Account"
  description  = "Service account for Virtuoso Cloud Functions"
}

# IAM roles for Cloud Functions service account
resource "google_project_iam_member" "cloud_functions_roles" {
  for_each = toset([
    "roles/datastore.viewer",
    "roles/pubsub.publisher",
    "roles/secretmanager.secretAccessor",
    "roles/logging.logWriter",
    "roles/cloudtrace.agent",
    "roles/monitoring.metricWriter",
  ])

  project = var.project_id
  role    = each.value
  member  = "serviceAccount:${google_service_account.cloud_functions.email}"
}

# Service Account for Cloud Tasks
resource "google_service_account" "cloud_tasks" {
  account_id   = "virtuoso-tasks-sa"
  display_name = "Virtuoso Cloud Tasks Service Account"
  description  = "Service account for Cloud Tasks to invoke services"
}

# Allow Cloud Tasks to invoke Cloud Run
resource "google_cloud_run_v2_service_iam_member" "tasks_invoker" {
  project  = google_cloud_run_v2_service.api.project
  location = google_cloud_run_v2_service.api.location
  name     = google_cloud_run_v2_service.api.name
  role     = "roles/run.invoker"
  member   = "serviceAccount:${google_service_account.cloud_tasks.email}"
}

# Service Account for Cloud Scheduler
resource "google_service_account" "cloud_scheduler" {
  account_id   = "virtuoso-scheduler-sa"
  display_name = "Virtuoso Cloud Scheduler Service Account"
  description  = "Service account for Cloud Scheduler jobs"
}

# Allow Cloud Scheduler to invoke Cloud Functions
resource "google_project_iam_member" "scheduler_functions_invoker" {
  project = var.project_id
  role    = "roles/cloudfunctions.invoker"
  member  = "serviceAccount:${google_service_account.cloud_scheduler.email}"
}

# Service Account for CI/CD (Cloud Build)
resource "google_service_account" "cloud_build" {
  account_id   = "virtuoso-build-sa"
  display_name = "Virtuoso Cloud Build Service Account"
  description  = "Service account for CI/CD pipeline"
}

# IAM roles for Cloud Build service account
resource "google_project_iam_member" "cloud_build_roles" {
  for_each = toset([
    "roles/cloudbuild.builds.builder",
    "roles/run.developer",
    "roles/storage.admin",
    "roles/artifactregistry.writer",
    "roles/secretmanager.secretAccessor",
  ])

  project = var.project_id
  role    = each.value
  member  = "serviceAccount:${google_service_account.cloud_build.email}"
}

# Identity Platform configuration for API key management
resource "google_identity_platform_config" "default" {
  count = var.enable_identity_platform ? 1 : 0

  project = var.project_id

  # Enable API key authentication
  authorized_domains = var.api_domains

  # Multi-factor authentication (optional)
  mfa {
    state = var.environment == "production" ? "ENABLED" : "DISABLED"
  }

  # Monitoring
  monitoring {
    request_logging {
      enabled = true
    }
  }

  depends_on = [google_project_service.apis["identitytoolkit.googleapis.com"]]
}

# Organization Policy for security best practices
resource "google_organization_policy" "security_policies" {
  for_each = var.enable_org_policies ? {
    "compute.requireShieldedVm" = "true"
    "compute.requireOsLogin"    = "true"
    "iam.disableServiceAccountKeyCreation" = "true"
    "storage.uniformBucketLevelAccess" = "true"
  } : {}

  org_id     = var.organization_id
  constraint = each.key

  boolean_policy {
    enforced = each.value == "true"
  }
}

# VPC Service Controls (optional, for enhanced security)
resource "google_access_context_manager_service_perimeter" "api_perimeter" {
  count = var.enable_vpc_service_controls ? 1 : 0

  parent = "accessPolicies/${var.access_policy_id}"
  name   = "accessPolicies/${var.access_policy_id}/servicePerimeters/virtuoso_api_perimeter"
  title  = "Virtuoso API Service Perimeter"

  status {
    resources = [
      "projects/${data.google_project.project.number}",
    ]

    restricted_services = [
      "firestore.googleapis.com",
      "redis.googleapis.com",
      "storage.googleapis.com",
      "pubsub.googleapis.com",
      "cloudtasks.googleapis.com",
      "secretmanager.googleapis.com",
    ]

    # Allow access from VPC
    vpc_accessible_services {
      enable_restriction = true
      allowed_services   = ["RESTRICTED-SERVICES"]
    }
  }
}

# Binary Authorization (optional, for container security)
resource "google_binary_authorization_policy" "policy" {
  count = var.enable_binary_authorization ? 1 : 0

  project = var.project_id

  admission_whitelist_patterns {
    name_pattern = "gcr.io/${var.project_id}/*"
  }

  admission_whitelist_patterns {
    name_pattern = "${var.region}-docker.pkg.dev/${var.project_id}/*"
  }

  default_admission_rule {
    evaluation_mode  = "REQUIRE_ATTESTATION"
    enforcement_mode = "ENFORCED_BLOCK_AND_AUDIT_LOG"

    require_attestations_by = [
      google_binary_authorization_attestor.prod_attestor[0].name,
    ]
  }

  # Allow Google-provided system images
  global_policy_evaluation_mode = "ENABLE"
}

resource "google_binary_authorization_attestor" "prod_attestor" {
  count = var.enable_binary_authorization ? 1 : 0

  name = "prod-attestor"

  attestation_authority_note {
    note_reference = google_container_analysis_note.attestor_note[0].name

    public_keys {
      id = data.google_kms_crypto_key_version.attestor_key[0].id

      pkix_public_key {
        public_key_pem      = data.google_kms_crypto_key_version.attestor_key[0].public_key[0].pem
        signature_algorithm = data.google_kms_crypto_key_version.attestor_key[0].public_key[0].algorithm
      }
    }
  }
}

# Container Analysis Note for Binary Authorization
resource "google_container_analysis_note" "attestor_note" {
  count = var.enable_binary_authorization ? 1 : 0

  name = "prod-attestor-note"

  attestation_authority {
    hint {
      human_readable_name = "Production Attestor"
    }
  }
}

# KMS key for Binary Authorization (if enabled)
resource "google_kms_key_ring" "attestor_key_ring" {
  count    = var.enable_binary_authorization ? 1 : 0
  name     = "attestor-key-ring"
  location = var.region
}

resource "google_kms_crypto_key" "attestor_key" {
  count           = var.enable_binary_authorization ? 1 : 0
  name            = "attestor-key"
  key_ring        = google_kms_key_ring.attestor_key_ring[0].id
  purpose         = "ASYMMETRIC_SIGN"

  version_template {
    algorithm = "RSA_SIGN_PSS_4096_SHA512"
  }
}

data "google_kms_crypto_key_version" "attestor_key" {
  count      = var.enable_binary_authorization ? 1 : 0
  crypto_key = google_kms_crypto_key.attestor_key[0].id
}

# Security Command Center notifications (optional)
resource "google_scc_notification_config" "security_findings" {
  count = var.enable_security_center ? 1 : 0

  config_id    = "virtuoso-security-findings"
  organization = var.organization_id

  pubsub_topic = google_pubsub_topic.security_alerts[0].id

  streaming_config {
    filter = "category=\"VULNERABILITY\" OR category=\"THREAT\""
  }
}

# Pub/Sub topic for security alerts
resource "google_pubsub_topic" "security_alerts" {
  count = var.enable_security_center ? 1 : 0

  name = "virtuoso-security-alerts"

  message_retention_duration = "604800s" # 7 days

  labels = merge(var.common_labels, {
    component = "security"
    purpose   = "alerts"
  })
}

# Outputs
output "service_accounts" {
  value = {
    cloud_run       = google_service_account.cloud_run.email
    cloud_functions = google_service_account.cloud_functions.email
    cloud_tasks     = google_service_account.cloud_tasks.email
    cloud_scheduler = google_service_account.cloud_scheduler.email
    cloud_build     = google_service_account.cloud_build.email
  }
  description = "Service account emails"
}

output "secret_ids" {
  value = {
    virtuoso_api_key = google_secret_manager_secret.virtuoso_api_key.secret_id
    redis_url        = google_secret_manager_secret.redis_url.secret_id
    jwt_secret       = google_secret_manager_secret.additional["jwt-secret"].secret_id
    webhook_secret   = google_secret_manager_secret.additional["webhook-secret"].secret_id
    encryption_key   = google_secret_manager_secret.additional["encryption-key"].secret_id
  }
  description = "Secret Manager secret IDs"
}
