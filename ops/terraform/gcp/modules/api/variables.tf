variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "region" {
  description = "GCP region"
  type        = string
}

variable "environment" {
  description = "Environment name"
  type        = string
}

variable "app_name" {
  description = "Application name"
  type        = string
}

variable "labels" {
  description = "Labels to apply to resources"
  type        = map(string)
  default     = {}
}

variable "api_domain" {
  description = "API domain"
  type        = string
}

variable "min_instances" {
  description = "Minimum number of Cloud Run instances"
  type        = number
}

variable "max_instances" {
  description = "Maximum number of Cloud Run instances"
  type        = number
}

variable "cpu" {
  description = "CPU allocation for Cloud Run"
  type        = string
}

variable "memory" {
  description = "Memory allocation for Cloud Run"
  type        = string
}

variable "vpc_connector_id" {
  description = "VPC connector ID for Cloud Run"
  type        = string
}

variable "service_account_email" {
  description = "Service account email for Cloud Run"
  type        = string
}

variable "db_connection_name" {
  description = "Cloud SQL connection name"
  type        = string
}

variable "db_host" {
  description = "Database host address"
  type        = string
}

variable "db_name" {
  description = "Database name"
  type        = string
}

variable "db_user" {
  description = "Database username"
  type        = string
  sensitive   = true
}

variable "db_password_secret_id" {
  description = "Secret Manager secret ID for database password"
  type        = string
}

variable "jwt_access_secret_id" {
  description = "Secret Manager secret ID for JWT access secret"
  type        = string
}

variable "jwt_refresh_secret_id" {
  description = "Secret Manager secret ID for JWT refresh secret"
  type        = string
}
