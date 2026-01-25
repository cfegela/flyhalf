# Random suffix for instance name (Cloud SQL names cannot be reused for ~1 week after deletion)
resource "random_id" "db_suffix" {
  byte_length = 4
}

# Cloud SQL PostgreSQL Instance
resource "google_sql_database_instance" "main" {
  name             = "${var.app_name}-${var.environment}-db-${random_id.db_suffix.hex}"
  database_version = "POSTGRES_16"
  region           = var.region

  deletion_protection = true  # Prevent accidental deletion

  settings {
    tier              = var.db_tier
    availability_type = var.db_availability_type
    disk_size         = 10  # GB (minimum), will auto-increase
    disk_type         = "PD_SSD"  # Required for db-f1-micro
    disk_autoresize   = true

    backup_configuration {
      enabled                        = true
      start_time                     = "03:00"  # 3 AM UTC
      point_in_time_recovery_enabled = false  # Disabled for cost savings
      transaction_log_retention_days = 1      # Minimum (reduced from 7)
      backup_retention_settings {
        retained_backups = 3  # Reduced from 7 for cost savings
      }
    }

    ip_configuration {
      ipv4_enabled    = false  # No public IP
      private_network = var.vpc_id
      require_ssl     = true
    }

    maintenance_window {
      day          = 7  # Sunday
      hour         = 4  # 4 AM UTC
      update_track = "stable"
    }

    insights_config {
      query_insights_enabled  = false  # Disabled for cost savings
      query_string_length     = 1024
      record_application_tags = false
    }

    database_flags {
      name  = "max_connections"
      value = "25"  # Reduced from 100 for minimal resource usage
    }

    user_labels = var.labels
  }

  project = var.project_id

  depends_on = [var.private_vpc_connection_id]
}

# Database
resource "google_sql_database" "database" {
  name     = var.db_name
  instance = google_sql_database_instance.main.name

  project = var.project_id
}

# Database User
resource "google_sql_user" "user" {
  name     = var.db_user
  instance = google_sql_database_instance.main.name
  password = data.google_secret_manager_secret_version.db_password.secret_data

  project = var.project_id
}

# Fetch database password from Secret Manager
data "google_secret_manager_secret_version" "db_password" {
  secret = var.db_password_secret_id
}
