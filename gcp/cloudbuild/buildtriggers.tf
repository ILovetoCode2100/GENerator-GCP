# Terraform configuration for Cloud Build triggers
# This defines automated CI/CD pipelines for the Virtuoso API CLI

terraform {
  required_version = ">= 1.6.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "region" {
  description = "Default region for resources"
  type        = string
  default     = "us-central1"
}

variable "github_owner" {
  description = "GitHub repository owner"
  type        = string
}

variable "github_repo" {
  description = "GitHub repository name"
  type        = string
  default     = "virtuoso-GENerator"
}

# Cloud Build API enablement
resource "google_project_service" "cloudbuild" {
  service = "cloudbuild.googleapis.com"
}

# Service account for Cloud Build
resource "google_service_account" "cloudbuild" {
  account_id   = "virtuoso-cloudbuild"
  display_name = "Virtuoso Cloud Build Service Account"
  description  = "Service account for Cloud Build operations"
}

# IAM roles for Cloud Build service account
resource "google_project_iam_member" "cloudbuild_roles" {
  for_each = toset([
    "roles/cloudbuild.builds.builder",
    "roles/run.admin",
    "roles/storage.admin",
    "roles/artifactregistry.admin",
    "roles/cloudfunctions.admin",
    "roles/logging.logWriter",
    "roles/iam.serviceAccountUser"
  ])

  project = var.project_id
  role    = each.key
  member  = "serviceAccount:${google_service_account.cloudbuild.email}"
}

# GitHub connection (requires manual approval in Console)
resource "google_cloudbuildv2_connection" "github" {
  project  = var.project_id
  location = var.region
  name     = "virtuoso-github"

  github_config {
    app_installation_id = 0 # Set after manual connection
    authorizer_credential {
      oauth_token_secret_version = "projects/${var.project_id}/secrets/github-oauth-token/versions/latest"
    }
  }
}

resource "google_cloudbuildv2_repository" "repo" {
  project           = var.project_id
  location          = var.region
  connection        = google_cloudbuildv2_connection.github.name
  name              = var.github_repo
  remote_uri        = "https://github.com/${var.github_owner}/${var.github_repo}.git"
  parent_connection = google_cloudbuildv2_connection.github.id
}

# Trigger 1: Main branch deployment
resource "google_cloudbuild_trigger" "main_deploy" {
  name        = "virtuoso-main-deploy"
  description = "Deploy to production on merge to main"
  project     = var.project_id
  location    = var.region

  repository_event_config {
    repository = google_cloudbuildv2_repository.repo.id
    push {
      branch = "^main$"
    }
  }

  filename = "gcp/cloudbuild/cloudbuild.yaml"

  substitutions = {
    _ENVIRONMENT     = "prod"
    _ENABLE_APPROVAL = "true"
  }

  service_account = google_service_account.cloudbuild.id

  included_files = [
    "cmd/**",
    "pkg/**",
    "go.mod",
    "go.sum",
    "Dockerfile"
  ]

  ignored_files = [
    "*.md",
    "docs/**",
    "examples/**"
  ]
}

# Trigger 2: Pull request builds
resource "google_cloudbuild_trigger" "pr_build" {
  name        = "virtuoso-pr-build"
  description = "Build and test pull requests"
  project     = var.project_id
  location    = var.region

  repository_event_config {
    repository = google_cloudbuildv2_repository.repo.id
    pull_request {
      branch = ".*"
    }
  }

  filename = "gcp/cloudbuild/cloudbuild-pr.yaml"

  substitutions = {
    _PR_NUMBER = "$_PR_NUMBER"
  }

  service_account = google_service_account.cloudbuild.id

  included_files = [
    "cmd/**",
    "pkg/**",
    "go.mod",
    "go.sum",
    "Dockerfile",
    "gcp/cloudbuild/**"
  ]
}

# Trigger 3: Tag-based releases
resource "google_cloudbuild_trigger" "release" {
  name        = "virtuoso-release"
  description = "Create releases from tags"
  project     = var.project_id
  location    = var.region

  repository_event_config {
    repository = google_cloudbuildv2_repository.repo.id
    push {
      tag = "^v[0-9]+\\.[0-9]+\\.[0-9]+$"
    }
  }

  filename = "gcp/cloudbuild/cloudbuild.yaml"

  substitutions = {
    _ENVIRONMENT     = "prod"
    _ENABLE_APPROVAL = "false" # Auto-deploy releases
  }

  service_account = google_service_account.cloudbuild.id
}

# Trigger 4: Staging deployment
resource "google_cloudbuild_trigger" "staging_deploy" {
  name        = "virtuoso-staging-deploy"
  description = "Deploy to staging on merge to staging branch"
  project     = var.project_id
  location    = var.region

  repository_event_config {
    repository = google_cloudbuildv2_repository.repo.id
    push {
      branch = "^staging$"
    }
  }

  filename = "gcp/cloudbuild/cloudbuild.yaml"

  substitutions = {
    _ENVIRONMENT     = "staging"
    _ENABLE_APPROVAL = "false"
  }

  service_account = google_service_account.cloudbuild.id
}

# Trigger 5: Development deployment
resource "google_cloudbuild_trigger" "dev_deploy" {
  name        = "virtuoso-dev-deploy"
  description = "Deploy to dev on merge to develop branch"
  project     = var.project_id
  location    = var.region

  repository_event_config {
    repository = google_cloudbuildv2_repository.repo.id
    push {
      branch = "^develop$"
    }
  }

  filename = "gcp/cloudbuild/cloudbuild.yaml"

  substitutions = {
    _ENVIRONMENT     = "dev"
    _ENABLE_APPROVAL = "false"
  }

  service_account = google_service_account.cloudbuild.id
}

# Trigger 6: Terraform plan on PR
resource "google_cloudbuild_trigger" "terraform_plan" {
  name        = "virtuoso-terraform-plan"
  description = "Run Terraform plan on infrastructure changes"
  project     = var.project_id
  location    = var.region

  repository_event_config {
    repository = google_cloudbuildv2_repository.repo.id
    pull_request {
      branch = ".*"
    }
  }

  filename = "gcp/cloudbuild/cloudbuild-terraform.yaml"

  substitutions = {
    _ACTION      = "plan"
    _ENVIRONMENT = "dev" # Plan against dev by default
  }

  service_account = google_service_account.cloudbuild.id

  included_files = [
    "gcp/terraform/**",
    "gcp/cloudbuild/cloudbuild-terraform.yaml"
  ]
}

# Trigger 7: Terraform apply on merge
resource "google_cloudbuild_trigger" "terraform_apply" {
  name        = "virtuoso-terraform-apply"
  description = "Apply Terraform changes on merge to main"
  project     = var.project_id
  location    = var.region

  repository_event_config {
    repository = google_cloudbuildv2_repository.repo.id
    push {
      branch = "^main$"
    }
  }

  filename = "gcp/cloudbuild/cloudbuild-terraform.yaml"

  substitutions = {
    _ACTION      = "apply"
    _ENVIRONMENT = "prod"
  }

  service_account = google_service_account.cloudbuild.id

  included_files = [
    "gcp/terraform/**"
  ]
}

# Trigger 8: Nightly integration tests
resource "google_cloudbuild_trigger" "nightly_tests" {
  name        = "virtuoso-nightly-tests"
  description = "Run comprehensive tests nightly"
  project     = var.project_id
  location    = var.region

  repository_event_config {
    repository = google_cloudbuildv2_repository.repo.id
    push {
      branch = "^main$"
    }
  }

  filename = "gcp/cloudbuild/cloudbuild-pr.yaml"

  # Run at 2 AM UTC daily
  trigger_template {
    branch_name = "main"
    repo_name   = google_cloudbuildv2_repository.repo.name
  }

  # Use Cloud Scheduler instead for true nightly runs
  # This is just a placeholder
  disabled = true

  service_account = google_service_account.cloudbuild.id
}

# Cloud Scheduler for nightly builds
resource "google_cloud_scheduler_job" "nightly_trigger" {
  name        = "virtuoso-nightly-trigger"
  description = "Trigger nightly integration tests"
  schedule    = "0 2 * * *" # 2 AM UTC daily
  time_zone   = "UTC"
  region      = var.region

  http_target {
    uri         = "https://cloudbuild.googleapis.com/v1/projects/${var.project_id}/locations/${var.region}/triggers/${google_cloudbuild_trigger.nightly_tests.trigger_id}:run"
    http_method = "POST"

    oauth_token {
      service_account_email = google_service_account.cloudbuild.email
    }

    body = base64encode(jsonencode({
      branchName = "main"
    }))

    headers = {
      "Content-Type" = "application/json"
    }
  }
}

# Build notifications Pub/Sub topic
resource "google_pubsub_topic" "build_notifications" {
  name = "virtuoso-build-notifications"
}

# Cloud Function for build notifications
resource "google_cloudfunctions2_function" "build_notifier" {
  name        = "virtuoso-build-notifier"
  location    = var.region
  description = "Send build notifications to Slack/Teams"

  build_config {
    runtime     = "go121"
    entry_point = "HandleBuildNotification"
    source {
      storage_source {
        bucket = "virtuoso-functions-source"
        object = "build-notifier.zip"
      }
    }
  }

  service_config {
    max_instance_count = 10
    timeout_seconds    = 60
    environment_variables = {
      SLACK_WEBHOOK_URL = "YOUR_SLACK_WEBHOOK"
      TEAMS_WEBHOOK_URL = "YOUR_TEAMS_WEBHOOK"
    }
  }

  event_trigger {
    trigger_region = var.region
    event_type     = "google.cloud.pubsub.topic.v1.messagePublished"
    pubsub_topic   = google_pubsub_topic.build_notifications.id
  }
}

# Storage buckets for build artifacts
resource "google_storage_bucket" "build_artifacts" {
  name          = "${var.project_id}-build-artifacts"
  location      = var.region
  force_destroy = false

  lifecycle_rule {
    condition {
      age = 30
    }
    action {
      type = "Delete"
    }
  }

  versioning {
    enabled = true
  }
}

resource "google_storage_bucket" "pr_artifacts" {
  name          = "${var.project_id}-pr-artifacts"
  location      = var.region
  force_destroy = false

  lifecycle_rule {
    condition {
      age = 7
    }
    action {
      type = "Delete"
    }
  }
}

resource "google_storage_bucket" "build_cache" {
  name          = "${var.project_id}-build-cache"
  location      = var.region
  force_destroy = false

  lifecycle_rule {
    condition {
      age = 7
    }
    action {
      type = "Delete"
    }
  }
}

# Outputs
output "cloudbuild_service_account" {
  value = google_service_account.cloudbuild.email
}

output "build_triggers" {
  value = {
    main_deploy    = google_cloudbuild_trigger.main_deploy.id
    pr_build       = google_cloudbuild_trigger.pr_build.id
    release        = google_cloudbuild_trigger.release.id
    staging_deploy = google_cloudbuild_trigger.staging_deploy.id
    dev_deploy     = google_cloudbuild_trigger.dev_deploy.id
  }
}
