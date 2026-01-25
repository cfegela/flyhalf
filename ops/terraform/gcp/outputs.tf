# Frontend Outputs
output "frontend_ip" {
  description = "IP address for frontend DNS A record"
  value       = module.frontend.load_balancer_ip
}

output "frontend_bucket" {
  description = "Cloud Storage bucket for frontend files"
  value       = module.frontend.bucket_name
}

output "frontend_url" {
  description = "Frontend URL"
  value       = "https://${var.frontend_domain}"
}

# API Outputs
output "api_url" {
  description = "Cloud Run API URL"
  value       = module.api.service_url
}

output "api_service_name" {
  description = "Cloud Run service name"
  value       = module.api.service_name
}

output "artifact_registry_url" {
  description = "Artifact Registry repository URL for pushing images"
  value       = module.api.artifact_registry_url
}

# Database Outputs
output "db_connection_name" {
  description = "Cloud SQL connection name"
  value       = module.database.connection_name
}

output "db_private_ip" {
  description = "Cloud SQL private IP address"
  value       = module.database.private_ip
  sensitive   = true
}

# Networking Outputs
output "vpc_connector_id" {
  description = "VPC connector ID for Cloud Run"
  value       = module.networking.vpc_connector_id
}

# Service Account Outputs
output "api_service_account" {
  description = "Service account email for API"
  value       = module.iam.api_service_account_email
}

# Instructions
output "deployment_instructions" {
  description = "Next steps for deployment"
  value       = <<-EOT

    ===== DEPLOYMENT COMPLETE =====

    Estimated Monthly Cost: $15-25/month
    - Cloud SQL (db-f1-micro): ~$7/month
    - VPC Connector (e2-micro x2): ~$8-12/month
    - Cloud Run (0 min instances): ~$0-5/month
    - Storage + Load Balancer: ~$0-2/month
    - Secret Manager: ~$0.06/month

    NEXT STEPS:

    1. Configure DNS:
       - Create A record: ${var.frontend_domain} -> ${module.frontend.load_balancer_ip}
       - Create CNAME record: ${var.api_domain} -> ${trimsuffix(trimprefix(module.api.service_url, "https://"), "")}

    2. Build and push API Docker image:
       gcloud auth configure-docker ${var.region}-docker.pkg.dev
       docker build -t ${module.api.artifact_registry_url}/flyhalf-api:latest .
       docker push ${module.api.artifact_registry_url}/flyhalf-api:latest

    3. Deploy new Cloud Run revision (happens automatically on image push)

    4. Upload frontend files to Cloud Storage:
       gsutil -m rsync -r -d web/ gs://${module.frontend.bucket_name}/

    5. Test the deployment:
       - Health check: curl https://${var.api_domain}/health
       - Frontend: https://${var.frontend_domain}

    COST OPTIMIZATION NOTES:
    - 0 min instances = cold starts on first request (~3-5s delay)
    - CDN disabled by default (enable with enable_cdn = true)
    - Point-in-time recovery disabled (manual backups only)
    - To reduce VPC Connector cost (~50% of total), consider using Cloud SQL Proxy instead

  EOT
}
