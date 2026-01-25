output "bucket_name" {
  description = "Cloud Storage bucket name for frontend files"
  value       = google_storage_bucket.frontend.name
}

output "bucket_url" {
  description = "Cloud Storage bucket URL"
  value       = google_storage_bucket.frontend.url
}

output "load_balancer_ip" {
  description = "Load balancer IP address for DNS A record"
  value       = google_compute_global_address.frontend.address
}

output "config_js_content" {
  description = "Generated config.js content for frontend"
  value       = local.config_js_content
}
