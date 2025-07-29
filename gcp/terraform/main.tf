# Main Terraform configuration for Virtuoso API CLI on GCP
# This file sets up the provider and enables required APIs

terraform {
  required_version = ">= 1.5.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.10"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "~> 5.10"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }

  # Backend configuration for remote state
  backend "gcs" {
    # These values should be provided via backend config file or CLI flags
    # bucket = "virtuoso-terraform-state"
    # prefix = "terraform/state"
  }
}

# Configure the Google Cloud Provider
provider "google" {
  project = var.project_id
  region  = var.region
}

provider "google-beta" {
  project = var.project_id
  region  = var.region
}

# Data source for project information
data "google_project" "project" {
  project_id = var.project_id
}

# Enable required APIs
resource "google_project_service" "apis" {
  for_each = toset([
    # Core services
    "compute.googleapis.com",
    "container.googleapis.com",
    "run.googleapis.com",
    "cloudfunctions.googleapis.com",

    # Storage and databases
    "storage.googleapis.com",
    "firestore.googleapis.com",
    "redis.googleapis.com",

    # Networking
    "vpcaccess.googleapis.com",
    "servicenetworking.googleapis.com",
    "networkmanagement.googleapis.com",

    # Security
    "secretmanager.googleapis.com",
    "iap.googleapis.com",
    "identitytoolkit.googleapis.com",

    # Messaging and async
    "pubsub.googleapis.com",
    "cloudtasks.googleapis.com",
    "cloudscheduler.googleapis.com",

    # Operations
    "logging.googleapis.com",
    "monitoring.googleapis.com",
    "cloudtrace.googleapis.com",
    "clouderrorreporting.googleapis.com",

    # Build and deploy
    "cloudbuild.googleapis.com",
    "artifactregistry.googleapis.com",
    "containerregistry.googleapis.com",

    # Other services
    "cloudresourcemanager.googleapis.com",
    "iam.googleapis.com",
    "serviceusage.googleapis.com",
  ])

  project            = var.project_id
  service            = each.value
  disable_on_destroy = false

  timeouts {
    create = "30m"
    update = "40m"
  }
}

# Random suffix for globally unique resource names
resource "random_id" "suffix" {
  byte_length = 4
}

# Create a Cloud Storage bucket for application data
resource "google_storage_bucket" "app_data" {
  name          = "${var.project_id}-virtuoso-data-${random_id.suffix.hex}"
  location      = var.region
  force_destroy = var.environment != "production"

  uniform_bucket_level_access = true

  versioning {
    enabled = true
  }

  lifecycle_rule {
    condition {
      age = 30
      matches_prefix = ["logs/"]
    }
    action {
      type = "Delete"
    }
  }

  lifecycle_rule {
    condition {
      age = 7
      matches_prefix = ["artifacts/"]
    }
    action {
      type          = "SetStorageClass"
      storage_class = "NEARLINE"
    }
  }

  lifecycle_rule {
    condition {
      age = 90
      matches_prefix = ["backups/"]
    }
    action {
      type          = "SetStorageClass"
      storage_class = "COLDLINE"
    }
  }

  labels = merge(var.common_labels, {
    component = "storage"
    purpose   = "app-data"
  })

  depends_on = [google_project_service.apis["storage.googleapis.com"]]
}

# Create folders in the bucket
resource "google_storage_bucket_object" "folders" {
  for_each = toset([
    "logs/",
    "artifacts/",
    "artifacts/test-results/",
    "artifacts/reports/",
    "backups/",
    "backups/firestore/",
    "backups/configs/"
  ])

  name    = each.value
  content = ""
  bucket  = google_storage_bucket.app_data.name
}

# Artifact Registry for container images
resource "google_artifact_registry_repository" "containers" {
  location      = var.region
  repository_id = "virtuoso-containers"
  description   = "Container images for Virtuoso API CLI"
  format        = "DOCKER"

  labels = merge(var.common_labels, {
    component = "registry"
  })

  depends_on = [google_project_service.apis["artifactregistry.googleapis.com"]]
}

# Output project number for use in other resources
output "project_number" {
  value = data.google_project.project.number
}

# Output the random suffix for use in other modules
output "random_suffix" {
  value = random_id.suffix.hex
}

# Output the storage bucket name
output "storage_bucket" {
  value = google_storage_bucket.app_data.name
}

# Output the artifact registry URL
output "artifact_registry_url" {
  value = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.containers.repository_id}"
}
