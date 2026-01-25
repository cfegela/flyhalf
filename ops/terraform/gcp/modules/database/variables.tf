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

variable "db_name" {
  description = "Database name"
  type        = string
}

variable "db_user" {
  description = "Database username"
  type        = string
  sensitive   = true
}

variable "db_tier" {
  description = "Cloud SQL instance tier"
  type        = string
}

variable "db_availability_type" {
  description = "Cloud SQL availability type (REGIONAL for HA, ZONAL for single zone)"
  type        = string
}

variable "vpc_id" {
  description = "VPC network ID"
  type        = string
}

variable "private_vpc_connection_id" {
  description = "Private VPC connection ID"
  type        = string
}

variable "db_password_secret_id" {
  description = "Secret Manager secret ID for database password"
  type        = string
}
