# Input variables for Virtuoso API CLI Terraform configuration

# Project and Environment
variable "project_id" {
  description = "GCP project ID"
  type        = string
}

variable "organization_id" {
  description = "GCP organization ID (optional, for org policies)"
  type        = string
  default     = ""
}

variable "environment" {
  description = "Environment name (development, staging, production)"
  type        = string
  default     = "development"

  validation {
    condition     = contains(["development", "staging", "production"], var.environment)
    error_message = "Environment must be one of: development, staging, production"
  }
}

variable "region" {
  description = "GCP region for resources"
  type        = string
  default     = "us-central1"
}

variable "backup_location" {
  description = "GCS location for backups (multi-region)"
  type        = string
  default     = "US"
}

variable "bigquery_location" {
  description = "BigQuery dataset location"
  type        = string
  default     = "US"
}

variable "firestore_location" {
  description = "Firestore location"
  type        = string
  default     = "us-central"
}

# Networking
variable "subnet_cidr" {
  description = "CIDR range for main subnet"
  type        = string
  default     = "10.0.0.0/24"
}

variable "vpc_connector_cidr" {
  description = "CIDR range for serverless VPC connector"
  type        = string
  default     = "10.1.0.0/28"
}

variable "vpc_connector_machine_type" {
  description = "Machine type for VPC connector"
  type        = string
  default     = "e2-micro"
}

# Cloud Run Configuration
variable "cloud_run_min_instances" {
  description = "Minimum number of Cloud Run instances"
  type        = number
  default     = 0
}

variable "cloud_run_max_instances" {
  description = "Maximum number of Cloud Run instances"
  type        = number
  default     = 1000
}

variable "cloud_run_cpu" {
  description = "CPU allocation for Cloud Run"
  type        = string
  default     = "2"
}

variable "cloud_run_memory" {
  description = "Memory allocation for Cloud Run"
  type        = string
  default     = "4Gi"
}

variable "api_image_url" {
  description = "Container image URL for API service (optional, will build if not provided)"
  type        = string
  default     = ""
}

# Redis Configuration
variable "redis_tier" {
  description = "Redis tier (BASIC or STANDARD_HA)"
  type        = string
  default     = "BASIC"

  validation {
    condition     = contains(["BASIC", "STANDARD_HA"], var.redis_tier)
    error_message = "Redis tier must be BASIC or STANDARD_HA"
  }
}

variable "redis_memory_gb" {
  description = "Redis memory size in GB"
  type        = number
  default     = 1
}

# Security
variable "virtuoso_api_key" {
  description = "Virtuoso API key (sensitive)"
  type        = string
  sensitive   = true
  default     = ""
}

variable "api_domains" {
  description = "List of domains for SSL certificate and API access"
  type        = list(string)
  default     = []
}

variable "cors_origins" {
  description = "CORS allowed origins"
  type        = list(string)
  default     = ["*"]
}

variable "allow_public_access" {
  description = "Allow public access to Cloud Run service"
  type        = bool
  default     = true
}

# Feature Flags
variable "enable_cloud_armor" {
  description = "Enable Cloud Armor DDoS protection"
  type        = bool
  default     = true
}

variable "enable_identity_platform" {
  description = "Enable Identity Platform for API key management"
  type        = bool
  default     = false
}

variable "enable_vpc_service_controls" {
  description = "Enable VPC Service Controls"
  type        = bool
  default     = false
}

variable "access_policy_id" {
  description = "Access policy ID for VPC Service Controls"
  type        = string
  default     = ""
}

variable "enable_binary_authorization" {
  description = "Enable Binary Authorization for container security"
  type        = bool
  default     = false
}

variable "enable_security_center" {
  description = "Enable Security Command Center integration"
  type        = bool
  default     = false
}

variable "enable_private_service_connect" {
  description = "Enable Private Service Connect for private endpoints"
  type        = bool
  default     = false
}

variable "enable_org_policies" {
  description = "Enable organization policies (requires org ID)"
  type        = bool
  default     = false
}

variable "enable_workflows" {
  description = "Enable Google Workflows for complex orchestration"
  type        = bool
  default     = false
}

variable "enable_bigquery" {
  description = "Enable BigQuery for analytics"
  type        = bool
  default     = true
}

variable "enable_slos" {
  description = "Enable Service Level Objectives monitoring"
  type        = bool
  default     = false
}

# Data Retention
variable "data_retention_days" {
  description = "Number of days to retain data"
  type        = string
  default     = "30"
}

# Monitoring and Alerts
variable "alert_email_addresses" {
  description = "Email addresses for monitoring alerts"
  type        = list(string)
  default     = []
}

variable "slack_webhook_url" {
  description = "Slack webhook URL for alerts"
  type        = string
  default     = ""
  sensitive   = true
}

variable "slack_channel_name" {
  description = "Slack channel name for alerts"
  type        = string
  default     = "#alerts"
}

variable "slack_auth_token" {
  description = "Slack auth token"
  type        = string
  default     = ""
  sensitive   = true
}

variable "notification_channels" {
  description = "List of notification channel IDs for alerts"
  type        = list(string)
  default     = []
}

# Labels
variable "common_labels" {
  description = "Common labels to apply to all resources"
  type        = map(string)
  default = {
    app         = "virtuoso-api-cli"
    managed_by  = "terraform"
  }
}

# Cost Optimization
variable "use_preemptible_instances" {
  description = "Use preemptible instances where possible"
  type        = bool
  default     = true
}

# Function Source Placeholders
variable "function_source_path" {
  description = "Path to Cloud Functions source code"
  type        = string
  default     = "./functions"
}
