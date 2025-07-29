# Outputs for Virtuoso API CLI Terraform configuration

# Service URLs
output "api_url" {
  value       = google_cloud_run_v2_service.api.uri
  description = "URL of the Cloud Run API service"
}

output "load_balancer_url" {
  value       = "https://${google_compute_global_address.api_ip.address}"
  description = "URL of the load balancer (requires DNS configuration)"
}

output "static_ip_address" {
  value       = google_compute_global_address.api_ip.address
  description = "Static IP address for the load balancer"
}

# Storage
output "app_data_bucket" {
  value       = google_storage_bucket.app_data.name
  description = "Name of the application data bucket"
}

output "backup_bucket" {
  value       = google_storage_bucket.backups.name
  description = "Name of the backup bucket"
}

output "static_assets_bucket" {
  value       = google_storage_bucket.static_assets.name
  description = "Name of the static assets bucket"
}

# Database
output "firestore_database_name" {
  value       = google_firestore_database.main.name
  description = "Firestore database name"
}

output "redis_connection_string" {
  value = format(
    "redis://:%s@%s:%s",
    google_redis_instance.cache.auth_string,
    google_redis_instance.cache.host,
    google_redis_instance.cache.port
  )
  description = "Redis connection string"
  sensitive   = true
}

# Networking
output "vpc_network_name" {
  value       = google_compute_network.vpc.name
  description = "VPC network name"
}

output "vpc_connector_id" {
  value       = google_vpc_access_connector.connector.id
  description = "Serverless VPC connector ID"
}

# Functions
output "cloud_function_urls" {
  value = {
    health_check    = google_cloudfunctions2_function.health_check.service_config[0].uri
    webhook_handler = google_cloudfunctions2_function.webhook_handler.service_config[0].uri
    cleanup         = google_cloudfunctions2_function.cleanup.service_config[0].uri
    analytics       = google_cloudfunctions2_function.analytics.service_config[0].uri
  }
  description = "Cloud Function URLs"
}

# Pub/Sub
output "pubsub_topics" {
  value = {
    command_events = google_pubsub_topic.command_events.name
    test_events    = google_pubsub_topic.test_events.name
    system_events  = google_pubsub_topic.system_events.name
  }
  description = "Pub/Sub topic names"
}

# Cloud Tasks
output "cloud_tasks_queues" {
  value = {
    command_execution = google_cloud_tasks_queue.command_execution.name
    test_execution    = google_cloud_tasks_queue.test_execution.name
    batch_processing  = google_cloud_tasks_queue.batch_processing.name
  }
  description = "Cloud Tasks queue names"
}

# Service Accounts
output "service_account_emails" {
  value = {
    cloud_run       = google_service_account.cloud_run.email
    cloud_functions = google_service_account.cloud_functions.email
    cloud_tasks     = google_service_account.cloud_tasks.email
    cloud_scheduler = google_service_account.cloud_scheduler.email
    cloud_build     = google_service_account.cloud_build.email
  }
  description = "Service account email addresses"
}

# Secrets
output "secret_manager_secrets" {
  value = {
    virtuoso_api_key = google_secret_manager_secret.virtuoso_api_key.name
    redis_url        = google_secret_manager_secret.redis_url.name
    jwt_secret       = google_secret_manager_secret.additional["jwt-secret"].name
    webhook_secret   = google_secret_manager_secret.additional["webhook-secret"].name
  }
  description = "Secret Manager secret names"
}

# Monitoring
output "monitoring_dashboard_url" {
  value       = "https://console.cloud.google.com/monitoring/dashboards/custom/${google_monitoring_dashboard.main.id}?project=${var.project_id}"
  description = "URL to the Cloud Monitoring dashboard"
}

output "uptime_check_id" {
  value       = google_monitoring_uptime_check_config.api_health.uptime_check_id
  description = "Uptime check ID"
}

# BigQuery
output "bigquery_datasets" {
  value = {
    logs      = google_bigquery_dataset.logs.dataset_id
    analytics = var.enable_bigquery ? google_bigquery_dataset.analytics[0].dataset_id : null
  }
  description = "BigQuery dataset IDs"
}

# Container Registry
output "container_registry_url" {
  value       = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.containers.repository_id}"
  description = "Artifact Registry URL for container images"
}

# Project Information
output "project_info" {
  value = {
    project_id     = var.project_id
    project_number = data.google_project.project.number
    region         = var.region
    environment    = var.environment
  }
  description = "Project information"
}

# Configuration Summary
output "deployment_summary" {
  value = {
    api_endpoint = var.allow_public_access ? google_cloud_run_v2_service.api.uri : "Private access only"
    load_balancer_ip = google_compute_global_address.api_ip.address
    redis_enabled = true
    firestore_enabled = true
    cloud_armor_enabled = var.enable_cloud_armor
    binary_authorization_enabled = var.enable_binary_authorization
    workflows_enabled = var.enable_workflows
    bigquery_analytics_enabled = var.enable_bigquery
  }
  description = "Deployment configuration summary"
}

# DNS Configuration Instructions
output "dns_configuration" {
  value = var.api_domains != [] ? {
    instructions = "Configure your DNS A records to point to the following IP address:"
    ip_address   = google_compute_global_address.api_ip.address
    domains      = var.api_domains
  } : {
    instructions = "No custom domains configured. Use the Cloud Run URL directly."
    cloud_run_url = google_cloud_run_v2_service.api.uri
  }
  description = "DNS configuration instructions"
}

# Next Steps
output "next_steps" {
  value = [
    "1. Configure DNS records if using custom domains",
    "2. Set the Virtuoso API key in Secret Manager: ${google_secret_manager_secret.virtuoso_api_key.name}",
    "3. Configure monitoring notification channels if needed",
    "4. Deploy your application container to: ${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.containers.repository_id}",
    "5. Test the health check endpoint: ${google_cloud_run_v2_service.api.uri}/health"
  ]
  description = "Next steps after deployment"
}
