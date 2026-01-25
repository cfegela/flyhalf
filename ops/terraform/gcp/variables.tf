# Project Configuration
variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "region" {
  description = "GCP region for resources"
  type        = string
  default     = "us-central1"
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  default     = "prod"

  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be dev, staging, or prod."
  }
}

# Domain Configuration
variable "frontend_domain" {
  description = "Frontend domain (e.g., app.flyhalf.io)"
  type        = string
}

variable "api_domain" {
  description = "API domain (e.g., api.flyhalf.io)"
  type        = string
}

# Database Configuration
variable "db_user" {
  description = "Database username"
  type        = string
  sensitive   = true
}

variable "db_password" {
  description = "Database password"
  type        = string
  sensitive   = true
}

variable "db_name" {
  description = "Database name"
  type        = string
  default     = "flyhalf"
}

variable "db_tier" {
  description = "Cloud SQL instance tier"
  type        = string
  default     = "db-f1-micro"
}

variable "db_availability_type" {
  description = "Cloud SQL availability type (REGIONAL for HA, ZONAL for single zone)"
  type        = string
  default     = "ZONAL"

  validation {
    condition     = contains(["REGIONAL", "ZONAL"], var.db_availability_type)
    error_message = "Availability type must be REGIONAL or ZONAL."
  }
}

# JWT Secrets
variable "jwt_access_secret" {
  description = "JWT access token secret"
  type        = string
  sensitive   = true
}

variable "jwt_refresh_secret" {
  description = "JWT refresh token secret"
  type        = string
  sensitive   = true
}

# Cloud Run Configuration
variable "api_min_instances" {
  description = "Minimum number of Cloud Run instances"
  type        = number
  default     = 0
}

variable "api_max_instances" {
  description = "Maximum number of Cloud Run instances"
  type        = number
  default     = 3
}

variable "api_cpu" {
  description = "CPU allocation for Cloud Run (1000m = 1 vCPU)"
  type        = string
  default     = "1000m"  # Minimum for Cloud Run Gen 2
}

variable "api_memory" {
  description = "Memory allocation for Cloud Run"
  type        = string
  default     = "512Mi"  # Minimum required for 1 vCPU (1000m)
}

# Networking
variable "vpc_cidr" {
  description = "CIDR range for VPC subnet"
  type        = string
  default     = "10.0.0.0/24"
}

variable "enable_cdn" {
  description = "Enable Cloud CDN for frontend"
  type        = bool
  default     = false  # Disabled by default for cost savings
}
