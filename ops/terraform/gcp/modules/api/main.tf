# Artifact Registry Repository
resource "google_artifact_registry_repository" "repo" {
  location      = var.region
  repository_id = "${var.app_name}-${var.environment}"
  description   = "Docker repository for ${var.app_name} API (${var.environment})"
  format        = "DOCKER"

  labels = var.labels

  project = var.project_id
}

# Cloud Run Service
resource "google_cloud_run_v2_service" "api" {
  name     = "${var.app_name}-${var.environment}-api"
  location = var.region
  ingress  = "INGRESS_TRAFFIC_ALL"

  template {
    service_account = var.service_account_email

    scaling {
      min_instance_count = var.min_instances
      max_instance_count = var.max_instances
    }

    vpc_access {
      connector = var.vpc_connector_id
      egress    = "PRIVATE_RANGES_ONLY"
    }

    annotations = {
      "run.googleapis.com/cloudsql-instances" = var.db_connection_name
    }

    containers {
      # Initial deployment uses a placeholder image
      # After pushing your actual image, update with: gcloud run services update flyhalf-prod-api --image=IMAGE_URL
      image = "us-docker.pkg.dev/cloudrun/container/hello:latest"  # Public placeholder image

      resources {
        limits = {
          cpu    = var.cpu
          memory = var.memory
        }
      }

      ports {
        container_port = 8080
      }

      # Environment variables
      # Note: PORT is automatically set by Cloud Run to 8080
      env {
        name  = "ENVIRONMENT"
        value = var.environment
      }

      env {
        name  = "DB_HOST"
        value = "/cloudsql/${var.db_connection_name}"
      }

      env {
        name  = "DB_PORT"
        value = "5432"
      }

      env {
        name  = "DB_NAME"
        value = var.db_name
      }

      env {
        name  = "DB_USER"
        value = var.db_user
      }

      # Database password from Secret Manager
      env {
        name = "DB_PASSWORD"
        value_source {
          secret_key_ref {
            secret  = var.db_password_secret_id
            version = "latest"
          }
        }
      }

      # JWT Access Secret from Secret Manager
      env {
        name = "JWT_ACCESS_SECRET"
        value_source {
          secret_key_ref {
            secret  = var.jwt_access_secret_id
            version = "latest"
          }
        }
      }

      # JWT Refresh Secret from Secret Manager
      env {
        name = "JWT_REFRESH_SECRET"
        value_source {
          secret_key_ref {
            secret  = var.jwt_refresh_secret_id
            version = "latest"
          }
        }
      }

      # Health check probe
      startup_probe {
        http_get {
          path = "/health"
          port = 8080
        }
        initial_delay_seconds = 0
        timeout_seconds       = 1
        period_seconds        = 3
        failure_threshold     = 3
      }

      liveness_probe {
        http_get {
          path = "/health"
          port = 8080
        }
        initial_delay_seconds = 0
        timeout_seconds       = 1
        period_seconds        = 10
        failure_threshold     = 3
      }
    }

    # Timeout for request processing (reduced for cost savings)
    timeout = "60s"
  }

  labels = var.labels

  project = var.project_id

  lifecycle {
    ignore_changes = [
      template[0].containers[0].image,  # Allow image updates without Terraform
      client,
      client_version,
    ]
  }
}

# Allow unauthenticated access to Cloud Run service
resource "google_cloud_run_v2_service_iam_member" "public_invoker" {
  location = google_cloud_run_v2_service.api.location
  name     = google_cloud_run_v2_service.api.name
  role     = "roles/run.invoker"
  member   = "allUsers"

  project = var.project_id
}

# Domain Mapping for custom domain
resource "google_cloud_run_domain_mapping" "api_domain" {
  location = var.region
  name     = var.api_domain

  metadata {
    namespace = var.project_id

    labels = var.labels
  }

  spec {
    route_name = google_cloud_run_v2_service.api.name
  }

  project = var.project_id
}
