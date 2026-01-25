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

variable "frontend_domain" {
  description = "Frontend domain"
  type        = string
}

variable "api_domain" {
  description = "API domain"
  type        = string
}

variable "enable_cdn" {
  description = "Enable Cloud CDN"
  type        = bool
}
