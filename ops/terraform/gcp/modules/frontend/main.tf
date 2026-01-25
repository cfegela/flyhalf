# Cloud Storage Bucket for Frontend
resource "google_storage_bucket" "frontend" {
  name          = "${var.project_id}-${var.app_name}-${var.environment}-frontend"
  location      = "US"  # Multi-region for better availability
  force_destroy = false

  uniform_bucket_level_access = true

  website {
    main_page_suffix = "index.html"
    not_found_page   = "index.html"  # SPA routing
  }

  cors {
    origin          = ["https://${var.frontend_domain}"]
    method          = ["GET", "HEAD"]
    response_header = ["*"]
    max_age_seconds = 3600
  }

  labels = var.labels

  project = var.project_id
}

# Make bucket public
resource "google_storage_bucket_iam_member" "public_read" {
  bucket = google_storage_bucket.frontend.name
  role   = "roles/storage.objectViewer"
  member = "allUsers"
}

# Backend bucket for Load Balancer
resource "google_compute_backend_bucket" "frontend" {
  name        = "${var.app_name}-${var.environment}-frontend-backend"
  bucket_name = google_storage_bucket.frontend.name
  enable_cdn  = var.enable_cdn

  cdn_policy {
    cache_mode        = "CACHE_ALL_STATIC"
    client_ttl        = 3600
    default_ttl       = 3600
    max_ttl           = 86400
    negative_caching  = true
    serve_while_stale = 86400
  }

  project = var.project_id
}

# Reserve static IP address
resource "google_compute_global_address" "frontend" {
  name    = "${var.app_name}-${var.environment}-frontend-ip"
  project = var.project_id
}

# URL Map
resource "google_compute_url_map" "frontend" {
  name            = "${var.app_name}-${var.environment}-frontend-url-map"
  default_service = google_compute_backend_bucket.frontend.id

  project = var.project_id
}

# Managed SSL Certificate
resource "google_compute_managed_ssl_certificate" "frontend" {
  name = "${var.app_name}-${var.environment}-frontend-cert"

  managed {
    domains = [var.frontend_domain]
  }

  project = var.project_id
}

# HTTPS Proxy
resource "google_compute_target_https_proxy" "frontend" {
  name             = "${var.app_name}-${var.environment}-frontend-https-proxy"
  url_map          = google_compute_url_map.frontend.id
  ssl_certificates = [google_compute_managed_ssl_certificate.frontend.id]

  project = var.project_id
}

# Global Forwarding Rule (HTTPS)
resource "google_compute_global_forwarding_rule" "frontend_https" {
  name       = "${var.app_name}-${var.environment}-frontend-https"
  target     = google_compute_target_https_proxy.frontend.id
  port_range = "443"
  ip_address = google_compute_global_address.frontend.address

  project = var.project_id
}

# HTTP to HTTPS Redirect
resource "google_compute_url_map" "frontend_http_redirect" {
  name = "${var.app_name}-${var.environment}-frontend-http-redirect"

  default_url_redirect {
    https_redirect         = true
    redirect_response_code = "MOVED_PERMANENTLY_DEFAULT"
    strip_query            = false
  }

  project = var.project_id
}

resource "google_compute_target_http_proxy" "frontend_http_redirect" {
  name    = "${var.app_name}-${var.environment}-frontend-http-proxy"
  url_map = google_compute_url_map.frontend_http_redirect.id

  project = var.project_id
}

resource "google_compute_global_forwarding_rule" "frontend_http" {
  name       = "${var.app_name}-${var.environment}-frontend-http"
  target     = google_compute_target_http_proxy.frontend_http_redirect.id
  port_range = "80"
  ip_address = google_compute_global_address.frontend.address

  project = var.project_id
}

# Generate config.js for frontend
locals {
  config_js_content = templatefile("${path.module}/templates/config.js.tpl", {
    api_domain  = var.api_domain
    environment = var.environment
    timestamp   = timestamp()
  })
}

# Upload config.js to bucket
resource "google_storage_bucket_object" "config_js" {
  name    = "js/config.js"
  bucket  = google_storage_bucket.frontend.name
  content = local.config_js_content

  content_type = "application/javascript"

  cache_control = "public, max-age=300"  # 5 minutes cache for config

  lifecycle {
    ignore_changes = [
      detect_md5hash,
      metadata
    ]
  }
}
