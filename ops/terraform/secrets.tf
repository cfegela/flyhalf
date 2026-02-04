# AWS Secrets Manager secrets for sensitive configuration

# Database password secret
resource "aws_secretsmanager_secret" "db_password" {
  name                    = "${var.project_name}-db-password"
  description             = "PostgreSQL database password"
  recovery_window_in_days = 7

  tags = {
    Name = "${var.project_name}-db-password"
  }
}

resource "aws_secretsmanager_secret_version" "db_password" {
  secret_id     = aws_secretsmanager_secret.db_password.id
  secret_string = var.db_password
}

# JWT access secret
resource "aws_secretsmanager_secret" "jwt_access_secret" {
  name                    = "${var.project_name}-jwt-access-secret"
  description             = "JWT access token secret"
  recovery_window_in_days = 7

  tags = {
    Name = "${var.project_name}-jwt-access-secret"
  }
}

resource "aws_secretsmanager_secret_version" "jwt_access_secret" {
  secret_id     = aws_secretsmanager_secret.jwt_access_secret.id
  secret_string = var.jwt_access_secret
}

# JWT refresh secret
resource "aws_secretsmanager_secret" "jwt_refresh_secret" {
  name                    = "${var.project_name}-jwt-refresh-secret"
  description             = "JWT refresh token secret"
  recovery_window_in_days = 7

  tags = {
    Name = "${var.project_name}-jwt-refresh-secret"
  }
}

resource "aws_secretsmanager_secret_version" "jwt_refresh_secret" {
  secret_id     = aws_secretsmanager_secret.jwt_refresh_secret.id
  secret_string = var.jwt_refresh_secret
}
