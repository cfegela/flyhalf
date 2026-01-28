# VPC Outputs
output "vpc_id" {
  description = "ID of the VPC"
  value       = aws_vpc.main.id
}

output "private_subnet_ids" {
  description = "IDs of private subnets"
  value       = [aws_subnet.private_a.id, aws_subnet.private_b.id]
}

output "public_subnet_ids" {
  description = "IDs of public subnets"
  value       = [aws_subnet.public_a.id, aws_subnet.public_b.id]
}

# ECR Outputs
output "ecr_repository_url" {
  description = "URL of the ECR repository for API images"
  value       = aws_ecr_repository.api.repository_url
}

output "ecr_repository_arn" {
  description = "ARN of the ECR repository"
  value       = aws_ecr_repository.api.arn
}

# RDS Outputs
output "rds_endpoint" {
  description = "RDS instance endpoint"
  value       = aws_db_instance.postgres.endpoint
}

output "rds_address" {
  description = "RDS instance address (hostname)"
  value       = aws_db_instance.postgres.address
}

output "db_connection_string" {
  description = "Database connection endpoint via Route53"
  value       = "postgresql://${var.db_username}@${aws_route53_record.db.name}:5432/${var.db_name}"
  sensitive   = true
}

# ECS Outputs
output "ecs_cluster_name" {
  description = "Name of the ECS cluster"
  value       = aws_ecs_cluster.main.name
}

output "ecs_cluster_arn" {
  description = "ARN of the ECS cluster"
  value       = aws_ecs_cluster.main.arn
}

output "ecs_service_name" {
  description = "Name of the ECS service"
  value       = aws_ecs_service.api.name
}

output "ecs_task_definition_arn" {
  description = "ARN of the ECS task definition"
  value       = aws_ecs_task_definition.api.arn
}

# ALB Outputs
output "alb_dns_name" {
  description = "DNS name of the Application Load Balancer"
  value       = aws_lb.api.dns_name
}

output "alb_arn" {
  description = "ARN of the Application Load Balancer"
  value       = aws_lb.api.arn
}

output "alb_zone_id" {
  description = "Zone ID of the Application Load Balancer"
  value       = aws_lb.api.zone_id
}

# CloudFront Outputs
output "cloudfront_distribution_id" {
  description = "ID of the CloudFront distribution"
  value       = aws_cloudfront_distribution.web.id
}

output "cloudfront_domain_name" {
  description = "Domain name of the CloudFront distribution"
  value       = aws_cloudfront_distribution.web.domain_name
}

# S3 Outputs
output "web_bucket_name" {
  description = "Name of the S3 bucket for web files"
  value       = aws_s3_bucket.web.id
}

output "web_bucket_arn" {
  description = "ARN of the S3 bucket for web files"
  value       = aws_s3_bucket.web.arn
}

# DNS Outputs
output "api_url" {
  description = "URL of the API"
  value       = "https://${aws_route53_record.api.name}"
}

output "web_url" {
  description = "URL of the web application"
  value       = "https://${aws_route53_record.web.name}"
}

output "db_url" {
  description = "Database URL via Route53"
  value       = aws_route53_record.db.name
}
