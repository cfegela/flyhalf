# Flyhalf GCP Infrastructure
# Deploys Go API + vanilla JS frontend + PostgreSQL to GCP

locals {
  app_name = "flyhalf"
  labels = {
    app         = local.app_name
    environment = var.environment
    managed_by  = "terraform"
  }
}

# Enable required GCP APIs
resource "google_project_service" "required_apis" {
  for_each = toset([
    "compute.googleapis.com",
    "servicenetworking.googleapis.com",
    "vpcaccess.googleapis.com",
    "sqladmin.googleapis.com",
    "run.googleapis.com",
    "artifactregistry.googleapis.com",
    "secretmanager.googleapis.com",
    "cloudresourcemanager.googleapis.com",
    "iam.googleapis.com",
  ])

  service            = each.key
  disable_on_destroy = false
}

# IAM - Service Accounts and Roles
module "iam" {
  source = "./modules/iam"

  project_id  = var.project_id
  environment = var.environment
  app_name    = local.app_name
  labels      = local.labels

  depends_on = [google_project_service.required_apis]
}

# Secrets - Secret Manager for sensitive values
module "secrets" {
  source = "./modules/secrets"

  project_id  = var.project_id
  environment = var.environment
  labels      = local.labels

  db_password         = var.db_password
  jwt_access_secret   = var.jwt_access_secret
  jwt_refresh_secret  = var.jwt_refresh_secret

  api_service_account_email = module.iam.api_service_account_email

  depends_on = [google_project_service.required_apis]
}

# Networking - VPC, Subnets, VPC Connector
module "networking" {
  source = "./modules/networking"

  project_id  = var.project_id
  region      = var.region
  environment = var.environment
  app_name    = local.app_name
  labels      = local.labels

  vpc_cidr = var.vpc_cidr

  depends_on = [google_project_service.required_apis]
}

# Database - Cloud SQL PostgreSQL
module "database" {
  source = "./modules/database"

  project_id  = var.project_id
  region      = var.region
  environment = var.environment
  app_name    = local.app_name
  labels      = local.labels

  db_name             = var.db_name
  db_user             = var.db_user
  db_tier             = var.db_tier
  db_availability_type = var.db_availability_type

  vpc_id                    = module.networking.vpc_id
  private_vpc_connection_id = module.networking.private_vpc_connection_id
  db_password_secret_id     = module.secrets.db_password_secret_id

  depends_on = [
    google_project_service.required_apis,
    module.networking
  ]
}

# API - Cloud Run and Artifact Registry
module "api" {
  source = "./modules/api"

  project_id  = var.project_id
  region      = var.region
  environment = var.environment
  app_name    = local.app_name
  labels      = local.labels

  api_domain      = var.api_domain
  min_instances   = var.api_min_instances
  max_instances   = var.api_max_instances
  cpu             = var.api_cpu
  memory          = var.api_memory

  vpc_connector_id          = module.networking.vpc_connector_id
  service_account_email     = module.iam.api_service_account_email
  db_connection_name        = module.database.connection_name
  db_host                   = module.database.private_ip
  db_name                   = var.db_name
  db_user                   = var.db_user
  db_password_secret_id     = module.secrets.db_password_secret_id
  jwt_access_secret_id      = module.secrets.jwt_access_secret_id
  jwt_refresh_secret_id     = module.secrets.jwt_refresh_secret_id

  depends_on = [
    google_project_service.required_apis,
    module.networking,
    module.database
  ]
}

# Frontend - Cloud Storage, CDN, Load Balancer
module "frontend" {
  source = "./modules/frontend"

  project_id  = var.project_id
  region      = var.region
  environment = var.environment
  app_name    = local.app_name
  labels      = local.labels

  frontend_domain = var.frontend_domain
  api_domain      = var.api_domain
  enable_cdn      = var.enable_cdn

  depends_on = [google_project_service.required_apis]
}
