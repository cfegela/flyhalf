# Flyhalf GCP Deployment

Cost-optimized Terraform configuration for deploying Flyhalf to Google Cloud Platform.

## Architecture

```
Internet
    │
    ├──→ Cloud CDN (Load Balancer)        Frontend
    │    └──→ Cloud Storage               (Static files)
    │
    └──→ Cloud Run (API)                  Go API
         └──→ VPC Connector
              └──→ Cloud SQL              PostgreSQL 16
                   (Private IP)
```

## Estimated Costs

### Minimum Configuration (Default)
**~$15-25/month**

| Component | Specification | Monthly Cost |
|-----------|--------------|--------------|
| Cloud SQL | db-f1-micro (0.6GB RAM) | ~$7 |
| VPC Connector | e2-micro × 2 instances | ~$8-12 |
| Cloud Run | 0 min instances, pay per request | ~$0-5 |
| Cloud Storage + LB | Minimal traffic | ~$0-2 |
| Secret Manager | 3 secrets | ~$0.06 |

### Performance Configuration
**~$60-80/month** (see terraform.tfvars.example for settings)

- Cloud SQL: db-custom-1-3840 (1 vCPU, 3.75GB RAM)
- Cloud Run: 1 min instance (no cold starts)
- CDN enabled for faster frontend

## Cost Optimization Features

✅ **Database:**
- Smallest shared-core instance (db-f1-micro)
- ZONAL availability (single zone)
- Point-in-time recovery disabled
- Reduced backup retention (3 days)
- Query insights disabled
- Max 25 connections

✅ **Cloud Run:**
- 0 minimum instances (cold starts accepted)
- 256Mi memory allocation
- 60s request timeout
- Max 3 instances

✅ **Networking:**
- Minimal VPC connector (e2-micro × 2)
- Private IP only for database

✅ **Frontend:**
- CDN disabled by default
- HTTP→HTTPS redirect
- Static file caching

## Prerequisites

1. GCP account with billing enabled
2. gcloud CLI installed and authenticated
3. Terraform >= 1.5.0

## Quick Start

### 1. Configure Variables

```bash
cd ops/terraform/gcp
cp terraform.tfvars.example terraform.tfvars
```

Edit `terraform.tfvars`:

```hcl
project_id      = "your-gcp-project-id"
frontend_domain = "app.example.com"
api_domain      = "api.example.com"
db_user         = "flyhalf_app"
```

Generate secrets:
```bash
openssl rand -base64 32  # Use for db_password
openssl rand -base64 32  # Use for jwt_access_secret
openssl rand -base64 32  # Use for jwt_refresh_secret
```

### 2. Deploy Infrastructure

```bash
# Initialize Terraform
terraform init

# Review planned changes
terraform plan

# Deploy (takes ~10-15 minutes)
terraform apply
```

### 3. Build and Deploy API

```bash
# Configure Docker authentication
gcloud auth configure-docker us-central1-docker.pkg.dev

# Build Docker image
docker build -t flyhalf-api:latest .

# Tag for Artifact Registry
docker tag flyhalf-api:latest \
  us-central1-docker.pkg.dev/YOUR-PROJECT/flyhalf-prod/flyhalf-api:latest

# Push to registry
docker push \
  us-central1-docker.pkg.dev/YOUR-PROJECT/flyhalf-prod/flyhalf-api:latest
```

### 4. Deploy Frontend

```bash
# Upload to Cloud Storage
gsutil -m rsync -r -d web/ gs://YOUR-BUCKET-NAME/
```

### 5. Configure DNS

Get the IP addresses from Terraform output:

```bash
terraform output frontend_ip
terraform output api_url
```

Create DNS records:
- **A Record:** `app.example.com` → `<frontend_ip>`
- **CNAME Record:** `api.example.com` → `<cloud_run_url_without_https>`

### 6. Verify Deployment

```bash
# Test API health
curl https://api.example.com/health

# Test frontend
open https://app.example.com
```

## Reducing Costs Further

### Option 1: Use Cloud SQL Proxy (Save ~$8-12/month)
Replace VPC Connector with Cloud SQL Proxy sidecar in Cloud Run. This eliminates the VPC Connector cost but requires code changes.

### Option 2: Development Environment
For non-production use:
- Use Cloud Run with public Cloud SQL access (no VPC needed)
- Disable automated backups
- Use on-demand scaling only

### Option 3: Share Resources
Use the same VPC, Cloud SQL instance, and Artifact Registry across multiple environments (dev/staging/prod) with separate databases.

## Scaling Up

When your application needs more resources:

1. **Increase Database Size:**
   ```hcl
   db_tier = "db-custom-1-3840"  # 1 vCPU, 3.75GB RAM
   ```

2. **Enable High Availability:**
   ```hcl
   db_availability_type = "REGIONAL"  # Automatic failover
   ```

3. **Eliminate Cold Starts:**
   ```hcl
   api_min_instances = 1  # Keep 1 instance warm
   ```

4. **Enable CDN:**
   ```hcl
   enable_cdn = true  # Faster frontend loads
   ```

5. **Increase API Resources:**
   ```hcl
   api_memory = "512Mi"
   api_max_instances = 10
   ```

## Monitoring Costs

View current spending:
```bash
gcloud billing accounts list
gcloud billing projects describe YOUR-PROJECT-ID
```

Set up budget alerts in GCP Console:
- Billing → Budgets & alerts
- Create budget for ~$30/month
- Alert at 50%, 90%, 100%

## Cleanup

To destroy all resources:

```bash
terraform destroy
```

**Note:** Cloud SQL instances cannot be immediately recreated with the same name. Wait ~1 week or use a different name.

## Troubleshooting

### Cloud Run can't connect to Cloud SQL
- Verify VPC connector is in same region
- Check firewall rules allow port 5432
- Ensure Cloud SQL is in private IP mode

### Frontend shows old content
- Clear CDN cache: `gcloud compute url-maps invalidate-cdn-cache`
- Check config.js was uploaded to Cloud Storage
- Verify DNS propagation: `dig app.example.com`

### High costs
- Check Cloud SQL is db-f1-micro, not larger tier
- Verify Cloud Run min_instances = 0
- Review VPC connector instance count (should be 2)
- Disable point-in-time recovery if enabled

## Support

For issues specific to this Terraform configuration, check:
1. Terraform plan output for errors
2. GCP Console → Cloud Logging for runtime errors
3. Cloud Run logs: `gcloud run services logs read flyhalf-prod-api`

## License

See project root LICENSE file.
