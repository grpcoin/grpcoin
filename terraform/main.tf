/**
 * Copyright 2021 Ahmet Alp Balkan
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */


variable "project" {
  default = "grpcoin"
}

variable "billing_account" {
  default = "0050EC-505932-A9F334"
}

variable "region" {
  default = "us-west2"
}

variable "image" {
  description = "main server container image"
  type        = string
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

resource "google_project_service" "run" {
  project = var.project
  service = "run.googleapis.com"
}

resource "google_project_service" "vpcaccess" {
  project = var.project
  service = "vpcaccess.googleapis.com"
}

resource "google_app_engine_application" "app" {
  project       = var.project
  location_id   = var.region
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
  region             = var.region
  memory_size_gb     = 1
  authorized_network = google_compute_network.vpc.name
}

resource "google_service_account" "sa" {
  account_id   = "grpcoin"
  display_name = "grpc main server identity"
}

resource "google_project_iam_binding" "firestore-access" {
  project = var.project
  role    = "roles/datastore.user"

  members = [
    "serviceAccount:${google_service_account.sa.email}",
  ]
}

resource "google_project_iam_binding" "tracing-access" {
  project = var.project
  role    = "roles/cloudtrace.agent"

  members = [
    "serviceAccount:${google_service_account.sa.email}",
  ]
}

resource "google_vpc_access_connector" "default" {
  depends_on = [
    google_project_service.vpcaccess
  ]
  name          = "vpc-connector"
  region        = var.region
  project       = var.project
  network       = google_compute_network.vpc.name
  ip_cidr_range = "10.8.0.0/28"
}

resource "google_cloud_run_service" "default" {
  depends_on = [
    google_project_service.run
  ]
  project  = var.project
  location = var.region
  name     = "grpcoin-main"

  template {
    metadata {
      annotations = {
        "run.googleapis.com/vpc-access-connector" : google_vpc_access_connector.default.name
      }
    }
    spec {
      service_account_name  = google_service_account.sa.email
      container_concurrency = 100
      timeout_seconds       = 900

      containers {
        image = var.image
        ports {
          name = "h2c"
          container_port = 8080
        }
        resources {
          limits = {
            cpu    = "2"
            memory = "512Mi"
          }
        }
        env {
          name  = "REDIS_IP"
          value = google_redis_instance.cache.host
        }
      }
    }
  }
}

data "google_iam_policy" "noauth" {
  binding {
    role = "roles/run.invoker"
    members = [
      "allUsers",
    ]
  }
}

resource "google_cloud_run_service_iam_policy" "default-noauth" {
  location    = google_cloud_run_service.default.location
  project     = google_cloud_run_service.default.project
  service     = google_cloud_run_service.default.name
  policy_data = data.google_iam_policy.noauth.policy_data
}
