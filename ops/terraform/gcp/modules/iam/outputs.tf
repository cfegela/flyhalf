output "api_service_account_email" {
  description = "Service account email for Cloud Run API"
  value       = google_service_account.api.email
}

output "api_service_account_id" {
  description = "Service account ID for Cloud Run API"
  value       = google_service_account.api.id
}
