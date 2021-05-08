
variable "project" {
  default = "grpcoin"
}

variable "billing_account" {
  default = "0050EC-505932-A9F334"
}

provider "google" {
  project = var.project
}

resource "google_project" "default" {
  name            = var.project
  project_id      = var.project
  billing_account = var.billing_account
}

resource "google_project_service" "compute" {
  project = var.project
  service = "compute.googleapis.com"
}

resource "google_project_service" "firestore" {
  project = var.project
  service = "firestore.googleapis.com"
}

resource "google_project_service" "redis" {
  project = var.project
  service = "redis.googleapis.com"
}

# TODO enable later when setting up a clean project
resource "google_app_engine_application" "app" {
  project       = var.project
  location_id   = "us-west2"
  database_type = "CLOUD_FIRESTORE"
}

resource "google_compute_network" "vpc" {
  project = var.project
  name    = "vpc"
  depends_on = [
    google_project_service.compute
  ]
}

resource "google_redis_instance" "cache" {
  depends_on = [
    google_project_service.redis
  ]
  project            = var.project
  name               = "cache"
  region             = "us-west2"
  memory_size_gb     = 1
  authorized_network = google_compute_network.vpc.name
}
