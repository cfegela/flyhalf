# Service Account for Cloud Run API
resource "google_service_account" "api" {
  account_id   = "${var.app_name}-${var.environment}-api"
  display_name = "${var.app_name} API Service Account (${var.environment})"
  description  = "Service account for ${var.app_name} Cloud Run API"

  project = var.project_id
}

# Grant Cloud SQL Client role to API service account
resource "google_project_iam_member" "api_sql_client" {
  project = var.project_id
  role    = "roles/cloudsql.client"
  member  = "serviceAccount:${google_service_account.api.email}"
}

# Grant Secret Manager Secret Accessor role to API service account
resource "google_project_iam_member" "api_secret_accessor" {
  project = var.project_id
  role    = "roles/secretmanager.secretAccessor"
  member  = "serviceAccount:${google_service_account.api.email}"
}

# Note: Cloud Run Invoker role is managed in the API module
