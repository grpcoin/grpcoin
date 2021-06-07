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

resource "google_compute_global_address" "default" {
  depends_on = [
    google_project_service.compute
  ]
  name = "lb-ip"
  description = "load balancer ip"
}

resource "google_compute_region_network_endpoint_group" "frontend" {
  depends_on = [
    google_project_service.compute
  ]
  project = var.project
  region  = var.region

  name                  = "cr-fe-neg"
  network_endpoint_type = "SERVERLESS"
  cloud_run {
    service = google_cloud_run_service.frontend.name
  }
}

resource "google_compute_region_network_endpoint_group" "apiserver" {
  depends_on = [
    google_project_service.compute
  ]
  project = var.project
  region  = var.region

  name                  = "cr-api-neg"
  network_endpoint_type = "SERVERLESS"
  cloud_run {
    service = google_cloud_run_service.apiserver.name
  }
}

resource "google_compute_managed_ssl_certificate" "default" {
  depends_on = [
    google_project_service.compute
  ]

  name = "grpcoin-cert"
  managed {
    domains = ["grpco.in", "api.grpco.in"]
  }
}

resource "google_compute_backend_service" "frontend" {
  name      = "grpcoin-frontend"

  protocol  = "HTTP"
  port_name = "http"
  timeout_sec = 30

  backend {
    group = google_compute_region_network_endpoint_group.frontend.id
  }
}

resource "google_compute_backend_service" "apiserver" {
  name      = "grpcoin-apiserver"

  protocol  = "HTTP"
  port_name = "http"
  backend {
    group = google_compute_region_network_endpoint_group.apiserver.id
  }
}

resource "google_compute_url_map" "default" {
  project         = var.project
  name            = "grpcoin-urlmap"

  default_service = google_compute_backend_service.frontend.self_link
  host_rule {
    hosts        = ["api.grpco.in"]
    path_matcher = "api"
  }
  path_matcher {
    name            = "api"
    default_service = google_compute_backend_service.apiserver.self_link
  }
}

resource "google_compute_target_https_proxy" "default" {
  name   = "lb-https-proxy"

  url_map          = google_compute_url_map.default.id
  ssl_certificates = [
    google_compute_managed_ssl_certificate.default.id
  ]
}

resource "google_compute_global_forwarding_rule" "default" {
  name   = "lb-https-fwdrule"

  port_range = "443"
  ip_address = google_compute_global_address.default.address
  target = google_compute_target_https_proxy.default.id
}

resource "google_compute_url_map" "https_redirect" {
  name            = "grpcoin-https-redirect"

  default_url_redirect {
    https_redirect         = true
    redirect_response_code = "MOVED_PERMANENTLY_DEFAULT"
    strip_query            = false
  }
}

resource "google_compute_target_http_proxy" "https_redirect" {
  name   = "lb-http-proxy"
  url_map          = google_compute_url_map.https_redirect.id
}

resource "google_compute_global_forwarding_rule" "https_redirect" {
  name   = "lb-http-fwdrule"

  port_range = "80"
  ip_address = google_compute_global_address.default.address
  target = google_compute_target_http_proxy.https_redirect.id
}
