# Networking configuration for Virtuoso API CLI
# Includes VPC, Load Balancer, CDN, and connectivity

# VPC Network
resource "google_compute_network" "vpc" {
  name                    = "virtuoso-vpc-${var.environment}"
  auto_create_subnetworks = false

  depends_on = [google_project_service.apis["compute.googleapis.com"]]
}

# Subnet for Cloud Run and other services
resource "google_compute_subnetwork" "main" {
  name          = "virtuoso-subnet-${var.region}"
  ip_cidr_range = var.subnet_cidr
  region        = var.region
  network       = google_compute_network.vpc.id

  # Enable Private Google Access
  private_ip_google_access = true

  # Enable Flow Logs for monitoring
  log_config {
    aggregation_interval = "INTERVAL_5_SEC"
    flow_sampling        = 0.5
    metadata             = "INCLUDE_ALL_METADATA"
  }
}

# Cloud Router for NAT
resource "google_compute_router" "router" {
  name    = "virtuoso-router-${var.region}"
  region  = var.region
  network = google_compute_network.vpc.id
}

# Cloud NAT for outbound connectivity
resource "google_compute_router_nat" "nat" {
  name                               = "virtuoso-nat-${var.region}"
  router                             = google_compute_router.router.name
  region                             = var.region
  nat_ip_allocate_option             = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"

  log_config {
    enable = true
    filter = "ERRORS_ONLY"
  }
}

# Serverless VPC Connector for Cloud Run
resource "google_vpc_access_connector" "connector" {
  name          = "virtuoso-connector"
  region        = var.region
  network       = google_compute_network.vpc.id
  ip_cidr_range = var.vpc_connector_cidr

  # Machine type for the connector
  machine_type = var.vpc_connector_machine_type

  # Scaling
  min_instances = 2
  max_instances = 10

  depends_on = [google_project_service.apis["vpcaccess.googleapis.com"]]
}

# Private VPC connection for services like Redis
resource "google_compute_global_address" "private_ip_range" {
  name          = "virtuoso-private-ip-range"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = google_compute_network.vpc.id
}

resource "google_service_networking_connection" "private_vpc_connection" {
  network                 = google_compute_network.vpc.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.private_ip_range.name]

  depends_on = [google_project_service.apis["servicenetworking.googleapis.com"]]
}

# SSL Certificate for HTTPS
resource "google_compute_managed_ssl_certificate" "api_cert" {
  name = "virtuoso-api-cert"

  managed {
    domains = var.api_domains
  }

  lifecycle {
    create_before_destroy = true
  }
}

# Backend Service for Cloud Run
resource "google_compute_backend_service" "api_backend" {
  name                  = "virtuoso-api-backend"
  protocol              = "HTTPS"
  load_balancing_scheme = "EXTERNAL_MANAGED"

  backend {
    group = google_compute_region_network_endpoint_group.api_neg.id
  }

  # CDN configuration
  enable_cdn = true
  cdn_policy {
    cache_mode = "CACHE_MODE_AUTO"
    default_ttl = 300
    client_ttl  = 300
    max_ttl     = 3600

    negative_caching = true
    negative_caching_policy {
      code = 404
      ttl  = 120
    }
    negative_caching_policy {
      code = 500
      ttl  = 10
    }

    cache_key_policy {
      include_host         = true
      include_protocol     = true
      include_query_string = true

      query_string_whitelist = [
        "session_id",
        "checkpoint_id",
        "format",
        "page",
        "limit"
      ]
    }
  }

  # Health check
  health_checks = [google_compute_health_check.api_health.id]

  # Session affinity
  session_affinity = "CLIENT_IP"

  # Timeout
  timeout_sec = 300

  # Logging
  log_config {
    enable      = true
    sample_rate = 1.0
  }

  # Security policy (Cloud Armor)
  dynamic "security_policy" {
    for_each = var.enable_cloud_armor ? [1] : []
    content {
      security_policy = google_compute_security_policy.api_security[0].id
    }
  }
}

# Network Endpoint Group for Cloud Run
resource "google_compute_region_network_endpoint_group" "api_neg" {
  name                  = "virtuoso-api-neg"
  network_endpoint_type = "SERVERLESS"
  region                = var.region

  cloud_run {
    service = google_cloud_run_v2_service.api.name
  }
}

# Health Check
resource "google_compute_health_check" "api_health" {
  name               = "virtuoso-api-health"
  check_interval_sec = 10
  timeout_sec        = 5

  https_health_check {
    port         = 443
    request_path = "/health"
  }
}

# URL Map for routing
resource "google_compute_url_map" "api_urlmap" {
  name            = "virtuoso-api-urlmap"
  default_service = google_compute_backend_service.api_backend.id

  # Host rules for different domains
  dynamic "host_rule" {
    for_each = length(var.api_domains) > 1 ? var.api_domains : []
    content {
      hosts        = [host_rule.value]
      path_matcher = "api-paths"
    }
  }

  # Path matcher
  path_matcher {
    name            = "api-paths"
    default_service = google_compute_backend_service.api_backend.id

    # Route health checks to Cloud Functions
    path_rule {
      paths   = ["/health/detailed"]
      service = google_compute_backend_service.health_function_backend[0].id
    }

    # API routes
    path_rule {
      paths   = ["/api/*", "/docs", "/redoc", "/openapi.json"]
      service = google_compute_backend_service.api_backend.id
    }
  }
}

# HTTPS Proxy
resource "google_compute_target_https_proxy" "api_proxy" {
  name             = "virtuoso-api-proxy"
  url_map          = google_compute_url_map.api_urlmap.id
  ssl_certificates = [google_compute_managed_ssl_certificate.api_cert.id]

  # Enable QUIC for better performance
  quic_override = "ENABLE"
}

# Global Forwarding Rule (Load Balancer Frontend)
resource "google_compute_global_forwarding_rule" "api_forwarding" {
  name                  = "virtuoso-api-lb"
  ip_protocol           = "TCP"
  load_balancing_scheme = "EXTERNAL_MANAGED"
  port_range            = "443"
  target                = google_compute_target_https_proxy.api_proxy.id
  ip_address            = google_compute_global_address.api_ip.id
}

# Static IP for Load Balancer
resource "google_compute_global_address" "api_ip" {
  name = "virtuoso-api-ip"
}

# HTTP to HTTPS redirect
resource "google_compute_url_map" "http_redirect" {
  name = "virtuoso-http-redirect"

  default_url_redirect {
    https_redirect         = true
    redirect_response_code = "MOVED_PERMANENTLY_DEFAULT"
    strip_query            = false
  }
}

resource "google_compute_target_http_proxy" "http_proxy" {
  name    = "virtuoso-http-proxy"
  url_map = google_compute_url_map.http_redirect.id
}

resource "google_compute_global_forwarding_rule" "http_forwarding" {
  name                  = "virtuoso-http-lb"
  ip_protocol           = "TCP"
  load_balancing_scheme = "EXTERNAL_MANAGED"
  port_range            = "80"
  target                = google_compute_target_http_proxy.http_proxy.id
  ip_address            = google_compute_global_address.api_ip.id
}

# Cloud Armor Security Policy (optional)
resource "google_compute_security_policy" "api_security" {
  count = var.enable_cloud_armor ? 1 : 0

  name        = "virtuoso-api-security"
  description = "Security policy for Virtuoso API"

  # Default rule
  rule {
    action   = "allow"
    priority = "2147483647"
    match {
      versioned_expr = "SRC_IPS_V1"
      config {
        src_ip_ranges = ["*"]
      }
    }
    description = "Default allow rule"
  }

  # Rate limiting rule
  rule {
    action   = "rate_based_ban"
    priority = "1000"

    match {
      versioned_expr = "SRC_IPS_V1"
      config {
        src_ip_ranges = ["*"]
      }
    }

    rate_limit_options {
      conform_action = "allow"
      exceed_action  = "deny(429)"

      rate_limit_threshold {
        count        = 1000
        interval_sec = 60
      }

      ban_duration_sec = 600 # 10 minutes
    }

    description = "Rate limiting rule"
  }

  # Block common attack patterns
  rule {
    action   = "deny(403)"
    priority = "900"

    match {
      expr {
        expression = "request.path.matches('(?i)(\\.\\./)|(passwd)|(etc/shadow)')"
      }
    }

    description = "Block path traversal attempts"
  }

  # Adaptive protection (DDoS)
  adaptive_protection_config {
    layer_7_ddos_defense_config {
      enable = true
    }
  }
}

# Firewall Rules
resource "google_compute_firewall" "allow_health_checks" {
  name    = "virtuoso-allow-health-checks"
  network = google_compute_network.vpc.name

  allow {
    protocol = "tcp"
    ports    = ["80", "443", "8080"]
  }

  source_ranges = [
    "35.191.0.0/16",  # Google Health Check IPs
    "130.211.0.0/22", # Google Health Check IPs
  ]

  target_tags = ["health-check"]
}

# Private Service Connect (optional, for private endpoints)
resource "google_compute_address" "psc_address" {
  count        = var.enable_private_service_connect ? 1 : 0
  name         = "virtuoso-psc-address"
  address_type = "INTERNAL"
  purpose      = "PRIVATE_SERVICE_CONNECT"
  network      = google_compute_network.vpc.id
  subnetwork   = google_compute_subnetwork.main.id
}

# Outputs
output "load_balancer_ip" {
  value       = google_compute_global_address.api_ip.address
  description = "Global IP address of the load balancer"
}

output "vpc_network" {
  value       = google_compute_network.vpc.name
  description = "VPC network name"
}

output "vpc_connector" {
  value       = google_vpc_access_connector.connector.id
  description = "VPC connector ID for serverless services"
}
