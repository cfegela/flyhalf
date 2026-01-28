variable "aws_region" {
  description = "AWS region for resources"
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "production"
}

variable "project_name" {
  description = "Project name for resource naming"
  type        = string
  default     = "flyhalf"
}

# Existing AWS Resources
variable "acm_certificate_arn" {
  description = "ARN of the ACM certificate for flyhalf.app"
  type        = string
}

variable "route53_zone_id" {
  description = "Route53 hosted zone ID for flyhalf.app"
  type        = string
}

variable "terraform_state_bucket" {
  description = "S3 bucket name for Terraform remote state"
  type        = string
}

# Database Configuration
variable "db_name" {
  description = "PostgreSQL database name"
  type        = string
  default     = "flyhalf"
}

variable "db_username" {
  description = "PostgreSQL master username"
  type        = string
  default     = "flyhalf"
}

variable "db_password" {
  description = "PostgreSQL master password"
  type        = string
  sensitive   = true
}

variable "db_instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t4g.micro"
}

# ECS Configuration
variable "api_cpu" {
  description = "CPU units for API container (256 = 0.25 vCPU)"
  type        = number
  default     = 256
}

variable "api_memory" {
  description = "Memory for API container in MB"
  type        = number
  default     = 512
}

variable "api_desired_count" {
  description = "Desired number of API tasks"
  type        = number
  default     = 1
}

# JWT Configuration
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

variable "jwt_access_expiry_min" {
  description = "JWT access token expiry in minutes"
  type        = number
  default     = 15
}

variable "jwt_refresh_expiry_day" {
  description = "JWT refresh token expiry in days"
  type        = number
  default     = 7
}

# Domain Configuration
variable "domain_name" {
  description = "Base domain name"
  type        = string
  default     = "flyhalf.app"
}

variable "api_subdomain" {
  description = "API subdomain"
  type        = string
  default     = "api"
}

variable "web_subdomain" {
  description = "Web subdomain"
  type        = string
  default     = "demo"
}

variable "db_subdomain" {
  description = "Database subdomain"
  type        = string
  default     = "db"
}
