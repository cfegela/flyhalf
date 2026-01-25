# VPC Network
resource "google_compute_network" "vpc" {
  name                    = "${var.app_name}-${var.environment}-vpc"
  auto_create_subnetworks = false
  routing_mode            = "REGIONAL"

  project = var.project_id
}

# Subnet for VPC Connector
resource "google_compute_subnetwork" "subnet" {
  name          = "${var.app_name}-${var.environment}-subnet"
  ip_cidr_range = var.vpc_cidr
  region        = var.region
  network       = google_compute_network.vpc.id

  project = var.project_id
}

# Allocate IP range for Private Services Access (Cloud SQL)
resource "google_compute_global_address" "private_ip_address" {
  name          = "${var.app_name}-${var.environment}-private-ip"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = google_compute_network.vpc.id

  project = var.project_id
}

# Private VPC Connection for Cloud SQL
resource "google_service_networking_connection" "private_vpc_connection" {
  network                 = google_compute_network.vpc.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.private_ip_address.name]
}

# VPC Access Connector for Cloud Run to access Cloud SQL
resource "google_vpc_access_connector" "connector" {
  name          = "${var.app_name}-${var.environment}-connector"
  region        = var.region
  network       = google_compute_network.vpc.name
  ip_cidr_range = "10.8.0.0/28"  # Separate /28 range for connector

  # Machine type for connector
  machine_type = "e2-micro"

  min_instances = 2
  max_instances = 3

  project = var.project_id

  depends_on = [google_compute_subnetwork.subnet]
}

# Firewall rule to allow Cloud Run to Cloud SQL
resource "google_compute_firewall" "allow_cloud_run_to_sql" {
  name    = "${var.app_name}-${var.environment}-allow-cloudrun-sql"
  network = google_compute_network.vpc.name

  allow {
    protocol = "tcp"
    ports    = ["5432"]
  }

  source_ranges = [var.vpc_cidr, "10.8.0.0/28"]
  target_tags   = ["cloudsql"]

  project = var.project_id
}
