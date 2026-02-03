# RDS Subnet Group
resource "aws_db_subnet_group" "main" {
  name       = "${var.project_name}-db-subnet-group"
  subnet_ids = [aws_subnet.private_a.id, aws_subnet.private_b.id]

  tags = {
    Name = "${var.project_name}-db-subnet-group"
  }
}

# RDS PostgreSQL Instance
resource "aws_db_instance" "postgres" {
  identifier        = "${var.project_name}-db"
  engine            = "postgres"
  engine_version    = "18.1"
  instance_class    = var.db_instance_class
  allocated_storage = 20
  storage_type      = "gp3"
  storage_encrypted = true

  db_name  = var.db_name
  username = var.db_username
  password = var.db_password

  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [aws_security_group.rds.id]

  # Allow major version upgrades
  allow_major_version_upgrade = true
  apply_immediately           = true

  # Single AZ for basic setup (cost-effective)
  multi_az                = false
  publicly_accessible     = false
  skip_final_snapshot     = true
  backup_retention_period = 7
  backup_window           = "03:00-04:00"
  maintenance_window      = "sun:04:00-sun:05:00"

  # Performance Insights
  enabled_cloudwatch_logs_exports = ["postgresql", "upgrade"]

  tags = {
    Name = "${var.project_name}-postgres"
  }
}

# Route53 Record for Database (CNAME to RDS endpoint)
resource "aws_route53_record" "db" {
  zone_id = var.route53_zone_id
  name    = "${var.db_subdomain}.${var.domain_name}"
  type    = "CNAME"
  ttl     = 300
  records = [aws_db_instance.postgres.address]
}
